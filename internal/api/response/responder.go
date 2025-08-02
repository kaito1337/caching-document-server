package response

import (
	"document-server/internal/api/models"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hedhyw/semerr/pkg/v1/httperr"
)

func respondJSON(w http.ResponseWriter, httpStatus int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(payload)
}

func RespondWithError(w http.ResponseWriter, err error) {
	if err == nil {
		err = errors.New("unknown error")
	}

	status := httperr.Code(err)

	response := models.APIResponse{
		Error: &models.APIError{
			Code: status,
			Text: err.Error(),
		},
	}

	respondJSON(w, status, response)
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
