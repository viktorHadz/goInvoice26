package storage

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const DefaultRootDir = "./uploads"

type LocalStore struct {
	rootDir string
}

func NewLocalStore(rootDir string) *LocalStore {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		rootDir = DefaultRootDir
	}

	return &LocalStore{rootDir: filepath.Clean(rootDir)}
}

func (s *LocalStore) RootDir() string {
	return s.rootDir
}

func (s *LocalStore) WriteTemp(r io.Reader, ext string) (string, error) {
	if ext == "" || ext[0] != '.' {
		return "", fmt.Errorf("write temp: invalid extension %q", ext)
	}

	dir := filepath.Join(s.rootDir, "tmp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("write temp: create tmp dir: %w", err)
	}

	tempPath := filepath.Join(dir, uuid.NewString()+ext+".tmp")
	f, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("write temp: create temp file: %w", err)
	}

	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("write temp: copy file data: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("write temp: close temp file: %w", err)
	}

	return tempPath, nil
}

func (s *LocalStore) NewStorageKey(accountID int64, kind, ext string) string {
	dir := kindDirectory(kind)
	filename := uuid.NewString() + ext

	return path.Join("accounts", strconv.FormatInt(accountID, 10), dir, filename)
}

func (s *LocalStore) PromoteTemp(tempPath, storageKey string) error {
	destPath := s.Path(storageKey)
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("promote temp: create destination dir: %w", err)
	}
	if err := os.Rename(tempPath, destPath); err != nil {
		return fmt.Errorf("promote temp: move file into place: %w", err)
	}
	return nil
}

func (s *LocalStore) Delete(storageKey string) error {
	if storageKey == "" {
		return nil
	}

	err := os.Remove(s.Path(storageKey))
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("delete file: %w", err)
}

func (s *LocalStore) Open(storageKey string) (*os.File, error) {
	f, err := os.Open(s.Path(storageKey))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

func (s *LocalStore) Path(storageKey string) string {
	cleanKey := strings.Trim(strings.TrimSpace(storageKey), "/")
	if cleanKey == "" {
		return s.rootDir
	}
	return filepath.Join(s.rootDir, filepath.FromSlash(cleanKey))
}

func (s *LocalStore) CleanupTemp() error {
	tmpDir := filepath.Join(s.rootDir, "tmp")
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("cleanup temp: %w", err)
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("cleanup temp recreate dir: %w", err)
	}
	return nil
}

func kindDirectory(kind string) string {
	switch strings.TrimSpace(kind) {
	case "logo":
		return "logos"
	default:
		return "files"
	}
}
