package res

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// FieldError represents a single field-level validation error.
type FieldError struct {
	Field   string         `json:"field"`
	Code    string         `json:"code"`
	Message string         `json:"message,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// APIError is a JSON error response payload.
type APIError struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

type envelope struct {
	Error APIError `json:"error"`
}

// JSON writes v as JSON with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode json response", "status", status, "err", err)
	}
}

// NoContent sends a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes a structured JSON error response.
func Error(w http.ResponseWriter, status int, code, message string) {
	if code == "" {
		code = "INTERNAL"
	}
	if message == "" {
		message = "Internal server error"
	}

	JSON(w, status, envelope{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	})
}

// Validation writes a 400 validation error response.
func Validation(w http.ResponseWriter, fields ...FieldError) {
	JSON(w, http.StatusBadRequest, envelope{
		Error: APIError{
			Code:    "VALIDATION_FAILED",
			Message: "Validation failed",
			Fields:  fields,
		},
	})
}

// NotFound writes a 404 NOT_FOUND response.
func NotFound(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "Not found"
	}

	Error(w, http.StatusNotFound, "NOT_FOUND", msg)
}

// BadJSON writes a 400 BAD_JSON response.
func BadJSON(w http.ResponseWriter) {
	Error(w, http.StatusBadRequest, "BAD_JSON", "Invalid JSON payload")
}

// DecodeJSON decodes the request body into dst.
// Returns false and writes the error response automatically on failure.
// Rejects unknown fields and multiple JSON values.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		slog.Warn("invalid json payload", "err", err)
		BadJSON(w)
		return false
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		slog.Warn("request body contained multiple json values", "err", fmt.Errorf("multiple JSON values: %w", err))
		BadJSON(w)
		return false
	}

	return true
}

// Invalid returns an INVALID field error.
func Invalid(field, msg string) FieldError {
	if msg == "" {
		msg = "invalid"
	}
	return FieldError{
		Field:   field,
		Code:    "INVALID",
		Message: msg,
	}
}

// Required returns a REQUIRED field error.
func Required(field string) FieldError {
	return FieldError{
		Field:   field,
		Code:    "REQUIRED",
		Message: "is required",
	}
}

// MaxLen returns a MAX_LENGTH field error.
func MaxLen(field string, max int) FieldError {
	return FieldError{
		Field:   field,
		Code:    "MAX_LENGTH",
		Message: "too long",
		Meta:    map[string]any{"max": max},
	}
}

// MinLen returns a MIN_LENGTH field error.
func MinLen(field string, min int) FieldError {
	return FieldError{
		Field:   field,
		Code:    "MIN_LENGTH",
		Message: "too short",
		Meta:    map[string]any{"min": min},
	}
}
