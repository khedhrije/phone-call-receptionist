// Package errors defines domain-specific error types used across the application.
package errors

import "fmt"

// NotFoundError indicates that a requested resource was not found.
type NotFoundError struct {
	// Resource is the type of resource that was not found.
	Resource string
	// ID is the identifier that was searched for.
	ID string
}

// Error returns the error message.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// NewNotFound creates a new NotFoundError for the given resource and ID.
func NewNotFound(resource string, id string) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

// ValidationError indicates that input data failed validation.
type ValidationError struct {
	// Field is the name of the field that failed validation.
	Field string
	// Message describes the validation failure.
	Message string
}

// Error returns the error message.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidation creates a new ValidationError for the given field and message.
func NewValidation(field string, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// ForbiddenError indicates that the caller does not have permission to perform the action.
type ForbiddenError struct {
	// Message describes why access was denied.
	Message string
}

// Error returns the error message.
func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: %s", e.Message)
}

// NewForbidden creates a new ForbiddenError with the given message.
func NewForbidden(message string) *ForbiddenError {
	return &ForbiddenError{Message: message}
}

// ConflictError indicates that the operation conflicts with existing state.
type ConflictError struct {
	// Message describes the conflict.
	Message string
}

// Error returns the error message.
func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %s", e.Message)
}

// NewConflict creates a new ConflictError with the given message.
func NewConflict(message string) *ConflictError {
	return &ConflictError{Message: message}
}

// ServiceUnavailableError indicates that an external service is not available.
type ServiceUnavailableError struct {
	// Service is the name of the unavailable service.
	Service string
	// Err is the underlying error.
	Err error
}

// Error returns the error message.
func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf("service unavailable (%s): %v", e.Service, e.Err)
}

// Unwrap returns the underlying error.
func (e *ServiceUnavailableError) Unwrap() error {
	return e.Err
}

// NewServiceUnavailable creates a new ServiceUnavailableError for the given service and error.
func NewServiceUnavailable(service string, err error) *ServiceUnavailableError {
	return &ServiceUnavailableError{Service: service, Err: err}
}

// IsNotFound checks whether the given error is a NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// IsValidation checks whether the given error is a ValidationError.
func IsValidation(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsForbidden checks whether the given error is a ForbiddenError.
func IsForbidden(err error) bool {
	_, ok := err.(*ForbiddenError)
	return ok
}

// IsConflict checks whether the given error is a ConflictError.
func IsConflict(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}

// IsServiceUnavailable checks whether the given error is a ServiceUnavailableError.
func IsServiceUnavailable(err error) bool {
	_, ok := err.(*ServiceUnavailableError)
	return ok
}
