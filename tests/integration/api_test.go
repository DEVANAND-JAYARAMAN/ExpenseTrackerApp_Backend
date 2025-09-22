package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	app   *echo.Echo
	token string
}

func (suite *APITestSuite) SetupSuite() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "expense_tracker_test")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("JWT_SECRET", "test-secret")

	suite.app = echo.New()
	// Initialize your app routes here
	// setupRoutes(suite.app)
}

func (suite *APITestSuite) TestUserRegistrationAndLogin() {
	// Test Registration
	regReq := map[string]string{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "password123",
	}
	regBody, _ := json.Marshal(regReq)
	
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(regBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	// Test Login
	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	
	req = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(loginBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var loginResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &loginResp)
	suite.token = loginResp["token"].(string)
	assert.NotEmpty(suite.T(), suite.token)
}

func (suite *APITestSuite) TestExpenseWorkflow() {
	// Create Category
	catReq := map[string]string{
		"name": "Test Category",
	}
	catBody, _ := json.Marshal(catReq)
	
	req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(catBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec := httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var catResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &catResp)
	categoryID := catResp["category"].(map[string]interface{})["id"].(string)

	// Create Expense
	expReq := map[string]interface{}{
		"title":       "Test Expense",
		"amount":      25.50,
		"category_id": categoryID,
		"date":        "2024-01-15",
	}
	expBody, _ := json.Marshal(expReq)
	
	req = httptest.NewRequest(http.MethodPost, "/api/expenses", bytes.NewReader(expBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec = httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var expResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &expResp)
	expenseID := expResp["expense"].(map[string]interface{})["id"].(string)

	// Get Expenses
	req = httptest.NewRequest(http.MethodGet, "/api/expenses", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec = httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	// Update Expense
	updateReq := map[string]interface{}{
		"title":       "Updated Test Expense",
		"amount":      30.00,
		"category_id": categoryID,
		"date":        "2024-01-15",
	}
	updateBody, _ := json.Marshal(updateReq)
	
	req = httptest.NewRequest(http.MethodPut, "/api/expenses/"+expenseID, bytes.NewReader(updateBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec = httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	// Delete Expense
	req = httptest.NewRequest(http.MethodDelete, "/api/expenses/"+expenseID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec = httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *APITestSuite) TestDashboard() {
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec := httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var dashResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &dashResp)
	assert.Contains(suite.T(), dashResp, "total_spent")
	assert.Contains(suite.T(), dashResp, "month_spent")
	assert.Contains(suite.T(), dashResp, "week_spent")
	assert.Contains(suite.T(), dashResp, "today_spent")
}

func (suite *APITestSuite) TestProfile() {
	// Get Profile
	req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec := httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	// Update Profile
	updateReq := map[string]string{
		"name": "Updated Test User",
	}
	updateBody, _ := json.Marshal(updateReq)
	
	req = httptest.NewRequest(http.MethodPut, "/api/profile", bytes.NewReader(updateBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec = httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *APITestSuite) TestLogout() {
	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	rec := httptest.NewRecorder()
	
	suite.app.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}