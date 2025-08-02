package api

import (
	"document-server/internal/api/controller"
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router
	userAuthMiddleware mux.MiddlewareFunc
}

func NewRouter() (*Router, error) {
	r := mux.NewRouter()

	return &Router{
		Router: r,
	}, nil
}

func (r *Router) SetUserRoutes(controller *controller.UserController) {
	r.HandleFunc("/api/auth", controller.Authenticate).Methods(http.MethodPost)

	registerSubrouter := r.PathPrefix("/api/register").Subrouter()
	registerSubrouter.HandleFunc("", controller.Register).Methods(http.MethodPost)

	logoutSubrouter := r.PathPrefix("/api/auth/{token}").Subrouter()
	logoutSubrouter.HandleFunc("", controller.Logout).Methods(http.MethodDelete)
}

func (r *Router) SetDocsRoutes(controller *controller.DocumentController) {
	docsSubrouter := r.PathPrefix("/api/docs").Subrouter()
	docsSubrouter.Use(r.userAuthMiddleware)

	docsSubrouter.HandleFunc("", controller.GetDocuments).Methods(http.MethodGet, http.MethodHead)
	docsSubrouter.HandleFunc("", controller.UploadDocument).Methods(http.MethodPost)
	docsSubrouter.HandleFunc("/{id}", controller.GetDocument).Methods(http.MethodGet, http.MethodHead)
	docsSubrouter.HandleFunc("/{id}", controller.DeleteDocument).Methods(http.MethodDelete)
}
