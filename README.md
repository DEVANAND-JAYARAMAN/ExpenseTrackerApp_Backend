# 💰 Expense Tracker App

A simple and powerful expense tracking application built with Go and PostgreSQL. This app helps you keep track of your daily expenses, categorize them, and manage your personal finances effectively.

## 🎯 What Does This App Do?

Imagine you want to know where your money goes every month. This app helps you:

- **Track Your Spending**: Record every expense you make (like buying coffee, groceries, or paying bills)
- **Organize by Categories**: Group your expenses (Food, Transportation, Entertainment, etc.)
- **Secure Access**: Only you can see your expenses with secure login
- **History Tracking**: See when and what you spent money on

Think of it as a digital notebook for your money, but smarter!

## 🚀 Quick Start Guide

### Prerequisites
- Go programming language (version 1.21 or higher)
- PostgreSQL database installed and running

### Step-by-Step Setup

1. **Install dependencies:**

   go mod tidy
  

2. **Set up your database:**
  
   CREATE DATABASE expense_tracker;
   

3. **Configure your environment:**
   The `.env` file contains your database settings. Update the password if needed:
   
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=expense_tracker
   DB_USER=postgres
   DB_PASSWORD=your_password_here
   PORT=3000
   JWT_SECRET=your_jwt_secret_key_here
   

4. **Start the application:**
  
   go run .
   

   You'll see: `Server starting on port 3000`

## 📊 Database Structure

The app automatically creates these tables when you first run it:

- **users** - Stores user account information
- **categories** - Different expense categories (Food, Transport, etc.)
- **expenses** - Your actual expense records
- **expense_categories** - Links expenses to categories
- **login_history** - Tracks when you log in
- **sessions** - Manages your login sessions

## 🔌 API Endpoints

### User Management

#### Create Account (Register)
- **Endpoint**: `POST /api/register`
- **What it does**: Creates a new user account
- **When to use**: First time using the app
- **Example**: Sign up with your name, email, and password

#### Login
- **Endpoint**: `POST /api/login`
- **What it does**: Logs you into your account and gives you a security token + session ID
- **When to use**: Every time you want to access your expenses
- **Example**: Enter your email and password to get access
- **Returns**: JWT token and session ID for tracking

#### Logout
- **Endpoint**: `POST /api/logout`
- **What it does**: Safely logs you out and deactivates your session
- **When to use**: When you're done using the app
- **Authentication**: Requires login token

### Category Management

#### Get Categories
- **Endpoint**: `GET /api/categories`
- **What it does**: Gets all available categories for the dropdown menu
- **When to use**: When you need to select a category for an expense
- **Example**: Shows Food, Transportation, Entertainment, Shopping, Bills
- **Authentication**: Requires login token

#### Create New Category
- **Endpoint**: `POST /api/categories`
- **What it does**: Creates a new custom category when existing ones don't fit
- **When to use**: When you need a category that doesn't exist (like "Fuel", "Medical", etc.)
- **Example**: Create "Fuel" category for gas expenses
- **Authentication**: Requires login token

### Expense Management

#### Add New Expense
- **Endpoint**: `POST /api/expenses`
- **What it does**: Creates a new expense record with title, amount, category, and date/time
- **When to use**: Every time you spend money and want to track it
- **Example**: Record buying groceries for $50 in the "Food" category
- **Authentication**: Requires login token

#### Update Expense
- **Endpoint**: `PUT /api/expenses/:id`
- **What it does**: Updates an existing expense with new information
- **When to use**: When you need to correct or modify an expense record
- **Example**: Change the amount from $50 to $45 if you remembered the exact price
- **Authentication**: Requires login token

#### Delete Expense
- **Endpoint**: `DELETE /api/expenses/:id`
- **What it does**: Permanently removes an expense record
- **When to use**: When you accidentally added a duplicate or wrong expense
- **Example**: Remove that coffee expense you added twice
- **Authentication**: Requires login token

#### View Expense List
- **Endpoint**: `GET /api/expenses`
- **What it does**: Returns all your expenses ordered by date (newest first)
- **When to use**: To see your complete expense history
- **Authentication**: Requires login token

#### Filter Expenses
- **Endpoint**: `GET /api/expenses?category_id=abc&start_date=01-08-2024&end_date=31-08-2024&min_amount=50&max_amount=200`
- **What it does**: Filters expenses by category, date range, and amount
- **When to use**: To analyze specific spending patterns
- **Parameters**: category_id, start_date, end_date, min_amount, max_amount (all optional)
- **Authentication**: Requires login token

#### Monthly Expense Summary
- **Endpoint**: `GET /api/expenses/summary/monthly`
- **What it does**: Returns monthly expense totals for chart display
- **When to use**: To visualize spending trends over time
- **Returns**: Array of {month: "Aug 2025", total: 1200.00}
- **Authentication**: Requires login token

## 🧪 Testing the App

### Using Postman (Recommended)

1. **Register a new user:**
   - Method: POST
   - URL: `http://localhost:3000/api/register`
   - Body:
     ```json
     {
       "name": "Your Name",
       "email": "your.email@example.com",
       "password": "YourPassword123"
     }
     ```

2. **Login with your account:**
   - Method: POST
   - URL: `http://localhost:3000/api/login`
   - Body:
     ```json
     {
       "email": "your.email@example.com",
       "password": "YourPassword123"
     }
     ```
   - **Save the token** from the response for expense operations!

3. **Add an expense:**
   - Method: POST
   - URL: `http://localhost:3000/api/expenses`
   - Headers: `Authorization: Bearer your_jwt_token`
   - Body:
     ```json
     {
       "title": "Groceries",
       "description": "Weekly grocery shopping",
       "amount": 75.50,
       "category_id": "your-category-uuid",
       "expense_date": "2024-01-15",
       "expense_time": "14:30"
     }
     ```

4. **Update an expense:**
   - Method: PUT
   - URL: `http://localhost:3000/api/expenses/expense-id`
   - Headers: `Authorization: Bearer your_jwt_token`
   - Body: Same as add expense

5. **Delete an expense:**
   - Method: DELETE
   - URL: `http://localhost:3000/api/expenses/expense-id`
   - Headers: `Authorization: Bearer your_jwt_token`

For detailed testing examples, check the `POSTMAN_ENDPOINTS.md` file.

## 🛠️ Technology Stack

- **Backend**: Go (Golang) with Echo framework
- **Database**: PostgreSQL
- **Authentication**: JWT (JSON Web Tokens)
- **Password Security**: bcrypt hashing

## 📝 Project Status

This is an active project. Currently implemented:
- ✅ User registration and login with JWT authentication
- ✅ Session management (create/deactivate sessions)
- ✅ Secure password handling with bcrypt
- ✅ Database setup with foreign key constraints
- ✅ Category management with dropdown support
- ✅ Add new expenses with category validation
- ✅ Update existing expenses with ownership checks
- ✅ Delete expenses with proper authorization
- ✅ View expense list ordered by date
- ✅ Filter expenses by category, date range, and amount
- ✅ Monthly expense summary for analytics
- ✅ JWT-based authentication for all protected routes
- ✅ Login history tracking

Recently completed:
- ✅ View expense history and lists
- ✅ Advanced expense filtering by category, date, and amount
- ✅ Monthly expense summary for charts and analytics

Coming next:
- 🔄 Category-wise expense breakdown
- 🔄 Export functionality
- 🔄 Expense budgeting features

## 🤝 Contributing

Feel free to contribute to this project! Whether it's bug fixes, new features, or documentation improvements, all contributions are welcome.

---

*Happy expense tracking! 💸*
