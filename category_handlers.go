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
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	categories, err := h.getAllCategories(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch categories"})
	}

	// Build response slice with explicit boolean is_default
	resp := make([]map[string]interface{}, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, map[string]interface{}{
			"id":         cat.ID,
			"name":       cat.Name,
			"is_default": cat.IsDefault,
			"created_at": cat.CreatedAt,
			"updated_at": cat.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "Categories retrieved successfully",
		"categories": resp,
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

// UpdateCategory updates a user-owned (non-default) category name
func (h *CategoryHandler) UpdateCategory(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	catID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid category ID"})
	}

	var req CreateCategoryRequest // reuse struct for name
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}
	if strings.TrimSpace(req.Name) == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Category name is required"})
	}

	var ownerID *uuid.UUID
	var isDefault bool
	query := `SELECT user_id, is_default FROM categories WHERE id = $1`
	if err := h.db.QueryRow(query, catID).Scan(&ownerID, &isDefault); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Category not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
	}
	if isDefault || ownerID == nil || *ownerID != userID {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "Cannot update this category"})
	}

	_, err = h.db.Exec(`UPDATE categories SET name = $1, updated_at = $2 WHERE id = $3`, req.Name, time.Now(), catID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update category"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Category updated successfully",
		"category_id": catID,
		"name":        req.Name,
	})
}

// DeleteCategory deletes a user-owned (non-default) category
func (h *CategoryHandler) DeleteCategory(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	catID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid category ID"})
	}

	var ownerID *uuid.UUID
	var isDefault bool
	query := `SELECT user_id, is_default FROM categories WHERE id = $1`
	if err := h.db.QueryRow(query, catID).Scan(&ownerID, &isDefault); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Category not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
	}
	if isDefault || ownerID == nil || *ownerID != userID {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "Cannot delete this category"})
	}

	_, err = h.db.Exec(`DELETE FROM categories WHERE id = $1`, catID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete category"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Category deleted successfully",
		"category_id": catID,
	})
}