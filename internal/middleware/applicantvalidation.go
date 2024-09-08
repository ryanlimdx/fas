// Handles applicant validation.
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"fas/internal/models"
	"fas/internal/utils"
)

var (
    validEmploymentStatus = []string{"employed", "unemployed", "self-employed", "retired"}
    validMaritalStatus    = []string{"single", "married", "divorced", "widowed"}
    validSchoolLevels     = []string{"none", "primary", "secondary", "post-secondary", "university", "graduated"}
	validSex              = []string{"male", "female"}
    validRelationships    = []string{"parent", "son", "daughter", "sibling", "spouse", "other"}
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

		// Validate the applicant's fields
        if !utils.IsValid(validEmploymentStatus, applicant.EmploymentStatus) {
            http.Error(w, "Invalid employment status, " + 
                utils.FormatValidOptions(validEmploymentStatus), http.StatusBadRequest)
            return
        }
        if !utils.IsValid(validMaritalStatus, applicant.MaritalStatus) {
            http.Error(w, "Invalid marital status, " + 
                utils.FormatValidOptions(validMaritalStatus), http.StatusBadRequest)
            return
        }
        if !utils.IsValid(validSex, applicant.Sex) {
            http.Error(w, "Invalid applicant sex, " + 
                utils.FormatValidOptions(validSex), http.StatusBadRequest)
            return
        }

        // Validate household member(s) fields
        for _, member := range applicant.Household {
            if !utils.IsValid(validRelationships, member.Relationship) {
                http.Error(w, "Invalid household member relationship, " + 
                    utils.FormatValidOptions(validRelationships), http.StatusBadRequest)
                return
            }
            if !utils.IsValid(validSchoolLevels, member.SchoolLevel) {
                http.Error(w, "Invalid household member school level, " + 
                    utils.FormatValidOptions(validSchoolLevels), http.StatusBadRequest)
                return
            }
			if !utils.IsValid(validEmploymentStatus, member.EmploymentStatus) {
                http.Error(w, "Invalid household member employment status, " + 
                    utils.FormatValidOptions(validEmploymentStatus), http.StatusBadRequest)
                return
            }
			if !utils.IsValid(validSex, applicant.Sex) {
				http.Error(w, "Invalid household member sex, " + 
                    utils.FormatValidOptions(validSex), http.StatusBadRequest)
				return
			}
        }

		r.Body = io.NopCloser(bytes.NewBuffer(body))
        next.ServeHTTP(w, r)
    })
}

