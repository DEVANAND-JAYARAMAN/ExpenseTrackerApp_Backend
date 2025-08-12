package main

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db *sql.DB
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest

	// Parse request body
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate request
	if err := validateRegisterRequest(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	// Check if email already exists
	exists, err := h.emailExists(req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
		})
	}
	if exists {
		return c.JSON(http.StatusConflict, ErrorResponse{
			Error: "Email already exists",
		})
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
		})
	}

	// Generate UUID for user
	userID := uuid.New()

	// Insert user into database
	user, err := h.createUser(userID, req.Name, req.Email, string(passwordHash))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
		})
	}

	// Return success response
	return c.JSON(http.StatusCreated, RegisterResponse{
		Message: "User registered successfully.",
		User:    *user,
	})
}

// validateRegisterRequest validates the registration request
func validateRegisterRequest(req RegisterRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 255 {
		return echo.NewHTTPError(http.StatusBadRequest, "Name must be between 2 and 255 characters")
	}

	if strings.TrimSpace(req.Email) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email is required")
	}
	if !isValidEmail(req.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid email format")
	}

	if strings.TrimSpace(req.Password) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Password is required")
	}
	if len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "Password must be at least 8 characters")
	}

	return nil
}

// emailExists checks if an email already exists in the database
func (h *AuthHandler) emailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := h.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

// createUser creates a new user in the database
func (h *AuthHandler) createUser(id uuid.UUID, name, email, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (id, name, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, created_at, updated_at, is_active
	`

	user := &User{}
	err := h.db.QueryRow(query, id, name, email, passwordHash).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest

	// Parse request body
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate credentials
	user, err := h.validateCredentials(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Email or Password is Wrong",
		})
	}

	// Generate JWT token
	token, err := h.generateJWT(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
		})
	}

	// Record login history
	if err := h.recordLoginHistory(user.ID); err != nil {
		// Log error but don't fail the login
		// In production, you might want to log this properly
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Message: "Login successful.",
		Token:   token,
	})
}

// validateCredentials validates user credentials
func (h *AuthHandler) validateCredentials(email, password string) (*User, error) {
	var user User
	query := `SELECT id, name, email, password, created_at, updated_at, is_active FROM users WHERE email = $1 AND is_active = true`

	err := h.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}

	return &user, nil
}

// generateJWT generates a JWT token for the user
func (h *AuthHandler) generateJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 720).Unix(), // 30 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Get JWT secret from environment
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key" // Default secret - should be in .env
	}

	return token.SignedString([]byte(secret))
}

// recordLoginHistory records user login in history table
func (h *AuthHandler) recordLoginHistory(userID uuid.UUID) error {
	query := `INSERT INTO login_history (id, user_id, login_at) VALUES ($1, $2, $3)`
	_, err := h.db.Exec(query, uuid.New(), userID, time.Now())
	return err
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Basic email validation - you might want to use a more robust library
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
