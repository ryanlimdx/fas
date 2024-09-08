package utils

import (
	"strings"
)

// Helper function to check if an existing value is valid.
func IsValid(validVals []string, value string) bool {
    for _, item := range validVals {
        if strings.EqualFold(item, value) {
            return true
        }
    }
    return false
}
