package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ExpenseHandler handles expense-related requests
type ExpenseHandler struct {
	db *sql.DB
}

// NewExpenseHandler creates a new ExpenseHandler instance
func NewExpenseHandler(db *sql.DB) *ExpenseHandler {
	return &ExpenseHandler{db: db}
}

// AddExpense handles adding a new expense
func (h *ExpenseHandler) AddExpense(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	var req AddExpenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate request
	if err := validateAddExpenseRequest(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	// Validate category ID
	categoryName, exists := GetCategoryName(req.CategoryID)
	if !exists {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid category ID",
		})
	}

	// Parse date and time
	expenseDate, expenseTime, err := parseDateTime(req.ExpenseDate, req.ExpenseTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid date or time format",
		})
	}

	// Create expense
	expenseID := uuid.New()
	err = h.createExpense(expenseID, userID, req.Title, req.Description, req.Amount, req.CategoryID, categoryName, expenseDate, expenseTime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to create expense: %v", err),
		})
	}

	return c.JSON(http.StatusCreated, AddExpenseResponse{
		Message:   "Expense added successfully",
		ExpenseID: expenseID,
	})
}

// UpdateExpense handles updating an existing expense
func (h *ExpenseHandler) UpdateExpense(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	expenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid expense ID",
		})
	}

	var req UpdateExpenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate request
	if err := validateUpdateExpenseRequest(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	// Check if expense exists and belongs to user
	exists, err := h.expenseExistsForUser(expenseID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Database error: %v", err),
		})
	}
	if !exists {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: fmt.Sprintf("Expense %s not found for user %s", expenseID, userID),
		})
	}

	// Validate category ID
	categoryName, exists := GetCategoryName(req.CategoryID)
	if !exists {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid category ID",
		})
	}

	// Parse date and time
	expenseDate, expenseTime, err := parseDateTime(req.ExpenseDate, req.ExpenseTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid date or time format",
		})
	}

	// Update expense
	err = h.updateExpense(expenseID, req.Title, req.Description, req.Amount, req.CategoryID, categoryName, expenseDate, expenseTime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to update expense",
		})
	}

	return c.JSON(http.StatusOK, UpdateExpenseResponse{
		Message: "Expense updated successfully",
	})
}

// DeleteExpense handles deleting an expense
func (h *ExpenseHandler) DeleteExpense(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	expenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid expense ID",
		})
	}

	// Check if expense exists and belongs to user
	exists, err := h.expenseExistsForUser(expenseID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Database error: %v", err),
		})
	}
	if !exists {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: fmt.Sprintf("Expense %s not found for user %s", expenseID, userID),
		})
	}

	// Delete expense
	err = h.deleteExpense(expenseID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to delete expense",
		})
	}

	return c.JSON(http.StatusOK, DeleteExpenseResponse{
		Message: "Expense deleted successfully",
	})
}

// GetExpenses handles getting all expenses for a user
func (h *ExpenseHandler) GetExpenses(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	expenses, err := h.getUserExpenses(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get expenses: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":  "Expenses retrieved successfully",
		"count":    len(expenses),
		"expenses": expenses,
	})
}

// Helper functions

func validateAddExpenseRequest(req AddExpenseRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Title is required")
	}
	if req.Amount <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Amount must be greater than 0")
	}
	if req.CategoryID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Category ID is required")
	}
	if strings.TrimSpace(req.ExpenseDate) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Expense date is required")
	}
	if strings.TrimSpace(req.ExpenseTime) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Expense time is required")
	}
	return nil
}

func validateUpdateExpenseRequest(req UpdateExpenseRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Title is required")
	}
	if req.Amount <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Amount must be greater than 0")
	}
	if req.CategoryID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Category ID is required")
	}
	if strings.TrimSpace(req.ExpenseDate) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Expense date is required")
	}
	if strings.TrimSpace(req.ExpenseTime) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Expense time is required")
	}
	return nil
}

func parseDateTime(dateStr, timeStr string) (time.Time, time.Time, error) {
	// Parse date (YYYY-MM-DD)
	expenseDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// Parse time (HH:MM)
	expenseTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return expenseDate, expenseTime, nil
}

func (h *ExpenseHandler) createExpense(id, userID uuid.UUID, title string, description *string, amount float64, categoryID int, categoryName string, expenseDate, expenseTime time.Time) error {
	query := `
		INSERT INTO expenses (id, user_id, title, description, amount, category_id, category_name, expense_date, expense_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	now := time.Now()
	_, err := h.db.Exec(query, id, userID, title, description, amount, categoryID, categoryName, expenseDate, expenseTime, now, now)
	return err
}

func (h *ExpenseHandler) updateExpense(id uuid.UUID, title string, description *string, amount float64, categoryID int, categoryName string, expenseDate, expenseTime time.Time) error {
	query := `
		UPDATE expenses 
		SET title = $2, description = $3, amount = $4, category_id = $5, category_name = $6, expense_date = $7, expense_time = $8, updated_at = $9
		WHERE id = $1
	`
	_, err := h.db.Exec(query, id, title, description, amount, categoryID, categoryName, expenseDate, expenseTime, time.Now())
	return err
}

func (h *ExpenseHandler) deleteExpense(id uuid.UUID) error {
	query := `DELETE FROM expenses WHERE id = $1`
	_, err := h.db.Exec(query, id)
	return err
}

func (h *ExpenseHandler) expenseExistsForUser(expenseID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM expenses WHERE id = $1 AND user_id = $2)`
	err := h.db.QueryRow(query, expenseID, userID).Scan(&exists)
	return exists, err
}

func (h *ExpenseHandler) getUserExpenses(userID uuid.UUID) ([]map[string]interface{}, error) {
	query := `SELECT id, user_id, title, COALESCE(description, '') as description, amount, category_id, category_name, expense_date, expense_time, created_at, updated_at FROM expenses WHERE user_id = $1 ORDER BY created_at DESC`
	
	rows, err := h.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []map[string]interface{}
	for rows.Next() {
		var id, userID uuid.UUID
		var title, description, categoryName string
		var amount float64
		var categoryID int
		var expenseDate, expenseTime, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &userID, &title, &description, &amount, &categoryID, &categoryName, &expenseDate, &expenseTime, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		expense := map[string]interface{}{
			"id":            id,
			"user_id":       userID,
			"title":         title,
			"description":   description,
			"amount":        amount,
			"category_id":   categoryID,
			"category_name": categoryName,
			"expense_date":  expenseDate.Format("2006-01-02"),
			"expense_time":  expenseTime.Format("15:04"),
			"created_at":    createdAt,
			"updated_at":    updatedAt,
		}

		expenses = append(expenses, expense)
	}

	return expenses, nil
}