package postgres

import (
	"context"
	"fmt"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"

	"github.com/rs/zerolog"
)

// knowledgeDocumentAdapter implements port.KnowledgeDocument using PostgreSQL.
type knowledgeDocumentAdapter struct {
	client *Client
	logger *zerolog.Logger
}

// NewKnowledgeDocumentAdapter creates a new PostgreSQL knowledge document adapter.
func NewKnowledgeDocumentAdapter(client *Client, logger *zerolog.Logger) port.KnowledgeDocument {
	return &knowledgeDocumentAdapter{client: client, logger: logger}
}

// Create persists a new knowledge document to PostgreSQL.
func (a *knowledgeDocumentAdapter) Create(ctx context.Context, doc model.KnowledgeDocument) error {
	a.logger.Debug().Str("id", doc.ID).Str("filename", doc.Filename).Msg("[PostgresKnowledgeDocument] creating document")

	var db KnowledgeDocumentDB
	db.FromDomain(doc)

	query := `INSERT INTO knowledge_documents (id, filename, mime_type, file_path, chunk_count, status, indexed_at, created_at, updated_at)
	           VALUES (:id, :filename, :mime_type, :file_path, :chunk_count, :status, :indexed_at, :created_at, :updated_at)`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		a.logger.Error().Err(err).Str("id", doc.ID).Msg("[PostgresKnowledgeDocument] failed to create document")
		return fmt.Errorf("failed to create knowledge document: %w", err)
	}

	a.logger.Debug().Str("id", doc.ID).Msg("[PostgresKnowledgeDocument] document created")
	return nil
}

// FindByID retrieves a knowledge document by its unique identifier from PostgreSQL.
func (a *knowledgeDocumentAdapter) FindByID(ctx context.Context, id string) (model.KnowledgeDocument, error) {
	a.logger.Debug().Str("id", id).Msg("[PostgresKnowledgeDocument] finding document by ID")

	var db KnowledgeDocumentDB
	query := `SELECT id, filename, mime_type, file_path, chunk_count, status, indexed_at, created_at, updated_at
	           FROM knowledge_documents WHERE id = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, id); err != nil {
		a.logger.Error().Err(err).Str("id", id).Msg("[PostgresKnowledgeDocument] failed to find document by ID")
		return model.KnowledgeDocument{}, fmt.Errorf("failed to find knowledge document by id: %w", err)
	}

	a.logger.Debug().Str("id", id).Str("filename", db.Filename).Msg("[PostgresKnowledgeDocument] document found")
	return db.ToDomain(), nil
}

// List retrieves all knowledge documents from PostgreSQL.
func (a *knowledgeDocumentAdapter) List(ctx context.Context) ([]model.KnowledgeDocument, error) {
	a.logger.Debug().Msg("[PostgresKnowledgeDocument] listing documents")

	query := `SELECT id, filename, mime_type, file_path, chunk_count, status, indexed_at, created_at, updated_at
	           FROM knowledge_documents ORDER BY created_at DESC`

	var rows []KnowledgeDocumentDB
	if err := a.client.DB.SelectContext(ctx, &rows, query); err != nil {
		a.logger.Error().Err(err).Msg("[PostgresKnowledgeDocument] failed to list documents")
		return nil, fmt.Errorf("failed to list knowledge documents: %w", err)
	}

	docs := make([]model.KnowledgeDocument, len(rows))
	for i, row := range rows {
		docs[i] = row.ToDomain()
	}

	a.logger.Debug().Int("count", len(docs)).Msg("[PostgresKnowledgeDocument] documents listed")
	return docs, nil
}

// Update modifies an existing knowledge document's data in PostgreSQL.
func (a *knowledgeDocumentAdapter) Update(ctx context.Context, doc model.KnowledgeDocument) error {
	a.logger.Debug().Str("id", doc.ID).Str("status", doc.Status).Msg("[PostgresKnowledgeDocument] updating document")

	var db KnowledgeDocumentDB
	db.FromDomain(doc)

	query := `UPDATE knowledge_documents SET filename = :filename, mime_type = :mime_type,
	           file_path = :file_path, chunk_count = :chunk_count, status = :status,
	           indexed_at = :indexed_at, updated_at = :updated_at WHERE id = :id`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		a.logger.Error().Err(err).Str("id", doc.ID).Msg("[PostgresKnowledgeDocument] failed to update document")
		return fmt.Errorf("failed to update knowledge document: %w", err)
	}

	a.logger.Debug().Str("id", doc.ID).Msg("[PostgresKnowledgeDocument] document updated")
	return nil
}

// Delete removes a knowledge document from PostgreSQL.
func (a *knowledgeDocumentAdapter) Delete(ctx context.Context, id string) error {
	a.logger.Debug().Str("id", id).Msg("[PostgresKnowledgeDocument] deleting document")

	query := `DELETE FROM knowledge_documents WHERE id = $1`

	_, err := a.client.DB.ExecContext(ctx, query, id)
	if err != nil {
		a.logger.Error().Err(err).Str("id", id).Msg("[PostgresKnowledgeDocument] failed to delete document")
		return fmt.Errorf("failed to delete knowledge document: %w", err)
	}

	a.logger.Debug().Str("id", id).Msg("[PostgresKnowledgeDocument] document deleted")
	return nil
}
