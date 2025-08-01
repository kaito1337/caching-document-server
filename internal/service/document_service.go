package service

import (
	"context"
	"database/sql"
	"document-server/internal/api/models"
	"document-server/internal/cache"
	storage "document-server/internal/storage/document"
	"slices"

	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type DocumentService struct {
	documentStorage DocumentStorage
	userStorage     UserStorage
	tokenStorage    TokenStorage
	cache           *cache.InMemoryCache
	storagePath     string
}

func NewDocumentService(
	userStorage UserStorage, documentStorage DocumentStorage,
	tokenStorage TokenStorage,
	storagePath string,
	cache *cache.InMemoryCache,
) *DocumentService {
	return &DocumentService{
		documentStorage: documentStorage,
		userStorage:     userStorage,
		tokenStorage:    tokenStorage,
		storagePath:     storagePath,
		cache:           cache,
	}
}

func (s *DocumentService) UploadDocument(ctx context.Context, meta models.DocumentUploadMetaDTO, fileBytes []byte, filename string, jsonData []byte) (*models.DocumentResponseDTO, error) {
	userToken, err := s.tokenStorage.GetByToken(ctx, meta.Token)
	if err != nil {
		return nil, errors.New("token not found")
	}
	user, err := s.userStorage.GetUserByID(ctx, userToken.UserID)
	if err != nil {
		return nil, errors.New("user not found")
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
			return nil, err
		}
		doc.FilePath = sql.NullString{String: docPath, Valid: true}
	} else {
		if len(jsonData) > 0 {
			if !json.Valid(jsonData) {
				return nil, errors.New("invalid JSON data")
			}
			doc.JSONData = sql.NullString{String: string(jsonData), Valid: true}
		}
	}

	if err := s.documentStorage.Create(ctx, doc); err != nil {
		return nil, err
	}

	s.cache.Set("document:"+doc.ID.String(), &doc)

	return &models.DocumentResponseDTO{
		JSON: json.RawMessage(doc.JSONData.String),
		File: doc.Name,
	}, nil
}

func (s *DocumentService) ListDocuments(ctx context.Context, token, login, key, value, limitStr string) ([]models.DocumentListItemDTO, error) {
	userToken, err := s.tokenStorage.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.New("token not found")
	}

	user, err := s.userStorage.GetUserByID(ctx, userToken.UserID)
	if err != nil {
		return nil, errors.New("user not found")
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
		return nil, err
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
			return nil, err
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

	return result, nil
}

func (s *DocumentService) GetDocument(ctx context.Context, id string) (*storage.Document, []byte, string, error) {
	cacheKey := "document:" + id

	if docCached, ok := s.cache.Get(cacheKey); ok {
		if docCached.IsFile {
			data, err := os.ReadFile(docCached.FilePath.String)
			if err != nil {
				s.cache.Delete(cacheKey)
				return nil, nil, "", err
			}
			return docCached, data, docCached.MimeType, nil
		} else {
			return docCached, nil, "", nil
		}
	}

	doc, err := s.documentStorage.GetByID(ctx, id)
	if err != nil {
		return nil, nil, "", err
	}

	var data []byte
	var mime string

	if doc.IsFile && doc.FilePath.Valid {
		data, err = os.ReadFile(doc.FilePath.String)
		if err != nil {
			return nil, nil, "", err
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
		return errors.New("invalid id")
	}

	doc, err := s.documentStorage.GetByID(ctx, docUUID.String())
	if err != nil {
		return err
	}

	if doc.IsFile && doc.FilePath.Valid {
		_ = os.Remove(doc.FilePath.String)
	}

	s.cache.Delete("document:" + docUUID.String())

	return s.documentStorage.DeleteDocumentByID(ctx, docUUID)
}
