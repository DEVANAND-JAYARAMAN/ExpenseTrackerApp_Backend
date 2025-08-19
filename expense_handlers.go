package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
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
	if strings.TrimSpace(req.Title) == "" || req.Amount <= 0 || len(req.Categories) == 0 || strings.TrimSpace(req.ExpenseDate) == "" || strings.TrimSpace(req.ExpenseTime) == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing or invalid fields",
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
	now := time.Now()
	query := `INSERT INTO expenses (id, user_id, title, description, amount, expense_date, expense_time, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = h.db.Exec(query, expenseID, userID, req.Title, req.Description, req.Amount, expenseDate, expenseTime, now, now)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to create expense: %v", err),
		})
	}

	// Link categories
	var categoryDetails []ExpenseCategoryDetail
	for _, catID := range req.Categories {
		// Insert into expense_categories
		_, err := h.db.Exec(`INSERT INTO expense_categories (id, expense_id, category_id) VALUES ($1, $2, $3)`, uuid.New(), expenseID, catID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: fmt.Sprintf("Failed to link category: %v", err),
			})
		}
		// Get category details
		var cat ExpenseCategoryDetail
		err = h.db.QueryRow(`SELECT id, name, is_default FROM categories WHERE id = $1`, catID).Scan(&cat.ID, &cat.Name, &cat.IsDefault)
		if err == nil {
			categoryDetails = append(categoryDetails, cat)
		}
	}

	// Build response
	resp := ExpenseDetailResponse{}
	resp.Message = "Expense created successfully."
	resp.Expense.ID = expenseID
	resp.Expense.UserID = userID
	resp.Expense.Title = req.Title
	resp.Expense.Description = req.Description
	resp.Expense.Amount = req.Amount
	resp.Expense.ExpenseDate = req.ExpenseDate
	resp.Expense.ExpenseTime = req.ExpenseTime
	resp.Expense.CreatedAt = now.Format(time.RFC3339)
	resp.Expense.UpdatedAt = now.Format(time.RFC3339)
	resp.Expense.Categories = categoryDetails

	return c.JSON(http.StatusCreated, resp)
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
	if strings.TrimSpace(req.Title) == "" || req.Amount <= 0 || len(req.Categories) == 0 || strings.TrimSpace(req.ExpenseDate) == "" || strings.TrimSpace(req.ExpenseTime) == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing or invalid fields",
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

	// Parse date and time
	expenseDate, expenseTime, err := parseDateTime(req.ExpenseDate, req.ExpenseTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid date or time format",
		})
	}

	// Update expense fields
	query := `UPDATE expenses SET title = $2, description = $3, amount = $4, expense_date = $5, expense_time = $6, updated_at = $7 WHERE id = $1`
	_, err = h.db.Exec(query, expenseID, req.Title, req.Description, req.Amount, expenseDate, expenseTime, time.Now())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to update expense",
		})
	}

	// Update categories: remove old links, add new ones
	_, err = h.db.Exec(`DELETE FROM expense_categories WHERE expense_id = $1`, expenseID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to update expense categories",
		})
	}
	var categoryDetails []ExpenseCategoryDetail
	for _, catID := range req.Categories {
		_, err := h.db.Exec(`INSERT INTO expense_categories (id, expense_id, category_id) VALUES ($1, $2, $3)`, uuid.New(), expenseID, catID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: fmt.Sprintf("Failed to link category: %v", err),
			})
		}
		var cat ExpenseCategoryDetail
		err = h.db.QueryRow(`SELECT id, name, is_default FROM categories WHERE id = $1`, catID).Scan(&cat.ID, &cat.Name, &cat.IsDefault)
		if err == nil {
			categoryDetails = append(categoryDetails, cat)
		}
	}

	// Build response
	resp := ExpenseDetailResponse{}
	resp.Message = "Expense updated successfully."
	resp.Expense.ID = expenseID
	resp.Expense.UserID = userID
	resp.Expense.Title = req.Title
	resp.Expense.Description = req.Description
	resp.Expense.Amount = req.Amount
	resp.Expense.ExpenseDate = req.ExpenseDate
	resp.Expense.ExpenseTime = req.ExpenseTime
	resp.Expense.UpdatedAt = time.Now().Format(time.RFC3339)
	resp.Expense.CreatedAt = ""
	resp.Expense.Categories = categoryDetails

	return c.JSON(http.StatusOK, resp)
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

// GetExpenses handles getting expenses with optional filtering by category, date range, and amount
func (h *ExpenseHandler) GetExpenses(c echo.Context) error {
	// Extract user ID from JWT token in request context
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	// Parse and validate query parameters for filtering
	filters, err := h.parseExpenseFilters(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	// Fetch expenses with applied filters
	expenses, err := h.getUserExpensesWithFilters(userID, filters)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get expenses: %v", err),
		})
	}

	// Return filtered expenses list with count and success message
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
	if len(req.Categories) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "At least one category is required")
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
	if len(req.Categories) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "At least one category is required")
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
	// Parse date (DD-MM-YYYY)
	expenseDate, err := time.Parse("02-01-2006", dateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// Parse time (HH:MM AM/PM)
	expenseTime, err := time.Parse("03:04 PM", timeStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return expenseDate, expenseTime, nil
}

// createExpense / updateExpense legacy helpers removed (multi-category handled separately)

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

// ExpenseFilters holds the filtering criteria for expense queries
type ExpenseFilters struct {
	CategoryID *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	MinAmount  *float64
	MaxAmount  *float64
}

// parseExpenseFilters extracts and validates filter parameters from query string
func (h *ExpenseHandler) parseExpenseFilters(c echo.Context) (*ExpenseFilters, error) {
	filters := &ExpenseFilters{}

	// Parse category_id filter if provided
	if categoryStr := c.QueryParam("category_id"); categoryStr != "" {
		categoryID, err := uuid.Parse(categoryStr)
		if err != nil {
			return nil, fmt.Errorf("invalid category ID format")
		}
		filters.CategoryID = &categoryID
	}

	// Parse start_date filter with validation
	if startDateStr := c.QueryParam("start_date"); startDateStr != "" {
		startDate, err := time.Parse("02-01-2006", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format. Use dd-mm-yyyy")
		}
		filters.StartDate = &startDate
	}

	// Parse end_date filter with validation
	if endDateStr := c.QueryParam("end_date"); endDateStr != "" {
		endDate, err := time.Parse("02-01-2006", endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format. Use dd-mm-yyyy")
		}
		filters.EndDate = &endDate
	}

	// Parse min_amount filter
	if minAmountStr := c.QueryParam("min_amount"); minAmountStr != "" {
		var minAmount float64
		if _, err := fmt.Sscanf(minAmountStr, "%f", &minAmount); err != nil || minAmount < 0 {
			return nil, fmt.Errorf("invalid min_amount value")
		}
		filters.MinAmount = &minAmount
	}

	// Parse max_amount filter
	if maxAmountStr := c.QueryParam("max_amount"); maxAmountStr != "" {
		var maxAmount float64
		if _, err := fmt.Sscanf(maxAmountStr, "%f", &maxAmount); err != nil || maxAmount < 0 {
			return nil, fmt.Errorf("invalid max_amount value")
		}
		filters.MaxAmount = &maxAmount
	}

	// Validate date range if both dates are provided
	if filters.StartDate != nil && filters.EndDate != nil && filters.StartDate.After(*filters.EndDate) {
		return nil, fmt.Errorf("start_date cannot be after end_date")
	}

	// Validate amount range if both amounts are provided
	if filters.MinAmount != nil && filters.MaxAmount != nil && *filters.MinAmount > *filters.MaxAmount {
		return nil, fmt.Errorf("min_amount cannot be greater than max_amount")
	}

	return filters, nil
}

// getUserExpensesWithFilters retrieves user expenses with applied filters
func (h *ExpenseHandler) getUserExpensesWithFilters(userID uuid.UUID, filters *ExpenseFilters) ([]map[string]interface{}, error) {
	// Build dynamic query based on provided filters
	queryBuilder := strings.Builder{}
	args := []interface{}{userID}
	argIndex := 2

	// Base query with potential JOIN for category filtering
	if filters.CategoryID != nil {
		queryBuilder.WriteString(`
			SELECT DISTINCT e.id, e.user_id, e.title, COALESCE(e.description, '') as description, 
			       e.amount, e.expense_date, e.expense_time, e.created_at, e.updated_at 
			FROM expenses e 
			JOIN expense_categories ec ON e.id = ec.expense_id 
			WHERE e.user_id = $1`)
	} else {
		queryBuilder.WriteString(`
			SELECT e.id, e.user_id, e.title, COALESCE(e.description, '') as description, 
			       e.amount, e.expense_date, e.expense_time, e.created_at, e.updated_at 
			FROM expenses e 
			WHERE e.user_id = $1`)
	}

	// Add category filter if specified
	if filters.CategoryID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND ec.category_id = $%d", argIndex))
		args = append(args, *filters.CategoryID)
		argIndex++
	}

	// Add date range filters
	if filters.StartDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND e.expense_date >= $%d", argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}
	if filters.EndDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND e.expense_date <= $%d", argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}

	// Add amount range filters
	if filters.MinAmount != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND e.amount >= $%d", argIndex))
		args = append(args, *filters.MinAmount)
		argIndex++
	}
	if filters.MaxAmount != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND e.amount <= $%d", argIndex))
		args = append(args, *filters.MaxAmount)
		argIndex++
	}

	// Always order by creation date (newest first)
	queryBuilder.WriteString(" ORDER BY e.created_at DESC")

	// Execute the dynamically built query
	rows, err := h.db.Query(queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results same as before
	expenses := make([]map[string]interface{}, 0)
	idOrder := make([]uuid.UUID, 0)
	indexByID := make(map[uuid.UUID]int)

	for rows.Next() {
		var expID, uID uuid.UUID
		var title, description string
		var amount float64
		var expenseDate, expenseTime, createdAt, updatedAt time.Time
		if err := rows.Scan(&expID, &uID, &title, &description, &amount, &expenseDate, &expenseTime, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		expMap := map[string]interface{}{
			"id":           expID,
			"user_id":      uID,
			"title":        title,
			"description":  description,
			"amount":       amount,
			"expense_date": expenseDate.Format("02-01-2006"),
			"expense_time": expenseTime.Format("03:04 PM"),
			"created_at":   createdAt.Format("02-01-2006 03:04:05 PM"),
			"updated_at":   updatedAt.Format("02-01-2006 03:04:05 PM"),
			"categories":   []map[string]interface{}{},
		}
		indexByID[expID] = len(expenses)
		idOrder = append(idOrder, expID)
		expenses = append(expenses, expMap)
	}

	if len(idOrder) == 0 {
		return expenses, nil
	}

	// Fetch and attach category information for each expense
	if err := h.attachCategoriesToExpenses(expenses, idOrder, indexByID); err != nil {
		return nil, err
	}

	return expenses, nil
}

// attachCategoriesToExpenses fetches and attaches category data to expense records
func (h *ExpenseHandler) attachCategoriesToExpenses(expenses []map[string]interface{}, idOrder []uuid.UUID, indexByID map[uuid.UUID]int) error {
	// Build dynamic IN clause for categories
	placeholders := make([]string, len(idOrder))
	args := make([]interface{}, len(idOrder))
	for i, id := range idOrder {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	catQuery := fmt.Sprintf(`SELECT ec.expense_id, c.id, c.name, c.is_default FROM expense_categories ec JOIN categories c ON c.id = ec.category_id WHERE ec.expense_id IN (%s) ORDER BY c.name ASC`, strings.Join(placeholders, ","))

	catRows, err := h.db.Query(catQuery, args...)
	if err != nil {
		return err
	}
	defer catRows.Close()

	// Attach categories to their respective expenses
	for catRows.Next() {
		var expID, catID uuid.UUID
		var catName string
		var isDefault bool
		if err := catRows.Scan(&expID, &catID, &catName, &isDefault); err != nil {
			return err
		}
		if idx, ok := indexByID[expID]; ok {
			expense := expenses[idx]
			cats := expense["categories"].([]map[string]interface{})
			cats = append(cats, map[string]interface{}{
				"id":         catID,
				"name":       catName,
				"is_default": isDefault,
			})
			expense["categories"] = cats
		}
	}

	return nil
}

// getUserExpenses retrieves all expenses for a user (backward compatibility)
func (h *ExpenseHandler) getUserExpenses(userID uuid.UUID) ([]map[string]interface{}, error) {
	// Use the new filtering function with empty filters for backward compatibility
	return h.getUserExpensesWithFilters(userID, &ExpenseFilters{})
}