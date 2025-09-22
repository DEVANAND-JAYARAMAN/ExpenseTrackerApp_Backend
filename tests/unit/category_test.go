package unit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// CategoryHandler struct for testing
type CategoryHandler struct {
	db *sql.DB
}

// Category struct for testing
type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	UserID    uuid.UUID `json:"user_id"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func TestCategoryHandler_GetCategories(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	categoryHandler := &CategoryHandler{db: db}
	userID := uuid.New()

	// Mock categories query
	rows := sqlmock.NewRows([]string{"id", "name", "user_id", "is_default", "created_at", "updated_at"}).
		AddRow(uuid.New(), "Food", userID, true, time.Now(), time.Now()).
		AddRow(uuid.New(), "Transportation", userID, true, time.Now(), time.Now())

	mock.ExpectQuery("SELECT.*FROM categories").WillReturnRows(rows)

	categories, err := categoryHandler.getCategories(userID)
	assert.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "Food", categories[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryHandler_CategoryExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	categoryHandler := &CategoryHandler{db: db}
	userID := uuid.New()

	// Mock category exists check
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := categoryHandler.categoryExists("Food", userID)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryHandler_CreateCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	categoryHandler := &CategoryHandler{db: db}
	userID := uuid.New()
	categoryID := uuid.New()

	// Mock category creation
	mock.ExpectQuery("INSERT INTO categories").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "user_id", "is_default", "created_at", "updated_at"}).
			AddRow(categoryID, "Custom Category", userID, false, time.Now(), time.Now()))

	category, err := categoryHandler.createCategory(categoryID, "Custom Category", userID, false)
	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, "Custom Category", category.Name)
	assert.False(t, category.IsDefault)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	categoryHandler := &CategoryHandler{db: db}
	categoryID := uuid.New()

	// Mock category deletion
	mock.ExpectExec("DELETE FROM categories").WillReturnResult(sqlmock.NewResult(1, 1))

	err = categoryHandler.deleteCategory(categoryID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper functions for testing
func (h *CategoryHandler) getCategories(userID uuid.UUID) ([]Category, error) {
	query := `SELECT id, name, user_id, is_default, created_at, updated_at FROM categories WHERE user_id = $1 OR is_default = true ORDER BY name`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.UserID, &cat.IsDefault, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

func (h *CategoryHandler) categoryExists(name string, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND (user_id = $2 OR is_default = true))`
	err := h.db.QueryRow(query, name, userID).Scan(&exists)
	return exists, err
}

func (h *CategoryHandler) createCategory(id uuid.UUID, name string, userID uuid.UUID, isDefault bool) (*Category, error) {
	query := `INSERT INTO categories (id, name, user_id, is_default) VALUES ($1, $2, $3, $4) RETURNING id, name, user_id, is_default, created_at, updated_at`

	category := &Category{}
	err := h.db.QueryRow(query, id, name, userID, isDefault).Scan(
		&category.ID, &category.Name, &category.UserID, &category.IsDefault, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (h *CategoryHandler) deleteCategory(id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := h.db.Exec(query, id)
	return err
}
