package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIncidentPayload_Validation(t *testing.T) {
	tests := []struct {
		name     string
		incident IncidentPayload
		wantErr  bool
		errField string
	}{
		{
			name: "valid incident payload",
			incident: IncidentPayload{
				ID:          "INC-2024-001",
				Title:       "Database connection timeout issue",
				Description: "Users are experiencing timeout errors when connecting to the main database server",
				Severity:    "high",
				Status:      "open",
				Priority:    3,
				Category:    "bug",
				Environment: "production",
				ReportedBy:  "john.doe@example.com",
				AssignedTo:  "jane.smith@example.com",
				ReportedAt:  time.Now(),
				UpdatedAt:   time.Now(),
				Tags:        []string{"database", "timeout", "urgent"},
				Impact:      "high",
			},
			wantErr: false,
		},
		{
			name: "missing required ID",
			incident: IncidentPayload{
				Title:       "Database connection timeout issue",
				Description: "Users are experiencing timeout errors when connecting to the main database server",
				Severity:    "high",
				Status:      "open",
				Priority:    3,
				Category:    "bug",
				Environment: "production",
				ReportedBy:  "john.doe@example.com",
				ReportedAt:  time.Now(),
			},
			wantErr:  true,
			errField: "ID",
		},
		{
			name: "invalid severity",
			incident: IncidentPayload{
				ID:          "INC-2024-001",
				Title:       "Database connection timeout issue",
				Description: "Users are experiencing timeout errors when connecting to the main database server",
				Severity:    "invalid",
				Status:      "open",
				Priority:    3,
				Category:    "bug",
				Environment: "production",
				ReportedBy:  "john.doe@example.com",
				ReportedAt:  time.Now(),
			},
			wantErr:  true,
			errField: "Severity",
		},
		{
			name: "title too short",
			incident: IncidentPayload{
				ID:          "INC-2024-001",
				Title:       "Short",
				Description: "Users are experiencing timeout errors when connecting to the main database server",
				Severity:    "high",
				Status:      "open",
				Priority:    3,
				Category:    "bug",
				Environment: "production",
				ReportedBy:  "john.doe@example.com",
				ReportedAt:  time.Now(),
			},
			wantErr:  true,
			errField: "Title",
		},
		{
			name: "priority out of range",
			incident: IncidentPayload{
				ID:          "INC-2024-001",
				Title:       "Database connection timeout issue",
				Description: "Users are experiencing timeout errors when connecting to the main database server",
				Severity:    "high",
				Status:      "open",
				Priority:    10,
				Category:    "bug",
				Environment: "production",
				ReportedBy:  "john.doe@example.com",
				ReportedAt:  time.Now(),
			},
			wantErr:  true,
			errField: "Priority",
		},
		{
			name: "invalid environment",
			incident: IncidentPayload{
				ID:          "INC-2024-001",
				Title:       "Database connection timeout issue",
				Description: "Users are experiencing timeout errors when connecting to the main database server",
				Severity:    "high",
				Status:      "open",
				Priority:    3,
				Category:    "bug",
				Environment: "invalid",
				ReportedBy:  "john.doe@example.com",
				ReportedAt:  time.Now(),
			},
			wantErr:  true,
			errField: "Environment",
		},
		{
			name: "invalid tag",
			incident: IncidentPayload{
				ID:          "INC-2024-001",
				Title:       "Database connection timeout issue",
				Description: "Users are experiencing timeout errors when connecting to the main database server",
				Severity:    "high",
				Status:      "open",
				Priority:    3,
				Category:    "bug",
				Environment: "production",
				ReportedBy:  "john.doe@example.com",
				ReportedAt:  time.Now(),
				Tags:        []string{"a"}, // Too short
			},
			wantErr:  true,
			errField: "Tags",
		},
	}

	validator := getTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.incident)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncidentPayload validation error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				// Check if the error is for the expected field
				errStr := err.Error()
				if tt.errField != "" && !containsField(errStr, tt.errField) {
					t.Errorf("Expected validation error for field %s, got: %v", tt.errField, err)
				}
			}
		})
	}
}

func TestIncidentPayload_FieldValidation(t *testing.T) {
	t.Run("ID validation", func(t *testing.T) {
		tests := []struct {
			id      string
			wantErr bool
		}{
			{"INC-2024-001", false},
			{"INC", false},                 // Actually valid - min=3
			{"", true},                     // Empty
			{"a", true},                    // Too short
			{generateLongString(51), true}, // Too long
		}

		for _, test := range tests {
			incident := getValidIncident()
			incident.ID = test.id

			err := getTestValidator().Struct(incident)
			if (err != nil) != test.wantErr {
				t.Errorf("ID validation for '%s': error = %v, wantErr %v", test.id, err, test.wantErr)
			}
		}
	})

	t.Run("Priority validation", func(t *testing.T) {
		tests := []struct {
			priority int
			wantErr  bool
		}{
			{1, false},
			{3, false},
			{5, false},
			{0, true},  // Too low
			{6, true},  // Too high
			{-1, true}, // Negative
		}

		for _, test := range tests {
			incident := getValidIncident()
			incident.Priority = test.priority

			err := getTestValidator().Struct(incident)
			if (err != nil) != test.wantErr {
				t.Errorf("Priority validation for %d: error = %v, wantErr %v", test.priority, err, test.wantErr)
			}
		}
	})
}

func TestIncidentPayload_JSONMarshaling(t *testing.T) {
	incident := getValidIncident()

	// Test JSON marshaling
	data, err := json.Marshal(incident)
	if err != nil {
		t.Fatalf("Failed to marshal incident: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled IncidentPayload
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal incident: %v", err)
	}

	// Compare key fields
	if incident.ID != unmarshaled.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, incident.ID)
	}
	if incident.Title != unmarshaled.Title {
		t.Errorf("Title mismatch: got %s, want %s", unmarshaled.Title, incident.Title)
	}
}

func BenchmarkIncidentPayload_Validation(b *testing.B) {
	incident := getValidIncident()
	validator := getTestValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Struct(incident)
	}
}

// Helper functions

func getValidIncident() IncidentPayload {
	return IncidentPayload{
		ID:          "INC-2024-001",
		Title:       "Database connection timeout issue",
		Description: "Users are experiencing timeout errors when connecting to the main database server",
		Severity:    "high",
		Status:      "open",
		Priority:    3,
		Category:    "bug",
		Environment: "production",
		ReportedBy:  "john.doe@example.com",
		AssignedTo:  "jane.smith@example.com",
		ReportedAt:  time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        []string{"database", "timeout"},
		Impact:      "high",
	}
}
