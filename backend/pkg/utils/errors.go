package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type AppError struct {
	Status  int
	Message string
	Fields  any
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func ErrBadRequest(message string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Message: message}
}

func ErrUnauthorized(message string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Message: message}
}

func ErrForbidden(message string) *AppError {
	return &AppError{Status: http.StatusForbidden, Message: message}
}

func ErrNotFound(message string) *AppError {
	return &AppError{Status: http.StatusNotFound, Message: message}
}

func ErrConflict(message string) *AppError {
	return &AppError{Status: http.StatusConflict, Message: message}
}

func ErrUnprocessable(message string, fields any) *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Message: message, Fields: fields}
}

func ErrInternal(err error) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Message: "internal server error", Err: err}
}

func MapError(err error) (int, JSONResponse) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		body := JSONResponse{Message: appErr.Message}
		if appErr.Fields != nil {
			body.Errors = appErr.Fields
		}
		return appErr.Status, body
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return http.StatusUnprocessableEntity, JSONResponse{
			Message: "validation failed",
			Errors:  toFieldErrors(validationErrs),
		}
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) {
		return http.StatusBadRequest, JSONResponse{Message: "invalid request body"}
	}

	if errors.Is(err, io.EOF) {
		return http.StatusBadRequest, JSONResponse{Message: "request body is required"}
	}

	return http.StatusInternalServerError, JSONResponse{Message: "internal server error"}
}

func toFieldErrors(errs validator.ValidationErrors) []FieldError {
	out := make([]FieldError, 0, len(errs))
	for _, e := range errs {
		out = append(out, FieldError{
			Field:   e.Field(),
			Message: validationMessage(e),
		})
	}
	return out
}

func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid url"
	case "min":
		return "must be at least " + e.Param()
	case "max":
		return "must be at most " + e.Param()
	case "gt":
		return "must be greater than " + e.Param()
	case "gte":
		return "must be greater than or equal to " + e.Param()
	case "oneof":
		return "must be one of: " + e.Param()
	default:
		return "is invalid"
	}
}
