// Package validations contains GitLab-specific validation logic and business rules.
// This module implements custom validators and business logic for GitLab webhook payloads.
package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github-data-validator/models"
	"github.com/go-playground/validator/v10"
)

// GitLabValidator provides GitLab-specific validation functionality.
type GitLabValidator struct {
	validator *validator.Validate
}

// NewGitLabValidator creates a new GitLab validator instance.
func NewGitLabValidator() *GitLabValidator {
	v := validator.New()

	// Register GitLab-specific custom validators
	v.RegisterValidation("gitlab_username", validateGitLabUsername)
	v.RegisterValidation("hexcolor", validateGitLabHexColor)

	return &GitLabValidator{validator: v}
}

// ValidatePayload validates a GitLab webhook payload with comprehensive rules.
func (gv *GitLabValidator) ValidatePayload(payload models.GitLabPayload) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "GitLabPayload",
		Provider:  "gitlab_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := gv.validator.Struct(payload); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatGitLabValidationError(fieldError),
					Code:       fieldError.Tag(),
					Value:      fieldError.Value(),
					Expected:   fieldError.Param(),
					Constraint: fieldError.Tag(),
					Path:       fieldError.Namespace(),
					Severity:   "error",
				})
			}
		}
	}

	// Perform business logic validation
	if result.IsValid {
		warnings := ValidateGitLabBusinessLogic(payload)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "gitlab_validator",
		FieldCount:         countGitLabStructFields(payload),
		RuleCount:          gv.getRuleCount(),
	}

	return result
}

// validateGitLabUsername validates GitLab username format.
func validateGitLabUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// GitLab username rules:
	// - 1-255 characters
	// - alphanumeric characters, dots, dashes, underscores
	// - cannot start or end with dot, dash, or underscore
	// - cannot have consecutive dots, dashes, or underscores
	if len(username) == 0 || len(username) > 255 {
		return false
	}

	// Check for valid characters and patterns
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9._-])*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`, username)
	if !matched {
		return false
	}

	// Check for consecutive special characters
	if strings.Contains(username, "..") || strings.Contains(username, "--") || strings.Contains(username, "__") {
		return false
	}

	return true
}

// validateGitLabHexColor validates hexadecimal color codes for GitLab.
func validateGitLabHexColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()

	// Remove # if present
	if strings.HasPrefix(color, "#") {
		color = color[1:]
	}

	// Must be 6 hex characters for GitLab API
	if len(color) != 6 {
		return false
	}

	matched, _ := regexp.MatchString(`^[0-9a-fA-F]{6}$`, color)
	return matched
}

// ValidateGitLabBusinessLogic performs GitLab-specific business logic validation.
func ValidateGitLabBusinessLogic(payload models.GitLabPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for WIP (Work in Progress) indicators
	warnings = append(warnings, checkGitLabWIPIndicators(payload)...)

	// Check for large changesets in merge requests
	warnings = append(warnings, checkGitLabLargeChangeset(payload)...)

	// Check for missing description
	warnings = append(warnings, checkGitLabMissingDescription(payload)...)

	// Check for security-related concerns
	warnings = append(warnings, checkGitLabSecurityConcerns(payload)...)

	// Check for project health indicators
	warnings = append(warnings, checkGitLabProjectHealth(payload)...)

	// Check for collaboration patterns
	warnings = append(warnings, checkGitLabCollaborationPatterns(payload)...)

	// Check for merge request patterns
	warnings = append(warnings, checkGitLabMergeRequestPatterns(payload)...)

	return warnings
}

// checkGitLabWIPIndicators checks for work-in-progress indicators in MR title.
func checkGitLabWIPIndicators(payload models.GitLabPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	if payload.ObjectAttributes == nil {
		return warnings
	}

	title := strings.ToLower(payload.ObjectAttributes.Title)
	wipPatterns := []string{
		"wip:",
		"[wip]",
		"work in progress",
		"do not merge",
		"dnm:",
		"[dnm]",
		"draft:",
		"[draft]",
		"temporary",
		"temp:",
		"fixup!",
		"squash!",
	}

	for _, pattern := range wipPatterns {
		if strings.Contains(title, pattern) {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "ObjectAttributes.Title",
				Message:    fmt.Sprintf("Merge request title contains WIP indicator: '%s'", pattern),
				Code:       "WIP_DETECTED",
				Value:      payload.ObjectAttributes.Title,
				Suggestion: "Consider marking as WIP or updating title when ready for review",
				Category:   "workflow",
			})
			break
		}
	}

	// Check if WIP flag conflicts with title
	if payload.ObjectAttributes.WorkInProgress && !strings.Contains(title, "wip") && !strings.Contains(title, "draft") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ObjectAttributes.WorkInProgress",
			Message:    "Merge request is marked as WIP but title doesn't indicate WIP status",
			Code:       "WIP_TITLE_MISMATCH",
			Suggestion: "Consider adding WIP or Draft indicator to title for clarity",
			Category:   "workflow",
		})
	}

	return warnings
}

// checkGitLabLargeChangeset checks for potentially problematic large changesets.
func checkGitLabLargeChangeset(payload models.GitLabPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// This would typically require additional API calls to get changeset information
	// For now, we'll check based on commit count
	commitCount := len(payload.Commits)

	if commitCount > 50 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Commits",
			Message:    fmt.Sprintf("Large number of commits detected: %d commits", commitCount),
			Code:       "LARGE_COMMIT_COUNT",
			Value:      commitCount,
			Suggestion: "Consider breaking large changes into smaller, focused merge requests",
			Category:   "maintainability",
		})
	}

	return warnings
}

// checkGitLabMissingDescription checks for missing or inadequate MR descriptions.
func checkGitLabMissingDescription(payload models.GitLabPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	if payload.ObjectAttributes == nil {
		return warnings
	}

	if payload.ObjectAttributes.Description == nil || len(*payload.ObjectAttributes.Description) < 10 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ObjectAttributes.Description",
			Message:    "Merge request description is missing or too short",
			Code:       "MISSING_DESCRIPTION",
			Suggestion: "Add a detailed description explaining the changes, motivation, and testing approach",
			Category:   "documentation",
		})
	} else {
		description := strings.ToLower(*payload.ObjectAttributes.Description)

		// Check for template sections that might be missing
		templateSections := []string{"summary", "changes", "testing", "checklist"}
		missingSections := []string{}

		for _, section := range templateSections {
			if !strings.Contains(description, section) {
				missingSections = append(missingSections, section)
			}
		}

		if len(missingSections) > 2 {
			warnings = append(warnings, models.ValidationWarning{
				Field: "ObjectAttributes.Description",
				Message: fmt.Sprintf("MR description may be missing standard sections: %s",
					strings.Join(missingSections, ", ")),
				Code:       "INCOMPLETE_TEMPLATE",
				Suggestion: "Consider using the MR template to ensure all necessary information is provided",
				Category:   "documentation",
			})
		}
	}

	return warnings
}

// checkGitLabSecurityConcerns checks for potential security-related issues.
func checkGitLabSecurityConcerns(payload models.GitLabPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	if payload.ObjectAttributes == nil {
		return warnings
	}

	// Check MR title and description for security keywords
	securityKeywords := []string{
		"password", "secret", "key", "token", "credential", "auth",
		"security", "vulnerability", "exploit", "backdoor", "hardcode",
		"api_key", "private_key", "access_token", "secret_key",
	}

	titleLower := strings.ToLower(payload.ObjectAttributes.Title)
	var descriptionLower string
	if payload.ObjectAttributes.Description != nil {
		descriptionLower = strings.ToLower(*payload.ObjectAttributes.Description)
	}

	foundKeywords := []string{}
	for _, keyword := range securityKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(descriptionLower, keyword) {
			foundKeywords = append(foundKeywords, keyword)
		}
	}

	if len(foundKeywords) > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field: "ObjectAttributes.Content",
			Message: fmt.Sprintf("Security-related keywords detected: %s",
				strings.Join(foundKeywords, ", ")),
			Code:       "SECURITY_KEYWORDS",
			Suggestion: "Ensure no sensitive information is exposed and consider security review",
			Category:   "security",
		})
	}

	// Check for potentially dangerous file patterns
	if strings.Contains(titleLower, "dockerfile") ||
		strings.Contains(titleLower, "docker-compose") ||
		strings.Contains(titleLower, ".env") ||
		strings.Contains(titleLower, "config") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ObjectAttributes.Title",
			Message:    "Configuration or deployment files may be modified",
			Code:       "CONFIG_FILE_CHANGES",
			Suggestion: "Review configuration changes carefully and ensure no secrets are exposed",
			Category:   "security",
		})
	}

	return warnings
}

// checkGitLabProjectHealth checks project health indicators.
func checkGitLabProjectHealth(payload models.GitLabPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	if payload.Project == nil {
		return warnings
	}

	project := *payload.Project

	// Check project visibility and security
	if project.VisibilityLevel == 20 { // Public visibility
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Project.VisibilityLevel",
			Message:    "Project is publicly visible",
			Code:       "PUBLIC_PROJECT",
			Suggestion: "Ensure no sensitive information is exposed in public repositories",
			Category:   "security",
		})
	}

	return warnings
}

// checkGitLabCollaborationPatterns checks for collaboration and workflow patterns.
func checkGitLabCollaborationPatterns(payload models.GitLabPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	if payload.ObjectAttributes == nil {
		return warnings
	}

	mr := *payload.ObjectAttributes

	// Check for missing reviewers on significant changes
	if len(payload.Reviewers) == 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Reviewers",
			Message:    "No reviewers assigned to merge request",
			Code:       "NO_REVIEWERS",
			Suggestion: "Request appropriate reviewers for code quality and knowledge sharing",
			Category:   "workflow",
		})
	}

	// Check for long-lived MR (created more than 7 days ago)
	if time.Since(mr.CreatedAt).Hours() > 24*7 {
		warnings = append(warnings, models.ValidationWarning{
			Field: "ObjectAttributes.CreatedAt",
			Message: fmt.Sprintf("Merge request is %d days old",
				int(time.Since(mr.CreatedAt).Hours()/24)),
			Code:       "STALE_MR",
			Suggestion: "Consider rebasing, updating, or closing if no longer relevant",
			Category:   "workflow",
		})
	}

	return warnings
}

// checkGitLabMergeRequestPatterns checks for merge request specific patterns.
func checkGitLabMergeRequestPatterns(payload models.GitLabPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	if payload.ObjectAttributes == nil {
		return warnings
	}

	mr := *payload.ObjectAttributes

	// Check for target branch patterns
	if mr.TargetBranch == "main" || mr.TargetBranch == "master" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ObjectAttributes.TargetBranch",
			Message:    fmt.Sprintf("Merge request targets main branch: %s", mr.TargetBranch),
			Code:       "MAIN_BRANCH_TARGET",
			Suggestion: "Ensure proper review process for changes to main branch",
			Category:   "workflow",
		})
	}

	// Check for squash recommendations
	if !mr.Squash && len(payload.Commits) > 5 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ObjectAttributes.Squash",
			Message:    fmt.Sprintf("Consider squashing %d commits for cleaner history", len(payload.Commits)),
			Code:       "SQUASH_RECOMMENDED",
			Suggestion: "Enable squash option to maintain clean commit history",
			Category:   "maintainability",
		})
	}

	// Check for merge conflicts
	if mr.MergeStatus == "cannot_be_merged" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ObjectAttributes.MergeStatus",
			Message:    "Merge request has conflicts that need resolution",
			Code:       "MERGE_CONFLICTS",
			Suggestion: "Resolve merge conflicts before the merge request can be completed",
			Category:   "workflow",
		})
	}

	return warnings
}

// formatGitLabValidationError formats validation errors with GitLab-specific context.
func formatGitLabValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required for GitLab webhook validation", fe.Field())
	case "gitlab_username":
		return fmt.Sprintf("Field '%s' must be a valid GitLab username (1-255 chars, alphanumeric with dots, dashes, underscores)", fe.Field())
	case "hexcolor":
		return fmt.Sprintf("Field '%s' must be a valid 6-character hexadecimal color code", fe.Field())
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL format", fe.Field())
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", fe.Field())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", fe.Field(), fe.Param())
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("Field '%s' must be greater than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("Field '%s' must be greater than or equal to %s", fe.Field(), fe.Param())
	case "len":
		return fmt.Sprintf("Field '%s' must be exactly %s characters long", fe.Field(), fe.Param())
	case "hexadecimal":
		return fmt.Sprintf("Field '%s' must contain only hexadecimal characters", fe.Field())
	default:
		return fmt.Sprintf("Field '%s' failed validation: %s", fe.Field(), fe.Tag())
	}
}

// countGitLabStructFields counts the number of fields in a GitLab struct for metrics.
func countGitLabStructFields(payload models.GitLabPayload) int {
	// This is a simplified count - in practice, you might use reflection
	// to count all nested fields for more accurate metrics
	return 45 // Approximate field count for GitLab payload
}

// getRuleCount returns the number of validation rules applied.
func (gv *GitLabValidator) getRuleCount() int {
	// Return approximate number of validation rules
	return 30
}
