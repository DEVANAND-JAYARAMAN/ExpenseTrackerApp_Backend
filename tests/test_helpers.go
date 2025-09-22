package tests

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// Test data structures
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Expense struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Title        string    `json:"title"`
	Amount       float64   `json:"amount"`
	Date         string    `json:"date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CategoryName string    `json:"category_name,omitempty"`
}

// Helper functions
func NewMockDB() (*sql.DB, sqlmock.Sqlmock, error) {
	return sqlmock.New()
}

// AnyTime matches any time.Time value
type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

// AnyUUID matches any UUID value
type AnyUUID struct{}

func (a AnyUUID) Match(v driver.Value) bool {
	_, ok := v.(uuid.UUID)
	if !ok {
		// Try string representation
		if s, ok := v.(string); ok {
			_, err := uuid.Parse(s)
			return err == nil
		}
	}
	return ok
}

// Test utilities
func CreateTestUser() User {
	return User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}
}

func CreateTestCategory() Category {
	return Category{
		ID:        uuid.New(),
		Name:      "Test Category",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func CreateTestExpense(userID uuid.UUID) Expense {
	return Expense{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     "Test Expense",
		Amount:    25.50,
		Date:      "2024-01-15",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}