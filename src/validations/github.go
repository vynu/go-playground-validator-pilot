// Package validations contains GitHub-specific validation logic and business rules.
// This module implements custom validators and business logic for GitHub webhook payloads.
package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"goplayground-data-validator/models"
)

// GitHubValidator provides GitHub-specific validation functionality.
type GitHubValidator struct {
	validator *validator.Validate
}

// NewGitHubValidator creates a new GitHub validator instance.
func NewGitHubValidator() *GitHubValidator {
	v := validator.New()

	// Register GitHub-specific custom validators
	v.RegisterValidation("github_username", validateGitHubUsername)
	v.RegisterValidation("hexcolor", validateHexColor)

	return &GitHubValidator{validator: v}
}

// ValidatePayload validates a GitHub webhook payload with comprehensive rules.
func (gv *GitHubValidator) ValidatePayload(payload models.GitHubPayload) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "GitHubPayload",
		Provider:  "github_validator",
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
					Message:    formatGitHubValidationError(fieldError),
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
		warnings := ValidateGitHubBusinessLogic(payload)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "github_validator",
		FieldCount:         countStructFields(payload),
		RuleCount:          gv.getRuleCount(),
	}

	return result
}

// validateGitHubUsername validates GitHub username format.
func validateGitHubUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// GitHub username rules:
	// - 1-39 characters
	// - alphanumeric characters and hyphens
	// - cannot start or end with hyphen
	// - cannot have consecutive hyphens
	if len(username) == 0 || len(username) > 39 {
		return false
	}

	// Check for valid characters and patterns
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9]|-(?!-))*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`, username)
	return matched
}

// validateHexColor validates hexadecimal color codes.
func validateHexColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()

	// Remove # if present
	if strings.HasPrefix(color, "#") {
		color = color[1:]
	}

	// Must be 6 hex characters for GitHub API
	if len(color) != 6 {
		return false
	}

	matched, _ := regexp.MatchString(`^[0-9a-fA-F]{6}$`, color)
	return matched
}

// ValidateGitHubBusinessLogic performs GitHub-specific business logic validation.
func ValidateGitHubBusinessLogic(payload models.GitHubPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for WIP (Work in Progress) indicators
	warnings = append(warnings, checkWIPIndicators(payload)...)

	// Check for large changesets
	warnings = append(warnings, checkLargeChangeset(payload)...)

	// Check for missing description
	warnings = append(warnings, checkMissingDescription(payload)...)

	// Check for security-related concerns
	warnings = append(warnings, checkSecurityConcerns(payload)...)

	// Check for repository health indicators
	warnings = append(warnings, checkRepositoryHealth(payload)...)

	// Check for collaboration patterns
	warnings = append(warnings, checkCollaborationPatterns(payload)...)

	return warnings
}

// checkWIPIndicators checks for work-in-progress indicators in PR title.
func checkWIPIndicators(payload models.GitHubPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	title := strings.ToLower(payload.PullRequest.Title)
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
				Field:      "PullRequest.Title",
				Message:    fmt.Sprintf("Pull request title contains WIP indicator: '%s'", pattern),
				Code:       "WIP_DETECTED",
				Value:      payload.PullRequest.Title,
				Suggestion: "Consider marking as draft or updating title when ready for review",
				Category:   "workflow",
			})
			break
		}
	}

	// Check if draft flag conflicts with title
	if payload.PullRequest.Draft && !strings.Contains(title, "wip") && !strings.Contains(title, "draft") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.Draft",
			Message:    "Pull request is marked as draft but title doesn't indicate WIP status",
			Code:       "DRAFT_TITLE_MISMATCH",
			Suggestion: "Consider adding WIP or Draft indicator to title for clarity",
			Category:   "workflow",
		})
	}

	return warnings
}

// checkLargeChangeset checks for potentially problematic large changesets.
func checkLargeChangeset(payload models.GitHubPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	totalChanges := payload.PullRequest.Additions + payload.PullRequest.Deletions
	changedFiles := payload.PullRequest.ChangedFiles

	// Large changeset thresholds
	if totalChanges > 1000 {
		severity := "warning"
		if totalChanges > 5000 {
			severity = "high"
		}

		warnings = append(warnings, models.ValidationWarning{
			Field: "PullRequest.Changes",
			Message: fmt.Sprintf("Large changeset detected: %d total changes across %d files (severity: %s)",
				totalChanges, changedFiles, severity),
			Code:       "LARGE_CHANGESET",
			Value:      totalChanges,
			Suggestion: "Consider breaking large changes into smaller, focused pull requests",
			Category:   "maintainability",
		})
	}

	// Too many files changed
	if changedFiles > 50 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.ChangedFiles",
			Message:    fmt.Sprintf("Many files changed: %d files modified", changedFiles),
			Code:       "MANY_FILES_CHANGED",
			Value:      changedFiles,
			Suggestion: "Consider splitting changes across multiple focused pull requests",
			Category:   "maintainability",
		})
	}

	// High deletion ratio might indicate refactoring
	if totalChanges > 0 && float64(payload.PullRequest.Deletions)/float64(totalChanges) > 0.7 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.Deletions",
			Message:    "High deletion ratio detected - this appears to be a major refactoring",
			Code:       "HIGH_DELETION_RATIO",
			Value:      float64(payload.PullRequest.Deletions) / float64(totalChanges),
			Suggestion: "Ensure thorough testing and consider gradual rollout for major refactors",
			Category:   "risk-assessment",
		})
	}

	return warnings
}

// checkMissingDescription checks for missing or inadequate PR descriptions.
func checkMissingDescription(payload models.GitHubPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	if payload.PullRequest.Body == nil || len(*payload.PullRequest.Body) < 10 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.Body",
			Message:    "Pull request description is missing or too short",
			Code:       "MISSING_DESCRIPTION",
			Suggestion: "Add a detailed description explaining the changes, motivation, and testing approach",
			Category:   "documentation",
		})
	} else {
		body := strings.ToLower(*payload.PullRequest.Body)

		// Check for template sections that might be missing
		templateSections := []string{"summary", "changes", "testing", "checklist"}
		missingSections := []string{}

		for _, section := range templateSections {
			if !strings.Contains(body, section) {
				missingSections = append(missingSections, section)
			}
		}

		if len(missingSections) > 2 {
			warnings = append(warnings, models.ValidationWarning{
				Field: "PullRequest.Body",
				Message: fmt.Sprintf("PR description may be missing standard sections: %s",
					strings.Join(missingSections, ", ")),
				Code:       "INCOMPLETE_TEMPLATE",
				Suggestion: "Consider using the PR template to ensure all necessary information is provided",
				Category:   "documentation",
			})
		}
	}

	return warnings
}

// checkSecurityConcerns checks for potential security-related issues.
func checkSecurityConcerns(payload models.GitHubPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check PR title and body for security keywords
	securityKeywords := []string{
		"password", "secret", "key", "token", "credential", "auth",
		"security", "vulnerability", "exploit", "backdoor", "hardcode",
		"api_key", "private_key", "access_token", "secret_key",
	}

	titleLower := strings.ToLower(payload.PullRequest.Title)
	var bodyLower string
	if payload.PullRequest.Body != nil {
		bodyLower = strings.ToLower(*payload.PullRequest.Body)
	}

	foundKeywords := []string{}
	for _, keyword := range securityKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(bodyLower, keyword) {
			foundKeywords = append(foundKeywords, keyword)
		}
	}

	if len(foundKeywords) > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field: "PullRequest.Content",
			Message: fmt.Sprintf("Security-related keywords detected: %s",
				strings.Join(foundKeywords, ", ")),
			Code:       "SECURITY_KEYWORDS",
			Suggestion: "Ensure no sensitive information is exposed and consider security review",
			Category:   "security",
		})
	}

	// Check for potentially dangerous file patterns in commits
	if payload.PullRequest.ChangedFiles > 0 {
		// This would typically require additional API calls to get file lists
		// For now, we'll check based on common patterns
		if strings.Contains(titleLower, "dockerfile") ||
			strings.Contains(titleLower, "docker-compose") ||
			strings.Contains(titleLower, ".env") ||
			strings.Contains(titleLower, "config") {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "PullRequest.Title",
				Message:    "Configuration or deployment files may be modified",
				Code:       "CONFIG_FILE_CHANGES",
				Suggestion: "Review configuration changes carefully and ensure no secrets are exposed",
				Category:   "security",
			})
		}
	}

	return warnings
}

// checkRepositoryHealth checks repository health indicators.
func checkRepositoryHealth(payload models.GitHubPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	repo := payload.Repository

	// Check for fork-based contributions
	if repo.Fork {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Repository.Fork",
			Message:    "Pull request is from a forked repository",
			Code:       "FORK_CONTRIBUTION",
			Suggestion: "Review external contributions carefully for security and quality",
			Category:   "security",
		})
	}

	// Check repository activity indicators
	if repo.StargazersCount == 0 && repo.ForksCount == 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Repository.Activity",
			Message:    "Repository has no stars or forks, indicating low community engagement",
			Code:       "LOW_ENGAGEMENT",
			Suggestion: "Consider repository visibility and community building strategies",
			Category:   "community",
		})
	}

	// Check for high number of open issues relative to activity
	if repo.OpenIssuesCount > 100 && repo.StargazersCount < 50 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Repository.OpenIssues",
			Message:    "High number of open issues relative to repository popularity",
			Code:       "HIGH_ISSUE_RATIO",
			Suggestion: "Consider issue triage and maintenance practices",
			Category:   "maintenance",
		})
	}

	return warnings
}

// checkCollaborationPatterns checks for collaboration and workflow patterns.
func checkCollaborationPatterns(payload models.GitHubPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	pr := payload.PullRequest

	// Check for self-assigned PRs in team repositories
	if pr.User.Login == payload.Sender.Login && len(pr.Assignees) == 1 {
		if pr.Assignees[0].Login == pr.User.Login {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "PullRequest.Assignees",
				Message:    "Pull request author assigned to their own PR",
				Code:       "SELF_ASSIGNED",
				Suggestion: "Consider assigning to a team member for review",
				Category:   "workflow",
			})
		}
	}

	// Check for missing reviewers on significant changes
	if len(pr.RequestedReviewers) == 0 && len(pr.RequestedTeams) == 0 {
		totalChanges := pr.Additions + pr.Deletions
		if totalChanges > 100 || pr.ChangedFiles > 5 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "PullRequest.Reviewers",
				Message:    "No reviewers requested for significant changes",
				Code:       "NO_REVIEWERS",
				Suggestion: "Request appropriate reviewers for code quality and knowledge sharing",
				Category:   "workflow",
			})
		}
	}

	// Check for long-lived PR (created more than 7 days ago)
	if time.Since(pr.CreatedAt).Hours() > 24*7 {
		warnings = append(warnings, models.ValidationWarning{
			Field: "PullRequest.CreatedAt",
			Message: fmt.Sprintf("Pull request is %d days old",
				int(time.Since(pr.CreatedAt).Hours()/24)),
			Code:       "STALE_PR",
			Suggestion: "Consider rebasing, updating, or closing if no longer relevant",
			Category:   "workflow",
		})
	}

	// Check for excessive commits (might indicate poor commit hygiene)
	if pr.Commits > 20 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.Commits",
			Message:    fmt.Sprintf("Pull request contains %d commits", pr.Commits),
			Code:       "MANY_COMMITS",
			Value:      pr.Commits,
			Suggestion: "Consider squashing related commits for cleaner history",
			Category:   "maintainability",
		})
	}

	return warnings
}

// formatGitHubValidationError formats validation errors with GitHub-specific context.
func formatGitHubValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required for GitHub webhook validation", fe.Field())
	case "github_username":
		return fmt.Sprintf("Field '%s' must be a valid GitHub username (1-39 chars, alphanumeric and hyphens)", fe.Field())
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

// countStructFields counts the number of fields in a struct for metrics.
func countStructFields(payload models.GitHubPayload) int {
	// This is a simplified count - in practice, you might use reflection
	// to count all nested fields for more accurate metrics
	return 50 // Approximate field count for GitHub payload
}

// getRuleCount returns the number of validation rules applied.
func (gv *GitHubValidator) getRuleCount() int {
	// Return approximate number of validation rules
	return 25
}
