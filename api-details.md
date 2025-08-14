# API Details

This document describes all API routes, their payloads, and responses for the Expense Tracker API.

---

## 1. User Registration

**Endpoint:** `POST /api/register`

**Payload:**

```json
{
  "name": "string (2-255 chars)",
  "email": "string (valid email)",
  "password": "string (min 8 chars)"
}
```

**Success Response:**

- **Status:** 201 Created

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

**Error Responses:**

- **Status:** 400 Bad Request

```json
{
  "error": "Invalid request body"
}
```

- **Status:** 409 Conflict

```json
{
  "error": "Email already exists"
}
```

---

## 2. User Login

**Endpoint:** `POST /api/login`

**Payload:**

```json
{
  "email": "string (valid email)",
  "password": "string"
}
```

**Success Response:**

- **Status:** 200 OK

```json
{
  "message": "Login successful.",
  "token": "jwt-token-string"
}
```

**Error Responses:**

- **Status:** 401 Unauthorized

```json
{
  "error": "Invalid email or password"
}
```

---

## 3. Get User Profile

**Endpoint:** `GET /api/profile`

**Headers:**

- `Authorization: Bearer <token>`

**Success Response:**

- **Status:** 200 OK

```json
{
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

**Error Responses:**

- **Status:** 401 Unauthorized

```json
{
  "error": "Invalid or expired token"
}
```

---

## 4. Create Expense

**Endpoint:** `POST /api/expenses`

**Headers:**

- `Authorization: Bearer <token>`

**Payload:**

```json
{
  "title": "string",
  "description": "string|null",
  "amount": "number",
  "expense_date": "YYYY-MM-DD",
  "expense_time": "HH:MM:SS",
  "categories": ["uuid", ...]
}
```

**Success Response:**

- **Status:** 201 Created

```json
{
  "message": "Expense created successfully.",
  "expense": {
    "id": "uuid",
    "user_id": "uuid",
    "title": "string",
    "description": "string|null",
    "amount": "number",
    "expense_date": "YYYY-MM-DD",
    "expense_time": "HH:MM:SS",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "categories": [
      {
        "id": "uuid",
        "name": "string"
      }
    ]
  }
}
```

**Error Responses:**

- **Status:** 400 Bad Request

```json
{
  "error": "Invalid expense data"
}
```

---

## 5. Get Expenses

**Endpoint:** `GET /api/expenses`

**Headers:**

- `Authorization: Bearer <token>`

**Success Response:**

- **Status:** 200 OK

```json
{
  "expenses": [
    {
      "id": "uuid",
      "title": "string",
      "amount": "number",
      "expense_date": "YYYY-MM-DD",
      "expense_time": "HH:MM:SS",
      "categories": [
        {
          "id": "uuid",
          "name": "string"
        }
      ]
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

## 6. Create Category

**Endpoint:** `POST /api/categories`

**Headers:**

- `Authorization: Bearer <token>`

**Payload:**

```json
{
  "name": "string"
}
```

**Success Response:**

- **Status:** 201 Created

```json
{
  "message": "Category created successfully.",
  "category": {
    "id": "uuid",
    "name": "string",
    "is_default": false,
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

**Error Responses:**

- **Status:** 400 Bad Request

```json
{
  "error": "Invalid category data"
}
```

---

## 7. Get Categories

**Endpoint:** `GET /api/categories`

**Headers:**

- `Authorization: Bearer <token>`

**Success Response:**

- **Status:** 200 OK

```json
{
  "categories": [
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

**Error Responses:**

- **Status:** 401 Unauthorized

```json
{
  "error": "Invalid or expired token"
}
```

---

## 8. Delete Expense

**Endpoint:** `DELETE /api/expenses/{id}`

**Headers:**

- `Authorization: Bearer <token>`

**Success Response:**

- **Status:** 200 OK

```json
{
  "message": "Expense deleted successfully."
}
```

**Error Responses:**

- **Status:** 404 Not Found

```json
{
  "error": "Expense not found"
}
```

---

## 9. Delete Category

**Endpoint:** `DELETE /api/categories/{id}`

**Headers:**

- `Authorization: Bearer <token>`

**Success Response:**

- **Status:** 200 OK

```json
{
  "message": "Category deleted successfully."
}
```

**Error Responses:**

- **Status:** 404 Not Found

```json
{
  "error": "Category not found"
}
```

---

## 10. Update User Profile

**Endpoint:** `PUT /api/profile`

**Headers:**

- `Authorization: Bearer <token>`

**Payload:**

```json
{
  "name": "string",
  "profile_image": "string|null"
}
```

**Success Response:**

- **Status:** 200 OK

```json
{
  "message": "Profile updated successfully.",
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

**Error Responses:**

- **Status:** 400 Bad Request

```json
{
  "error": "Invalid profile data"
}
```

---

## 11. Deactivate User

**Endpoint:** `POST /api/deactivate`

**Headers:**

- `Authorization: Bearer <token>`

**Success Response:**

- **Status:** 200 OK

```json
{
  "message": "User deactivated successfully."
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

## 12. Reactivate User

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
