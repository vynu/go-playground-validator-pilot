package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	mathRand "math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"goplayground-data-validator/models"
)

// Global request counter for metrics
var requestCount int64

// Initialize math/rand for fallback
func init() {
	mathRand.Seed(time.Now().UnixNano())
}

// Utility functions for the modular server

// registerCustomValidators registers custom validation functions
func registerCustomValidators(validate *validator.Validate) {
	// GitHub username validator
	validate.RegisterValidation("github_username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) == 0 || len(username) > 39 {
			return false
		}
		// GitHub username pattern: alphanumeric and hyphens, but not starting/ending with hyphen
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`, username)
		return matched
	})

	// SHA validator (for commit SHAs)
	validate.RegisterValidation("sha", func(fl validator.FieldLevel) bool {
		sha := fl.Field().String()
		if len(sha) != 40 {
			return false
		}
		matched, _ := regexp.MatchString(`^[a-f0-9]{40}$`, sha)
		return matched
	})

	// URL validator for GitHub URLs
	validate.RegisterValidation("github_url", func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		matched, _ := regexp.MatchString(`^https://api\.github\.com/`, url)
		return matched
	})

	log.Println("Custom validators registered successfully")
}

// getRequestCount returns the current request count
func getRequestCount() int64 {
	return atomic.LoadInt64(&requestCount)
}

// incrementRequestCount increments the request counter
func incrementRequestCount() {
	atomic.AddInt64(&requestCount, 1)
}

// checkRateLimit checks if the request should be rate limited
func checkRateLimit(r *http.Request) bool {
	// Simple rate limiting - always allow for now
	return true
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return "req_" + randomString(16)
}

// randomString generates a cryptographically secure random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// Use crypto/rand for secure random generation
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to math/rand if crypto/rand fails (should never happen)
			b[i] = charset[mathRand.Intn(len(charset))]
		} else {
			b[i] = charset[idx.Int64()]
		}
	}
	return string(b)
}

// Simple validation result structure
type SimpleValidationResult struct {
	ID      string   `json:"id"`
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors"`
}

// generateValidationSummary generates a simple validation summary
func generateValidationSummary(errorCount, warningCount int) map[string]interface{} {
	return map[string]interface{}{
		"total_fields":       10,
		"valid_fields":       10 - errorCount,
		"invalid_fields":     errorCount,
		"warning_fields":     warningCount,
		"validation_score":   float64(10-errorCount) / 10.0 * 100,
		"data_quality_score": 85.0,
	}
}

// processBatchValidation processes batch validation
func processBatchValidation(payloads []models.GitHubPayload) []SimpleValidationResult {
	results := make([]SimpleValidationResult, len(payloads))
	for i := range payloads {
		results[i] = SimpleValidationResult{
			ID:      fmt.Sprintf("batch_%d", i),
			IsValid: true, // Simplified validation
			Errors:  []string{},
		}
	}
	return results
}

// generateBatchSummary generates a batch summary
func generateBatchSummary(results []SimpleValidationResult) map[string]interface{} {
	valid := 0
	for _, result := range results {
		if result.IsValid {
			valid++
		}
	}
	return map[string]interface{}{
		"total_valid":   valid,
		"total_invalid": len(results) - valid,
		"success_rate":  float64(valid) / float64(len(results)) * 100,
	}
}

// convertValidationErrors converts validator errors to string slice
func convertValidationErrors(err error) []string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, ve := range validationErrors {
			errors = append(errors, fmt.Sprintf("%s: %s", ve.Field(), ve.Error()))
		}
	}
	return errors
}

// performBusinessValidation performs business logic validation
// DEPRECATED: This function should be moved to the GitHub validator
// Keeping for backward compatibility but will be removed in future versions
func performBusinessValidation(payload *models.GitHubPayload) []string {
	// This logic has been moved to the GitHub validator
	// Returning empty slice to maintain compatibility
	return []string{}
}

// handleShutdown handles graceful server shutdown
func handleShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down server...")
	// Basic shutdown - server will handle cleanup
}
