package service

import (
	"context"
	documentStorage "document-server/internal/storage/document"
	storage "document-server/internal/storage/document"
	tokenStorage "document-server/internal/storage/token"
	userStorage "document-server/internal/storage/user"

	"github.com/google/uuid"
)

type UserStorage interface {
	Create(ctx context.Context, user userStorage.User) error
	GetUserByLogin(ctx context.Context, login string) (userStorage.User, error)
	GetUserByID(ctx context.Context, uuid uuid.UUID) (userStorage.User, error)
}

type DocumentStorage interface {
	Create(ctx context.Context, doc documentStorage.Document) error
	GetByID(ctx context.Context, id string) (*documentStorage.Document, error)
	GetDocumentsByIDs(ctx context.Context, ids []string) ([]storage.Document, error)
	ListDocumentIDs(ctx context.Context, currentLogin string, filterLogin string, key string, value string, limit int) ([]string, error)
	DeleteDocumentByID(ctx context.Context, id uuid.UUID) error
}

type TokenStorage interface {
	Create(ctx context.Context, token tokenStorage.UserToken) error
	GetByToken(ctx context.Context, token string) (tokenStorage.UserToken, error)
	Delete(ctx context.Context, token string) error
}

type Cache interface {
	Get(key string) (*documentStorage.Document, bool)
	Set(key string, doc *documentStorage.Document)
	Delete(key string)
}
