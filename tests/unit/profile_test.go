package unit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// ProfileHandler struct for testing
type ProfileHandler struct {
	db *sql.DB
}

// Profile struct for testing
type Profile struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	ProfileImage *string    `json:"profile_image,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func TestProfileHandler_GetProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	profileHandler := &ProfileHandler{db: db}
	userID := uuid.New()

	// Mock profile query
	mock.ExpectQuery("SELECT id, name, email, profile_image, is_active, created_at, updated_at").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "profile_image", "is_active", "created_at", "updated_at"}).
			AddRow(userID, "John Doe", "john@example.com", nil, true, time.Now(), time.Now()))

	profile, err := profileHandler.getProfile(userID)
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "John Doe", profile.Name)
	assert.Equal(t, "john@example.com", profile.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	profileHandler := &ProfileHandler{db: db}
	userID := uuid.New()

	// Mock profile update
	mock.ExpectQuery("UPDATE users SET name").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "profile_image", "is_active", "created_at", "updated_at"}).
			AddRow(userID, "Jane Doe", "jane@example.com", nil, true, time.Now(), time.Now()))

	profile, err := profileHandler.updateProfile(userID, "Jane Doe", nil)
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "Jane Doe", profile.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_ChangePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	profileHandler := &ProfileHandler{db: db}
	userID := uuid.New()

	currentPassword := "oldpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

	// Mock current password check
	mock.ExpectQuery("SELECT password FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(string(hashedPassword)))

	// Mock password update
	mock.ExpectExec("UPDATE users SET password").WillReturnResult(sqlmock.NewResult(1, 1))

	err = profileHandler.changePassword(userID, currentPassword, "newpassword123")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_ChangePassword_WrongCurrent(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	profileHandler := &ProfileHandler{db: db}
	userID := uuid.New()

	currentPassword := "correctpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

	// Mock current password check
	mock.ExpectQuery("SELECT password FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(string(hashedPassword)))

	err = profileHandler.changePassword(userID, "wrongpassword", "newpassword123")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper functions for testing
func (h *ProfileHandler) getProfile(userID uuid.UUID) (*Profile, error) {
	query := `SELECT id, name, email, profile_image, is_active, created_at, updated_at FROM users WHERE id = $1`
	
	profile := &Profile{}
	err := h.db.QueryRow(query, userID).Scan(
		&profile.ID, &profile.Name, &profile.Email, &profile.ProfileImage, 
		&profile.IsActive, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (h *ProfileHandler) updateProfile(userID uuid.UUID, name string, profileImage *string) (*Profile, error) {
	query := `UPDATE users SET name = $2, profile_image = $3, updated_at = $4 WHERE id = $1 
			  RETURNING id, name, email, profile_image, is_active, created_at, updated_at`
	
	profile := &Profile{}
	err := h.db.QueryRow(query, userID, name, profileImage, time.Now()).Scan(
		&profile.ID, &profile.Name, &profile.Email, &profile.ProfileImage,
		&profile.IsActive, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (h *ProfileHandler) changePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	// Get current password hash
	var currentHash string
	query := `SELECT password FROM users WHERE id = $1`
	err := h.db.QueryRow(query, userID).Scan(&currentHash)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(currentPassword)); err != nil {
		return err
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	updateQuery := `UPDATE users SET password = $2, updated_at = $3 WHERE id = $1`
	_, err = h.db.Exec(updateQuery, userID, string(newHash), time.Now())
	return err
}