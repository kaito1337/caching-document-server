CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL DEFAULT '',
    mime_type VARCHAR(255) NOT NULL DEFAULT '',
    file BOOLEAN NOT NULL DEFAULT FALSE,
    public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    content_path TEXT,
    json_content JSONB,
    granted_to TEXT[]
);

CREATE INDEX idx_documents_name_created_at ON documents (name, created_at DESC);