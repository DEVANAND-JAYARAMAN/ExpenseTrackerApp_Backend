package main

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Email          string     `json:"email" db:"email"`
	Password       string     `json:"-" db:"password"`
	ProfileImage   *string    `json:"profile_image,omitempty" db:"profile_image"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	DeactivatedAt  *time.Time `json:"deactivated_at,omitempty" db:"deactivated_at"`
}

// Category represents an expense category
type Category struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	IsDefault bool      `json:"is_default" db:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Expense represents an expense record
type Expense struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description,omitempty" db:"description"`
	Amount      float64   `json:"amount" db:"amount"`
	ExpenseDate time.Time `json:"expense_date" db:"expense_date"`
	ExpenseTime time.Time `json:"expense_time" db:"expense_time"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ExpenseCategory represents the many-to-many relationship between expenses and categories
type ExpenseCategory struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ExpenseID  uuid.UUID `json:"expense_id" db:"expense_id"`
	CategoryID uuid.UUID `json:"category_id" db:"category_id"`
}

// LoginHistory represents user login history
type LoginHistory struct {
	ID      uuid.UUID `json:"id" db:"id"`
	UserID  uuid.UUID `json:"user_id" db:"user_id"`
	LoginAt time.Time `json:"login_at" db:"login_at"`
}

// Session represents user session data
type Session struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsActive  bool       `json:"is_active" db:"is_active"`
}

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterResponse represents the response for successful user registration
type RegisterResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
