package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"bytes"
    "io"
	"fas/internal/models"
)

var (
    validEmploymentStatus = []string{"employed", "unemployed", "self-employed", "retired"}
    validMaritalStatus    = []string{"single", "married", "divorced", "widowed"}
    validSchoolLevels       = []string{"none", "primary", "secondary", "post-secondary", "university"}
	validSex 				= []string{"male", "female"}
    validRelationships      = []string{"parent", "son", "daughter", "sibling", "spouse", "other"}
)

func ValidateApplicant(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "could not read request body", http.StatusBadRequest)
            return
        }
        r.Body = io.NopCloser(bytes.NewBuffer(body)) 

        var applicant models.Applicant
        if err := json.NewDecoder(r.Body).Decode(&applicant); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

		// Validate the applicant
        if !isValid(validEmploymentStatus, applicant.EmploymentStatus) {
            http.Error(w, "Invalid employment status", http.StatusBadRequest)
            return
        }
        if !isValid(validMaritalStatus, applicant.MaritalStatus) {
            http.Error(w, "Invalid marital status", http.StatusBadRequest)
            return
        }
        if !isValid(validSex, applicant.Sex) {
            http.Error(w, "Invalid applicant sex", http.StatusBadRequest)
            return
        }

        // Validate household member(s)
        for _, member := range applicant.Household {
            if !isValid(validRelationships, member.Relationship) {
                http.Error(w, "Invalid household member relationship", http.StatusBadRequest)
                return
            }
            if !isValid(validSchoolLevels, member.SchoolLevel) {
                http.Error(w, "Invalid household member school level", http.StatusBadRequest)
                return
            }
			if !isValid(validEmploymentStatus, member.EmploymentStatus) {
                http.Error(w, "Invalid household member school level", http.StatusBadRequest)
                return
            }
			if !isValid(validSex, applicant.Sex) {
				http.Error(w, "Invalid household member sex", http.StatusBadRequest)
				return
			}
        }

		r.Body = io.NopCloser(bytes.NewBuffer(body))
        
		// Call the next handler
        next.ServeHTTP(w, r)
    })
}

// Helper function to check if an existing value is valid.
func isValid(validVals []string, value string) bool {
    for _, item := range validVals {
        if strings.EqualFold(item, value) {
            return true
        }
    }
    return false
}
