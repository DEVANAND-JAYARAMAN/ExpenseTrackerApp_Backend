## Expense Tracker API (Current Code Alignment)

This documentation reflects ONLY the endpoints actually implemented in the current Go code (`handlers.go`, `category_handlers.go`, `expense_handlers.go`, `main.go`). Removed any unused / stale sections (reactivate, login-history not implemented in handlers, etc.). Formats and field names match handler responses.

---

## Authentication

### Register

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

### Login

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

### Logout

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

## Categories

### Get Categories

GET /api/categories (Bearer)

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

- The response contains all categories from the database for the user, with `is_default` as a boolean (`true` or `false`).
  Errors: 401 Unauthorized

### Create Category

POST /api/categories

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
- 409 Category already exists

### Update Category

PUT /api/categories/:id

Request
{ "name": "string", "is_default": false }

Success 200
{ "message": "Category updated successfully", "category_id": "uuid", "name": "string", "is_default": false }

Errors

- 400 Invalid request body / ID
- 401 Unauthorized
- 403 Cannot update this category
- 404 Category not found

### Delete Category

DELETE /api/categories/:id

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

---

## Expenses

### Add Expense

POST /api/expenses (Bearer)

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

### Get Expenses

GET /api/expenses (Bearer)

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

### Update Expense

PUT /api/expenses/:id (Bearer)

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

### Delete Expense

DELETE /api/expenses/:id (Bearer)

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

## Error Format

All error responses:

```json
{
  "error": "Error message here"
}
```

---

## Notes / Gaps

- GetExpenses currently does not return categories array (only basic fields); create/update responses include categories list.
  -- Time format required: expense_time must be HH:MM AM/PM (12-hour, no seconds).
- UpdateExpense response leaves created_at blank (future improvement: fetch from DB).
- JWT expiration set to 30 days in code.
- Provide JWT_SECRET via environment variable in production.
