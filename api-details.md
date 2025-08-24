# API Endpoints

## Authentication
- **POST** `/api/register`
- **POST** `/api/login`
- **POST** `/api/logout`

## Categories
- **GET** `/api/categories`
- **POST** `/api/categories`

## Expenses
- **POST** `/api/expenses`
- **GET** `/api/expenses`
- **PUT** `/api/expenses/{id}`
- **DELETE** `/api/expenses/{id}`

## Summary APIs
- **GET** `/api/expenses/summary/daily?page=1&limit=10`
- **GET** `/api/expenses/summary/monthly?page=1&limit=12`
- **GET** `/api/expenses/summary/weekly?month=2024-08&page=1&limit=10`

## Dashboard
- **GET** `/api/dashboard`

## Profile
- **GET** `/api/profile`
- **PUT** `/api/profile`
- **PUT** `/api/profile/password`