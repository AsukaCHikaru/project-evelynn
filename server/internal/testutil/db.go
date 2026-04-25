package testutil

import (
	"context"
	"database/sql"
	"fmt"

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
