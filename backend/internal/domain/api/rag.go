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
	"phone-call-receptionist/backend/pkg/helpers"
)

// RAGApi provides business operations for the Retrieval-Augmented Generation pipeline.
type RAGApi struct {
	docPort     port.KnowledgeDocument
	vectorDB    port.VectorDB
	embedding   port.Embedding
	llm         port.LLM
	fileStorage port.FileStorage
	logger      *zerolog.Logger
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
	r.logger.Info().Str("filename", filename).Str("mimeType", mimeType).Int("bytes", len(data)).Msg("[RAGApi] IngestDocument started")

	filePath, err := r.fileStorage.Create(ctx, filename, data)
	if err != nil {
		r.logger.Error().Err(err).Str("filename", filename).Msg("[RAGApi] Failed to save file to storage")
		return model.KnowledgeDocument{}, fmt.Errorf("failed to save file: %w", err)
	}
	r.logger.Info().Str("filename", filename).Str("filePath", filePath).Msg("[RAGApi] File saved to storage")

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
		r.logger.Error().Err(err).Str("documentId", doc.ID).Msg("[RAGApi] Failed to create document record")
		return model.KnowledgeDocument{}, fmt.Errorf("failed to create document: %w", err)
	}
	r.logger.Info().Str("documentId", doc.ID).Str("filename", filename).Msg("[RAGApi] Document record created, starting async indexing")

	go r.processDocument(context.Background(), doc, data)

	return doc, nil
}

// processDocument extracts text based on mime type, chunks it, embeds, and stores in vector DB.
func (r *RAGApi) processDocument(ctx context.Context, doc model.KnowledgeDocument, data []byte) {
	r.logger.Info().Str("documentId", doc.ID).Str("filename", doc.Filename).Str("mimeType", doc.MimeType).Msg("[RAGApi] processDocument started (async)")

	// Step 1: Extract text based on mime type
	pages, err := r.extractPages(doc, data)
	if err != nil {
		r.logger.Error().Err(err).Str("documentId", doc.ID).Msg("[RAGApi] Failed to extract text from document")
		r.markDocumentFailed(ctx, doc)
		return
	}
	r.logger.Info().Str("documentId", doc.ID).Int("pagesExtracted", len(pages)).Msg("[RAGApi] Text extraction completed")

	// Step 2: Chunk pages into smaller pieces
	chunks := r.chunkPages(pages)
	r.logger.Info().Str("documentId", doc.ID).Int("totalChunks", len(chunks)).Msg("[RAGApi] Text chunked from pages")

	if len(chunks) == 0 {
		r.logger.Warn().Str("documentId", doc.ID).Msg("[RAGApi] No chunks produced from document, marking as failed")
		r.markDocumentFailed(ctx, doc)
		return
	}

	// Step 3: Embed and store each chunk
	chunkCount := 0
	for i, chunk := range chunks {
		r.logger.Debug().Str("documentId", doc.ID).Int("chunk", i).Int("page", chunk.pageNumber).Int("chunkLen", len(chunk.content)).Msg("[RAGApi] Embedding chunk")

		emb, err := r.embedding.Embed(ctx, chunk.content)
		if err != nil {
			r.logger.Error().Err(err).Str("documentId", doc.ID).Int("chunk", i).Msg("[RAGApi] Failed to embed chunk")
			r.markDocumentFailed(ctx, doc)
			return
		}
		r.logger.Debug().Str("documentId", doc.ID).Int("chunk", i).Int("embeddingDims", len(emb)).Msg("[RAGApi] Chunk embedded successfully")

		modelChunk := model.Chunk{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			Content:    chunk.content,
			PageNumber: chunk.pageNumber,
			ChunkIndex: i,
			Embedding:  emb,
			CreatedAt:  time.Now().Format(time.RFC3339),
		}

		r.logger.Debug().Str("documentId", doc.ID).Int("chunk", i).Str("chunkId", modelChunk.ID).Msg("[RAGApi] Storing chunk in vector DB")
		if err := r.vectorDB.Create(ctx, modelChunk); err != nil {
			r.logger.Error().Err(err).Str("documentId", doc.ID).Int("chunk", i).Msg("[RAGApi] Failed to store chunk in vector DB")
			r.markDocumentFailed(ctx, doc)
			return
		}

		chunkCount++
		r.logger.Debug().Str("documentId", doc.ID).Int("chunk", i).Msg("[RAGApi] Chunk stored successfully")
	}

	// Step 4: Update document status
	now := time.Now().Format(time.RFC3339)
	doc.ChunkCount = chunkCount
	doc.Status = "indexed"
	doc.IndexedAt = now
	doc.UpdatedAt = now

	if err := r.docPort.Update(ctx, doc); err != nil {
		r.logger.Error().Err(err).Str("documentId", doc.ID).Msg("[RAGApi] Failed to update document status to indexed")
	}

	r.logger.Info().Str("documentId", doc.ID).Int("chunks", chunkCount).Msg("[RAGApi] Document indexed successfully")
}

// pageContent holds extracted text with its page number.
type pageContent struct {
	pageNumber int
	content    string
}

// chunkResult holds a chunk of text with metadata.
type chunkResult struct {
	content    string
	pageNumber int
}

// extractPages extracts text from the document based on its mime type.
func (r *RAGApi) extractPages(doc model.KnowledgeDocument, data []byte) ([]pageContent, error) {
	mimeType := strings.ToLower(doc.MimeType)
	r.logger.Debug().Str("documentId", doc.ID).Str("mimeType", mimeType).Msg("[RAGApi] Detecting document type for extraction")

	switch {
	case mimeType == "application/pdf":
		return r.extractPDFPages(doc, data)
	case strings.HasPrefix(mimeType, "text/"):
		return r.extractTextPages(doc, data)
	case mimeType == "application/json" || mimeType == "application/xml":
		return r.extractTextPages(doc, data)
	case mimeType == "application/octet-stream":
		// Try as text
		r.logger.Debug().Str("documentId", doc.ID).Msg("[RAGApi] Unknown binary type, attempting text extraction")
		return r.extractTextPages(doc, data)
	default:
		// Fallback: treat as plain text
		r.logger.Warn().Str("documentId", doc.ID).Str("mimeType", mimeType).Msg("[RAGApi] Unsupported mime type, falling back to text extraction")
		return r.extractTextPages(doc, data)
	}
}

// extractPDFPages uses the PDF library to extract text per page.
func (r *RAGApi) extractPDFPages(doc model.KnowledgeDocument, data []byte) ([]pageContent, error) {
	r.logger.Info().Str("documentId", doc.ID).Int("bytes", len(data)).Msg("[RAGApi] Extracting text from PDF")

	pdfPages, err := helpers.ExtractPDFPages(data)
	if err != nil {
		r.logger.Error().Err(err).Str("documentId", doc.ID).Msg("[RAGApi] PDF text extraction failed")
		return nil, fmt.Errorf("failed to extract PDF text: %w", err)
	}

	var pages []pageContent
	for _, p := range pdfPages {
		if strings.TrimSpace(p.Content) == "" {
			r.logger.Debug().Str("documentId", doc.ID).Int("page", p.Number).Msg("[RAGApi] Skipping empty PDF page")
			continue
		}
		r.logger.Debug().Str("documentId", doc.ID).Int("page", p.Number).Int("contentLen", len(p.Content)).Msg("[RAGApi] PDF page extracted")
		pages = append(pages, pageContent{
			pageNumber: p.Number,
			content:    p.Content,
		})
	}

	if len(pages) == 0 {
		r.logger.Warn().Str("documentId", doc.ID).Msg("[RAGApi] No text content found in PDF, falling back to raw text")
		return r.extractTextPages(doc, data)
	}

	r.logger.Info().Str("documentId", doc.ID).Int("pagesWithContent", len(pages)).Int("totalPages", len(pdfPages)).Msg("[RAGApi] PDF text extraction completed")
	return pages, nil
}

// extractTextPages treats the entire file content as a single page of text.
func (r *RAGApi) extractTextPages(doc model.KnowledgeDocument, data []byte) ([]pageContent, error) {
	content := strings.TrimSpace(string(data))
	if content == "" {
		r.logger.Warn().Str("documentId", doc.ID).Msg("[RAGApi] Document has no text content")
		return nil, fmt.Errorf("document has no text content")
	}

	r.logger.Info().Str("documentId", doc.ID).Int("contentLen", len(content)).Msg("[RAGApi] Text content extracted")
	return []pageContent{
		{pageNumber: 1, content: content},
	}, nil
}

// chunkPages splits page contents into smaller chunks with overlap for better retrieval.
func (r *RAGApi) chunkPages(pages []pageContent) []chunkResult {
	const chunkSize = 500  // words per chunk
	const overlap = 50     // overlap words between chunks

	var allChunks []chunkResult
	for _, page := range pages {
		pageChunks := r.chunkText(page.content, chunkSize, overlap)
		for _, text := range pageChunks {
			allChunks = append(allChunks, chunkResult{
				content:    text,
				pageNumber: page.pageNumber,
			})
		}
	}
	return allChunks
}

func (r *RAGApi) markDocumentFailed(ctx context.Context, doc model.KnowledgeDocument) {
	r.logger.Warn().Str("documentId", doc.ID).Msg("[RAGApi] Marking document as failed")
	doc.Status = "failed"
	doc.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := r.docPort.Update(ctx, doc); err != nil {
		r.logger.Error().Err(err).Str("documentId", doc.ID).Msg("[RAGApi] Failed to mark document as failed")
	}
}

// chunkText splits text into chunks of approximately chunkSize words with overlap.
func (r *RAGApi) chunkText(text string, chunkSize int, overlap int) []string {
	// First try paragraph-based chunking for better semantic boundaries
	paragraphs := strings.Split(text, "\n\n")
	if len(paragraphs) > 1 {
		return r.chunkByParagraphs(paragraphs, chunkSize)
	}

	// Fallback to word-based chunking
	return r.chunkByWords(text, chunkSize, overlap)
}

// chunkByParagraphs groups paragraphs into chunks that don't exceed chunkSize words.
func (r *RAGApi) chunkByParagraphs(paragraphs []string, maxWords int) []string {
	var chunks []string
	var current []string
	currentWords := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		paraWords := len(strings.Fields(para))
		if paraWords == 0 {
			continue
		}

		// If single paragraph exceeds max, chunk it by words
		if paraWords > maxWords {
			if currentWords > 0 {
				chunks = append(chunks, strings.Join(current, "\n\n"))
				current = nil
				currentWords = 0
			}
			wordChunks := r.chunkByWords(para, maxWords, 50)
			chunks = append(chunks, wordChunks...)
			continue
		}

		// If adding this paragraph would exceed max, flush current
		if currentWords+paraWords > maxWords && currentWords > 0 {
			chunks = append(chunks, strings.Join(current, "\n\n"))
			current = nil
			currentWords = 0
		}

		current = append(current, para)
		currentWords += paraWords
	}

	if currentWords > 0 {
		chunks = append(chunks, strings.Join(current, "\n\n"))
	}

	return chunks
}

// chunkByWords splits text into chunks of chunkSize words with overlap.
func (r *RAGApi) chunkByWords(text string, chunkSize int, overlap int) []string {
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
	r.logger.Info().Str("question", question).Int("topK", topK).Msg("[RAGApi] QueryKnowledgeBase started")

	r.logger.Debug().Str("question", question).Msg("[RAGApi] Embedding question")
	emb, err := r.embedding.Embed(ctx, question)
	if err != nil {
		r.logger.Error().Err(err).Str("question", question).Msg("[RAGApi] Failed to embed question")
		return "", nil, "", 0, fmt.Errorf("failed to embed question: %w", err)
	}
	r.logger.Debug().Int("embeddingDims", len(emb)).Msg("[RAGApi] Question embedded")

	r.logger.Debug().Int("topK", topK).Msg("[RAGApi] Searching vector DB")
	chunks, err := r.vectorDB.Search(ctx, emb, topK)
	if err != nil {
		r.logger.Error().Err(err).Msg("[RAGApi] Failed to search vector DB")
		return "", nil, "", 0, fmt.Errorf("failed to search vector db: %w", err)
	}
	r.logger.Info().Int("chunksFound", len(chunks)).Msg("[RAGApi] Vector DB search completed")

	if len(chunks) == 0 {
		r.logger.Warn().Str("question", question).Msg("[RAGApi] No relevant chunks found")
		return "Je n'ai pas cette information spécifique. Souhaitez-vous que je vous mette en contact avec notre équipe ?", nil, r.llm.Provider(), 0, nil
	}

	var contextParts []string
	for _, chunk := range chunks {
		contextParts = append(contextParts, chunk.Content)
	}
	contextText := strings.Join(contextParts, "\n\n---\n\n")

	systemPrompt := `Tu es un assistant réceptionniste professionnel pour une entreprise de services informatiques, tu t'appelles "Alex".
Tu parles clairement et de manière concise — les réponses doivent faire moins de 2 phrases pour la diffusion vocale.
Réponds UNIQUEMENT en utilisant le contexte fourni. Si le contexte ne contient pas la réponse, dis exactement :
"Je n'ai pas cette information spécifique. Souhaitez-vous que je vous mette en contact avec notre équipe ?"
Ne JAMAIS inventer de prix, de détails de service ou de disponibilités.
Ne JAMAIS utiliser de formatage markdown — uniquement du texte parlé.
Réponds TOUJOURS en français.`

	userPrompt := fmt.Sprintf("Context:\n%s\n\nQuestion: %s", contextText, question)

	r.logger.Debug().Int("contextLen", len(contextText)).Msg("[RAGApi] Generating LLM response")
	answer, tokens, err := r.llm.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		r.logger.Error().Err(err).Msg("[RAGApi] Failed to generate LLM answer")
		return "", nil, "", 0, fmt.Errorf("failed to generate answer: %w", err)
	}

	r.logger.Info().Str("provider", r.llm.Provider()).Int("tokens", tokens).Int("answerLen", len(answer)).Msg("[RAGApi] LLM answer generated")
	return answer, chunks, r.llm.Provider(), tokens, nil
}

// DeleteDocument removes a document and its chunks from both the database and vector store.
func (r *RAGApi) DeleteDocument(ctx context.Context, id string) error {
	r.logger.Info().Str("documentId", id).Msg("[RAGApi] DeleteDocument started")

	if err := r.vectorDB.DeleteByDocumentID(ctx, id); err != nil {
		r.logger.Warn().Err(err).Str("documentId", id).Msg("[RAGApi] Failed to delete chunks from vector DB")
	} else {
		r.logger.Info().Str("documentId", id).Msg("[RAGApi] Chunks deleted from vector DB")
	}

	doc, err := r.docPort.FindByID(ctx, id)
	if err != nil {
		r.logger.Error().Err(err).Str("documentId", id).Msg("[RAGApi] Failed to find document for deletion")
		return fmt.Errorf("failed to find document: %w", err)
	}

	if doc.FilePath != "" {
		if err := r.fileStorage.Delete(ctx, doc.FilePath); err != nil {
			r.logger.Warn().Err(err).Str("path", doc.FilePath).Msg("[RAGApi] Failed to delete file from storage")
		} else {
			r.logger.Info().Str("path", doc.FilePath).Msg("[RAGApi] File deleted from storage")
		}
	}

	if err := r.docPort.Delete(ctx, id); err != nil {
		r.logger.Error().Err(err).Str("documentId", id).Msg("[RAGApi] Failed to delete document record")
		return fmt.Errorf("failed to delete document: %w", err)
	}

	r.logger.Info().Str("documentId", id).Msg("[RAGApi] Document deleted successfully")
	return nil
}

// ReindexDocument re-processes an existing document by deleting old chunks and re-ingesting.
func (r *RAGApi) ReindexDocument(ctx context.Context, id string) error {
	r.logger.Info().Str("documentId", id).Msg("[RAGApi] ReindexDocument started")

	doc, err := r.docPort.FindByID(ctx, id)
	if err != nil {
		r.logger.Error().Err(err).Str("documentId", id).Msg("[RAGApi] Failed to find document for reindex")
		return fmt.Errorf("failed to find document: %w", err)
	}

	if err := r.vectorDB.DeleteByDocumentID(ctx, id); err != nil {
		r.logger.Warn().Err(err).Str("documentId", id).Msg("[RAGApi] Failed to delete old chunks from vector DB")
	} else {
		r.logger.Info().Str("documentId", id).Msg("[RAGApi] Old chunks deleted from vector DB")
	}

	data, err := r.fileStorage.Find(ctx, doc.FilePath)
	if err != nil {
		r.logger.Error().Err(err).Str("documentId", id).Str("filePath", doc.FilePath).Msg("[RAGApi] Failed to load file for reindex")
		return fmt.Errorf("failed to load file: %w", err)
	}
	r.logger.Info().Str("documentId", id).Int("bytes", len(data)).Msg("[RAGApi] File loaded for reindex")

	doc.Status = "indexing"
	doc.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := r.docPort.Update(ctx, doc); err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	go r.processDocument(context.Background(), doc, data)

	r.logger.Info().Str("documentId", id).Msg("[RAGApi] Reindex started (async)")
	return nil
}
