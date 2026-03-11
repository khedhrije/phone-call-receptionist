package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/api"
	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/pkg/dtos/requests"
	"phone-call-receptionist/backend/pkg/dtos/responses"
)

// KnowledgeHandler handles knowledge base HTTP requests.
type KnowledgeHandler struct {
	kbApi  *api.KnowledgeBaseApi
	ragApi *api.RAGApi
	logger *zerolog.Logger
}

// NewKnowledgeHandler creates a new KnowledgeHandler with the given dependencies.
func NewKnowledgeHandler(kbApi *api.KnowledgeBaseApi, ragApi *api.RAGApi, logger *zerolog.Logger) *KnowledgeHandler {
	return &KnowledgeHandler{kbApi: kbApi, ragApi: ragApi, logger: logger}
}

// Upload godoc
//
//	@Summary		Upload a document
//	@Description	Upload a file to the knowledge base for indexing
//	@Tags			Knowledge
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Document file"
//	@Success		201		{object}	responses.KnowledgeDocumentResponse
//	@Security		BearerAuth
//	@Router			/knowledge/documents [post]
func (h *KnowledgeHandler) Upload(c *gin.Context) {
	h.logger.Info().Msg("[KnowledgeHandler] Upload request received")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("[KnowledgeHandler] Upload failed to read form file")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	h.logger.Info().
		Str("filename", header.Filename).
		Str("contentType", header.Header.Get("Content-Type")).
		Int64("size", header.Size).
		Msg("[KnowledgeHandler] Upload processing file")

	data, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error().Err(err).Str("filename", header.Filename).Msg("[KnowledgeHandler] Upload failed to read file data")
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{Error: "failed to read file"})
		return
	}

	doc, err := h.kbApi.Upload(c.Request.Context(), header.Filename, header.Header.Get("Content-Type"), data)
	if err != nil {
		h.logger.Error().Err(err).Str("filename", header.Filename).Msg("[KnowledgeHandler] Upload failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("documentID", doc.ID).Str("filename", doc.Filename).Msg("[KnowledgeHandler] Upload succeeded")

	c.JSON(http.StatusCreated, toDocResponse(doc))
}

// List godoc
//
//	@Summary		List documents
//	@Description	Returns all knowledge base documents
//	@Tags			Knowledge
//	@Produce		json
//	@Success		200	{array}	responses.KnowledgeDocumentResponse
//	@Security		BearerAuth
//	@Router			/knowledge/documents [get]
func (h *KnowledgeHandler) List(c *gin.Context) {
	h.logger.Info().Msg("[KnowledgeHandler] List request received")

	docs, err := h.kbApi.List(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("[KnowledgeHandler] List failed")
		HandleError(c, err)
		return
	}

	var items []responses.KnowledgeDocumentResponse
	for _, d := range docs {
		items = append(items, toDocResponse(d))
	}

	h.logger.Info().Int("count", len(items)).Msg("[KnowledgeHandler] List succeeded")

	c.JSON(http.StatusOK, items)
}

// FindByID godoc
//
//	@Summary		Get document detail
//	@Description	Returns a knowledge document by ID
//	@Tags			Knowledge
//	@Produce		json
//	@Param			id	path	string	true	"Document ID"
//	@Success		200	{object}	responses.KnowledgeDocumentResponse
//	@Security		BearerAuth
//	@Router			/knowledge/documents/{id} [get]
func (h *KnowledgeHandler) FindByID(c *gin.Context) {
	docID := c.Param("id")
	h.logger.Info().Str("documentID", docID).Msg("[KnowledgeHandler] FindByID request received")

	doc, err := h.kbApi.FindByID(c.Request.Context(), docID)
	if err != nil {
		h.logger.Error().Err(err).Str("documentID", docID).Msg("[KnowledgeHandler] FindByID failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("documentID", docID).Msg("[KnowledgeHandler] FindByID succeeded")

	c.JSON(http.StatusOK, toDocResponse(doc))
}

// Delete godoc
//
//	@Summary		Delete document
//	@Description	Delete a knowledge document and its chunks
//	@Tags			Knowledge
//	@Produce		json
//	@Param			id	path	string	true	"Document ID"
//	@Success		204
//	@Security		BearerAuth
//	@Router			/knowledge/documents/{id} [delete]
func (h *KnowledgeHandler) Delete(c *gin.Context) {
	docID := c.Param("id")
	h.logger.Info().Str("documentID", docID).Msg("[KnowledgeHandler] Delete request received")

	if err := h.kbApi.Delete(c.Request.Context(), docID); err != nil {
		h.logger.Error().Err(err).Str("documentID", docID).Msg("[KnowledgeHandler] Delete failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("documentID", docID).Msg("[KnowledgeHandler] Delete succeeded")

	c.Status(http.StatusNoContent)
}

// Reindex godoc
//
//	@Summary		Reindex document
//	@Description	Re-process document through the RAG pipeline
//	@Tags			Knowledge
//	@Produce		json
//	@Param			id	path	string	true	"Document ID"
//	@Success		200	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/knowledge/documents/{id}/reindex [post]
func (h *KnowledgeHandler) Reindex(c *gin.Context) {
	docID := c.Param("id")
	h.logger.Info().Str("documentID", docID).Msg("[KnowledgeHandler] Reindex request received")

	if err := h.kbApi.Reindex(c.Request.Context(), docID); err != nil {
		h.logger.Error().Err(err).Str("documentID", docID).Msg("[KnowledgeHandler] Reindex failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("documentID", docID).Msg("[KnowledgeHandler] Reindex succeeded")

	c.JSON(http.StatusOK, gin.H{"message": "reindexing started"})
}

// Search godoc
//
//	@Summary		Search knowledge base
//	@Description	Semantic search across the knowledge base
//	@Tags			Knowledge
//	@Accept			json
//	@Produce		json
//	@Param			request	body	requests.SearchKnowledgeRequest	true	"Search query"
//	@Success		200	{object}	responses.KnowledgeSearchResponse
//	@Security		BearerAuth
//	@Router			/knowledge/search [post]
func (h *KnowledgeHandler) Search(c *gin.Context) {
	h.logger.Info().Msg("[KnowledgeHandler] Search request received")

	var req requests.SearchKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("[KnowledgeHandler] Search failed to bind request body")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	topK := req.TopK
	if topK == 0 {
		topK = 5
	}

	h.logger.Info().Str("query", req.Query).Int("topK", topK).Msg("[KnowledgeHandler] Search processing query")

	answer, chunks, provider, tokens, err := h.ragApi.QueryKnowledgeBase(c.Request.Context(), req.Query, topK)
	if err != nil {
		h.logger.Error().Err(err).Str("query", req.Query).Msg("[KnowledgeHandler] Search failed")
		HandleError(c, err)
		return
	}

	var sources []responses.SearchResultResponse
	for _, chunk := range chunks {
		sources = append(sources, responses.SearchResultResponse{
			ChunkID:    chunk.ID,
			DocumentID: chunk.DocumentID,
			Content:    chunk.Content,
			PageNumber: chunk.PageNumber,
		})
	}

	h.logger.Info().
		Int("sourceCount", len(sources)).
		Str("provider", provider).
		Int("tokens", tokens).
		Msg("[KnowledgeHandler] Search succeeded")

	c.JSON(http.StatusOK, responses.KnowledgeSearchResponse{
		Answer: answer, Sources: sources, Provider: provider, Tokens: tokens,
	})
}

func toDocResponse(d model.KnowledgeDocument) responses.KnowledgeDocumentResponse {
	return responses.KnowledgeDocumentResponse{
		ID: d.ID, Filename: d.Filename, MimeType: d.MimeType,
		ChunkCount: d.ChunkCount, Status: d.Status, IndexedAt: d.IndexedAt,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}
