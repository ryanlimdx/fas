// Handles all the requests related to applicants.
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"fas/internal/models"
	"fas/internal/utils"
)

// GetApplicants retrieves all applicants from the database, returning them in JSON format.
func GetApplicants(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT id, name, employment_status, marital_status, sex, date_of_birth 
			FROM applicants
		`)
		if err != nil {
			http.Error(w, "Failed to retrieve applicants", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var applicants []models.Applicant

		// Parse applicants
		for rows.Next() {
			var applicant models.Applicant
			err := rows.Scan(
				&applicant.ID, 
				&applicant.Name, 
				&applicant.EmploymentStatus, 
				&applicant.MaritalStatus, 
				&applicant.Sex, 
				&applicant.DateOfBirth,
			)
			if err != nil {
				http.Error(w, "Failed to scan applicants", http.StatusInternalServerError)
				return
			}

			// Get household members for the current applicant
			householdMembers, err := getHouseholdMembers(db, applicant.ID)
			if err != nil {
				http.Error(w, "Failed to retrieve household members", http.StatusInternalServerError)
				return
			}
			applicant.Household = householdMembers

			applicants = append(applicants, applicant)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating over applicants", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(applicants)
	}
}

// getHouseholdMembers retrieves the household members for a given applicant ID
func getHouseholdMembers(db *sql.DB, applicantID string) ([]models.Household, error) {
	rows, err := db.Query(
		`SELECT id, applicant_id, name, relationship, sex, school_level, employment_status, date_of_birth 
		FROM household WHERE applicant_id = ?`, 
		applicantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var householdMembers []models.Household

	// Parse household members
	for rows.Next() {
		var member models.Household
		err := rows.Scan(&member.ID, &member.ApplicantID, &member.Name, &member.Relationship, &member.Sex, &member.SchoolLevel, &member.EmploymentStatus, &member.DateOfBirth)
		if err != nil {
			return nil, err
		}

		householdMembers = append(householdMembers, member)
	}

	return householdMembers, rows.Err()
}

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
		for i := range applicant.Household {
			member := &applicant.Household[i]
			fmt.Println("Inserting household member:", *member)
			member.ID = uuid.New().String()
			member.ApplicantID = applicant.ID
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

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(applicant)
	}
}

// UpdateApplicant updates an existing applicant and their household members in the database.
func UpdateApplicant(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		// Validate the applicant
        vars := mux.Vars(r)
        applicantID := vars["id"]
        if err := checkApplicant(db, applicantID); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

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

		// Update the applicant
        _, err = tx.Exec(`UPDATE applicants SET name=?, employment_status=?, marital_status=?, sex=?, date_of_birth=? WHERE id=?`,
            applicant.Name, applicant.EmploymentStatus, applicant.MaritalStatus, applicant.Sex, applicant.DateOfBirth, applicantID)
        if err != nil {
            http.Error(w, "Failed to update applicant", http.StatusInternalServerError)
            return
        }

        // Delete all existing household members
        _, err = tx.Exec(`DELETE FROM household WHERE applicant_id=?`, applicantID)
        if err != nil {
            http.Error(w, "Failed to delete existing household members", http.StatusInternalServerError)
            return
        }

        // Insert new household members (if provided)
        for i := range applicant.Household {
			member := &applicant.Household[i]
            member.ID = uuid.New().String()
            _, err = tx.Exec(`INSERT INTO household (id, applicant_id, name, relationship, sex, school_level, employment_status, date_of_birth) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
                member.ID, applicantID, member.Name, member.Relationship, member.Sex, member.SchoolLevel, member.EmploymentStatus, member.DateOfBirth)
            if err != nil {
                utils.HandleInsertError(w, err, "household member")
                return
            }
        }

		// Commit the transaction
        if err = tx.Commit(); err != nil {
            http.Error(w, "Failed to commit", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}

// DeleteApplicant removes an applicant from the database.
func DeleteApplicant(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Validate the applicant
		vars := mux.Vars(r)
        applicantID := vars["id"]
        if err := checkApplicant(db, applicantID); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

		// Begin transaction
        tx, err := db.Begin()
        if err != nil {
            http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
            return
        }
        defer tx.Rollback()

		// Delete the applicant
        _, err = tx.Exec(`DELETE FROM applicants WHERE id=?`, applicantID)
        if err != nil {
            http.Error(w, "Failed to delete applicant", http.StatusInternalServerError)
            return
        }

		// Commit the transaction
        err = tx.Commit()
        if err != nil {
            http.Error(w, "Failed to commit", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}

// checkApplicant validates the UUID and checks if an applicant exists in the database.
func checkApplicant(db *sql.DB, applicantID string) error {
    // Validate the UUID for security
    if err := utils.ValidateUUID(applicantID); err != nil {
        return fmt.Errorf("invalid UUID: %w", err)
    }

    // Check if the applicant exists
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM applicants WHERE id = ?)", applicantID).Scan(&exists)
    if err != nil {
        return fmt.Errorf("error checking applicant existence: %w", err)
    }
    if !exists {
        return fmt.Errorf("applicant not found")
    }

    return nil
}
