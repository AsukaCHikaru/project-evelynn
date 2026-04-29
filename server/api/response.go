package api

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Data  any       `json:"data"`
	Error *APIError `json:"error"`
}

type APIError struct {
	Code    APIErrorCode `json:"code"`
	Message string       `json:"message"`
}

func ReturnAPISuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		APIResponse{
			Data:  data,
			Error: nil,
		})
}

func ReturnAPIError(w http.ResponseWriter, status int, e APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(
		APIResponse{
			Data:  nil,
			Error: &e,
		})
}
