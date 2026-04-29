package api

type APIErrorCode string

const (
	ErrInvalidRequestBody APIErrorCode = "E00001"
	ErrServerError        APIErrorCode = "E00002"
	ErrInvalidUserProfile APIErrorCode = "E01001"
	ErrUserNotFound       APIErrorCode = "E01002"
)
