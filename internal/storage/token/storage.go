package storage

import (
	"context"
	"database/sql"
	"document-server/internal/storage"
	"errors"

	"github.com/jmoiron/sqlx"
)

type TokenStorage struct {
	db *sqlx.DB
}

func NewTokenStorage(db *sqlx.DB) *TokenStorage {
	return &TokenStorage{db: db}
}

func (s *TokenStorage) Create(ctx context.Context, token UserToken) error {
	query := `INSERT INTO user_tokens (token, user_id, expires_at)
              VALUES ($1, $2, $3)`
	_, err := s.db.ExecContext(ctx, query, token.Token, token.UserID, token.ExpiresAt)
	return err
}

func (s *TokenStorage) GetByToken(ctx context.Context, tokenValue string) (UserToken, error) {
	var token UserToken
	query := "SELECT token, user_id, expires_at FROM user_tokens WHERE token=$1"
	err := s.db.GetContext(ctx, &token, query, tokenValue)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserToken{}, storage.ErrTokenNotFound
		}
		return UserToken{}, err
	}
	return token, nil
}

func (s *TokenStorage) Delete(ctx context.Context, tokenValue string) error {
	query := "DELETE FROM user_tokens WHERE token=$1"
	_, err := s.db.ExecContext(ctx, query, tokenValue)
	return err
}
