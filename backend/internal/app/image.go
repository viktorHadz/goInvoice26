package app

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

const maxLogoUploadSize = 5 << 20 // 5 MiB

func LogoUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxLogoUploadSize)

		if err := r.ParseMultipartForm(maxLogoUploadSize); err != nil {
			slog.ErrorContext(r.Context(),
				"parse multipart form failed",
				"err", err,
			)
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Invalid multipart form data")
			return
		}

		file, header, err := r.FormFile("user_logo")
		if err != nil {
			slog.ErrorContext(r.Context(),
				"missing uploaded file",
				"field", "user_logo",
				"err", err,
			)
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
		if err != nil && err != io.EOF {
			slog.ErrorContext(r.Context(),
				"read uploaded file header failed",
				"err", err,
			)
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
			slog.ErrorContext(r.Context(),
				"rewind uploaded file failed",
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to process upload")
			return
		}

		uploadDir := "./uploads"
		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			slog.ErrorContext(r.Context(),
				"create upload directory failed",
				"dir", uploadDir,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
			return
		}

		filename := uuid.NewString() + ext
		destPath := filepath.Join(uploadDir, filename)

		destFile, err := os.Create(destPath)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"create destination file failed",
				"path", destPath,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
			return
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, file); err != nil {
			slog.ErrorContext(r.Context(),
				"save uploaded file failed",
				"path", destPath,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
			return
		}

		logoURL := "/uploads/" + filename

		slog.DebugContext(r.Context(),
			"logo uploaded successfully",
			"file", filename,
			"content_type", contentType,
			"size", header.Size,
		)

		res.JSON(w, http.StatusOK, map[string]string{
			"logoUrl": logoURL,
		})
	}
}
