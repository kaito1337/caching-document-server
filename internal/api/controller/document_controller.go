package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"document-server/internal/api/models"
	"document-server/internal/api/response"
	"document-server/internal/service"

	"github.com/gorilla/mux"
)

type DocumentController struct {
	documentService *service.DocumentService
}

func NewDocumentController(documentService *service.DocumentService) *DocumentController {
	return &DocumentController{documentService: documentService}
}

func (c *DocumentController) UploadDocument(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}

	metaStr := r.FormValue("meta")
	var meta models.DocumentUploadMetaDTO
	if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid meta JSON")
		return
	}

	var fileBytes []byte
	var filename string
	var jsonData []byte

	if meta.File {
		file, header, err := r.FormFile("file")
		if err != nil {
			response.RespondWithError(w, http.StatusBadRequest, "Missing file")
			return
		}
		defer file.Close()

		filename = header.Filename
		fileBytes, err = io.ReadAll(file)
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "Error reading file")
			return
		}
	} else {
		jsonFile, _, err := r.FormFile("json")
		if err == nil {
			defer jsonFile.Close()
			jsonData, _ = io.ReadAll(jsonFile)
		}
	}

	doc, err := c.documentService.UploadDocument(r.Context(), meta, fileBytes, filename, jsonData)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondWithData(w, http.StatusCreated, map[string]interface{}{
		"data": doc,
	})
}

func (c *DocumentController) GetDocuments(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	login := r.URL.Query().Get("login")
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	limitStr := r.URL.Query().Get("limit")

	docs, err := c.documentService.ListDocuments(r.Context(), token, login, key, value, limitStr)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondWithData(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"docs": docs,
		},
	})
}

func (c *DocumentController) GetDocument(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	doc, content, mimeType, err := c.documentService.GetDocument(r.Context(), id)
	if err != nil {
		response.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if doc.IsFile {
		w.Header().Set("Content-Type", mimeType)
		w.Write(content)
		return
	}

	response.RespondWithData(w, http.StatusOK, map[string]interface{}{
		"data": doc.JSONData})
}

func (c *DocumentController) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	err := c.documentService.DeleteDocument(r.Context(), id)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondWithData(w, http.StatusOK, map[string]interface{}{
		"response": map[string]bool{id: true},
	})
}
