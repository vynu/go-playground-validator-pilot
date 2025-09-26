package main

import (
	"goplayground-data-validator/models"
)

// performBusinessValidation performs business logic validation
// DEPRECATED: This function should be moved to the GitHub validator
// Keeping for backward compatibility but will be removed in future versions
func performBusinessValidation(payload *models.GitHubPayload) []string {
	// This logic has been moved to the GitHub validator
	// Returning empty slice to maintain compatibility
	return []string{}
}
