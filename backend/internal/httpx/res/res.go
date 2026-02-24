package res

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// FieldError represents a single validation problem for one specific input field.
//
// Usage: build a FieldError slice during validation and return via Validation(...)
type FieldError struct {
	Field   string         `json:"field"`
	Code    string         `json:"code"`
	Message string         `json:"message,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

/*
It is both:
- a Go error (implements error interface via Error())
- a structured payload to send client

Usage:
- Whenever returning a non-2xx response in a consistent JSON shape

Notes:
- ID - for users to report errors to then find in logs - sent to client .
- Code - error code - sent to the client.
- Message - for displaying errors on the client.
- Fields - used in multiple field validation - sent to client.
- Status is HTTP transport detail - not sent in JSON.
- Cause is an internal error for logging - not sent in JSON.
*/
type APIError struct {
	ID      string       `json:"id,omitempty"`     // Generated if missing
	Code    string       `json:"code"`             // Stable code for frontend logic
	Message string       `json:"message"`          // Safe summary to show user
	Fields  []FieldError `json:"fields,omitempty"` // Used only with validation

	Status int   `json:"-"` // HTTP status code
	Cause  error `json:"-"` // Underlying error - for logs only never returned to client
}

// Error makes APIError implement the built-in error interface.
// In logs, this will show the error Code only unless logged via Message separately.
func (e *APIError) Error() string { return e.Code }

// envelope is the JSON shape the API returns for errors.
// Clients can always expect: { "error": { ... } }
type envelope struct {
	Error *APIError `json:"error"`
}

/*
Writes a successful JSON response.

Usage:
- request succeeded (2xx / 3xx)
- need to return a JSON body

Examples:

- res.JSON(w, http.StatusOK, client)

- res.JSON(w, http.StatusCreated, map[string]any{"id": id})
*/
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

/*
NoContent writes a 204 No Content response.

Usage:
- operation succeeded but there is no body to return (common for DELETE/PATCH)

Examples:
- PATCH updated successfully, nothing else to return
- DELETE successful
*/
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

/*
Error writes an error response in the standard JSON envelope and returns the error ID.

Usage:
- exiting a handler early with an error response

Important behavior:
- If `err` is an *APIError, it writes it.
- If `err` is a normal error, it is converted into an INTERNAL 500 error.
- It fills missing defaults (ID/status/code/message)

Return value:
- the correlation id used for the response - for logging/support tickets

Example:

- res.Error(w, res.NotFound("client not found"))

- res.Error(w, res.Validation(errs...))

- res.Error(w, res.Database(err))
*/
func Error(w http.ResponseWriter, err error) string {
	ae := AsAPIError(err)

	if ae.ID == "" {
		ae.ID = newID()
	}
	if ae.Status == 0 {
		ae.Status = http.StatusInternalServerError
	}
	if ae.Code == "" {
		ae.Code = "INTERNAL"
	}
	if ae.Message == "" {
		ae.Message = "Internal server error"
	}

	JSON(w, ae.Status, envelope{Error: ae})
	return ae.ID
}

/*
AsAPIError converts any error into an *APIError.

Usage:
- to normalize error into the API error type
- to inspect status/code before writing/logging

Behavior:
- If err already wraps/is an *APIError, it returns a safe to modify COPY of it.
- Otherwise it wraps it as Internal(err).

Example:
- ae := res.AsAPIError(err)
- if ae.Status >= 500 { slog.ErrorContext(ctx, "server error", "cause", ae.Cause) }
*/
func AsAPIError(err error) *APIError {
	var ae *APIError
	if errors.As(err, &ae) {
		// copy to avoid mutating shared pointers
		out := *ae
		return &out
	}
	return Internal(err)
}

// ---- constructors

/*
BadJSON creates a 400 error for invalid JSON payloads.

Usage:
- json decoding fails
- request body is not valid JSON
- request contains unknown fields (if DisallowUnknownFields() is enabled)

Example:
- dec := json.NewDecoder(r.Body) dec.DisallowUnknownFields()
- if err := dec.Decode(&dst); err != nil { res.Error(w, res.BadJSON()); return }
*/
func BadJSON() *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "BAD_JSON",
		Message: "Invalid JSON payload",
	}
}

/*
Checks JSON received via frontend (FE) is a single JSON object

Throws error if receiving multiple JSON objects
USAGE:

	var client models.CreateClient
	if ok := res.DecodeJSON(w, r, &client); !ok {
	  return
	}
*/
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		Error(w, BadJSON())
		return false
	}

	// Must be exactly one JSON value
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		Error(w, BadJSON())
		return false
	}

	return true
}

/*
Validation creates a 400 error for field validation problems.

Usage:
- the request was syntactically valid JSON
- but values fail the rules (required, max length, invalid email, etc.)

Example:
- if len(errs) > 0 { res.Error(w, res.Validation(errs...)); return }
*/
func Validation(fields ...FieldError) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "VALIDATION_FAILED",
		Message: "Validation failed",
		Fields:  fields,
	}
}

/*
NotFound creates a 404 error.

Usage:
- the requested resource doesn't exist
- an update/delete affected 0 rows for an id that should exist

Example:
- if affected == 0 { res.Error(w, res.NotFound("client not found")); return }
*/
func NotFound(msg string) *APIError {
	if msg == "" {
		msg = "Not found"
	}
	return &APIError{
		Status:  http.StatusNotFound,
		Code:    "NOT_FOUND",
		Message: msg,
	}
}

/*
Database creates a 500 error representing a DB failure.

Usage:
- the DB call returned an error - query failed, connection issue...
- Hides details from the client but keeps the cause for logs

Important:
- This is for unexpected DB errors (500).

Example:
- if err != nil { slog.ErrorContext(ctx, "db failed", "err", err); res.Error(w, res.Database(err)); return }
*/
func Database(err error) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    "DATABASE_ERROR",
		Message: "Database error",
		Cause:   err,
	}
}

/*
Internal creates a 500 INTERNAL error.

Usage
- something unexpected happened
- there is a raw error and it need a safe response for client

Example:
- return res.Internal(err)
*/
func Internal(err error) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL",
		Message: "Internal server error",
		Cause:   err,
	}
}

// ---- field error helpers

// Invalid builds a validation FieldError for "this value is invalid".
// Examples:
//
// - res.Invalid("email", "email format is invalid")
//
// - res.Invalid("id", "invalid route param")
func Invalid(field, msg string) FieldError {
	if msg == "" {
		msg = "invalid"
	}
	return FieldError{Field: field, Code: "INVALID", Message: msg}
}

func Required(field string) FieldError {
	return FieldError{Field: field, Code: "REQUIRED", Message: "is required"}
}

func MaxLen(field string, max int) FieldError {
	return FieldError{
		Field:   field,
		Code:    "MAX_LENGTH",
		Message: "too long",
		Meta:    map[string]any{"max": max},
	}
}
func MinLen(field string, min int) FieldError {
	return FieldError{
		Field:   field,
		Code:    "MIN_LENGTH",
		Message: "too short",
		Meta:    map[string]any{"min": min},
	}
}
func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
