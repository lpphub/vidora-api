package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"vidora-api/shared/errs"
)

type MemoryStorage struct {
	baseURL string
	objects map[string][]byte
	mu      sync.RWMutex
}

func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		baseURL: baseURL,
		objects: make(map[string][]byte),
	}
}

func (m *MemoryStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	m.objects[objectName] = data
	m.mu.Unlock()

	return m.GetURL(objectName), nil
}

func (m *MemoryStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	m.mu.RLock()
	data, exists := m.objects[objectName]
	m.mu.RUnlock()

	if !exists {
		return nil, errs.ErrObjectNotFound
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MemoryStorage) Delete(ctx context.Context, objectName string) error {
	m.mu.Lock()
	delete(m.objects, objectName)
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) Exists(ctx context.Context, objectName string) (bool, error) {
	m.mu.RLock()
	_, exists := m.objects[objectName]
	m.mu.RUnlock()
	return exists, nil
}

func (m *MemoryStorage) GetURL(objectName string) string {
	return fmt.Sprintf("%s/%s", m.baseURL, objectName)
}

func (m *MemoryStorage) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	return m.GetURL(objectName), nil
}

func (m *MemoryStorage) Get(objectName string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, exists := m.objects[objectName]
	return data, exists
}

func (m *MemoryStorage) ComposeObject(ctx context.Context, sources []string, dest string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var combined []byte
	for _, src := range sources {
		data, exists := m.objects[src]
		if !exists {
			return errs.ErrObjectNotFound
		}
		combined = append(combined, data...)
	}

	m.objects[dest] = combined
	return nil
}
