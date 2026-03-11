// Package weaviate implements the VectorDB port using Weaviate as the backing store.
package weaviate

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
)

const className = "KnowledgeChunk"

// Adapter implements port.VectorDB using Weaviate.
type Adapter struct {
	client *weaviate.Client
	logger *zerolog.Logger
}

// NewWeaviateAdapter creates a new Weaviate adapter and ensures the KnowledgeChunk class exists.
func NewWeaviateAdapter(client *weaviate.Client, logger *zerolog.Logger) port.VectorDB {
	a := &Adapter{
		client: client,
		logger: logger,
	}
	return a
}

// EnsureClass creates the KnowledgeChunk class in Weaviate if it does not already exist.
func (a *Adapter) EnsureClass(ctx context.Context) error {
	a.logger.Debug().Str("class", className).Msg("[WeaviateAdapter] ensuring class exists")

	exists, err := a.client.Schema().ClassExistenceChecker().WithClassName(className).Do(ctx)
	if err != nil {
		a.logger.Error().Err(err).Str("class", className).Msg("[WeaviateAdapter] failed to check class existence")
		return fmt.Errorf("failed to check class existence: %w", err)
	}
	if exists {
		a.logger.Debug().Str("class", className).Msg("Weaviate class already exists")
		return nil
	}

	classObj := &models.Class{
		Class:       className,
		Description: "Knowledge document chunks with embeddings for RAG retrieval",
		Properties: []*models.Property{
			{
				Name:        "documentID",
				DataType:    []string{"text"},
				Description: "The identifier of the parent knowledge document",
			},
			{
				Name:        "content",
				DataType:    []string{"text"},
				Description: "The text content of the chunk",
			},
			{
				Name:        "pageNumber",
				DataType:    []string{"int"},
				Description: "The page number where the chunk originated",
			},
			{
				Name:        "chunkIndex",
				DataType:    []string{"int"},
				Description: "The sequential index of this chunk within the document",
			},
		},
		VectorIndexType: "hnsw",
		Vectorizer:      "none",
	}

	err = a.client.Schema().ClassCreator().WithClass(classObj).Do(ctx)
	if err != nil {
		a.logger.Error().Err(err).Str("class", className).Msg("[WeaviateAdapter] failed to create class")
		return fmt.Errorf("failed to create weaviate class: %w", err)
	}

	a.logger.Info().Str("class", className).Msg("[WeaviateAdapter] created class")
	return nil
}

// Create persists a chunk with its embedding to the Weaviate vector database.
func (a *Adapter) Create(ctx context.Context, chunk model.Chunk) error {
	a.logger.Debug().Str("chunkID", chunk.ID).Str("documentID", chunk.DocumentID).Msg("[WeaviateAdapter] storing chunk")

	properties := map[string]interface{}{
		"documentID": chunk.DocumentID,
		"content":    chunk.Content,
		"pageNumber": chunk.PageNumber,
		"chunkIndex": chunk.ChunkIndex,
	}

	_, err := a.client.Data().Creator().
		WithClassName(className).
		WithID(chunk.ID).
		WithProperties(properties).
		WithVector(chunk.Embedding).
		Do(ctx)
	if err != nil {
		a.logger.Error().Err(err).Str("chunkID", chunk.ID).Str("documentID", chunk.DocumentID).Msg("[WeaviateAdapter] failed to store chunk")
		return fmt.Errorf("failed to store chunk in weaviate: %w", err)
	}

	a.logger.Debug().
		Str("chunkID", chunk.ID).
		Str("documentID", chunk.DocumentID).
		Msg("[WeaviateAdapter] stored chunk")

	return nil
}

// Search finds the most similar chunks to the given embedding vector.
func (a *Adapter) Search(ctx context.Context, embedding []float32, topK int) ([]model.Chunk, error) {
	a.logger.Debug().Int("topK", topK).Int("embeddingLen", len(embedding)).Msg("[WeaviateAdapter] searching chunks")
	nearVector := a.client.GraphQL().NearVectorArgBuilder().WithVector(embedding)

	fields := []graphql.Field{
		{Name: "_additional { id distance }"},
		{Name: "documentID"},
		{Name: "content"},
		{Name: "pageNumber"},
		{Name: "chunkIndex"},
	}

	result, err := a.client.GraphQL().Get().
		WithClassName(className).
		WithFields(fields...).
		WithNearVector(nearVector).
		WithLimit(topK).
		Do(ctx)
	if err != nil {
		a.logger.Error().Err(err).Int("topK", topK).Msg("[WeaviateAdapter] failed to search chunks")
		return nil, fmt.Errorf("failed to search weaviate: %w", err)
	}

	if result.Errors != nil && len(result.Errors) > 0 {
		a.logger.Error().Str("error", result.Errors[0].Message).Msg("[WeaviateAdapter] search returned errors")
		return nil, fmt.Errorf("failed to search weaviate: %s", result.Errors[0].Message)
	}

	getData, ok := result.Data["Get"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	classData, ok := getData[className].([]interface{})
	if !ok {
		return nil, nil
	}

	chunks := make([]model.Chunk, 0, len(classData))
	for _, item := range classData {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		chunk := model.Chunk{
			Content:    stringFromMap(obj, "content"),
			DocumentID: stringFromMap(obj, "documentID"),
			PageNumber: intFromMap(obj, "pageNumber"),
			ChunkIndex: intFromMap(obj, "chunkIndex"),
		}

		if additional, ok := obj["_additional"].(map[string]interface{}); ok {
			chunk.ID = stringFromMap(additional, "id")
		}

		chunks = append(chunks, chunk)
	}

	a.logger.Debug().Int("results", len(chunks)).Msg("[WeaviateAdapter] search completed")
	return chunks, nil
}

// DeleteByDocumentID removes all chunks belonging to a specific document.
func (a *Adapter) DeleteByDocumentID(ctx context.Context, documentID string) error {
	a.logger.Debug().Str("documentID", documentID).Msg("[WeaviateAdapter] deleting chunks by document ID")
	where := filters.Where().
		WithPath([]string{"documentID"}).
		WithOperator(filters.Equal).
		WithValueText(documentID)

	result, err := a.client.Batch().ObjectsBatchDeleter().
		WithClassName(className).
		WithWhere(where).
		Do(ctx)
	if err != nil {
		a.logger.Error().Err(err).Str("documentID", documentID).Msg("[WeaviateAdapter] failed to delete chunks")
		return fmt.Errorf("failed to delete chunks from weaviate: %w", err)
	}

	a.logger.Info().
		Str("documentID", documentID).
		Int64("deleted", result.Results.Matches).
		Msg("[WeaviateAdapter] deleted chunks")

	return nil
}

// stringFromMap safely extracts a string value from a map.
func stringFromMap(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// intFromMap safely extracts an int value from a map.
func intFromMap(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}
