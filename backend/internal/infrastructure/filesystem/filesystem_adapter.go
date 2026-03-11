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

	logger.Info().Str("basePath", basePath).Msg("[FilesystemAdapter] storage initialized")

	return &Adapter{
		basePath: basePath,
		logger:   logger,
	}, nil
}

// Create persists file data to storage and returns the full storage path.
func (a *Adapter) Create(_ context.Context, filename string, data []byte) (string, error) {
	a.logger.Debug().Str("filename", filename).Int("bytes", len(data)).Msg("[FilesystemAdapter] saving file")

	fullPath := filepath.Join(a.basePath, filename)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		a.logger.Error().Err(err).Str("filename", filename).Msg("[FilesystemAdapter] failed to create directory")
		return "", fmt.Errorf("failed to create directory for %s: %w", filename, err)
	}

	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		a.logger.Error().Err(err).Str("filename", filename).Msg("[FilesystemAdapter] failed to write file")
		return "", fmt.Errorf("failed to save file %s: %w", filename, err)
	}

	a.logger.Debug().
		Str("filename", filename).
		Int("bytes", len(data)).
		Msg("[FilesystemAdapter] file saved")

	return fullPath, nil
}

// Find retrieves file data from storage by filename.
// Returns a NotFoundError if the file does not exist.
func (a *Adapter) Find(_ context.Context, filename string) ([]byte, error) {
	a.logger.Debug().Str("filename", filename).Msg("[FilesystemAdapter] loading file")

	fullPath := filepath.Join(a.basePath, filename)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			a.logger.Debug().Str("filename", filename).Msg("[FilesystemAdapter] file not found")
			return nil, errors.NewNotFound("file", filename)
		}
		a.logger.Error().Err(err).Str("filename", filename).Msg("[FilesystemAdapter] failed to read file")
		return nil, fmt.Errorf("failed to load file %s: %w", filename, err)
	}

	a.logger.Debug().
		Str("filename", filename).
		Int("bytes", len(data)).
		Msg("[FilesystemAdapter] file loaded")

	return data, nil
}

// Delete removes a file from storage by filename.
func (a *Adapter) Delete(_ context.Context, filename string) error {
	a.logger.Debug().Str("filename", filename).Msg("[FilesystemAdapter] deleting file")

	fullPath := filepath.Join(a.basePath, filename)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			a.logger.Debug().Str("filename", filename).Msg("[FilesystemAdapter] file already absent, skipping delete")
			return nil
		}
		a.logger.Error().Err(err).Str("filename", filename).Msg("[FilesystemAdapter] failed to delete file")
		return fmt.Errorf("failed to delete file %s: %w", filename, err)
	}

	a.logger.Debug().Str("filename", filename).Msg("[FilesystemAdapter] file deleted")
	return nil
}
