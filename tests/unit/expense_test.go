package unit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ExpenseHandler struct for testing
type ExpenseHandler struct {
	db *sql.DB
}

func TestExpenseHandler_ExpenseExistsForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	expenseHandler := &ExpenseHandler{db: db}
	userID := uuid.New()
	expenseID := uuid.New()

	// Mock expense exists check
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := expenseHandler.expenseExistsForUser(expenseID, userID)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseHandler_DeleteExpense(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	expenseHandler := &ExpenseHandler{db: db}
	expenseID := uuid.New()

	// Mock expense deletion
	mock.ExpectExec("DELETE FROM expenses").WillReturnResult(sqlmock.NewResult(1, 1))

	err = expenseHandler.deleteExpense(expenseID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseHandler_GetMonthlySummary(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	expenseHandler := &ExpenseHandler{db: db}
	userID := uuid.New()

	// Mock monthly summary query
	rows := sqlmock.NewRows([]string{"month", "total"}).
		AddRow("Jan 2024", 150.50).
		AddRow("Feb 2024", 200.75)

	mock.ExpectQuery("SELECT.*TO_CHAR.*month.*SUM.*amount.*total").WillReturnRows(rows)

	summary, err := expenseHandler.getMonthlyExpenseSummary(userID)
	assert.NoError(t, err)
	assert.Len(t, summary, 2)
	assert.Equal(t, "Jan 2024", summary[0]["month"])
	assert.Equal(t, 150.50, summary[0]["total"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseHandler_GetDashboard_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	expenseHandler := &ExpenseHandler{db: db}
	userID := uuid.New()

	// Mock dashboard queries
	mock.ExpectQuery("SELECT COUNT.*COALESCE.*SUM.*amount").WillReturnRows(
		sqlmock.NewRows([]string{"count", "total"}).AddRow(10, 150.50))
	mock.ExpectQuery("SELECT COUNT.*COALESCE.*SUM.*amount").WillReturnRows(
		sqlmock.NewRows([]string{"count", "total"}).AddRow(5, 75.25))
	mock.ExpectQuery("SELECT COUNT.*COALESCE.*SUM.*amount").WillReturnRows(
		sqlmock.NewRows([]string{"count", "total"}).AddRow(3, 25.00))
	mock.ExpectQuery("SELECT COUNT.*COALESCE.*SUM.*amount").WillReturnRows(
		sqlmock.NewRows([]string{"count", "total"}).AddRow(2, 10.50))

	// Mock monthly summary
	mock.ExpectQuery("SELECT.*TO_CHAR.*month.*SUM.*amount.*total").WillReturnRows(
		sqlmock.NewRows([]string{"month", "total"}).AddRow("Jan 2024", 150.50))

	// Mock recent expenses
	mock.ExpectQuery("SELECT id, title, amount, expense_date, expense_time").WillReturnRows(
		sqlmock.NewRows([]string{"id", "title", "amount", "expense_date", "expense_time"}).
			AddRow(uuid.New(), "Coffee", 4.50, time.Now(), time.Now()))

	dashboard, err := expenseHandler.getDashboardData(userID)
	assert.NoError(t, err)
	assert.NotNil(t, dashboard)
	assert.Contains(t, dashboard, "summary")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper functions for testing
func (h *ExpenseHandler) expenseExistsForUser(expenseID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM expenses WHERE id = $1 AND user_id = $2)`
	err := h.db.QueryRow(query, expenseID, userID).Scan(&exists)
	return exists, err
}

func (h *ExpenseHandler) deleteExpense(id uuid.UUID) error {
	query := `DELETE FROM expenses WHERE id = $1`
	_, err := h.db.Exec(query, id)
	return err
}

func (h *ExpenseHandler) getMonthlyExpenseSummary(userID uuid.UUID) ([]map[string]interface{}, error) {
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

	summary := make([]map[string]interface{}, 0)
	for rows.Next() {
		var month string
		var total float64
		if err := rows.Scan(&month, &total); err != nil {
			return nil, err
		}

		summary = append(summary, map[string]interface{}{
			"month": month,
			"total": total,
		})
	}

	return summary, nil
}

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
			"total_expenses":       totalCount,
			"total_amount":         totalAmount,
			"current_month_count":  currentMonthCount,
			"current_month_amount": currentMonthAmount,
			"current_week_count":   currentWeekCount,
			"current_week_amount":  currentWeekAmount,
			"today_count":          todayCount,
			"today_amount":         todayAmount,
		},
		"monthly_summary": monthlySummary,
		"recent_expenses": recentExpenses,
	}

	return dashboard, nil
}