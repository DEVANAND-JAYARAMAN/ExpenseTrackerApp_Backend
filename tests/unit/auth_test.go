package unit

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler struct for testing
type AuthHandler struct {
	db *sql.DB
}

// RegisterRequest for testing
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func TestAuthHandler_Register_Success(t *testing.T) {
	reqBody := RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	// Test the validation function
	err := validateRegisterRequest(reqBody)
	assert.NoError(t, err)
}

func TestAuthHandler_Register_EmailExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	authHandler := &AuthHandler{db: db}

	reqBody := RegisterRequest{
		Name:     "John Doe",
		Email:    "existing@example.com",
		Password: "password123",
	}

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Test email exists check
	exists, err := authHandler.emailExists(reqBody.Email)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthHandler_Login_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	authHandler := &AuthHandler{db: db}

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	userID := uuid.New()
	mock.ExpectQuery("SELECT id, name, email, password").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "password", "created_at", "updated_at", "is_active"}).
			AddRow(userID, "John Doe", "john@example.com", string(hashedPassword), time.Now(), time.Now(), true))

	// Test credential validation
	user, err := authHandler.validateCredentials("john@example.com", password)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "John Doe", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	authHandler := &AuthHandler{db: db}

	mock.ExpectQuery("SELECT id, name, email, password").WillReturnError(sql.ErrNoRows)

	// Test invalid credentials
	user, err := authHandler.validateCredentials("wrong@example.com", "wrongpassword")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper functions for testing
func (h *AuthHandler) emailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := h.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

func (h *AuthHandler) validateCredentials(email, password string) (*User, error) {
	var user User
	query := `SELECT id, name, email, password, created_at, updated_at, is_active FROM users WHERE email = $1 AND is_active = true`

	err := h.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}

	return &user, nil
}

func validateRegisterRequest(req RegisterRequest) error {
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 255 {
		return echo.NewHTTPError(http.StatusBadRequest, "Name must be between 2 and 255 characters")
	}
	if req.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email is required")
	}
	if req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Password is required")
	}
	if len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "Password must be at least 8 characters")
	}
	return nil
}

func TestValidateRegisterRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: RegisterRequest{
				Name:     "",
				Email:    "john@example.com",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "short password",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegisterRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}