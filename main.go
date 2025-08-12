package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize handlers
	authHandler := NewAuthHandler(db)
	expenseHandler := NewExpenseHandler(db)

	// Routes
	api := e.Group("/api")
	
	// Public routes
	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)
	
	// Protected routes
	protected := api.Group("", JWTMiddleware())
	protected.POST("/expenses", expenseHandler.AddExpense)
	protected.GET("/expenses", expenseHandler.GetExpenses)
	protected.PUT("/expenses/:id", expenseHandler.UpdateExpense)
	protected.DELETE("/expenses/:id", expenseHandler.DeleteExpense)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
