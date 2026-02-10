// Package response provides a single, consistent way to write API errors as responses
//
//	Usage:
//		response.Error(w, response.Validation(
//		    response.Required("name"),
//		    response.MaxLen("email", 320),
//		))
//
//		response.Error(w, response.NotFound("client"))
//		response.Error(w, response.Unauthorized())
//		response.Error(w, response.Forbidden())
//		response.Error(w, response.Conflict("email already exists"))
//		response.Error(w, err) // unknown err -> INTERNAL_ERROR
//
//	Success helpers:
//		response.JSON(w, http.StatusOK, payload)
//		response.NoContent(w)
//
//
//
//	Validation Pattern:
//		errs := []res.FieldError{}
//
//		if client.Name == "" {
//			errs = append(errs, res.Required("name"))
//		} else if len(client.Name) > 50 {
//			errs = append(errs, res.MaxLen("name", 50))
//		}
//
//		if len(errs) > 0 {
//			res.Error(w, res.Validation(errs...))
//			return
//		}
//
//	Error response JSON shape:
//
//		{
//		  "error": {
//		    "id": "a1b2c3d4e5f6a7b8",
//		    "code": "VALIDATION_FAILED",
//		    "message": "Validation failed",
//		    "fields": [
//		      { "field": "name", "code": "REQUIRED", "message": "is required" },
//		      { "field": "email", "code": "MAX_LENGTH", "message": "too long", "meta": { "max": 320 } }
//		    ]
//		  }
//		}
package res

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
)

// --------------------------
// Success writers (optional)
// --------------------------

// JSON writes a JSON success payload.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// NoContent writes a 204 with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// --------------------------
// Single error writer
// --------------------------

// Error is the ONLY error writer that should be used in the app
//
// Pass:
//
// - one of the builders (Validation, NotFound,..)
//
// - a random error (it will be treated as a safe default -  INTERNAL_ERROR)
//
//	func Error(w http.ResponseWriter, err error) {
//		var api *APIError
//		if errors.As(err, &api) {
//			writeAPIError(w, api)
//			return
//		}
//		// Unknown/untyped errors are INTERNAL_ERROR (safe default)
//		writeAPIError(w, Internal())
//	}
func Error(w http.ResponseWriter, err error) {
	var api *APIError
	if errors.As(err, &api) {
		writeAPIError(w, api)
		return
	}
	// Unknown/untyped errors are INTERNAL_ERROR (safe default)
	writeAPIError(w, Internal())
}

func writeAPIError(w http.ResponseWriter, e *APIError) {
	if e.ID == "" {
		e.ID = newErrorID()
	}
	if e.Status == 0 {
		e.Status = http.StatusInternalServerError
	}
	if e.Code == "" {
		e.Code = CodeInternalError
	}
	if e.Message == "" {
		e.Message = defaultMessage(e.Code)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{Error: *e})
}

// --------------------------
// Error response model
// --------------------------

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

type APIError struct {
	// Correlation id for logs and potential tickets support tickets.
	ID string `json:"id,omitempty"`

	// HTTP status is NOT included in JSON (transport detail).
	Status int `json:"-"`

	// Stable error code
	Code string `json:"code"`

	// Human readable summary (use for toasts, frontend, status messages etc).
	Message string `json:"message,omitempty"`

	// Per-field validation errors (if any).
	Fields []FieldError `json:"fields,omitempty"`
}

func (e *APIError) Error() string { return e.Code }

type FieldError struct {
	Field   string         `json:"field"`
	Code    string         `json:"code"`
	Message string         `json:"message,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// --------------------------
// Grouped error codes (top-level)
// --------------------------

// Validation (400)
const (
	CodeValidationFailed = "VALIDATION_FAILED"
)

// Auth (401)
const (
	CodeAuthRequired     = "AUTH_REQUIRED"
	CodeAuthMissingToken = "AUTH_MISSING_TOKEN"
	CodeAuthInvalidToken = "AUTH_INVALID_TOKEN"
	CodeAuthExpiredToken = "AUTH_EXPIRED_TOKEN"
)

// Permission (403)
const (
	CodePermissionDenied = "PERMISSION_DENIED"
)

// Not found (404)
const (
	CodeNotFound = "NOT_FOUND"
)

// Conflict (409)
const (
	CodeConflict = "CONFLICT"
)

// Rate limit (429)
const (
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

// Server (500)
const (
	CodeInternalError = "INTERNAL_ERROR"
	CodeDatabaseError = "DATABASE_ERROR"
)

// --------------------------
// Validation field codes
// --------------------------

const (
	FieldRequired     = "REQUIRED"
	FieldMaxLength    = "MAX_LENGTH"
	FieldMinLength    = "MIN_LENGTH"
	FieldInvalid      = "INVALID"
	FieldInvalidFmt   = "INVALID_FORMAT"
	FieldInvalidEmail = "INVALID_EMAIL"
	FieldOutOfRange   = "OUT_OF_RANGE"
)

// --------------------------
// Error builders
// --------------------------

// Error builder:
//
// Returns a 400 VALIDATION_FAILED with a field error array.
func Validation(fields ...FieldError) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    CodeValidationFailed,
		Message: "Validation failed",
		Fields:  fields,
	}
}

// Error builder:
//
// Returns a 404 NOT_FOUND.
func NotFound(resource string) *APIError {
	msg := "Not found"
	if resource != "" {
		msg = resource + " not found"
	}
	return &APIError{
		Status:  http.StatusNotFound,
		Code:    CodeNotFound,
		Message: msg,
	}
}

// Error builder:
//
// Returns a 409 CONFLICT.
func Conflict(msg string) *APIError {
	if msg == "" {
		msg = "Conflict"
	}
	return &APIError{
		Status:  http.StatusConflict,
		Code:    CodeConflict,
		Message: msg,
	}
}

// Error builder:
//
// Returns a 401 AUTH_REQUIRED.
func Unauthorized() *APIError {
	return &APIError{
		Status:  http.StatusUnauthorized,
		Code:    CodeAuthRequired,
		Message: "Authentication required",
	}
}

// Error builder:
//
// Returns a 401 AUTH_MISSING_TOKEN.
func MissingToken() *APIError {
	return &APIError{
		Status:  http.StatusUnauthorized,
		Code:    CodeAuthMissingToken,
		Message: "Missing authentication token",
	}
}

// Error builder:
//
// Returns a 401 AUTH_INVALID_TOKEN.
func InvalidToken() *APIError {
	return &APIError{
		Status:  http.StatusUnauthorized,
		Code:    CodeAuthInvalidToken,
		Message: "Invalid authentication token",
	}
}

// Error builder:
//
// Returns a 401 AUTH_EXPIRED_TOKEN.
func ExpiredToken() *APIError {
	return &APIError{
		Status:  http.StatusUnauthorized,
		Code:    CodeAuthExpiredToken,
		Message: "Authentication token expired",
	}
}

// Error builder:
//
// Returns a 403 PERMISSION_DENIED.
func Forbidden() *APIError {
	return &APIError{
		Status:  http.StatusForbidden,
		Code:    CodePermissionDenied,
		Message: "Permission denied",
	}
}

// Error builder:
//
// Returns a 429 RATE_LIMIT_EXCEEDED.
func RateLimitExceeded() *APIError {
	return &APIError{
		Status:  http.StatusTooManyRequests,
		Code:    CodeRateLimitExceeded,
		Message: "Rate limit exceeded",
	}
}

// Error builder:
//
// Returns a 500 DATABASE_ERROR use when it's DB-related
//
//   - Client receives http.StatusInternalServerError
func Database() *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    CodeDatabaseError,
		Message: "Database error",
	}
}

// Error builder:
//
// Returns a 500 INTERNAL_ERROR (safe default).
func Internal() *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternalError,
		Message: "Internal server error",
	}
}

// --------------------------
// FieldError builders (use inside Validation())
// --------------------------

// Validation FieldError Builder: field is required
//
// Example:
//
//	res.Error(res.Validation(res.Required("name")))
//
// Error => "name is required"
func Required(field string) FieldError {
	return FieldError{
		Field:   field,
		Code:    FieldRequired,
		Message: "is required",
	}
}

// Validation FieldError Builder: string too long
//
// Example:
//
//	res.Error(res.Validation(res.MaxLen("email", 50)))
//
// Error => "email too long max length 50"
func MaxLen(field string, max int) FieldError {
	return FieldError{
		Field:   field,
		Code:    FieldMaxLength,
		Message: "too long",
		Meta: map[string]any{
			"max": max,
		},
	}
}

// Validation FieldError Builder: string too short
//
// Example:
//
//	res.Error(res.Validation(res.MinLen("name", 50)))
//
// Error => "name too short max length 2"
func MinLen(field string, min int) FieldError {
	return FieldError{
		Field:   field,
		Code:    FieldMinLength,
		Message: "too short",
		Meta: map[string]any{
			"min": min,
		},
	}
}

// Validation FieldError Builder: invalid value with custom message
//
// Example:
//
//	res.Error(res.Validation(res.Invalid("id", "id is invalid")))
//
// Error => "invalid id"
func Invalid(field string, msg string) FieldError {
	if msg == "" {
		msg = "invalid"
	}
	return FieldError{
		Field:   field,
		Code:    FieldInvalid,
		Message: msg,
	}
}

// Validation FieldError Builder: numeric range issue
//
// Example:
//
//	res.Error(res.Validation(res.OutOfRange("discount percentage", 0, 100)))
//
// Error => "discount percentage out of range min 0 max 100"
func OutOfRange(field string, min, max any) FieldError {
	return FieldError{
		Field:   field,
		Code:    FieldOutOfRange,
		Message: "out of range",
		Meta: map[string]any{
			"min": min,
			"max": max,
		},
	}
}

func newErrorID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func defaultMessage(code string) string {
	switch code {
	case CodeValidationFailed:
		return "Validation failed"
	case CodeAuthRequired:
		return "Authentication required"
	case CodeAuthMissingToken:
		return "Missing authentication token"
	case CodeAuthInvalidToken:
		return "Invalid authentication token"
	case CodeAuthExpiredToken:
		return "Authentication token expired"
	case CodePermissionDenied:
		return "Permission denied"
	case CodeNotFound:
		return "Not found"
	case CodeConflict:
		return "Conflict"
	case CodeRateLimitExceeded:
		return "Rate limit exceeded"
	case CodeDatabaseError:
		return "Database error"
	case CodeInternalError:
		return "Internal server error"
	default:
		return "Error"
	}
}
