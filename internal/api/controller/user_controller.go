package controller

import (
	"document-server/internal/api/models"
	"document-server/internal/api/response"
	"document-server/internal/service"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hedhyw/semerr/pkg/v1/semerr"
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
		response.RespondWithError(w, semerr.NewBadRequestError(err))
		return
	}

	err := c.userService.RegisterUser(r.Context(), req.Login, req.Password, req.Token)
	if err != nil {
		response.RespondWithError(w, err)
		return
	}

	response.RespondWithConfirm(w, http.StatusCreated, models.RegisterResponseDTO{Login: req.Login})
}

func (c *UserController) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, semerr.NewBadRequestError(err))
		return
	}

	token, err := c.userService.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		response.RespondWithError(w, err)
		return
	}

	response.RespondWithConfirm(w, http.StatusOK, models.AuthResponseDTO{Token: token})
}

func (c *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	if token == "" {
		response.RespondWithError(w, semerr.NewBadRequestError(errors.New("missing token")))
		return
	}

	if err := c.userService.Logout(r.Context(), token); err != nil {
		response.RespondWithError(w, err)
		return
	}

	response.RespondWithConfirm(w, http.StatusOK, map[string]bool{
		token: true,
	})
}
