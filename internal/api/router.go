package api

import (
	"document-server/internal/api/controller"
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router
}

func NewRouter() (*Router, error) {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	return &Router{
		Router: api,
	}, nil
}

func (r *Router) SetUserRoutes(controller *controller.UserController) {
	r.HandleFunc("/auth", controller.Authenticate).Methods(http.MethodPost)

	registerSub := r.PathPrefix("/register").Subrouter()
	registerSub.HandleFunc("", controller.Register).Methods(http.MethodPost)

	logoutSub := r.PathPrefix("/auth/{token}").Subrouter()
	logoutSub.HandleFunc("", controller.Logout).Methods(http.MethodDelete)
}

func (r *Router) SetDocsRoutes(controller *controller.DocumentController) {
	docs := r.PathPrefix("/docs").Subrouter()

	docs.HandleFunc("", controller.GetDocuments).Methods(http.MethodGet, http.MethodHead)
	docs.HandleFunc("", controller.UploadDocument).Methods(http.MethodPost)
	docs.HandleFunc("/{id}", controller.GetDocument).Methods(http.MethodGet, http.MethodHead)
	docs.HandleFunc("/{id}", controller.DeleteDocument).Methods(http.MethodDelete)
}
