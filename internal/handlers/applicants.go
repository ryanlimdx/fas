// Handles all the requests related to applicants
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/google/uuid"
	"fas/internal/utils"
	"fas/internal/models"
)

// CreateApplicant creates a new applicant in the database from the JSON input.
func CreateApplicant(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var applicant models.Applicant
		if err := json.NewDecoder(r.Body).Decode(&applicant); err != nil {
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

		// Insert the applicant
		fmt.Println("Inserting applicant:", applicant)
		applicant.ID = uuid.New().String()
		_, err = tx.Exec(`INSERT INTO applicants (id, name, employment_status, marital_status, sex, date_of_birth) 
			VALUES (?, ?, ?, ?, ?, ?)`, 
			applicant.ID, applicant.Name, applicant.EmploymentStatus, applicant.MaritalStatus, applicant.Sex, applicant.DateOfBirth)

		if err != nil {
			utils.HandleInsertError(w, err, "applicant")
			return
		}

		// Insert household members (if any)
		for _, member := range applicant.Household {
			fmt.Println("Inserting household member:", member)
			member.ID = uuid.New().String()
			_, err = tx.Exec(`INSERT INTO household (id, applicant_id, name, relationship, sex, school_level, employment_status, date_of_birth)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				member.ID, applicant.ID, member.Name, member.Relationship, member.Sex, member.SchoolLevel, member.EmploymentStatus, member.DateOfBirth)
			
			if err != nil {
				utils.HandleInsertError(w, err, "household member")
				return
			}
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Failed to commit", http.StatusInternalServerError)
			return
		}

		// Respond with the created applicant
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(applicant)
	}
}


