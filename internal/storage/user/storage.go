package storage

import (
	"context"
	"database/sql"
	"document-server/internal/storage"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserStorage struct {
	db *sqlx.DB
}

func NewUserStorage(db *sqlx.DB) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) Create(ctx context.Context, user User) error {
	query := "INSERT INTO users (login, password_hash) VALUES ($1, $2)"
	_, err := s.db.Exec(query, user.Login, user.PasswordHash)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStorage) GetUserByLogin(ctx context.Context, login string) (User, error) {
	var user User
	query := "SELECT id, login, password_hash, created_at FROM users WHERE login=$1"
	err := s.db.GetContext(ctx, &user, query, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, storage.ErrUserNotFound
		}
		return User{}, err
	}

	return user, nil
}
func (s *UserStorage) GetUserByID(ctx context.Context, uuid uuid.UUID) (User, error) {
	var user User
	query := "SELECT id, login, password_hash, created_at FROM users WHERE id=$1"
	err := s.db.GetContext(ctx, &user, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, storage.ErrUserNotFound
		}
		return User{}, err
	}

	return user, nil
}
