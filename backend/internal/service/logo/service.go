package logo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/service/storage"
	"github.com/viktorHadz/goInvoice26/internal/transaction/settingsTx"
)

var ErrLogoNotFound = errors.New("logo not found")

type Service struct {
	db    *sql.DB
	store *storage.LocalStore
}

func NewService(db *sql.DB, store *storage.LocalStore) *Service {
	return &Service{
		db:    db,
		store: store,
	}
}

func (s *Service) Replace(ctx context.Context, accountID int64, src io.Reader, ext, contentType string) (models.Settings, error) {
	tempPath, err := s.store.WriteTemp(src, ext)
	if err != nil {
		return models.Settings{}, fmt.Errorf("stage logo upload: %w", err)
	}

	storageKey := s.store.NewStorageKey(accountID, settingsTx.StoredFileKindLogo, ext)
	if err := s.store.PromoteTemp(tempPath, storageKey); err != nil {
		_ = os.Remove(tempPath)
		return models.Settings{}, fmt.Errorf("promote staged logo: %w", err)
	}

	_, prev, err := settingsTx.ReplaceLogo(ctx, s.db, accountID, storageKey, contentType)
	if err != nil {
		_ = s.store.Delete(storageKey)
		return models.Settings{}, err
	}

	if err := s.cleanupDetachedAsset(ctx, prev); err != nil {
		return models.Settings{}, err
	}

	settings, err := settingsTx.Get(ctx, s.db, accountID)
	if err != nil {
		return models.Settings{}, fmt.Errorf("reload settings after logo replace: %w", err)
	}
	return settings, nil
}

func (s *Service) Remove(ctx context.Context, accountID int64) (models.Settings, error) {
	prev, err := settingsTx.RemoveLogo(ctx, s.db, accountID)
	if err != nil {
		return models.Settings{}, err
	}

	if err := s.cleanupDetachedAsset(ctx, prev); err != nil {
		return models.Settings{}, err
	}

	settings, err := settingsTx.Get(ctx, s.db, accountID)
	if err != nil {
		return models.Settings{}, fmt.Errorf("reload settings after logo delete: %w", err)
	}
	return settings, nil
}

func (s *Service) OpenCurrent(ctx context.Context, accountID int64) (settingsTx.StoredFile, *os.File, error) {
	file, ok, err := settingsTx.GetLogoFile(ctx, s.db, accountID)
	if err != nil {
		return settingsTx.StoredFile{}, nil, err
	}
	if !ok {
		return settingsTx.StoredFile{}, nil, ErrLogoNotFound
	}

	reader, err := s.store.Open(file.StorageKey)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return settingsTx.StoredFile{}, nil, ErrLogoNotFound
		}
		return settingsTx.StoredFile{}, nil, fmt.Errorf("open current logo: %w", err)
	}

	return file, reader, nil
}

func (s *Service) SweepPendingDeletes(ctx context.Context) error {
	files, err := settingsTx.ListDeletePendingFiles(ctx, s.db)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := s.store.Delete(file.StorageKey); err != nil {
			continue
		}
		if err := settingsTx.DeleteStoredFile(ctx, s.db, file.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) CleanupTemp() error {
	return s.store.CleanupTemp()
}

func (s *Service) MigrateLegacyLogo(ctx context.Context, accountID int64) error {
	_, ok, err := settingsTx.GetLogoFile(ctx, s.db, accountID)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	legacyValue, err := settingsTx.GetLegacyLogoURL(ctx, s.db, accountID)
	if err != nil {
		return err
	}

	legacyPath, ext, contentType, ok := s.resolveLegacyLogoPath(legacyValue)
	if !ok {
		return nil
	}

	storageKey := s.store.NewStorageKey(accountID, settingsTx.StoredFileKindLogo, ext)
	if err := s.store.PromoteTemp(legacyPath, storageKey); err != nil {
		return fmt.Errorf("move legacy logo into account storage: %w", err)
	}

	if _, _, err := settingsTx.ReplaceLogo(ctx, s.db, accountID, storageKey, contentType); err != nil {
		_ = s.store.Delete(storageKey)
		return fmt.Errorf("assign migrated legacy logo: %w", err)
	}

	if err := settingsTx.ClearLegacyLogoURL(ctx, s.db, accountID); err != nil {
		return err
	}

	return nil
}

func (s *Service) cleanupDetachedAsset(ctx context.Context, file *settingsTx.StoredFile) error {
	if file == nil {
		return nil
	}

	if err := s.store.Delete(file.StorageKey); err != nil {
		if markErr := settingsTx.MarkStoredFileDeletePending(ctx, s.db, file.ID); markErr != nil {
			return fmt.Errorf("delete old asset: %w (mark pending failed: %v)", err, markErr)
		}
		return nil
	}

	if err := settingsTx.DeleteStoredFile(ctx, s.db, file.ID); err != nil {
		return err
	}

	return nil
}

func (s *Service) resolveLegacyLogoPath(v string) (string, string, string, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", "", "", false
	}

	candidates := []string{}
	if filepath.IsAbs(v) {
		candidates = append(candidates, v)
	}

	trimmed := strings.TrimPrefix(v, "/")
	if strings.HasPrefix(trimmed, "uploads/") {
		trimmed = strings.TrimPrefix(trimmed, "uploads/")
		candidates = append(candidates, s.store.Path(trimmed))
	}
	if !filepath.IsAbs(v) {
		candidates = append(candidates, filepath.Clean(v))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}

		ext := strings.ToLower(filepath.Ext(candidate))
		contentType, ok := contentTypeForExtension(ext)
		if !ok {
			return "", "", "", false
		}
		return candidate, ext, contentType, true
	}

	return "", "", "", false
}

func contentTypeForExtension(ext string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".png":
		return "image/png", true
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".webp":
		return "image/webp", true
	default:
		return "", false
	}
}
