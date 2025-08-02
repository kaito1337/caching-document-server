package storage

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Document struct {
	ID        uuid.UUID      `db:"id"`
	Name      string         `db:"name"`
	MimeType  string         `db:"mime_type"`
	IsPublic  bool           `db:"public"`
	IsFile    bool           `db:"file"`
	FilePath  sql.NullString `db:"content_path"`
	JSONData  sql.NullString `db:"json_content"`
	CreatedAt time.Time      `db:"created_at"`
	GrantedTo pq.StringArray `db:"granted_to"`
}
