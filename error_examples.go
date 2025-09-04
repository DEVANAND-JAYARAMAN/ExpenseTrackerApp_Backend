package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

// Example usage of the standardized error system

// Example 1: Using predefined error codes
func ExampleSessionExpired(c echo.Context) error {
	// When JWT token is expired or session is invalid
	return SendStandardError(c, ErrorSessionExpired)
	// Returns: {"error": "session_expired", "message": "Session expired, redirecting to login page", "status_code": 401}
}

// Example 2: Using custom error messages
func ExampleCustomValidation(c echo.Context) error {
	// When you need a specific validation message
	return SendCustomError(c, ErrorValidationFailed, "Password must contain at least one uppercase letter", http.StatusBadRequest)
	// Returns: {"error": "validation_failed", "message": "Password must contain at least one uppercase letter", "status_code": 400}
}

// Example 3: Resource not found with custom message
func ExampleExpenseNotFound(c echo.Context, expenseID string) error {
	message := "Expense with ID " + expenseID + " not found"
	return SendCustomError(c, ErrorNotFound, message, http.StatusNotFound)
	// Returns: {"error": "not_found", "message": "Expense with ID abc123 not found", "status_code": 404}
}

// Example 4: Database connection error
func ExampleDatabaseError(c echo.Context) error {
	return SendStandardError(c, ErrorDatabaseError)
	// Returns: {"error": "database_error", "message": "Database operation failed", "status_code": 500}
}

// Example 5: Forbidden access to resource
func ExampleForbiddenAccess(c echo.Context) error {
	return SendCustomError(c, ErrorForbidden, "You can only access your own expenses", http.StatusForbidden)
	// Returns: {"error": "forbidden", "message": "You can only access your own expenses", "status_code": 403}
}

// Example 6: Creating a completely custom error
func ExampleCustomError(c echo.Context) error {
	return SendCustomError(c, "rate_limit_exceeded", "Too many requests. Please try again in 5 minutes", http.StatusTooManyRequests)
	// Returns: {"error": "rate_limit_exceeded", "message": "Too many requests. Please try again in 5 minutes", "status_code": 429}
}

// Example 7: Multiple validation errors
func ExampleMultipleValidationErrors(c echo.Context) error {
	errors := []string{
		"Name is required",
		"Email format is invalid", 
		"Password must be at least 8 characters",
	}
	
	message := "Validation failed: "
	for i, err := range errors {
		if i > 0 {
			message += ", "
		}
		message += err
	}
	
	return SendCustomError(c, ErrorValidationFailed, message, http.StatusBadRequest)
	// Returns: {"error": "validation_failed", "message": "Validation failed: Name is required, Email format is invalid, Password must be at least 8 characters", "status_code": 400}
}