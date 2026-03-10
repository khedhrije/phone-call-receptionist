package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
)

// RAGApi provides business operations for the Retrieval-Augmented Generation pipeline.
type RAGApi struct {
	docPort       port.KnowledgeDocument
	vectorDB      port.VectorDB
	embedding     port.Embedding
	llm           port.LLM
	fileStorage   port.FileStorage
	logger        *zerolog.Logger
}

// NewRAGApi creates a new RAGApi with the given dependencies.
func NewRAGApi(
	docPort port.KnowledgeDocument,
	vectorDB port.VectorDB,
	embedding port.Embedding,
	llm port.LLM,
	fileStorage port.FileStorage,
	logger *zerolog.Logger,
) *RAGApi {
	return &RAGApi{
		docPort:     docPort,
		vectorDB:    vectorDB,
		embedding:   embedding,
		llm:         llm,
		fileStorage: fileStorage,
		logger:      logger,
	}
}

// IngestDocument processes a file by extracting text, chunking, embedding, and storing vectors.
func (r *RAGApi) IngestDocument(ctx context.Context, filename string, mimeType string, data []byte) (model.KnowledgeDocument, error) {
	filePath, err := r.fileStorage.Save(ctx, filename, data)
	if err != nil {
		return model.KnowledgeDocument{}, fmt.Errorf("failed to save file: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	doc := model.KnowledgeDocument{
		ID:        uuid.New().String(),
		Filename:  filename,
		MimeType:  mimeType,
		FilePath:  filePath,
		Status:    "indexing",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := r.docPort.Create(ctx, doc); err != nil {
		return model.KnowledgeDocument{}, fmt.Errorf("failed to create document: %w", err)
	}

	go r.processDocument(context.Background(), doc, data)

	return doc, nil
}

func (r *RAGApi) processDocument(ctx context.Context, doc model.KnowledgeDocument, data []byte) {
	content := string(data)
	chunks := r.chunkText(content, 500, 50)

	chunkCount := 0
	for i, chunkText := range chunks {
		emb, err := r.embedding.Embed(ctx, chunkText)
		if err != nil {
			r.logger.Error().Err(err).Str("documentId", doc.ID).Int("chunk", i).Msg("Failed to embed chunk")
			r.markDocumentFailed(ctx, doc)
			return
		}

		chunk := model.Chunk{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			Content:    chunkText,
			PageNumber: 1,
			ChunkIndex: i,
			Embedding:  emb,
			CreatedAt:  time.Now().Format(time.RFC3339),
		}

		if err := r.vectorDB.Store(ctx, chunk); err != nil {
			r.logger.Error().Err(err).Str("documentId", doc.ID).Int("chunk", i).Msg("Failed to store chunk")
			r.markDocumentFailed(ctx, doc)
			return
		}

		chunkCount++
	}

	now := time.Now().Format(time.RFC3339)
	doc.ChunkCount = chunkCount
	doc.Status = "indexed"
	doc.IndexedAt = now
	doc.UpdatedAt = now

	if err := r.docPort.Update(ctx, doc); err != nil {
		r.logger.Error().Err(err).Str("documentId", doc.ID).Msg("Failed to update document status")
	}

	r.logger.Info().Str("documentId", doc.ID).Int("chunks", chunkCount).Msg("Document indexed successfully")
}

func (r *RAGApi) markDocumentFailed(ctx context.Context, doc model.KnowledgeDocument) {
	doc.Status = "failed"
	doc.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := r.docPort.Update(ctx, doc); err != nil {
		r.logger.Error().Err(err).Str("documentId", doc.ID).Msg("Failed to mark document as failed")
	}
}

func (r *RAGApi) chunkText(text string, chunkSize int, overlap int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var chunks []string
	for i := 0; i < len(words); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
		if end == len(words) {
			break
		}
	}

	return chunks
}

// QueryKnowledgeBase performs a RAG query: embeds the question, retrieves relevant chunks,
// and generates an answer using the LLM.
func (r *RAGApi) QueryKnowledgeBase(ctx context.Context, question string, topK int) (string, []model.Chunk, string, int, error) {
	emb, err := r.embedding.Embed(ctx, question)
	if err != nil {
		return "", nil, "", 0, fmt.Errorf("failed to embed question: %w", err)
	}

	chunks, err := r.vectorDB.Search(ctx, emb, topK)
	if err != nil {
		return "", nil, "", 0, fmt.Errorf("failed to search vector db: %w", err)
	}

	if len(chunks) == 0 {
		return "I don't have that specific information. Would you like me to connect you with our team?", nil, r.llm.Provider(), 0, nil
	}

	var contextParts []string
	for _, chunk := range chunks {
		contextParts = append(contextParts, chunk.Content)
	}
	contextText := strings.Join(contextParts, "\n\n---\n\n")

	systemPrompt := `You are a professional IT services receptionist assistant named "Alex".
You speak clearly and concisely — responses must be under 2 sentences for voice delivery.
Answer ONLY using the provided context. If the context does not contain the answer, say exactly:
"I don't have that specific information. Would you like me to connect you with our team?"
NEVER fabricate pricing, service details, or availability.
NEVER use markdown formatting — plain speech only.`

	userPrompt := fmt.Sprintf("Context:\n%s\n\nQuestion: %s", contextText, question)

	answer, tokens, err := r.llm.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", nil, "", 0, fmt.Errorf("failed to generate answer: %w", err)
	}

	return answer, chunks, r.llm.Provider(), tokens, nil
}

// DeleteDocument removes a document and its chunks from both the database and vector store.
func (r *RAGApi) DeleteDocument(ctx context.Context, id string) error {
	if err := r.vectorDB.DeleteByDocumentID(ctx, id); err != nil {
		r.logger.Warn().Err(err).Str("documentId", id).Msg("Failed to delete chunks from vector db")
	}

	doc, err := r.docPort.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find document: %w", err)
	}

	if doc.FilePath != "" {
		if err := r.fileStorage.Delete(ctx, doc.FilePath); err != nil {
			r.logger.Warn().Err(err).Str("path", doc.FilePath).Msg("Failed to delete file")
		}
	}

	if err := r.docPort.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	r.logger.Info().Str("documentId", id).Msg("Document deleted")
	return nil
}

// ReindexDocument re-processes an existing document by deleting old chunks and re-ingesting.
func (r *RAGApi) ReindexDocument(ctx context.Context, id string) error {
	doc, err := r.docPort.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find document: %w", err)
	}

	if err := r.vectorDB.DeleteByDocumentID(ctx, id); err != nil {
		r.logger.Warn().Err(err).Str("documentId", id).Msg("Failed to delete old chunks")
	}

	data, err := r.fileStorage.Load(ctx, doc.FilePath)
	if err != nil {
		return fmt.Errorf("failed to load file: %w", err)
	}

	doc.Status = "indexing"
	doc.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := r.docPort.Update(ctx, doc); err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	go r.processDocument(context.Background(), doc, data)

	return nil
}
