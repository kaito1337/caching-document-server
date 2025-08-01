package controller

import (
	"document-server/internal/api/models"
	"document-server/internal/api/response"
	"document-server/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(s *service.UserService) *UserController {
	return &UserController{userService: s}
}

func (c *UserController) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := c.userService.RegisterUser(r.Context(), req.Login, req.Password, req.Token)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondWithConfirm(w, http.StatusCreated, models.RegisterResponseDTO{Login: req.Login})
}

func (c *UserController) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	token, err := c.userService.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		slog.Error(err.Error())
		response.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	response.RespondWithConfirm(w, http.StatusOK, models.AuthResponseDTO{Token: token})
}

func (c *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	if token == "" {
		response.RespondWithError(w, http.StatusBadRequest, "Missing token in path")
		return
	}

	if err := c.userService.Logout(r.Context(), token); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	response.RespondWithConfirm(w, http.StatusOK, map[string]bool{
		token: true,
	})
}
