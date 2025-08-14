package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// initDB initializes the database and creates all necessary tables
func initDB() (*sql.DB, error) {
	// Database connection parameters
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "expense_tracker"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}

	// Connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Try to connect to the database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		// If database doesn't exist, try to create it
		if err := createDatabaseIfNotExists(host, port, user, password, dbname); err != nil {
			return nil, fmt.Errorf("failed to create database: %v", err)
		}
		
		// Try connecting again
		db.Close()
		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to open database after creation: %v", err)
		}
		
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %v", err)
		}
	}

	// Create all tables
	if err := createAllTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database connected and tables created successfully")
	return db, nil
}

// createAllTables creates all tables required for the application
func createAllTables(db *sql.DB) error {
	query := `
	-- Enable pgcrypto for gen_random_uuid if not already enabled
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	-- USERS TABLE
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(999) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password TEXT NOT NULL,
		profile_image TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT TRUE,
		deactivated_at TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

	-- CATEGORIES TABLE
	CREATE TABLE IF NOT EXISTS categories (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		is_default BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- EXPENSES TABLE (legacy category_id/name removed; multi-category via expense_categories)
	CREATE TABLE IF NOT EXISTS expenses (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		amount DECIMAL(10, 2) NOT NULL,
		expense_date DATE NOT NULL,
		expense_time TIME NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- EXPENSE_CATEGORIES TABLE
	CREATE TABLE IF NOT EXISTS expense_categories (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		expense_id UUID REFERENCES expenses(id) ON DELETE CASCADE,
		category_id UUID REFERENCES categories(id) ON DELETE CASCADE
	);

	-- LOGIN_HISTORY TABLE
	CREATE TABLE IF NOT EXISTS login_history (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		login_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- SESSIONS TABLE
	CREATE TABLE IF NOT EXISTS sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		token TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP,
		is_active BOOLEAN DEFAULT TRUE
	);
	`
	_, err := db.Exec(query)
	return err
}

// createDatabaseIfNotExists creates the database if it doesn't exist
func createDatabaseIfNotExists(host, port, user, password, dbname string) error {
	// Connect to postgres database to create our target database
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)
	
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer db.Close()
	
	if err := db.Ping(); err != nil {
		return err
	}
	
	// Create database
	query := fmt.Sprintf("CREATE DATABASE %s", dbname)
	_, err = db.Exec(query)
	if err != nil {
		// Ignore error if database already exists
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	
	log.Printf("Database %s created successfully", dbname)
	return nil
}
