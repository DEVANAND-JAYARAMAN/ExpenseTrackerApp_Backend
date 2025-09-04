package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

// Error codes for standardized error responses
const (
	// Authentication errors
	ErrorSessionExpired    = "session_expired"
	ErrorUnauthorized      = "unauthorized"
	ErrorInvalidToken      = "invalid_token"
	ErrorInvalidCredentials = "invalid_credentials"
	
	// Validation errors
	ErrorValidationFailed  = "validation_failed"
	ErrorInvalidRequest    = "invalid_request"
	ErrorMissingFields     = "missing_fields"
	
	// Resource errors
	ErrorNotFound          = "not_found"
	ErrorAlreadyExists     = "already_exists"
	ErrorForbidden         = "forbidden"
	
	// Server errors
	ErrorInternalServer    = "internal_server_error"
	ErrorDatabaseError     = "database_error"
	ErrorServiceUnavailable = "service_unavailable"
)

// Predefined error responses
var ErrorResponses = map[string]StandardErrorResponse{
	ErrorSessionExpired: {
		Error:      ErrorSessionExpired,
		Message:    "Session expired, redirecting to login page",
		StatusCode: http.StatusUnauthorized,
	},
	ErrorUnauthorized: {
		Error:      ErrorUnauthorized,
		Message:    "Access denied. Please login to continue",
		StatusCode: http.StatusUnauthorized,
	},
	ErrorInvalidToken: {
		Error:      ErrorInvalidToken,
		Message:    "Invalid or malformed authentication token",
		StatusCode: http.StatusUnauthorized,
	},
	ErrorInvalidCredentials: {
		Error:      ErrorInvalidCredentials,
		Message:    "Invalid email or password",
		StatusCode: http.StatusUnauthorized,
	},
	ErrorValidationFailed: {
		Error:      ErrorValidationFailed,
		Message:    "Request validation failed",
		StatusCode: http.StatusBadRequest,
	},
	ErrorInvalidRequest: {
		Error:      ErrorInvalidRequest,
		Message:    "Invalid request format or data",
		StatusCode: http.StatusBadRequest,
	},
	ErrorMissingFields: {
		Error:      ErrorMissingFields,
		Message:    "Required fields are missing",
		StatusCode: http.StatusBadRequest,
	},
	ErrorNotFound: {
		Error:      ErrorNotFound,
		Message:    "Requested resource not found",
		StatusCode: http.StatusNotFound,
	},
	ErrorAlreadyExists: {
		Error:      ErrorAlreadyExists,
		Message:    "Resource already exists",
		StatusCode: http.StatusConflict,
	},
	ErrorForbidden: {
		Error:      ErrorForbidden,
		Message:    "Access to this resource is forbidden",
		StatusCode: http.StatusForbidden,
	},
	ErrorInternalServer: {
		Error:      ErrorInternalServer,
		Message:    "Internal server error occurred",
		StatusCode: http.StatusInternalServerError,
	},
	ErrorDatabaseError: {
		Error:      ErrorDatabaseError,
		Message:    "Database operation failed",
		StatusCode: http.StatusInternalServerError,
	},
	ErrorServiceUnavailable: {
		Error:      ErrorServiceUnavailable,
		Message:    "Service temporarily unavailable",
		StatusCode: http.StatusServiceUnavailable,
	},
}

// NewStandardError creates a standardized error response
func NewStandardError(errorCode string) StandardErrorResponse {
	if response, exists := ErrorResponses[errorCode]; exists {
		return response
	}
	// Default to internal server error if code not found
	return ErrorResponses[ErrorInternalServer]
}

// NewCustomError creates a custom standardized error response
func NewCustomError(errorCode, message string, statusCode int) StandardErrorResponse {
	return StandardErrorResponse{
		Error:      errorCode,
		Message:    message,
		StatusCode: statusCode,
	}
}

// SendStandardError sends a standardized error response
func SendStandardError(c echo.Context, errorCode string) error {
	response := NewStandardError(errorCode)
	return c.JSON(response.StatusCode, response)
}

// SendCustomError sends a custom standardized error response
func SendCustomError(c echo.Context, errorCode, message string, statusCode int) error {
	response := NewCustomError(errorCode, message, statusCode)
	return c.JSON(statusCode, response)
}