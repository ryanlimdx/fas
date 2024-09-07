// Handles all the requests related to applicants
package handlers

import (
	"database/sql"
	"encoding/json"
	"fas/internal/models"
	"fas/internal/utils"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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

// GetApplicants retrieves all applicants from the database, returning them in JSON format.
func GetApplicants(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query the database for all applicants
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

		// Convert the list of applicants to JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(applicants)
	}
}

// getHouseholdMembers retrieves the household members for a given applicant ID
func getHouseholdMembers(db *sql.DB, applicantID string) ([]models.Household, error) {
	// Query the database for the applicant's household members
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
