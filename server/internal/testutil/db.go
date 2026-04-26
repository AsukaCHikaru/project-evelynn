package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func BootTestDB() (*sql.DB, func(), error) {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	terminate := func() {
		postgresContainer.Terminate(ctx)
	}

	connStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		return nil, terminate, fmt.Errorf("failed to get connection string: %w", err)
	}
	conn, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, terminate, fmt.Errorf("failed to open database connection: %w", err)
	}
	err = conn.Ping()
	if err != nil {
		return nil, terminate, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, terminate, nil
}

func MigrateTestDB(conn *sql.DB) error {
	_, filename, _, _ := runtime.Caller(0)
	migrationFilePath := filepath.Join(filepath.Dir(filename), "../../migrations")
	driver, err := migratepg.WithInstance(conn, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationFilePath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func TruncateTestDB(conn *sql.DB) error {
	rows, err := conn.Query(`
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name != 'schema_migrations'
	`)
	if err != nil {
		return fmt.Errorf("failed to query table names: %w", err)
	}
	defer rows.Close()

	var tableList []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tableList = append(tableList, fmt.Sprintf(`"%s"`, tableName))
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over table names: %w", err)
	}

	if len(tableList) == 0 {
		return nil
	}

	sql := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tableList, ", "))
	_, err = conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}
