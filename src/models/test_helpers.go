package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Test helper functions used across model tests

// getTestValidator returns a configured validator instance for testing
func getTestValidator() *validator.Validate {
	v := validator.New()

	// Register custom GitHub username validator for testing
	v.RegisterValidation("github_username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) < 1 || len(username) > 39 {
			return false
		}
		// Basic GitHub username validation
		return !strings.HasPrefix(username, "-") && !strings.HasSuffix(username, "-")
	})

	// Register hex color validator for testing
	v.RegisterValidation("hexcolor", func(fl validator.FieldLevel) bool {
		color := fl.Field().String()
		if len(color) != 6 {
			return false
		}
		for _, char := range color {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
		return true
	})

	// Register API content type validator for testing
	v.RegisterValidation("api_content_type", func(fl validator.FieldLevel) bool {
		contentType := fl.Field().String()
		validTypes := []string{
			"application/json",
			"application/xml",
			"text/plain",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
		}
		for _, valid := range validTypes {
			if contentType == valid {
				return true
			}
		}
		return false
	})

	// Register API version validator for testing
	v.RegisterValidation("api_version", func(fl validator.FieldLevel) bool {
		version := fl.Field().String()
		// Simple version validation (v1, v1.0, v1.2.3, etc.)
		return strings.HasPrefix(version, "v") && len(version) > 1
	})

	return v
}

// containsField checks if error message contains reference to a specific field
func containsField(errStr, field string) bool {
	return strings.Contains(strings.ToLower(errStr), strings.ToLower(field))
}

// generateLongString generates a string of specified length for testing
func generateLongString(length int) string {
	return strings.Repeat("a", length)
}

// MarshalJSON methods for testing JSON serialization

func (i IncidentPayload) MarshalJSON() ([]byte, error) {
	type Alias IncidentPayload
	return json.Marshal((*Alias)(&i))
}

func (i *IncidentPayload) UnmarshalJSON(data []byte) error {
	type Alias IncidentPayload
	return json.Unmarshal(data, (*Alias)(i))
}

// Test data builders for complex structures

func getValidAPIRequest() APIRequest {
	now := time.Now()
	return APIRequest{
		Method: "POST",
		URL:    "https://api.example.com/v1/users",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		QueryParams: map[string]interface{}{
			"include": "profile",
			"format":  "json",
		},
		Body: map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		Timestamp: now,
		RequestID: "req-123456",
		RemoteIP:  "192.168.1.100",
		Source:    "web",
	}
}

// Helper function for getting the field count for testing
func getFieldCount(v interface{}) int {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.NumField()
}
