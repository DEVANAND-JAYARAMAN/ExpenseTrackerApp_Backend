package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// ProfileHandler handles user profile-related requests
type ProfileHandler struct {
	db *sql.DB
}

// NewProfileHandler creates a new ProfileHandler instance
func NewProfileHandler(db *sql.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

// GetProfile handles getting user profile information
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	// Extract user ID from JWT token
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	// Fetch user profile from database
	profile, err := h.getUserProfile(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get profile: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Profile retrieved successfully",
		"profile": profile,
	})
}

// UpdateProfile handles updating user profile information
func (h *ProfileHandler) UpdateProfile(c echo.Context) error {
	// Extract user ID from JWT token
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate profile data
	if strings.TrimSpace(req.Name) == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Name is required",
		})
	}

	// Update user profile in database
	err := h.updateUserProfile(userID, req.Name, req.ProfileImage)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to update profile: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Profile updated successfully",
	})
}

// ChangePassword handles changing user password
func (h *ProfileHandler) ChangePassword(c echo.Context) error {
	// Extract user ID from JWT token
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate password requirements
	if len(req.NewPassword) < 8 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "New password must be at least 8 characters",
		})
	}

	// Verify current password
	valid, err := h.verifyCurrentPassword(userID, req.CurrentPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to verify current password",
		})
	}
	if !valid {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Current password is incorrect",
		})
	}

	// Hash new password and update
	err = h.updateUserPassword(userID, req.NewPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to update password",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Password changed successfully",
	})
}

// Helper functions for profile management

// getUserProfile retrieves user profile information
func (h *ProfileHandler) getUserProfile(userID uuid.UUID) (map[string]interface{}, error) {
	query := `SELECT id, name, email, profile_image, created_at, updated_at FROM users WHERE id = $1 AND is_active = true`
	
	var id uuid.UUID
	var name, email string
	var profileImage *string
	var createdAt, updatedAt time.Time
	
	err := h.db.QueryRow(query, userID).Scan(&id, &name, &email, &profileImage, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	profile := map[string]interface{}{
		"id":         id,
		"name":       name,
		"email":      email,
		"created_at": createdAt.Format("02-01-2006 03:04:05 PM"),
		"updated_at": updatedAt.Format("02-01-2006 03:04:05 PM"),
	}

	if profileImage != nil {
		profile["profile_image"] = *profileImage
	}

	return profile, nil
}

// updateUserProfile updates user profile information
func (h *ProfileHandler) updateUserProfile(userID uuid.UUID, name string, profileImage *string) error {
	query := `UPDATE users SET name = $2, profile_image = $3, updated_at = $4 WHERE id = $1`
	_, err := h.db.Exec(query, userID, name, profileImage, time.Now())
	return err
}

// verifyCurrentPassword checks if the provided current password is correct
func (h *ProfileHandler) verifyCurrentPassword(userID uuid.UUID, currentPassword string) (bool, error) {
	var hashedPassword string
	query := `SELECT password FROM users WHERE id = $1 AND is_active = true`
	
	err := h.db.QueryRow(query, userID).Scan(&hashedPassword)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(currentPassword))
	return err == nil, nil
}

// updateUserPassword updates user password with new hashed password
func (h *ProfileHandler) updateUserPassword(userID uuid.UUID, newPassword string) error {
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password in database
	query := `UPDATE users SET password = $2, updated_at = $3 WHERE id = $1`
	_, err = h.db.Exec(query, userID, string(hashedPassword), time.Now())
	return err
}