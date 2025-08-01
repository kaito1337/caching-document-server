package storage

import (
	"context"
	"database/sql"
	"document-server/internal/storage"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type DocumentStorage struct {
	db *sqlx.DB
}

func NewDocumentStorage(db *sqlx.DB) *DocumentStorage {
	return &DocumentStorage{db: db}
}

func (s *DocumentStorage) Create(ctx context.Context, doc Document) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	docQuery := `
		INSERT INTO documents (id, name, mime_type, public, file, content_path, json_content, granted_to)
		VALUES (:id, :name, :mime_type, :public, :file, :content_path, :json_content, :granted_to)
	`

	_, err = tx.NamedExecContext(ctx, docQuery, doc)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *DocumentStorage) GetByID(ctx context.Context, id string) (*Document, error) {
	var doc Document
	query := "SELECT * FROM documents WHERE id=$1"
	if err := s.db.GetContext(ctx, &doc, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrDocumentNotFound
		}
		return nil, err
	}
	return &doc, nil
}

func (s *DocumentStorage) ListDocumentIDs(ctx context.Context, currentLogin string, filterLogin string, key string, value string, limit int) ([]string, error) {
	searchLogin := currentLogin
	if filterLogin != "" {
		searchLogin = filterLogin
	}

	query := `SELECT id FROM documents WHERE $1 = ANY(granted_to)`
	args := []interface{}{searchLogin}

	if key != "" && value != "" {
		switch key {
		case "name":
			query += " AND name = $2"
		case "mime_type":
			query += " AND mime_type = $2"
		default:
			return nil, errors.New("unsupported filter key")
		}
		args = append(args, value)
	}

	query += " ORDER BY name ASC, created_at DESC"
	if limit > 0 {
		query += " LIMIT $3"
		args = append(args, limit)
	}

	var ids []string
	if err := s.db.SelectContext(ctx, &ids, query, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *DocumentStorage) GetDocumentsByIDs(ctx context.Context, ids []string) ([]Document, error) {
	query := `SELECT * FROM documents WHERE id = ANY($1)`
	var docs []Document
	if err := s.db.SelectContext(ctx, &docs, query, pq.Array(ids)); err != nil {
		return nil, err
	}
	return docs, nil
}

func (s *DocumentStorage) DeleteDocumentByID(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}
