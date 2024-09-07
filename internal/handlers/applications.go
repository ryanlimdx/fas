package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
	"github.com/google/uuid"
    "fas/internal/models"
	"fas/internal/utils"
)

// CreateApplication creates a new application in the database
func CreateApplication(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var application models.Application
        if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
            http.Error(w, "Invalid input", http.StatusBadRequest)
            return
        }

		// Begin transaction
		tx, err := db.Begin()
        if err != nil {
            http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
            return
        }
        defer tx.Rollback()

        // Check if an application already exists
        if applicationExists(db, application.ApplicantID, application.SchemeID) {
            http.Error(w, "Application already exists", http.StatusConflict)
            return
        }

        application.ID = uuid.New().String()
        application.Status = "Pending"
		application.AppliedDate = time.Now().Format("2006-01-02")

        // Insert the application
        _, err = tx.Exec(`INSERT INTO applications (id, applicant_id, scheme_id, status, applied_date) 
			VALUES (?, ?, ?, ?, ?)`, 
			application.ID, application.ApplicantID, application.SchemeID, application.Status, application.AppliedDate)
        if err != nil {
            utils.HandleInsertError(w, err, "application")
            return
        }

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Failed to commit", http.StatusInternalServerError)
			return
		}

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(application)
    }
}

// Checks if an application already exists with the same applicant and scheme IDs
func applicationExists(db *sql.DB, applicantID, schemeID string) bool {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM applications WHERE applicant_id = ? AND scheme_id = ?)`
    err := db.QueryRow(query, applicantID, schemeID).Scan(&exists)
    if err != nil {
        return false
    }
    return exists
}

// GetApplications retrieves all applications from the database.
func GetApplications(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rows, err := db.Query("SELECT id, applicant_id, scheme_id, status, applied_date FROM applications")
        if err != nil {
            http.Error(w, "Failed to retrieve applications", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var applications []models.Application
        for rows.Next() {
            var application models.Application
            if err := rows.Scan(&application.ID, &application.ApplicantID, &application.SchemeID, &application.Status, &application.AppliedDate); err != nil {
                http.Error(w, "Failed to scan application", http.StatusInternalServerError)
                return
            }
            applications = append(applications, application)
        }
        if err := rows.Err(); err != nil {
            http.Error(w, "Failed to read application data", http.StatusInternalServerError)
            return
        }

        json.NewEncoder(w).Encode(applications)
    }
}
