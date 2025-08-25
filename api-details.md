## Expense Tracker API (Current Code Alignment):

This documentation reflects ALL endpoints actually implemented in the current Go code (`handlers.go`, `category_handlers.go`, `expense_handlers.go`, `profile_handlers.go`, `main.go`). All formats and field names match handler responses.

---

## Authentication:

### Register:

POST /api/register

Request

```json
{
  "name": "string (2-255 chars)",
  "email": "string (valid email)",
  "password": "string (min 8 chars)"
}
```

Success 201

```json
{
  "message": "User registered successfully.",
  "user": {
    "id": "uuid",
    "name": "string",
    "email": "string",
    "profile_image": "string|null",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "is_active": true,
    "deactivated_at": "timestamp|null"
  }
}
```

Errors

- 400 Invalid request body / validation
- 409 Email already exists

### Login:

POST /api/login

Request

```json
{
  "email": "string (valid email)",
  "password": "string"
}
```

Success 200

```json
{
  "message": "Login successful.",
  "token": "jwt-token-string",
  "session_id": "uuid"
}
```

Errors

- 400 Invalid request body
- 401 Email or Password is Wrong

### Logout:

POST /api/logout (Bearer token required)

Success 200

```json
{
  "message": "Logout successful"
}
```

Errors

- 400 Authorization header required
- 401 Unauthorized

---

## Profile Management:

### Get Profile:

GET /api/profile (Bearer token required)

Success 200

```json
{
  "message": "Profile retrieved successfully",
  "profile": {
    "id": "uuid",
    "name": "string",
    "email": "string",
    "profile_image": "string|null",
    "is_active": true,
    "created_at": "DD-MM-YYYY HH:MM:SS AM/PM",
    "updated_at": "DD-MM-YYYY HH:MM:SS AM/PM"
  }
}
```

Errors

- 401 Unauthorized
- 500 Failed to get profile

### Update Profile:

PUT /api/profile (Bearer token required)

Request

```json
{
  "name": "string (required)",
  "profile_image": "string|null"
}
```

Success 200

```json
{
  "message": "Profile updated successfully",
  "profile": {
    "id": "uuid",
    "name": "string",
    "email": "string",
    "profile_image": "string|null",
    "is_active": true,
    "created_at": "DD-MM-YYYY HH:MM:SS AM/PM",
    "updated_at": "DD-MM-YYYY HH:MM:SS AM/PM"
  }
}
```

Errors

- 400 Invalid request body / Name is required
- 401 Unauthorized
- 500 Failed to update profile

### Change Password:

PUT /api/profile/password (Bearer token required)

Request

```json
{
  "current_password": "string (required)",
  "new_password": "string (min 8 chars, required)"
}
```

Success 200

```json
{
  "message": "Password changed successfully"
}
```

Errors

- 400 Invalid request body / Current password and new password are required / New password must be at least 8 characters
- 401 Unauthorized / Current password is incorrect
- 500 Failed to change password

---

## Categories:

### Get Categories:

GET /api/categories (Bearer token required)

Success 200

```json
{
  "categories": [
    {
      "id": "uuid",
      "name": "string",
      "is_default": true,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    },
    {
      "id": "uuid",
      "name": "string",
      "is_default": false,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ]
}
```

Errors

- 401 Unauthorized

### Create Category:

POST /api/categories (Bearer token required)

Request

```json
{
  "name": "string (2-255 chars)",
  "is_default": false
}
```

Success 201

```json
{
  "message": "Category created successfully.",
  "category_id": "uuid",
  "name": "string",
  "is_default": false
}
```

Errors

- 400 Invalid request body / validation
- 401 Unauthorized
- 409 Category already exists

### Update Category:

PUT /api/categories/:id (Bearer token required)

Request

```json
{
  "name": "string",
  "is_default": false
}
```

Success 200

```json
{
  "message": "Category updated successfully",
  "category_id": "uuid",
  "name": "string",
  "is_default": false
}
```

Errors

- 400 Invalid request body / ID
- 401 Unauthorized
- 403 Cannot update this category
- 404 Category not found

### Delete Category:

DELETE /api/categories/:id (Bearer token required)

Success 200

```json
{
  "message": "Category deleted successfully."
}
```

Errors

- 400 Invalid category ID
- 401 Unauthorized
- 403 Cannot delete this category
- 404 Category not found

---

## Expenses:

### Add Expense:

POST /api/expenses (Bearer token required)

Request

```json
{
  "title": "string",
  "description": "string|null",
  "amount": "number",
  "expense_date": "DD-MM-YYYY",
  "expense_time": "HH:MM AM/PM",
  "categories": ["uuid1", "uuid2", ...]
}
```

Success 201

```json
{
  "message": "Expense created successfully.",
  "expense": {
    "id": "uuid",
    "user_id": "uuid",
    "title": "string",
    "description": "string|null",
    "amount": "number",
    "expense_date": "DD-MM-YYYY",
    "expense_time": "HH:MM AM/PM",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "categories": [
      {
        "id": "uuid1",
        "name": "string",
        "is_default": false
      },
      {
        "id": "uuid2",
        "name": "string",
        "is_default": false
      }
    ]
  }
}
```

Errors

- 400 Missing or invalid fields / Invalid date or time format
- 401 Unauthorized

### Get Expenses:

GET /api/expenses (Bearer token required)

Query Parameters (all optional):

- `category_id`: Filter by category UUID
- `start_date`: Start date in DD-MM-YYYY format
- `end_date`: End date in DD-MM-YYYY format
- `min_amount`: Minimum amount filter
- `max_amount`: Maximum amount filter

Success 200

```json
{
  "expenses": [
    {
      "id": "uuid",
      "title": "string",
      "description": "string|null",
      "amount": "number",
      "expense_date": "DD-MM-YYYY",
      "expense_time": "HH:MM AM/PM",
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ]
}
```

Errors

- 401 Unauthorized
- 400 Invalid filter parameters

### Dashboard:

GET /api/dashboard (Bearer token required)

Success 200

```json
{
  "dashboard": {
    "total_expenses": 2500.75,
    "total_count": 156,
    "current_month": {
      "total": 450.5,
      "count": 28,
      "average_per_day": 15.02
    },
    "current_week": {
      "total": 125.75,
      "count": 8,
      "average_per_day": 17.96
    },
    "today": {
      "total": 25.5,
      "count": 2
    },
    "top_categories": [
      {
        "category_id": "uuid",
        "category_name": "Food",
        "total_amount": 850.25,
        "percentage": 34.01
      },
      {
        "category_id": "uuid",
        "category_name": "Transport",
        "total_amount": 420.5,
        "percentage": 16.82
      }
    ],
    "monthly_trend": [
      {
        "month": "2024-01",
        "total": 450.5,
        "count": 28
      },
      {
        "month": "2024-02",
        "total": 520.25,
        "count": 32
      }
    ],
    "recent_expenses": [
      {
        "id": "uuid",
        "title": "Coffee",
        "amount": 5.5,
        "expense_date": "15-01-2024",
        "expense_time": "09:30 AM",
        "categories": ["Beverages"]
      }
    ]
  }
}
```

Errors

- 401 Unauthorized
- 500 Failed to get dashboard data

### Update Expense:

PUT /api/expenses/:id (Bearer token required)

Request

```json
{
  "title": "string",
  "description": "string|null",
  "amount": "number",
  "expense_date": "DD-MM-YYYY",
  "expense_time": "HH:MM AM/PM",
  "categories": ["uuid1", "uuid2", ...]
}
```

Success 200

```json
{
  "message": "Expense updated successfully.",
  "expense": {
    "id": "uuid",
    "user_id": "uuid",
    "title": "string",
    "description": "string|null",
    "amount": "number",
    "expense_date": "DD-MM-YYYY",
    "expense_time": "HH:MM AM/PM",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "categories": [
      {
        "id": "uuid1",
        "name": "string",
        "is_default": false
      },
      {
        "id": "uuid2",
        "name": "string",
        "is_default": false
      }
    ]
  }
}
```

Errors

- 400 Missing or invalid fields / Invalid date or time format / Invalid expense ID
- 401 Unauthorized
- 404 Expense <id> not found for user <user_id>

### Delete Expense:

DELETE /api/expenses/:id (Bearer token required)

Success 200

```json
{
  "message": "Expense deleted successfully."
}
```

Errors

- 400 Invalid expense ID
- 401 Unauthorized
- 404 Expense <id> not found for user <user_id>

---

## Error Format:

All error responses:

```json
{
  "error": "Error message here"
}
```

---

## Additional Info:

- **JWT Token**: Set to 30-day expiration in code
- **Time Format**: expense_time must be HH:MM AM/PM (12-hour format, no seconds)
- **Date Format**: expense_date must be DD-MM-YYYY
- **Profile Timestamps**: Formatted as DD-MM-YYYY HH:MM:SS AM/PM
- **Authentication**: All protected routes require Bearer token in Authorization header
- **Filtering**: GetExpenses supports filtering by category, date range, and amount range
- **Dashboard**: Provides comprehensive analytics with multiple time breakdowns
- **Profile Management**: Complete CRUD operations for user profile and password changes
- **Environment**: Provide JWT_SECRET via environment variable in production

---

## Note:

**Monthly Expense Summary endpoint** (`GET /api/expenses/summary/monthly`) mentioned in README.md is **NOT YET IMPLEMENTED** in the current codebase. This endpoint has been removed from this documentation to reflect only the actually available APIs.
