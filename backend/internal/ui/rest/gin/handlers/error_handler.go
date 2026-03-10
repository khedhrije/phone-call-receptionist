// Package handlers provides HTTP request handlers for the REST API.
package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "phone-call-receptionist/backend/internal/domain/errors"
	"phone-call-receptionist/backend/pkg/dtos/responses"
)

// HandleError maps domain errors to appropriate HTTP status codes and responses.
func HandleError(c *gin.Context, err error) {
	var notFound *domainErrors.NotFoundError
	var validation *domainErrors.ValidationError
	var forbidden *domainErrors.ForbiddenError
	var conflict *domainErrors.ConflictError
	var unavailable *domainErrors.ServiceUnavailableError

	switch {
	case errors.As(err, &notFound):
		c.JSON(http.StatusNotFound, responses.ErrorResponse{Error: err.Error()})
	case errors.As(err, &validation):
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
	case errors.As(err, &forbidden):
		c.JSON(http.StatusForbidden, responses.ErrorResponse{Error: err.Error()})
	case errors.As(err, &conflict):
		c.JSON(http.StatusConflict, responses.ErrorResponse{Error: err.Error()})
	case errors.As(err, &unavailable):
		c.JSON(http.StatusServiceUnavailable, responses.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{Error: "internal server error"})
	}
}
