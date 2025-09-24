// Package validations contains Incident-specific validation logic with custom business rules
package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github-data-validator/models"
	"github.com/go-playground/validator/v10"
)

// IncidentValidator handles validation for Incident payloads with 2 custom validations
type IncidentValidator struct {
	validator *validator.Validate
}

// NewIncidentValidator creates a new Incident validator instance
func NewIncidentValidator() *IncidentValidator {
	v := validator.New()
	return &IncidentValidator{validator: v}
}

// ValidatePayload validates an Incident payload and returns structured results
func (iv *IncidentValidator) ValidatePayload(payload models.IncidentPayload) models.ValidationResult {
	// Perform struct validation
	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "incident",
		Provider:  "go-playground",
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	if err := iv.validator.Struct(payload); err != nil {
		result.IsValid = false
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, ve := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   ve.Field(),
					Message: iv.getCustomErrorMessage(ve),
					Code:    "VALIDATION_FAILED",
					Value:   fmt.Sprintf("%v", ve.Value()),
				})
			}
		}
	}

	// Apply 2 custom validations only if basic validation passed
	if result.IsValid {
		// Custom Validation 1: ID format validation
		if err := iv.validateIncidentIDFormat(payload.ID); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "id",
				Message: err.Error(),
				Code:    "INVALID_ID_FORMAT",
				Value:   payload.ID,
			})
		}

		// Custom Validation 2: Priority vs Severity consistency
		if err := iv.validatePrioritySeverityConsistency(payload.Priority, payload.Severity); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "priority",
				Message: err.Error(),
				Code:    "PRIORITY_SEVERITY_MISMATCH",
				Value:   fmt.Sprintf("priority=%d, severity=%s", payload.Priority, payload.Severity),
			})
		}
	}

	// Add business logic warnings even if validation failed
	result.Warnings = iv.validateBusinessLogic(payload)

	return result
}

// Custom Validation 1: validateIncidentIDFormat ensures ID follows pattern INC-YYYYMMDD-NNNN
func (iv *IncidentValidator) validateIncidentIDFormat(id string) error {
	// Pattern: INC-YYYYMMDD-NNNN (e.g., INC-20240924-0001)
	pattern := `^INC-\d{8}-\d{4}$`
	matched, err := regexp.MatchString(pattern, id)
	if err != nil {
		return fmt.Errorf("error validating ID format: %v", err)
	}
	if !matched {
		return fmt.Errorf("incident ID must follow format INC-YYYYMMDD-NNNN (e.g., INC-20240924-0001), got: %s", id)
	}
	return nil
}

// Custom Validation 2: validatePrioritySeverityConsistency ensures priority aligns with severity
func (iv *IncidentValidator) validatePrioritySeverityConsistency(priority int, severity string) error {
	// Define expected priority ranges for each severity level
	expectedPriorities := map[string][]int{
		"low":      {1, 2}, // Priority 1-2 for low severity
		"medium":   {2, 3}, // Priority 2-3 for medium severity
		"high":     {3, 4}, // Priority 3-4 for high severity
		"critical": {4, 5}, // Priority 4-5 for critical severity
	}

	allowedPriorities, exists := expectedPriorities[severity]
	if !exists {
		return fmt.Errorf("unknown severity level: %s", severity)
	}

	// Check if priority is in allowed range
	for _, allowedPriority := range allowedPriorities {
		if priority == allowedPriority {
			return nil
		}
	}

	return fmt.Errorf("priority %d is inconsistent with severity '%s' (expected: %v)",
		priority, severity, allowedPriorities)
}

// getCustomErrorMessage provides friendly error messages for standard validators
func (iv *IncidentValidator) getCustomErrorMessage(ve validator.FieldError) string {
	switch ve.Tag() {
	case "oneof":
		return fmt.Sprintf("Field must be one of: %s", ve.Param())
	case "min":
		if ve.Kind().String() == "string" {
			return fmt.Sprintf("Field must be at least %s characters long", ve.Param())
		}
		return fmt.Sprintf("Field must be at least %s", ve.Param())
	case "max":
		if ve.Kind().String() == "string" {
			return fmt.Sprintf("Field must be at most %s characters long", ve.Param())
		}
		return fmt.Sprintf("Field must be at most %s", ve.Param())
	default:
		return ve.Error()
	}
}

// validateBusinessLogic performs incident-specific business validation checks
func (iv *IncidentValidator) validateBusinessLogic(payload models.IncidentPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Warning: Critical incidents should be assigned immediately
	if payload.Severity == "critical" && payload.AssignedTo == "" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "assigned_to",
			Message:    "Critical incident should be assigned to an engineer immediately",
			Code:       "CRITICAL_INCIDENT_UNASSIGNED",
			Suggestion: "Assign to on-call engineer or escalation team",
		})
	}

	// Warning: Production issues should have high priority
	if payload.Environment == "production" && payload.Priority < 3 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "priority",
			Message:    fmt.Sprintf("Production incident has low priority (%d), consider increasing", payload.Priority),
			Code:       "PRODUCTION_LOW_PRIORITY",
			Suggestion: "Review if priority should be 3 or higher for production issues",
		})
	}

	// Warning: Old open incidents should be reviewed
	if payload.Status == "open" || payload.Status == "investigating" {
		timeSince := time.Since(payload.ReportedAt)
		if timeSince.Hours() > 24 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "status",
				Message:    fmt.Sprintf("Incident has been %s for %.1f hours", payload.Status, timeSince.Hours()),
				Code:       "STALE_INCIDENT",
				Suggestion: "Review incident progress and update status or escalate",
			})
		}
	}

	// Warning: Generic titles for high-severity incidents
	if payload.Severity == "high" || payload.Severity == "critical" {
		genericWords := []string{"issue", "problem", "error", "bug", "broken", "down", "failure"}
		titleLower := strings.ToLower(payload.Title)

		for _, word := range genericWords {
			if strings.Contains(titleLower, word) && len(strings.Fields(payload.Title)) < 6 {
				warnings = append(warnings, models.ValidationWarning{
					Field:      "title",
					Message:    "High/Critical severity incident has generic title",
					Code:       "GENERIC_INCIDENT_TITLE",
					Suggestion: "Provide more specific description for high priority incidents",
				})
				break
			}
		}
	}

	// Warning: Security incidents should be tagged appropriately
	if payload.Category == "security" && len(payload.Tags) == 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "tags",
			Message:    "Security incidents should have relevant tags for tracking and reporting",
			Code:       "SECURITY_INCIDENT_NO_TAGS",
			Suggestion: "Add tags like 'security-breach', 'vulnerability', 'compliance', etc.",
		})
	}

	return warnings
}
