package storage

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrDocumentNotFound = errors.New("document not found")
	ErrUserExists       = errors.New("user with this login already exists")
	ErrTokenNotFound    = errors.New("token not found")
)
