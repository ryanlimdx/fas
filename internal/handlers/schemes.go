// Handles all the requests related to the schemes
package handlers

import (
	"database/sql"
	"encoding/json"
	"fas/internal/models"
	"fas/internal/utils"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

// GetSchemes retrieves all schemes from the database.
func GetSchemes(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Query to select all schemes
        rows, err := db.Query("SELECT id, name FROM schemes")
        if err != nil {
            http.Error(w, "Failed to retrieve schemes", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var schemes []models.Scheme
        for rows.Next() {
            var scheme models.Scheme
            if err := rows.Scan(&scheme.ID, &scheme.Name); err != nil {
                http.Error(w, "Failed to scan scheme", http.StatusInternalServerError)
                return
            }

            // Fetch criteria
            scheme.Criteria, err = getCriteriaForScheme(db, scheme.ID)
            if err != nil {
				http.Error(w, "Failed to retrieve criteria", http.StatusInternalServerError)
				return
			}

            // Fetch benefits
            scheme.Benefits, err = getBenefitsForScheme(db, scheme.ID)
            if err != nil {
				http.Error(w, "Failed to retrieve benefits", http.StatusInternalServerError)
				return
			}

            schemes = append(schemes, scheme)
        }

        if err := rows.Err(); err != nil {
            http.Error(w, "Failed to read scheme data", http.StatusInternalServerError)
            return
        }

        // Sending the retrieved schemes as a JSON array
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(schemes)
    }
}

// getCriteriaForScheme retrieves all criteria for a scheme.
func getCriteriaForScheme(db *sql.DB, schemeID string) ([]models.Criteria, error) {
    var criteria []models.Criteria
    rows, err := db.Query(`SELECT id, criteria_level, criteria_type, status FROM criteria 
                            JOIN scheme_criteria ON criteria.id = scheme_criteria.criteria_id 
                            WHERE scheme_criteria.scheme_id = ?`, schemeID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var criterion models.Criteria
        if err := rows.Scan(&criterion.ID, &criterion.CriteriaLevel, &criterion.CriteriaType, &criterion.Status); err != nil {
            return nil, err
        }
        criteria = append(criteria, criterion)
    }
    return criteria, nil
}

// getBenefitsForScheme retrieves all benefits for a scheme.
func getBenefitsForScheme(db *sql.DB, schemeID string) ([]models.Benefit, error) {
    var benefits []models.Benefit
    rows, err := db.Query(`SELECT id, name, amount FROM benefits 
                            JOIN scheme_benefits ON benefits.id = scheme_benefits.benefit_id 
                            WHERE scheme_benefits.scheme_id = ?`, schemeID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var benefit models.Benefit
        if err := rows.Scan(&benefit.ID, &benefit.Name, &benefit.Amount); err != nil {
            return nil, err
        }
        benefits = append(benefits, benefit)
    }
    return benefits, nil
}

// GetEligibleSchemes returns the schemes an applicant is eligible for
func GetEligibleSchemes(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        applicantID := r.URL.Query().Get("applicant")

        // Validate the UUID for security
        if err := utils.ValidateUUID(applicantID); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Check if applicant exist
        var exists bool
        err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM applicants WHERE id = ?)", applicantID).Scan(&exists)
        if err != nil || !exists {
            http.Error(w, "Applicant not found", http.StatusBadRequest)
            return
        }

        // Fetch schemes the applicant is eligible for
        schemes, err := fetchEligibleSchemes(db, applicantID)
        if err != nil {
            http.Error(w, "Error retrieving schemes", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(schemes)
    }
}

// fetchEligibleSchemes queries the database for schemes an applicant is eligible for.
func fetchEligibleSchemes(db *sql.DB, applicantID string) ([]models.Scheme, error) {
    query := `SELECT s.id, s.name
	FROM schemes s
	LEFT JOIN (
		SELECT sc.scheme_id
		FROM scheme_criteria sc
		JOIN criteria c ON sc.criteria_id = c.id
		LEFT JOIN applicants a ON a.id = ?
		LEFT JOIN household h ON h.applicant_id = a.id
		WHERE (
			(c.criteria_level = 'individual' AND c.criteria_type = 'employment_status' AND a.employment_status = c.status)
			OR (c.criteria_level = 'individual' AND c.criteria_type = 'marital_status' AND a.marital_status = c.status)
			OR (c.criteria_level = 'individual' AND c.criteria_type = 'has_children' AND EXISTS (
				SELECT 1 FROM household WHERE applicant_id = a.id AND (relationship = 'son' OR relationship = 'daughter')
			))
			OR (c.criteria_level = 'household' AND c.criteria_type = 'school_level' AND h.school_level = c.status)
			OR (c.criteria_level = 'household' AND c.criteria_type = 'employment_status' AND h.employment_status = c.status)
		)
		GROUP BY sc.scheme_id
		HAVING COUNT(DISTINCT c.id) = (
			SELECT COUNT(*) FROM scheme_criteria WHERE scheme_id = sc.scheme_id
		)
	) AS eligible_schemes ON s.id = eligible_schemes.scheme_id
	WHERE eligible_schemes.scheme_id IS NOT NULL OR NOT EXISTS (
		SELECT 1 FROM scheme_criteria WHERE scheme_id = s.id
	)
	`
	rows, err := db.Query(query, applicantID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
	
    // Parse schemes that the applicant is eligible for.
    var schemes []models.Scheme
    for rows.Next() {
        var scheme models.Scheme
        if err := rows.Scan(&scheme.ID, &scheme.Name); err != nil {
            return nil, err
        }
        schemes = append(schemes, scheme)
    }

    return schemes, nil
}

// UpdateScheme updates an existing scheme.
func UpdateScheme(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Validate the scheme
        vars := mux.Vars(r)
        schemeID := vars["id"]
        if err := checkScheme(db, schemeID); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

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

        // Update the scheme
        _, err = tx.Exec(`UPDATE schemes SET name=? WHERE id=?`, scheme.Name, schemeID)
        if err != nil {
            http.Error(w, "Failed to update scheme", http.StatusInternalServerError)
            return
        }

        // Delete all existing criteria
        _, err = tx.Exec(`DELETE FROM scheme_criteria WHERE scheme_id=?`, schemeID)
        if err != nil {
            http.Error(w, "Failed to delete existing criteria", http.StatusInternalServerError)
            return
        }

        // Delete all existing benefits
        _, err = tx.Exec(`DELETE FROM scheme_benefits WHERE scheme_id=?`, schemeID)
        if err != nil {
            http.Error(w, "Failed to delete existing benefits", http.StatusInternalServerError)
            return
        }

        // Insert Criteria
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
                schemeID, criteria.ID)
            if err != nil {
                http.Error(w, "Failed to link criteria to scheme", http.StatusInternalServerError)
                return
            }
        }

        // Insert Benefits
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
                schemeID, benefit.ID)
            if err != nil {
                http.Error(w, "Failed to link benefit to scheme", http.StatusInternalServerError)
                return
            }
        }

        // Commit the transaction
        if err = tx.Commit(); err != nil {
            http.Error(w, "Failed to commit", http.StatusInternalServerError)
            return
        }

        // Cleanup orphaned benefits and criteria
        if err := deleteOrphanedBenefits(db); err != nil {
            http.Error(w, "Failed to delete unused benefits", http.StatusInternalServerError)
            return
        }
        if err := deleteOrphanedCriteria(db); err != nil {
            http.Error(w, "Failed to delete unused criteria", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}

// DeleteScheme removes a scheme from the database.
func DeleteScheme(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Validate the scheme
        vars := mux.Vars(r)
        schemeID := vars["id"]
        if err := checkScheme(db, schemeID); err != nil {
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

        // Delete the scheme
        _, err = tx.Exec(`DELETE FROM schemes WHERE id=?`, schemeID)
        if err != nil {
            http.Error(w, "Failed to delete scheme", http.StatusInternalServerError)
            return
        }

        // Commit the transaction
        if err := tx.Commit(); err != nil {
            http.Error(w, "Failed to commit", http.StatusInternalServerError)
            return
        }

        // Cleanup orphaned benefits and criteria
        if err := deleteOrphanedBenefits(db); err != nil {
            http.Error(w, "Failed to delete unused benefits", http.StatusInternalServerError)
            return
        }
        if err := deleteOrphanedCriteria(db); err != nil {
            http.Error(w, "Failed to delete unused criteria", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}

// checkScheme validates the UUID and checks if a scheme exists.
func checkScheme(db *sql.DB, schemeID string) error {
    // Validate the UUID for security
    if err := utils.ValidateUUID(schemeID); err != nil {
        return fmt.Errorf("invalid UUID: %w", err)
    }

    // Check if scheme exists
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schemes WHERE id = ?)", schemeID).Scan(&exists)
    if err != nil {
        return fmt.Errorf("error checking scheme existence: %w", err)
    }
    if !exists {
        return fmt.Errorf("scheme not found")
    }

    return nil
}

// deleteOrphanedBenefits deletes benefits that are not linked to any scheme.
func deleteOrphanedBenefits(db *sql.DB) error {
    // Fetch the IDs of unused benefits
    rows, err := db.Query(`SELECT id FROM benefits WHERE id NOT IN (SELECT benefit_id FROM scheme_benefits)`)
    if err != nil {
        return fmt.Errorf("failed to fetch unused benefits: %v", err)
    }
    defer rows.Close()

    delete, err := db.Prepare(`DELETE FROM benefits WHERE id = ?`)
    if err != nil {
        return fmt.Errorf("failed to prepare delete statement: %v", err)
    }
    defer delete.Close()

    var id string
    // Execute a DELETE for each unused benefit
    for rows.Next() {
        if err := rows.Scan(&id); err != nil {
            return fmt.Errorf("failed to scan benefit id: %v", err)
        }

        if _, err := delete.Exec(id); err != nil {
            return fmt.Errorf("failed to delete benefit with id %s: %v", id, err)
        }
    }

    if err := rows.Err(); err != nil {
        return fmt.Errorf("iteration error: %v", err)
    }

    return nil
}

// deleteOrphanedCriteria deletes criteria that are not linked to any scheme.
func deleteOrphanedCriteria(db *sql.DB) error {
    // Fetch the IDs of unused criteria
    rows, err := db.Query(`SELECT id FROM criteria WHERE id NOT IN (SELECT criteria_id FROM scheme_criteria)`)
    if err != nil {
        return fmt.Errorf("failed to fetch unused benefits: %v", err)
    }
    defer rows.Close()

    delete, err := db.Prepare(`DELETE FROM criteria WHERE id = ?`)
    if err != nil {
        return fmt.Errorf("failed to prepare delete statement: %v", err)
    }
    defer delete.Close()

    var id string
    // Execute a DELETE for each unused criteria
    for rows.Next() {
        if err := rows.Scan(&id); err != nil {
            return fmt.Errorf("failed to scan criteria id: %v", err)
        }

        if _, err := delete.Exec(id); err != nil {
            return fmt.Errorf("failed to delete criteria with id %s: %v", id, err)
        }
    }

    if err := rows.Err(); err != nil {
        return fmt.Errorf("iteration error: %v", err)
    }

    return nil
}
