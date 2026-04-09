package blobfs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type Store struct {
	root string
}

func NewStore(root string) Store {
	return Store{root: root}
}

func (s Store) Put(_ context.Context, objectKey string, data []byte) error {
	path, err := s.objectPath(objectKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s Store) Get(_ context.Context, objectKey string) ([]byte, error) {
	path, err := s.objectPath(objectKey)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

func (s Store) objectPath(objectKey string) (string, error) {
	cleaned := filepath.Clean(strings.TrimSpace(objectKey))
	if cleaned == "." || cleaned == "" {
		return "", errors.New("object key is required")
	}
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", errors.New("object key must stay within blob store root")
	}
	return filepath.Join(s.root, filepath.FromSlash(cleaned)), nil
}
