// Database handles the setup and connection to the database
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

// SetupDB connects to the MySQL database, create the relevant tables and returns the database.
func SetupDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get the DSN from the environment variable
	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("DSN not set in environment.")
	}

	// Open a connection to the MySQL database.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MySQL.")

	// Verify the connection to the database.
	if err = db.Ping(); err != nil {
		return nil, err
	}

	createTables(db)

	fmt.Println("Database setup complete.")
	return db, nil
}

// createTables executes the SQL commands to create the necessary tables.
func createTables(db *sql.DB) {
	queries := []string{
		// Applicants table
		`CREATE TABLE IF NOT EXISTS applicants (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100),
			employment_status VARCHAR(50),
			sex VARCHAR(10),
			date_of_birth DATE
		);`,

		// Households table
		`CREATE TABLE IF NOT EXISTS household (
			id VARCHAR(36) PRIMARY KEY,
			applicant_id VARCHAR(36),
			name VARCHAR(100),
			relationship VARCHAR(50),
			sex VARCHAR(10),
			school_level VARCHAR(50),
			employment_status VARCHAR(50),
			date_of_birth DATE,
			FOREIGN KEY (applicant_id) REFERENCES applicants(id) ON DELETE CASCADE
		);`,

		// Schemes table
		`CREATE TABLE IF NOT EXISTS schemes (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100)
		);`,

		// Criteria table
		`CREATE TABLE IF NOT EXISTS criteria (
			id VARCHAR(36) PRIMARY KEY,
			criteria_level VARCHAR(50),
			criteria_type VARCHAR(100),
			status VARCHAR(50)
		);`,

		// Scheme_Criteria table
		`CREATE TABLE IF NOT EXISTS scheme_criteria (
			scheme_id VARCHAR(36),
			criteria_id VARCHAR(36),
			PRIMARY KEY (scheme_id, criteria_id),
			FOREIGN KEY (scheme_id) REFERENCES schemes(id) ON DELETE CASCADE,
			FOREIGN KEY (criteria_id) REFERENCES criteria(id) ON DELETE CASCADE
		);`,

		// Benefits table
		`CREATE TABLE IF NOT EXISTS benefits (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100),
			amount DECIMAL(10, 2)
		);`,

		// Scheme_Benefits table
		`CREATE TABLE IF NOT EXISTS scheme_benefits (
			scheme_id VARCHAR(36),
			benefit_id VARCHAR(36),
			PRIMARY KEY (scheme_id, benefit_id),
			FOREIGN KEY (scheme_id) REFERENCES schemes(id) ON DELETE CASCADE,
			FOREIGN KEY (benefit_id) REFERENCES benefits(id) ON DELETE CASCADE
		);`,

		// Applications table
		`CREATE TABLE IF NOT EXISTS applications (
			id VARCHAR(36) PRIMARY KEY,
			applicant_id VARCHAR(36),
			scheme_id VARCHAR(36),
			status VARCHAR(50),
			FOREIGN KEY (applicant_id) REFERENCES applicants(id) ON DELETE CASCADE,
			FOREIGN KEY (scheme_id) REFERENCES schemes(id) ON DELETE CASCADE
		);`,
	}

	// Execute each query to create the tables.
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Fatalf("error creating table: %v", err)
		}
	}
}