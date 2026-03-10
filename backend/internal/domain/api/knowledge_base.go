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
	doc, err := k.ragApi.IngestDocument(ctx, filename, mimeType, data)
	if err != nil {
		return model.KnowledgeDocument{}, fmt.Errorf("failed to ingest document: %w", err)
	}
	return doc, nil
}

// List retrieves all knowledge documents with their current status.
func (k *KnowledgeBaseApi) List(ctx context.Context) ([]model.KnowledgeDocument, error) {
	docs, err := k.docPort.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	return docs, nil
}

// FindByID retrieves a specific knowledge document by its ID.
func (k *KnowledgeBaseApi) FindByID(ctx context.Context, id string) (model.KnowledgeDocument, error) {
	doc, err := k.docPort.FindByID(ctx, id)
	if err != nil {
		return model.KnowledgeDocument{}, fmt.Errorf("failed to find document: %w", err)
	}
	return doc, nil
}

// Delete removes a document and all its associated chunks.
func (k *KnowledgeBaseApi) Delete(ctx context.Context, id string) error {
	if err := k.ragApi.DeleteDocument(ctx, id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// Reindex re-processes an existing document through the RAG pipeline.
func (k *KnowledgeBaseApi) Reindex(ctx context.Context, id string) error {
	if err := k.ragApi.ReindexDocument(ctx, id); err != nil {
		return fmt.Errorf("failed to reindex document: %w", err)
	}
	return nil
}
