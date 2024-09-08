// Contains helper functions.
package utils

import (
    "fmt"
	"strings"

    "github.com/google/uuid"
)

// IsValid checks if an existing value is valid.
func IsValid(validVals []string, value string) bool {
    for _, item := range validVals {
        if strings.EqualFold(item, value) {
            return true
        }
    }
    return false
}

// ValidateUUID checks if the provided string is a valid UUID and not empty.
func ValidateUUID(id string) (error) {
    if id == "" {
        return fmt.Errorf("ID is required")
    }
    if _, err := uuid.Parse(id); err != nil {
        return fmt.Errorf("invalid ID format")
    }
    return nil
}

// FormatValidOptions returns a string with the valid options.
func FormatValidOptions(options []string) string {
    return "valid options: [" + strings.Join(options, ", ") + "]. "
}