// Handles all the requests related to the schemes
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
	"fas/internal/utils"
	"fas/internal/models"
)

// CreateScheme creates a new scheme in the database.
func CreateScheme(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var scheme models.Scheme
        if err := json.NewDecoder(r.Body).Decode(&scheme); err != nil {
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

        // Insert the scheme
		scheme.ID = uuid.New().String()
        _, err = tx.Exec(`INSERT INTO schemes (id, name) VALUES (?, ?)`,
            scheme.ID, scheme.Name)
        if err != nil {
            utils.HandleInsertError(w, err, "scheme")
            return
        }

        // Criteria
        for i := range scheme.Criteria {
            criteria := &scheme.Criteria[i]
            // Check if the criteria already exists
            err = db.QueryRow(`SELECT id FROM criteria WHERE criteria_level = ? AND criteria_type = ? AND status = ?`,
                criteria.CriteriaLevel, criteria.CriteriaType, criteria.Status).Scan(&criteria.ID)
            if err == sql.ErrNoRows {
                // Criteria doesn't exist  (insert it)
                criteria.ID = uuid.New().String()
                _, err = tx.Exec(`INSERT INTO criteria (id, criteria_level, criteria_type, status) 
                    VALUES (?, ?, ?, ?)`,
                    criteria.ID, criteria.CriteriaLevel, criteria.CriteriaType, criteria.Status)
                if err != nil {
                    utils.HandleInsertError(w, err, "criteria")
                    return
                }
            } else if err != nil {
                http.Error(w, "Failed to check criteria", http.StatusInternalServerError)
                return
            }

            // Link criteria to the scheme
            _, err = tx.Exec(`INSERT INTO scheme_criteria (scheme_id, criteria_id) VALUES (?, ?)`,
                scheme.ID, criteria.ID)
            if err != nil {
                http.Error(w, "Failed to link criteria to scheme", http.StatusInternalServerError)
                return
            }
        }

        // Benefits
        for i := range scheme.Benefits {
            benefit := &scheme.Benefits[i]
            // Check if the benefit already exists
            err = db.QueryRow(`SELECT id FROM benefits WHERE name = ? AND amount = ?`,
                benefit.Name, benefit.Amount).Scan(&benefit.ID)
            if err == sql.ErrNoRows {
                // Benefit doesn't exist (insert it)
                benefit.ID = uuid.New().String()
                _, err = tx.Exec(`INSERT INTO benefits (id, name, amount) 
                    VALUES (?, ?, ?)`,
                    benefit.ID, benefit.Name, benefit.Amount)
                if err != nil {
					utils.HandleInsertError(w, err, "benefit")
                    return
                }
            } else if err != nil {
                http.Error(w, "Failed to check benefit", http.StatusInternalServerError)
                return
            }

            // Link benefit to the scheme
            _, err = tx.Exec(`INSERT INTO scheme_benefits (scheme_id, benefit_id) VALUES (?, ?)`,
                scheme.ID, benefit.ID)
            if err != nil {
                http.Error(w, "Failed to link benefit to scheme", http.StatusInternalServerError)
                return
            }
        }

        // Commit the transaction
        if err := tx.Commit(); err != nil {
            http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(scheme)
    }
}