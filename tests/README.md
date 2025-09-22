# 🧪 Expense Tracker API Tests

Comprehensive test suite for the Expense Tracker API with unit and integration tests.

## 📁 Test Structure

```
tests/
├── unit/                 # Unit tests for individual components
│   ├── auth_test.go      # Authentication handler tests
│   ├── expense_test.go   # Expense handler tests
│   ├── category_test.go  # Category handler tests
│   └── profile_test.go   # Profile handler tests
├── integration/          # End-to-end API tests
│   └── api_test.go       # Full workflow integration tests
├── test_helpers.go       # Test utilities and mock structures
├── Makefile             # Test automation commands
└── README.md            # This file
```

## 🚀 Quick Start

### Prerequisites
```bash
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/suite
go get github.com/DATA-DOG/go-sqlmock
```

### Running Tests

**All tests:**
```bash
make test
```

**Unit tests only:**
```bash
make test-unit
```

**Integration tests only:**
```bash
make test-integration
```

**With coverage report:**
```bash
make test-coverage
```

## 📊 Test Coverage

### Unit Tests Cover:

#### Authentication (`auth_test.go`)
- ✅ User registration success
- ✅ User registration with existing email
- ✅ User login success
- ✅ User login with invalid credentials
- ✅ Request validation (name, email, password)

#### Expenses (`expense_test.go`)
- ✅ Create expense success
- ✅ Get all expenses
- ✅ Update expense success
- ✅ Delete expense success
- ✅ Dashboard data retrieval

#### Categories (`category_test.go`)
- ✅ Get all categories
- ✅ Create new category
- ✅ Create duplicate category (conflict)
- ✅ Invalid category name validation

#### Profile (`profile_test.go`)
- ✅ Get user profile
- ✅ Update profile information
- ✅ Change password success
- ✅ Change password with wrong current password

### Integration Tests Cover:

#### Full API Workflows (`api_test.go`)
- ✅ Complete user registration → login flow
- ✅ Category creation → expense creation → CRUD operations
- ✅ Dashboard data consistency
- ✅ Profile management workflow
- ✅ Session management and logout

## 🛠️ Test Features

### Mocking Strategy
- **Database Mocking**: Uses `go-sqlmock` for database interaction testing
- **HTTP Mocking**: Uses `httptest` for HTTP request/response testing
- **JWT Mocking**: Tests authentication flows with mock tokens

### Test Utilities
- **Helper Functions**: Common test data creation
- **Mock Structures**: Simplified handler interfaces for testing
- **Custom Matchers**: UUID and time matching for flexible assertions

### Assertions
- **HTTP Status Codes**: Validates correct response codes
- **Response Bodies**: Checks JSON response structure
- **Database Interactions**: Verifies SQL queries and parameters
- **Error Handling**: Tests error scenarios and edge cases

## 📈 Running Specific Tests

**Run specific test function:**
```bash
go test ./tests/unit/... -run TestAuthHandler_Register_Success
```

**Run tests with race detection:**
```bash
make test-verbose
```

**Generate coverage HTML report:**
```bash
make test-coverage
# Opens coverage.html in browser
```

## 🔧 Test Configuration

### Environment Variables for Integration Tests
```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=expense_tracker_test
DB_USER=postgres
DB_PASSWORD=password
JWT_SECRET=test-secret
```

### Mock Database Setup
Tests use `go-sqlmock` to simulate database interactions without requiring a real database connection.

## 📝 Adding New Tests

### Unit Test Template
```go
func TestHandler_Method_Scenario(t *testing.T) {
    // Setup
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()
    
    handler := &Handler{db: db}
    
    // Mock expectations
    mock.ExpectQuery("SELECT ...").WillReturnRows(...)
    
    // Execute
    err = handler.Method(context)
    
    // Assert
    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

### Integration Test Template
```go
func (suite *APITestSuite) TestNewFeature() {
    req := httptest.NewRequest(http.MethodGet, "/api/endpoint", nil)
    req.Header.Set("Authorization", "Bearer "+suite.token)
    rec := httptest.NewRecorder()
    
    suite.app.ServeHTTP(rec, req)
    assert.Equal(suite.T(), http.StatusOK, rec.Code)
}
```

## 🎯 Best Practices

1. **Test Naming**: Use descriptive names following `TestHandler_Method_Scenario` pattern
2. **Isolation**: Each test should be independent and not rely on others
3. **Mocking**: Mock external dependencies (database, HTTP calls)
4. **Coverage**: Aim for high test coverage but focus on critical paths
5. **Edge Cases**: Test both success and failure scenarios
6. **Performance**: Keep tests fast and efficient

## 🚨 Common Issues

**Mock expectations not met:**
- Ensure all mocked database calls are actually executed
- Check SQL query formatting matches exactly

**Integration test failures:**
- Verify test database is accessible
- Check environment variables are set correctly

**Race conditions:**
- Use `-race` flag to detect concurrent access issues
- Ensure proper synchronization in concurrent tests