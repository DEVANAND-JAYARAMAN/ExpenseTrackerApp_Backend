# Expense Tracker App

Go-based expense tracking application with PostgreSQL database.

## Setup

1. Install dependencies:
```bash
go mod tidy
```

2. Configure `.env` file with your PostgreSQL credentials.

3. Create database:
```sql
CREATE DATABASE expense_tracker;
```

4. Run application:
```bash
go run .
```

## Database Schema

Complete PostgreSQL schema with 6 tables:
- users
- categories  
- expenses
- expense_categories
- login_history
- sessions

All tables created automatically on startup.

## API Endpoints

- `POST /api/register` - User registration

## Testing

Use Postman with endpoints from POSTMAN_ENDPOINTS.md file.