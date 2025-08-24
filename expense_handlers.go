package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings" // Add this line
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

// GetMonthlyExpenseSummary handles getting monthly expense totals for chart display
func (h *ExpenseHandler) GetMonthlyExpenseSummary(c echo.Context) error {
	// Verify user authentication
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	// Get monthly summary data from database
	summary, err := h.getMonthlyExpenseSummary(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get monthly summary: %v", err),
		})
	}

	// Return summary data directly as array
	return c.JSON(http.StatusOK, summary)
}

// getMonthlyExpenseSummary aggregates expenses by month for the user
func (h *ExpenseHandler) getMonthlyExpenseSummary(userID uuid.UUID) ([]map[string]interface{}, error) {
	// SQL query to group expenses by month and sum amounts
	query := `
		SELECT 
			TO_CHAR(expense_date, 'Mon YYYY') as month,
			SUM(amount) as total
		FROM expenses 
		WHERE user_id = $1 
		GROUP BY TO_CHAR(expense_date, 'YYYY-MM'), TO_CHAR(expense_date, 'Mon YYYY')
		ORDER BY TO_CHAR(expense_date, 'YYYY-MM') DESC
	`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Build response array with month and total pairs
	summary := make([]map[string]interface{}, 0)
	for rows.Next() {
		var month string
		var total float64
		if err := rows.Scan(&month, &total); err != nil {
			return nil, err
		}
		
		// Add formatted entry to summary
		summary = append(summary, map[string]interface{}{
			"month": month,
			"total": total,
		})
	}

	return summary, nil
}

// getWeeklySummary aggregates expenses by week for the last 4 weeks
func (h *ExpenseHandler) getWeeklySummary(userID uuid.UUID) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			'Week ' || EXTRACT(WEEK FROM expense_date) || ', ' || EXTRACT(YEAR FROM expense_date) as week,
			SUM(amount) as total
		FROM expenses 
		WHERE user_id = $1 
		AND expense_date >= CURRENT_DATE - INTERVAL '4 weeks'
		GROUP BY EXTRACT(YEAR FROM expense_date), EXTRACT(WEEK FROM expense_date)
		ORDER BY EXTRACT(YEAR FROM expense_date) DESC, EXTRACT(WEEK FROM expense_date) DESC
		LIMIT 4
	`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	weeklySummary := make([]map[string]interface{}, 0)
	for rows.Next() {
		var week string
		var total float64
		if err := rows.Scan(&week, &total); err != nil {
			return nil, err
		}
		
		weeklySummary = append(weeklySummary, map[string]interface{}{
			"week":  week,
			"total": total,
		})
	}

	return weeklySummary, nil
}

// getDailySummary aggregates expenses by day for the last 7 days
func (h *ExpenseHandler) getDailySummary(userID uuid.UUID) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			TO_CHAR(expense_date, 'DD Mon') as day,
			SUM(amount) as total
		FROM expenses 
		WHERE user_id = $1 
		AND expense_date >= CURRENT_DATE - INTERVAL '7 days'
		GROUP BY expense_date
		ORDER BY expense_date DESC
		LIMIT 7
	`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dailySummary := make([]map[string]interface{}, 0)
	for rows.Next() {
		var day string
		var total float64
		if err := rows.Scan(&day, &total); err != nil {
			return nil, err
		}
		
		dailySummary = append(dailySummary, map[string]interface{}{
			"day":   day,
			"total": total,
		})
	}

	return dailySummary, nil
}

// GetDailySummaryPaginated handles getting paginated daily expense summary
func (h *ExpenseHandler) GetDailySummaryPaginated(c echo.Context) error {
	// Verify user authentication
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	// Parse pagination parameters
	page := 1
	limit := 10
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get paginated daily summary
	summary, total, err := h.getDailySummaryPaginated(userID, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get daily summary: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        summary,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": (total + limit - 1) / limit,
	})
}

// GetMonthlySummaryPaginated handles getting paginated monthly expense summary
func (h *ExpenseHandler) GetMonthlySummaryPaginated(c echo.Context) error {
	// Verify user authentication
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	// Parse pagination parameters
	page := 1
	limit := 12
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get paginated monthly summary
	summary, total, err := h.getMonthlySummaryPaginated(userID, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get monthly summary: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        summary,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": (total + limit - 1) / limit,
	})
}

// GetWeeklySummaryPaginated handles getting paginated weekly expense summary for a specific month
func (h *ExpenseHandler) GetWeeklySummaryPaginated(c echo.Context) error {
	// Verify user authentication
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	// Parse month parameter (required)
	month := c.QueryParam("month") // Format: YYYY-MM
	if month == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Month parameter is required (format: YYYY-MM)",
		})
	}

	// Parse pagination parameters
	page := 1
	limit := 10
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get paginated weekly summary for the month
	summary, total, err := h.getWeeklySummaryPaginated(userID, month, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get weekly summary: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        summary,
		"month":       month,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": (total + limit - 1) / limit,
	})
}

// Helper functions for paginated summaries

// getDailySummaryPaginated gets paginated daily expense summary
func (h *ExpenseHandler) getDailySummaryPaginated(userID uuid.UUID, page, limit int) ([]map[string]interface{}, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT expense_date) 
		FROM expenses 
		WHERE user_id = $1
	`
	var total int
	err := h.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * limit
	query := `
		SELECT 
			TO_CHAR(expense_date, 'DD Mon YYYY') as day,
			expense_date,
			SUM(amount) as total,
			COUNT(*) as expense_count
		FROM expenses 
		WHERE user_id = $1 
		GROUP BY expense_date
		ORDER BY expense_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	summary := make([]map[string]interface{}, 0)
	for rows.Next() {
		var day string
		var expenseDate time.Time
		var total float64
		var count int
		if err := rows.Scan(&day, &expenseDate, &total, &count); err != nil {
			return nil, 0, err
		}
		
		summary = append(summary, map[string]interface{}{
			"day":           day,
			"date":          expenseDate.Format("2006-01-02"),
			"total":         total,
			"expense_count": count,
		})
	}

	return summary, total, nil
}

// getMonthlySummaryPaginated gets paginated monthly expense summary
func (h *ExpenseHandler) getMonthlySummaryPaginated(userID uuid.UUID, page, limit int) ([]map[string]interface{}, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT TO_CHAR(expense_date, 'YYYY-MM')) 
		FROM expenses 
		WHERE user_id = $1
	`
	var total int
	err := h.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * limit
	query := `
		SELECT 
			TO_CHAR(expense_date, 'Mon YYYY') as month,
			TO_CHAR(expense_date, 'YYYY-MM') as month_key,
			SUM(amount) as total,
			COUNT(*) as expense_count
		FROM expenses 
		WHERE user_id = $1 
		GROUP BY TO_CHAR(expense_date, 'YYYY-MM'), TO_CHAR(expense_date, 'Mon YYYY')
		ORDER BY TO_CHAR(expense_date, 'YYYY-MM') DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	summary := make([]map[string]interface{}, 0)
	for rows.Next() {
		var month, monthKey string
		var total float64
		var count int
		if err := rows.Scan(&month, &monthKey, &total, &count); err != nil {
			return nil, 0, err
		}
		
		summary = append(summary, map[string]interface{}{
			"month":         month,
			"month_key":     monthKey,
			"total":         total,
			"expense_count": count,
		})
	}

	return summary, total, nil
}

// getWeeklySummaryPaginated gets paginated weekly expense summary for a specific month
func (h *ExpenseHandler) getWeeklySummaryPaginated(userID uuid.UUID, month string, page, limit int) ([]map[string]interface{}, int, error) {
	// Get total count for the month
	countQuery := `
		SELECT COUNT(DISTINCT EXTRACT(WEEK FROM expense_date)) 
		FROM expenses 
		WHERE user_id = $1 AND TO_CHAR(expense_date, 'YYYY-MM') = $2
	`
	var total int
	err := h.db.QueryRow(countQuery, userID, month).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * limit
	query := `
		SELECT 
			'Week ' || EXTRACT(WEEK FROM expense_date) as week,
			EXTRACT(WEEK FROM expense_date) as week_number,
			SUM(amount) as total,
			COUNT(*) as expense_count,
			MIN(expense_date) as week_start,
			MAX(expense_date) as week_end
		FROM expenses 
		WHERE user_id = $1 AND TO_CHAR(expense_date, 'YYYY-MM') = $2
		GROUP BY EXTRACT(WEEK FROM expense_date)
		ORDER BY EXTRACT(WEEK FROM expense_date) DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := h.db.Query(query, userID, month, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	summary := make([]map[string]interface{}, 0)
	for rows.Next() {
		var week string
		var weekNumber int
		var total float64
		var count int
		var weekStart, weekEnd time.Time
		if err := rows.Scan(&week, &weekNumber, &total, &count, &weekStart, &weekEnd); err != nil {
			return nil, 0, err
		}
		
		summary = append(summary, map[string]interface{}{
			"week":          week,
			"week_number":   weekNumber,
			"total":         total,
			"expense_count": count,
			"week_start":    weekStart.Format("2006-01-02"),
			"week_end":      weekEnd.Format("2006-01-02"),
		})
	}

	return summary, total, nil
}

// GetDashboard handles getting comprehensive dashboard data for the user
func (h *ExpenseHandler) GetDashboard(c echo.Context) error {
	// Verify user authentication
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	// Get dashboard data from database
	dashboard, err := h.getDashboardData(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get dashboard data: %v", err),
		})
	}

	// Return comprehensive dashboard data
	return c.JSON(http.StatusOK, dashboard)
}

// getDashboardData aggregates all dashboard metrics for the user
func (h *ExpenseHandler) getDashboardData(userID uuid.UUID) (map[string]interface{}, error) {
	// Get total expenses count and amount
	totalQuery := `SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = $1`
	var totalCount int
	var totalAmount float64
	err := h.db.QueryRow(totalQuery, userID).Scan(&totalCount, &totalAmount)
	if err != nil {
		return nil, err
	}

	// Get current month expenses
	currentMonthQuery := `
		SELECT COUNT(*), COALESCE(SUM(amount), 0) 
		FROM expenses 
		WHERE user_id = $1 
		AND EXTRACT(MONTH FROM expense_date) = EXTRACT(MONTH FROM CURRENT_DATE)
		AND EXTRACT(YEAR FROM expense_date) = EXTRACT(YEAR FROM CURRENT_DATE)
	`
	var currentMonthCount int
	var currentMonthAmount float64
	err = h.db.QueryRow(currentMonthQuery, userID).Scan(&currentMonthCount, &currentMonthAmount)
	if err != nil {
		return nil, err
	}

	// Get current week expenses
	currentWeekQuery := `
		SELECT COUNT(*), COALESCE(SUM(amount), 0) 
		FROM expenses 
		WHERE user_id = $1 
		AND expense_date >= DATE_TRUNC('week', CURRENT_DATE)
		AND expense_date < DATE_TRUNC('week', CURRENT_DATE) + INTERVAL '1 week'
	`
	var currentWeekCount int
	var currentWeekAmount float64
	err = h.db.QueryRow(currentWeekQuery, userID).Scan(&currentWeekCount, &currentWeekAmount)
	if err != nil {
		return nil, err
	}

	// Get today's expenses
	todayQuery := `
		SELECT COUNT(*), COALESCE(SUM(amount), 0) 
		FROM expenses 
		WHERE user_id = $1 
		AND DATE(expense_date) = CURRENT_DATE
	`
	var todayCount int
	var todayAmount float64
	err = h.db.QueryRow(todayQuery, userID).Scan(&todayCount, &todayAmount)
	if err != nil {
		return nil, err
	}

	// Get monthly summary for charts
	monthlySummary, err := h.getMonthlyExpenseSummary(userID)
	if err != nil {
		return nil, err
	}

	// Get weekly summary for the last 4 weeks
	weeklySummary, err := h.getWeeklySummary(userID)
	if err != nil {
		return nil, err
	}

	// Get daily summary for the last 7 days
	dailySummary, err := h.getDailySummary(userID)
	if err != nil {
		return nil, err
	}

	// Get recent expenses (last 5)
	recentQuery := `
		SELECT id, title, amount, expense_date, expense_time 
		FROM expenses 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT 5
	`
	rows, err := h.db.Query(recentQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recentExpenses := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id uuid.UUID
		var title string
		var amount float64
		var expenseDate, expenseTime time.Time
		if err := rows.Scan(&id, &title, &amount, &expenseDate, &expenseTime); err != nil {
			return nil, err
		}
		recentExpenses = append(recentExpenses, map[string]interface{}{
			"id":           id,
			"title":        title,
			"amount":       amount,
			"expense_date": expenseDate.Format("02-01-2006"),
			"expense_time": expenseTime.Format("03:04 PM"),
		})
	}

	// Build comprehensive dashboard response
	dashboard := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_expenses":        totalCount,
			"total_amount":          totalAmount,
			"current_month_count":   currentMonthCount,
			"current_month_amount":  currentMonthAmount,
			"current_week_count":    currentWeekCount,
			"current_week_amount":   currentWeekAmount,
			"today_count":           todayCount,
			"today_amount":          todayAmount,
		},
		"monthly_summary":  monthlySummary,
		"weekly_summary":   weeklySummary,
		"daily_summary":    dailySummary,
		"recent_expenses": recentExpenses,
	}

	return dashboard, nil
}

// getUserExpenses retrieves all expenses for a user (backward compatibility)
func (h *ExpenseHandler) getUserExpenses(userID uuid.UUID) ([]map[string]interface{}, error) {
	// Use the new filtering function with empty filters for backward compatibility
	return h.getUserExpensesWithFilters(userID, &ExpenseFilters{})
}