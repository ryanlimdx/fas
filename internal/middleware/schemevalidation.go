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
	validCriteriaLevels = []string{"individual", "household"}
	validCriteriaTypes  = []string{
		"marital_status", // Note that this criteria type is only for the individual level
		"school_level", // Note that this criteria type is only for the household level
		"employment_status", 
		"has_children", // Note that this criteria type is only for the individual level
	}
)

func ValidateScheme(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read request body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		var scheme models.Scheme
        if err := json.NewDecoder(r.Body).Decode(&scheme); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

		// Validate scheme criteria
		for _, criteria := range scheme.Criteria {
            if !utils.IsValid(validCriteriaLevels, criteria.CriteriaLevel) {
                http.Error(w, "Invalid criteria level", http.StatusBadRequest)
                return
            }
			if !utils.IsValid(validCriteriaTypes, criteria.CriteriaType) {
                http.Error(w, "Invalid criteria type", http.StatusBadRequest)
                return
            }
		}

		// Validate scheme benefits
		for _, benefit := range scheme.Benefits {
			if benefit.Amount < 0 {
				http.Error(w, "Invalid benefit amount", http.StatusBadRequest)
				return
			}
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		next.ServeHTTP(w, r)                       
	})
}

