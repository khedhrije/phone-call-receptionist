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
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{Error: "failed to read file"})
		return
	}

	doc, err := h.kbApi.Upload(c.Request.Context(), header.Filename, header.Header.Get("Content-Type"), data)
	if err != nil {
		HandleError(c, err)
		return
	}

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
	docs, err := h.kbApi.List(c.Request.Context())
	if err != nil {
		HandleError(c, err)
		return
	}

	var items []responses.KnowledgeDocumentResponse
	for _, d := range docs {
		items = append(items, toDocResponse(d))
	}
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
	doc, err := h.kbApi.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleError(c, err)
		return
	}
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
	if err := h.kbApi.Delete(c.Request.Context(), c.Param("id")); err != nil {
		HandleError(c, err)
		return
	}
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
	if err := h.kbApi.Reindex(c.Request.Context(), c.Param("id")); err != nil {
		HandleError(c, err)
		return
	}
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
	var req requests.SearchKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	topK := req.TopK
	if topK == 0 {
		topK = 5
	}

	answer, chunks, provider, tokens, err := h.ragApi.QueryKnowledgeBase(c.Request.Context(), req.Query, topK)
	if err != nil {
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
