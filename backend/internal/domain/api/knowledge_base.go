package api

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
)

// KnowledgeBaseApi provides business operations for knowledge base document management.
type KnowledgeBaseApi struct {
	docPort port.KnowledgeDocument
	ragApi  *RAGApi
	logger  *zerolog.Logger
}

// NewKnowledgeBaseApi creates a new KnowledgeBaseApi with the given dependencies.
func NewKnowledgeBaseApi(docPort port.KnowledgeDocument, ragApi *RAGApi, logger *zerolog.Logger) *KnowledgeBaseApi {
	return &KnowledgeBaseApi{
		docPort: docPort,
		ragApi:  ragApi,
		logger:  logger,
	}
}

// Upload saves a file and triggers async ingestion via the RAG pipeline.
func (k *KnowledgeBaseApi) Upload(ctx context.Context, filename string, mimeType string, data []byte) (model.KnowledgeDocument, error) {
	k.logger.Info().Str("filename", filename).Str("mimeType", mimeType).Int("bytes", len(data)).Msg("[KnowledgeBaseApi] Upload started")
	doc, err := k.ragApi.IngestDocument(ctx, filename, mimeType, data)
	if err != nil {
		k.logger.Error().Err(err).Str("filename", filename).Msg("[KnowledgeBaseApi] Upload failed")
		return model.KnowledgeDocument{}, fmt.Errorf("failed to ingest document: %w", err)
	}
	k.logger.Info().Str("documentId", doc.ID).Str("filename", filename).Msg("[KnowledgeBaseApi] Upload completed, indexing in progress")
	return doc, nil
}

// List retrieves all knowledge documents with their current status.
func (k *KnowledgeBaseApi) List(ctx context.Context) ([]model.KnowledgeDocument, error) {
	k.logger.Debug().Msg("[KnowledgeBaseApi] Listing all documents")
	docs, err := k.docPort.List(ctx)
	if err != nil {
		k.logger.Error().Err(err).Msg("[KnowledgeBaseApi] Failed to list documents")
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	k.logger.Debug().Int("count", len(docs)).Msg("[KnowledgeBaseApi] Documents listed")
	return docs, nil
}

// FindByID retrieves a specific knowledge document by its ID.
func (k *KnowledgeBaseApi) FindByID(ctx context.Context, id string) (model.KnowledgeDocument, error) {
	k.logger.Debug().Str("documentId", id).Msg("[KnowledgeBaseApi] Finding document by ID")
	doc, err := k.docPort.FindByID(ctx, id)
	if err != nil {
		k.logger.Error().Err(err).Str("documentId", id).Msg("[KnowledgeBaseApi] Failed to find document")
		return model.KnowledgeDocument{}, fmt.Errorf("failed to find document: %w", err)
	}
	k.logger.Debug().Str("documentId", id).Str("status", doc.Status).Msg("[KnowledgeBaseApi] Document found")
	return doc, nil
}

// Delete removes a document and all its associated chunks.
func (k *KnowledgeBaseApi) Delete(ctx context.Context, id string) error {
	k.logger.Info().Str("documentId", id).Msg("[KnowledgeBaseApi] Delete started")
	if err := k.ragApi.DeleteDocument(ctx, id); err != nil {
		k.logger.Error().Err(err).Str("documentId", id).Msg("[KnowledgeBaseApi] Delete failed")
		return fmt.Errorf("failed to delete document: %w", err)
	}
	k.logger.Info().Str("documentId", id).Msg("[KnowledgeBaseApi] Document deleted")
	return nil
}

// Reindex re-processes an existing document through the RAG pipeline.
func (k *KnowledgeBaseApi) Reindex(ctx context.Context, id string) error {
	k.logger.Info().Str("documentId", id).Msg("[KnowledgeBaseApi] Reindex started")
	if err := k.ragApi.ReindexDocument(ctx, id); err != nil {
		k.logger.Error().Err(err).Str("documentId", id).Msg("[KnowledgeBaseApi] Reindex failed")
		return fmt.Errorf("failed to reindex document: %w", err)
	}
	k.logger.Info().Str("documentId", id).Msg("[KnowledgeBaseApi] Reindex triggered")
	return nil
}
