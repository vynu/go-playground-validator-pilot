// Package validations contains Deployment-specific validation logic
package validations

import (
	"fmt"
	"regexp"
	"strings"

	"github-data-validator/models"
	"github.com/go-playground/validator/v10"
)

// DeploymentValidator handles validation for Deployment payloads
type DeploymentValidator struct {
	validator *validator.Validate
}

// NewDeploymentValidator creates a new Deployment validator instance with custom validators
func NewDeploymentValidator() *DeploymentValidator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("deployment_name", validateDeploymentName)
	v.RegisterValidation("semver", validateDeploymentSemVer)

	return &DeploymentValidator{
		validator: v,
	}
}

// ValidatePayload validates a Deployment payload and returns structured results
func (dv *DeploymentValidator) ValidatePayload(payload interface{}) models.ValidationResult {
	deploymentPayload, ok := payload.(models.DeploymentPayload)
	if !ok {
		return models.ValidationResult{
			IsValid:   false,
			ModelType: "deployment",
			Provider:  "go-playground",
			Errors: []models.ValidationError{{
				Field:   "payload",
				Message: "payload is not a Deployment payload",
				Code:    "TYPE_MISMATCH",
				Value:   fmt.Sprintf("%T", payload),
			}},
		}
	}

	// Perform struct validation
	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "deployment",
		Provider:  "go-playground",
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	if err := dv.validator.Struct(deploymentPayload); err != nil {
		result.IsValid = false
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, ve := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   ve.Field(),
					Message: dv.getCustomErrorMessage(ve),
					Code:    "VALIDATION_FAILED",
					Value:   fmt.Sprintf("%v", ve.Value()),
				})
			}
		}
	}

	// Add business logic validation warnings
	if result.IsValid {
		result.Warnings = dv.validateBusinessLogic(deploymentPayload)
	}

	return result
}

// Custom Validation 1: validateDeploymentName - validates app name format
func validateDeploymentName(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Must start with a letter, can contain letters, numbers, hyphens
	// Must not start or end with hyphen
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9-]*[a-zA-Z0-9]$`, value)
	return matched || len(value) == 1 // Allow single character names
}

// Custom Validation 2: validateDeploymentSemVer - validates semantic versioning format
func validateDeploymentSemVer(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Semantic versioning: MAJOR.MINOR.PATCH (e.g., 1.0.0, 2.1.3)
	// Can optionally have pre-release: 1.0.0-alpha.1
	semverPattern := `^([0-9]+)\.([0-9]+)\.([0-9]+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`
	matched, _ := regexp.MatchString(semverPattern, value)
	return matched
}

// getCustomErrorMessage provides friendly error messages for custom validators
func (dv *DeploymentValidator) getCustomErrorMessage(ve validator.FieldError) string {
	switch ve.Tag() {
	case "deployment_name":
		return "App name must start with a letter and contain only letters, numbers, and hyphens"
	case "semver":
		return "Version must follow semantic versioning format (e.g., 1.0.0, 2.1.3-alpha.1)"
	case "hexadecimal":
		return "Commit hash must be a valid hexadecimal string"
	case "len":
		return fmt.Sprintf("Field must be exactly %s characters long", ve.Param())
	case "oneof":
		return fmt.Sprintf("Field must be one of: %s", ve.Param())
	default:
		return ve.Error()
	}
}

// validateBusinessLogic performs deployment-specific business validation checks
func (dv *DeploymentValidator) validateBusinessLogic(payload models.DeploymentPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Warning: Production deployments should be from main/master branch
	if payload.Environment == "production" {
		if payload.Branch != "main" && payload.Branch != "master" {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "branch",
				Message:    fmt.Sprintf("Production deployment from '%s' branch is not recommended", payload.Branch),
				Code:       "NON_MAIN_PROD_DEPLOY",
				Suggestion: "Consider deploying production from main/master branch",
			})
		}
	}

	// Warning: Check for rollback deployments
	if payload.Rollback {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "rollback",
			Message:    "This is a rollback deployment",
			Code:       "ROLLBACK_DEPLOYMENT",
			Suggestion: "Ensure the target version is stable",
		})
	}

	// Warning: Failed deployments should be investigated
	if payload.Status == "failed" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "status",
			Message:    "Deployment has failed status",
			Code:       "FAILED_DEPLOYMENT",
			Suggestion: "Check deployment logs and investigate failure cause",
		})
	}

	// Warning: Check for development/staging versions in production
	if payload.Environment == "production" {
		if strings.Contains(strings.ToLower(payload.Version), "dev") ||
			strings.Contains(strings.ToLower(payload.Version), "test") ||
			strings.Contains(strings.ToLower(payload.Version), "alpha") {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "version",
				Message:    fmt.Sprintf("Version '%s' appears to be a development/test version", payload.Version),
				Code:       "DEV_VERSION_IN_PROD",
				Suggestion: "Ensure you're deploying a stable production version",
			})
		}
	}

	return warnings
}
