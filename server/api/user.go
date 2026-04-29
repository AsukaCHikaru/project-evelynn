package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/asukachikaru/project-evelynn/server/db"
	"github.com/google/uuid"
)

type CreateUserRequest struct {
	DisplayName string `json:"display_name"`
}

type UserProfileResponse struct {
	DisplayName    string `json:"display_name"`
	DailyWordLimit int32  `json:"daily_word_limit"`
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ReturnAPIError(w, http.StatusBadRequest, APIError{
			Code:    ErrInvalidRequestBody,
			Message: "Invalid request body",
		},
		)
		return
	}
	err = req.Validate()
	if err != nil {
		ReturnAPIError(w, http.StatusBadRequest, APIError{
			Code:    ErrInvalidUserProfile,
			Message: "Invalid display name",
		})
		return
	}

	hashId := uuid.New().String()

	user, err := s.q.CreateUser(r.Context(), db.CreateUserParams{
		UserHashID:  hashId,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		ReturnAPIError(w, http.StatusInternalServerError, APIError{
			Code:    ErrServerError,
			Message: "Failed to create user",
		})
		return
	}

	ReturnAPISuccess(w, UserProfileResponse{
		DisplayName:    user.DisplayName,
		DailyWordLimit: user.DailyWordLimit,
	})
}

func (r *CreateUserRequest) Validate() error {
	if r.DisplayName == "" {
		return errors.New("Display name is empty")
	}
	return nil
}

func (s *Server) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	user, err := s.q.GetUser(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			ReturnAPIError(w, http.StatusNotFound, APIError{
				Code:    ErrUserNotFound,
				Message: "User not found",
			})
		default:
			ReturnAPIError(w, http.StatusInternalServerError, APIError{
				Code:    ErrServerError,
				Message: "Failed to get user profile",
			})
		}
		return
	}

	ReturnAPISuccess(w, UserProfileResponse{
		DisplayName:    user.DisplayName,
		DailyWordLimit: user.DailyWordLimit,
	})
}

type UpdateUserRequest struct {
	DisplayName    *string `json:"display_name,omitempty"`
	DailyWordLimit *int32  `json:"daily_word_limit,omitempty"`
}

func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req UpdateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ReturnAPIError(w, http.StatusBadRequest, APIError{
			Code:    ErrInvalidRequestBody,
			Message: "Invalid request body",
		})
		return
	}

	err = req.Validate()
	if err != nil {
		ReturnAPIError(w, http.StatusBadRequest, APIError{
			Code:    ErrInvalidUserProfile,
			Message: err.Error(),
		})
		return
	}

	var displayName sql.NullString
	if req.DisplayName != nil {
		displayName = sql.NullString{
			String: *req.DisplayName,
			Valid:  true,
		}
	}
	var dailyWordLimit sql.NullInt32
	if req.DailyWordLimit != nil {
		dailyWordLimit = sql.NullInt32{
			Int32: *req.DailyWordLimit,
			Valid: true,
		}
	}

	user, err := s.q.UpdateUser(r.Context(), db.UpdateUserParams{
		DisplayName:    displayName,
		DailyWordLimit: dailyWordLimit,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ReturnAPIError(w, http.StatusNotFound, APIError{
				Code:    ErrUserNotFound,
				Message: "User not found",
			})
			return
		}
		ReturnAPIError(w, http.StatusInternalServerError, APIError{
			Code:    ErrServerError,
			Message: "Failed to update user profile",
		})
		return
	}

	ReturnAPISuccess(w, UserProfileResponse{
		DisplayName:    user.DisplayName,
		DailyWordLimit: user.DailyWordLimit,
	})
}

func (r *UpdateUserRequest) Validate() error {
	if r.DailyWordLimit != nil && (*r.DailyWordLimit <= 0 || *r.DailyWordLimit > 20) {
		return errors.New("Daily word limit must be between 1 and 20")
	}
	if r.DisplayName != nil && *r.DisplayName == "" {
		return errors.New("Display name cannot be empty")
	}
	return nil
}
