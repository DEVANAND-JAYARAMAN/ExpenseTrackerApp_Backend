# Postman API Endpoints

## User Registration

**URL:** `http://localhost:3000/api/register`  
**Method:** POST  
**Headers:** `Content-Type: application/json`

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Success Response (201):**
```json
{
  "message": "User registered successfully.",
  "user": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "is_active": true
  }
}
```

**Error Response (409):**
```json
{
  "error": "Email already exists"
}
```

## User Login

**URL:** `http://localhost:3000/api/login`  
**Method:** POST  
**Headers:** `Content-Type: application/json`
**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "Password123"
}
```

**Success Response (200):**
```json
{
  "message": "Login successful.",
  "token": "jwt_token"
}
```

**Error Response (401):**
```json
{
  "error": "Email or Password is Wrong"
}
```

## Add Expense

**URL:** `http://localhost:3000/api/expenses`  
**Method:** POST  
**Headers:** 
- `Content-Type: application/json`
- `Authorization: Bearer jwt_token`

**Request Body:**
```json
{
  "title": "Groceries",
  "description": "Monthly grocery shopping",
  "amount": 2200.50,
  "category_name": "Food",
  "expense_date": "2025-08-06",
  "expense_time": "19:30"
}
```

**Success Response (201):**
```json
{
  "message": "Expense added successfully",
  "expense_id": "5a37ef20-8cd0-4c7e-b5ec-90a31712d710"
}
```

**Error Response (400):**
```json
{
  "error": "Missing or invalid fields"
}
```

## Update Expense

**URL:** `http://localhost:3000/api/expenses/:id`  
**Method:** PUT  
**Headers:** 
- `Content-Type: application/json`
- `Authorization: Bearer jwt_token`

**Request Body:**
```json
{
  "title": "Groceries Updated",
  "description": "Updated grocery note",
  "amount": 2100,
  "category_name": "Food",
  "expense_date": "2025-08-06",
  "expense_time": "20:00"
}
```

**Success Response (200):**
```json
{
  "message": "Expense updated successfully"
}
```

**Error Response (404):**
```json
{
  "error": "Unauthorized or not found"
}
```

## Delete Expense

**URL:** `http://localhost:3000/api/expenses/:id`  
**Method:** DELETE  
**Headers:** 
- `Authorization: Bearer jwt_token`

**Success Response (200):**
```json
{
  "message": "Expense deleted successfully"
}
```

**Error Response (404):**
```json
{
  "error": "Expense not found or unauthorized"
}
```

## Get Categories (for dropdown)

**URL:** `http://localhost:3000/api/categories`  
**Method:** GET  
**Headers:** `Authorization: Bearer jwt_token`

**Success Response (200):**
```json
{
  "message": "Categories retrieved successfully",
  "categories": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Food",
      "is_default": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "123e4567-e89b-12d3-a456-426614174001",
      "name": "Transportation",
      "is_default": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

## Create New Category

**URL:** `http://localhost:3000/api/categories`  
**Method:** POST  
**Headers:** 
- `Content-Type: application/json`
- `Authorization: Bearer jwt_token`

**Request Body:**
```json
{
  "name": "Fuel"
}
```

**Success Response (201):**
```json
{
  "message": "Category created successfully",
  "category_id": "new-category-uuid"
}
```

**Error Response (409):**
```json
{
  "error": "Category already exists"
}
```

## User Logout

**URL:** `http://localhost:3000/api/logout`  
**Method:** POST  
**Headers:** `Authorization: Bearer jwt_token`

**Success Response (200):**
```json
{
  "message": "Logout successful"
}
```






## Get User Profile
