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






## Get User Profile
