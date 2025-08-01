package response

import (
	"document-server/internal/api/models"
	"encoding/json"
	"net/http"
)

func respondJSON(w http.ResponseWriter, httpStatus int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(payload)
}

func RespondWithError(w http.ResponseWriter, httpStatus int, message string) {
	response := models.APIResponse{
		Error: &models.APIError{
			Code: httpStatus,
			Text: message,
		},
	}
	respondJSON(w, httpStatus, response)
}

func RespondWithData(w http.ResponseWriter, httpStatus int, dataPayload interface{}) {
	response := models.APIResponse{
		Data: dataPayload,
	}
	respondJSON(w, httpStatus, response)
}

func RespondWithConfirm(w http.ResponseWriter, httpStatus int, responsePayload interface{}) {
	response := models.APIResponse{
		Response: responsePayload,
	}
	respondJSON(w, httpStatus, response)
}
