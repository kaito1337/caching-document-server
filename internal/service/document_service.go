package service

import (
	"context"
	"database/sql"
	"document-server/internal/api/models"
	"document-server/internal/cache"
	storage "document-server/internal/storage/document"
	"log/slog"
	"slices"

	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hedhyw/semerr/pkg/v1/semerr"
)

type DocumentService struct {
	documentStorage DocumentStorage
	userStorage     UserStorage
	tokenStorage    TokenStorage
	logger          *slog.Logger
	cache           Cache
	storagePath     string
}

func NewDocumentService(
	userStorage UserStorage, documentStorage DocumentStorage,
	tokenStorage TokenStorage,
	logger *slog.Logger,
	storagePath string,
	cache *cache.InMemoryCache,
) *DocumentService {
	return &DocumentService{
		documentStorage: documentStorage,
		userStorage:     userStorage,
		tokenStorage:    tokenStorage,
		logger:          logger,
		storagePath:     storagePath,
		cache:           cache,
	}
}

func (s *DocumentService) UploadDocument(ctx context.Context, meta models.DocumentUploadMetaDTO, fileBytes []byte, filename string, jsonData []byte) (*models.DocumentResponseDTO, error) {
	userToken, err := s.tokenStorage.GetByToken(ctx, meta.Token)
	if err != nil {
		s.logger.Error("token not found", slog.String("token", meta.Token))
		return nil, semerr.NewBadRequestError(errors.New("token not found"))
	}
	user, err := s.userStorage.GetUserByID(ctx, userToken.UserID)
	if err != nil {
		s.logger.Error("user not found", slog.String("user_id", userToken.UserID.String()))
		return nil, semerr.NewBadRequestError(errors.New("user not found"))
	}

	if !slices.Contains(meta.Grant, user.Login) {
		meta.Grant = append(meta.Grant, user.Login)
	}

	doc := storage.Document{
		ID:        uuid.New(),
		Name:      meta.Name,
		MimeType:  meta.Mime,
		IsFile:    meta.File,
		IsPublic:  meta.Public,
		CreatedAt: time.Now(),
		GrantedTo: meta.Grant,
	}

	if meta.File {
		docPath := filepath.Join(s.storagePath, doc.ID.String()+"_"+filename)
		err := os.WriteFile(docPath, fileBytes, 0644)
		if err != nil {
			s.logger.Error("failed to write file", slog.String("path", docPath), slog.String("error", err.Error()))
			return nil, semerr.NewInternalServerError(err)
		}
		doc.FilePath = sql.NullString{String: docPath, Valid: true}
	} else {
		if len(jsonData) > 0 {
			if !json.Valid(jsonData) {
				s.logger.Error("invalid JSON data")
				return nil, semerr.NewBadRequestError(errors.New("invalid JSON data"))
			}
			doc.JSONData = sql.NullString{String: string(jsonData), Valid: true}
		}
	}

	if err := s.documentStorage.Create(ctx, doc); err != nil {
		s.logger.Error("failed to create document", slog.String("error", err.Error()))
		return nil, semerr.NewInternalServerError(err)
	}

	s.cache.Set("document:"+doc.ID.String(), &doc)
	s.logger.Info("document uploaded", slog.String("doc_id", doc.ID.String()), slog.String("user", user.Login))

	return &models.DocumentResponseDTO{
		JSON: json.RawMessage(doc.JSONData.String),
		File: doc.Name,
	}, nil
}

func (s *DocumentService) ListDocuments(ctx context.Context, token, login, key, value, limitStr string) ([]models.DocumentListItemDTO, error) {
	userToken, err := s.tokenStorage.GetByToken(ctx, token)
	if err != nil {
		s.logger.Error("token not found", slog.String("token", token))
		return nil, semerr.NewBadRequestError(errors.New("token not found"))
	}

	user, err := s.userStorage.GetUserByID(ctx, userToken.UserID)
	if err != nil {
		s.logger.Error("user not found", slog.String("user_id", userToken.UserID.String()))
		return nil, semerr.NewBadRequestError(errors.New("user not found"))
	}

	var limit int
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 20
		}
	} else {
		limit = 20
	}

	targetLogin := login
	if targetLogin == "" {
		targetLogin = user.Login
	}

	docIDs, err := s.documentStorage.ListDocumentIDs(ctx, user.Login, targetLogin, key, value, limit)
	if err != nil {
		s.logger.Error("failed to list document IDs", slog.String("error", err.Error()))
		return nil, semerr.NewInternalServerError(err)
	}

	var result []models.DocumentListItemDTO
	var idsToFetchFromDB []string

	for _, id := range docIDs {
		cacheKey := "document:" + id
		if cached, ok := s.cache.Get(cacheKey); ok {
			result = append(result, models.DocumentListItemDTO{
				ID:        cached.ID.String(),
				Name:      cached.Name,
				Mime:      cached.MimeType,
				File:      cached.IsFile,
				Public:    cached.IsPublic,
				CreatedAt: cached.CreatedAt,
				Grant:     cached.GrantedTo,
			})
		} else {
			idsToFetchFromDB = append(idsToFetchFromDB, id)
		}
	}

	if len(idsToFetchFromDB) > 0 {
		docs, err := s.documentStorage.GetDocumentsByIDs(ctx, idsToFetchFromDB)
		if err != nil {
			s.logger.Error("failed to fetch documents from DB", slog.String("error", err.Error()))
			return nil, semerr.NewInternalServerError(err)
		}

		for _, doc := range docs {
			cacheKey := "document:" + doc.ID.String()
			s.cache.Set(cacheKey, &doc)

			result = append(result, models.DocumentListItemDTO{
				ID:        doc.ID.String(),
				Name:      doc.Name,
				Mime:      doc.MimeType,
				File:      doc.IsFile,
				Public:    doc.IsPublic,
				CreatedAt: doc.CreatedAt,
				Grant:     doc.GrantedTo,
			})
		}
	}

	s.logger.Info("documents listed", slog.String("user", user.Login), slog.Int("count", len(result)))
	return result, nil
}

func (s *DocumentService) GetDocument(ctx context.Context, id string) (*storage.Document, []byte, string, error) {
	cacheKey := "document:" + id

	if docCached, ok := s.cache.Get(cacheKey); ok {
		if docCached.IsFile {
			data, err := os.ReadFile(docCached.FilePath.String)
			if err != nil {
				s.cache.Delete(cacheKey)
				s.logger.Error("cached file not found, removed from cache", slog.String("path", docCached.FilePath.String))
				return nil, nil, "", semerr.NewInternalServerError(err)
			}
			return docCached, data, docCached.MimeType, nil
		} else {
			return docCached, nil, "", nil
		}
	}

	doc, err := s.documentStorage.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("document not found", slog.String("id", id))
		return nil, nil, "", semerr.NewBadRequestError(errors.New("document not found"))
	}

	var data []byte
	var mime string

	if doc.IsFile && doc.FilePath.Valid {
		data, err = os.ReadFile(doc.FilePath.String)
		if err != nil {
			s.logger.Error("failed to read file", slog.String("path", doc.FilePath.String), slog.String("error", err.Error()))
			return nil, nil, "", semerr.NewInternalServerError(err)
		}
		mime = doc.MimeType
	}

	s.cache.Set(cacheKey, &storage.Document{
		ID:        doc.ID,
		Name:      doc.Name,
		MimeType:  doc.MimeType,
		IsFile:    doc.IsFile,
		IsPublic:  doc.IsPublic,
		CreatedAt: doc.CreatedAt,
		GrantedTo: doc.GrantedTo,
		FilePath:  doc.FilePath,
		JSONData:  doc.JSONData,
	})

	return doc, data, mime, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, id string) error {
	docUUID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid document ID", slog.String("id", id))
		return semerr.NewBadRequestError(errors.New("invalid document ID"))
	}

	doc, err := s.documentStorage.GetByID(ctx, docUUID.String())
	if err != nil {
		s.logger.Error("document not found", slog.String("id", id))
		return semerr.NewBadRequestError(errors.New("document not found"))
	}

	if doc.IsFile && doc.FilePath.Valid {
		err := os.Remove(doc.FilePath.String)
		if err != nil {
			s.logger.Error("failed to delete file", slog.String("path", doc.FilePath.String), slog.String("error", err.Error()))
		}
	}

	s.cache.Delete("document:" + docUUID.String())

	if err := s.documentStorage.DeleteDocumentByID(ctx, docUUID); err != nil {
		s.logger.Error("failed to delete document from DB", slog.String("id", id), slog.String("error", err.Error()))
		return semerr.NewInternalServerError(err)
	}

	s.logger.Info("document deleted", slog.String("id", id))
	return nil
}
