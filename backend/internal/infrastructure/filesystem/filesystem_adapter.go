// Package filesystem implements the FileStorage port using the local filesystem.
package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/errors"
	"phone-call-receptionist/backend/internal/domain/port"
)

// Adapter implements port.FileStorage using the local filesystem.
type Adapter struct {
	basePath string
	logger   *zerolog.Logger
}

// NewFilesystemAdapter creates a new filesystem storage adapter with the given base directory.
// It creates the base directory if it does not already exist.
func NewFilesystemAdapter(basePath string, logger *zerolog.Logger) (port.FileStorage, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory %s: %w", basePath, err)
	}

	logger.Info().Str("basePath", basePath).Msg("Filesystem storage initialized")

	return &Adapter{
		basePath: basePath,
		logger:   logger,
	}, nil
}

// Save persists file data to storage and returns the full storage path.
func (a *Adapter) Save(_ context.Context, filename string, data []byte) (string, error) {
	fullPath := filepath.Join(a.basePath, filename)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", filename, err)
	}

	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to save file %s: %w", filename, err)
	}

	a.logger.Debug().
		Str("filename", filename).
		Int("bytes", len(data)).
		Msg("File saved")

	return fullPath, nil
}

// Load retrieves file data from storage by filename.
// Returns a NotFoundError if the file does not exist.
func (a *Adapter) Load(_ context.Context, filename string) ([]byte, error) {
	fullPath := filepath.Join(a.basePath, filename)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewNotFound("file", filename)
		}
		return nil, fmt.Errorf("failed to load file %s: %w", filename, err)
	}

	a.logger.Debug().
		Str("filename", filename).
		Int("bytes", len(data)).
		Msg("File loaded")

	return data, nil
}

// Delete removes a file from storage by filename.
func (a *Adapter) Delete(_ context.Context, filename string) error {
	fullPath := filepath.Join(a.basePath, filename)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete file %s: %w", filename, err)
	}

	a.logger.Debug().Str("filename", filename).Msg("File deleted")
	return nil
}
