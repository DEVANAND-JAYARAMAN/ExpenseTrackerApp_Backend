# Expense Tracker API Documentation

This document describes all API routes, payloads, and responses for the Expense Tracker API, based on the latest codebase.

---

## Authentication

### Register

- **Endpoint:** `POST /api/register`
- **Payload:**

```json
{
  "name": "string (2-255 chars)",
  "email": "string (valid email)",
  "password": "string (min 8 chars)"
}
```

- **Success Response:**

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

- **Error Responses:**
  - 400: `{ "error": "Invalid request body" }`
  - 409: `{ "error": "Email already exists" }`

### Login

- **Endpoint:** `POST /api/login`
- **Payload:**

```json
{
  "email": "string (valid email)",
  "password": "string"
}
```

- **Success Response:**

```json
{
  "message": "Login successful.",
  "token": "jwt-token-string",
  "session_id": "uuid"
}
```

- **Error Responses:**
  - 400: `{ "error": "Invalid request body" }`
  - 401: `{ "error": "Email or Password is Wrong" }`

### Logout

- **Endpoint:** `POST /api/logout`
- **Headers:** `Authorization: Bearer <token>`
- **Success Response:**

```json
{
  "message": "Logout successful"
}
```

- **Error Responses:**
  - 401: `{ "error": "Unauthorized" }`
  - 400: `{ "error": "Authorization header required" }`

---

## Categories

### Get Categories

- **Endpoint:** `GET /api/categories`
- **Headers:** `Authorization: Bearer <token>`
- **Success Response:**

```json
{
  "categories": [
    {
      "id": "uuid",
      "name": "string",
      "is_default": true|false,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ]
}
```

### Create Category

- **Endpoint:** `POST /api/categories`
- **Headers:** `Authorization: Bearer <token>`
- **Payload:**

```json
{
  "name": "string (2-255 chars)"
}
```

- **Success Response:**

```json
{
  "message": "Category created successfully.",
  "category_id": "uuid"
}
```

- **Error Responses:**
  - 400: `{ "error": "Invalid category data" }`

---

## Expenses

### Add Expense

- **Endpoint:** `POST /api/expenses`
- **Headers:** `Authorization: Bearer <token>`
- **Payload:**

```json
{
  "title": "string",
  "description": "string|null",
  "amount": "number (>0)",
  "category_id": "integer",
  "expense_date": "YYYY-MM-DD",
  "expense_time": "HH:MM:SS"
}
```

- **Success Response:**

```json
{
  "message": "Expense added successfully.",
  "expense_id": "uuid"
}
```

- **Error Responses:**
  - 400: `{ "error": "Invalid expense data" }`

### Get Expenses

- **Endpoint:** `GET /api/expenses`
- **Headers:** `Authorization: Bearer <token>`
- **Success Response:**

```json
{
  "expenses": [
    {
      "id": "uuid",
      "title": "string",
      "description": "string|null",
      "amount": "number",
      "category_id": "integer",
      "category_name": "string",
      "expense_date": "YYYY-MM-DD",
      "expense_time": "HH:MM:SS",
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ]
}
```

### Update Expense

- **Endpoint:** `PUT /api/expenses/:id`
- **Headers:** `Authorization: Bearer <token>`
- **Payload:**

```json
{
  "title": "string",
  "description": "string|null",
  "amount": "number (>0)",
  "category_id": "integer",
  "expense_date": "YYYY-MM-DD",
  "expense_time": "HH:MM:SS"
}
```

- **Success Response:**

```json
{
  "message": "Expense updated successfully."
}
```

- **Error Responses:**
  - 400: `{ "error": "Invalid expense data" }`
  - 404: `{ "error": "Expense not found" }`

### Delete Expense

- **Endpoint:** `DELETE /api/expenses/:id`
- **Headers:** `Authorization: Bearer <token>`
- **Success Response:**

```json
{
  "message": "Expense deleted successfully."
}
```

- **Error Responses:**
  - 404: `{ "error": "Expense not found" }`

---

## Error Response Format

All error responses follow this format:

```json
{
  "error": "Error message here"
}
```

---

## Notes

- All protected routes require a valid JWT token in the `Authorization` header.
- Timestamps are in ISO 8601 format.
- All IDs are UUIDs unless otherwise specified.
- Default categories are created automatically for new users.
  **Endpoint:** `POST /api/reactivate`

**Headers:**

- `Authorization: Bearer <token>`

**Success Response:**

- **Status:** 200 OK

```json
{
  "message": "User reactivated successfully."
}
```

**Error Responses:**

- **Status:** 404 Not Found

```json
{
  "error": "User not found"
}
```

---

## 13. Get Login History

**Endpoint:** `GET /api/login-history`

**Headers:**

- `Authorization: Bearer <token>`

**Success Response:**

- **Status:** 200 OK

```json
{
  "login_history": [
    {
      "id": "uuid",
      "login_at": "timestamp"
    }
  ]
}
```

**Error Responses:**

- **Status:** 401 Unauthorized

```json
{
  "error": "Invalid or expired token"
}
```

---

## Error Response Format

All error responses follow this format:

```json
{
  "error": "Error message here"
}
```
