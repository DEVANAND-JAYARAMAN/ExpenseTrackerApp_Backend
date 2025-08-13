package main

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CategoryHandler handles category-related requests
type CategoryHandler struct {
	db *sql.DB
}

// NewCategoryHandler creates a new CategoryHandler instance
func NewCategoryHandler(db *sql.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// GetCategories handles getting all available categories for dropdown
func (h *CategoryHandler) GetCategories(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	categories, err := h.getAllCategories(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get categories",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "Categories retrieved successfully",
		"categories": categories,
	})
}

// getAllCategories gets all categories (default + user-specific)
func (h *CategoryHandler) getAllCategories(userID uuid.UUID) ([]Category, error) {
	query := `
		SELECT id, name, user_id, is_default, created_at, updated_at 
		FROM categories 
		WHERE is_default = true OR user_id = $1 
		ORDER BY is_default DESC, name ASC
	`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		var userIDPtr *uuid.UUID

		err := rows.Scan(
			&category.ID, &category.Name, &userIDPtr, &category.IsDefault,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle NULL user_id for default categories
		if userIDPtr != nil {
			category.UserID = *userIDPtr
		}

		categories = append(categories, category)
	}

	return categories, nil
}

// CreateCategory handles creating a new custom category
func (h *CategoryHandler) CreateCategory(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate category name
	if strings.TrimSpace(req.Name) == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Category name is required",
		})
	}

	// Check if category already exists for this user
	if h.categoryExists(userID, req.Name) {
		return c.JSON(http.StatusConflict, ErrorResponse{
			Error: "Category already exists",
		})
	}

	// Create new category
	categoryID := uuid.New()
	err := h.createCategory(categoryID, userID, req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create category",
		})
	}

	return c.JSON(http.StatusCreated, CreateCategoryResponse{
		Message:    "Category created successfully",
		CategoryID: categoryID,
	})
}

// categoryExists checks if a category name already exists for the user
func (h *CategoryHandler) categoryExists(userID uuid.UUID, name string) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE LOWER(name) = LOWER($1) AND (user_id = $2 OR is_default = true))`
	err := h.db.QueryRow(query, name, userID).Scan(&exists)
	return err == nil && exists
}

// createCategory creates a new user-specific category
func (h *CategoryHandler) createCategory(id, userID uuid.UUID, name string) error {
	query := `INSERT INTO categories (id, name, user_id, is_default, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	now := time.Now()
	_, err := h.db.Exec(query, id, name, userID, false, now, now)
	return err
}