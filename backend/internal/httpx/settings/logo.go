package settings

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/service/logo"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

const maxLogoUploadSize = 5 << 20 // 5 MiB

func GetLogo(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "get logo missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load logo")
			return
		}

		fileMeta, reader, err := a.Logos.OpenCurrent(r.Context(), accountID)
		switch {
		case errors.Is(err, logo.ErrLogoNotFound):
			res.NotFound(w, "Logo not found")
			return
		case err != nil:
			slog.ErrorContext(r.Context(), "open current logo failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load logo")
			return
		}
		defer reader.Close()

		info, err := reader.Stat()
		if err != nil {
			slog.ErrorContext(r.Context(), "stat current logo failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load logo")
			return
		}

		w.Header().Set("Content-Type", fileMeta.ContentType)
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		http.ServeContent(w, r, filepath.Base(fileMeta.StorageKey), info.ModTime(), reader)
	}
}

func PutLogo(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userscope.Role(r.Context()) != "owner" {
			res.Error(w, http.StatusForbidden, "SETTINGS_OWNER_ONLY", "Only the workspace admin can edit settings")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxLogoUploadSize)

		if err := r.ParseMultipartForm(maxLogoUploadSize); err != nil {
			slog.ErrorContext(r.Context(), "parse multipart form failed", "err", err)
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Invalid multipart form data")
			return
		}

		file, header, err := r.FormFile("user_logo")
		if err != nil {
			slog.ErrorContext(r.Context(), "missing uploaded file", "field", "user_logo", "err", err)
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Uploaded file is required")
			return
		}
		defer file.Close()

		if header.Size <= 0 {
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Uploaded file is empty")
			return
		}
		if header.Size > maxLogoUploadSize {
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "File too large")
			return
		}

		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			slog.ErrorContext(r.Context(), "read uploaded file header failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to process upload")
			return
		}

		contentType := http.DetectContentType(buf[:n])
		allowedTypes := map[string]string{
			"image/png":  ".png",
			"image/jpeg": ".jpg",
			"image/webp": ".webp",
		}

		ext, ok := allowedTypes[contentType]
		if !ok {
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Unsupported file type. Use PNG, JPG, or WebP.")
			return
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			slog.ErrorContext(r.Context(), "rewind uploaded file failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to process upload")
			return
		}

		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "replace logo missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to save logo")
			return
		}

		settings, err := a.Logos.Replace(r.Context(), accountID, file, ext, contentType)
		if err != nil {
			slog.ErrorContext(r.Context(), "replace logo failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to save logo")
			return
		}

		res.JSON(w, http.StatusOK, settings)
	}
}

func DeleteLogo(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userscope.Role(r.Context()) != "owner" {
			res.Error(w, http.StatusForbidden, "SETTINGS_OWNER_ONLY", "Only the workspace admin can edit settings")
			return
		}

		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "delete logo missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to delete logo")
			return
		}
		settings, err := a.Logos.Remove(r.Context(), accountID)
		if err != nil {
			slog.ErrorContext(r.Context(), "delete logo failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to delete logo")
			return
		}

		res.JSON(w, http.StatusOK, settings)
	}
}
