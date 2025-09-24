// Package validations contains Bitbucket-specific validation logic and business rules.
// This module implements custom validators and business logic for Bitbucket webhook payloads.
package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github-data-validator/models"
	"github.com/go-playground/validator/v10"
)

// BitbucketValidator provides Bitbucket-specific validation functionality.
type BitbucketValidator struct {
	validator *validator.Validate
}

// NewBitbucketValidator creates a new Bitbucket validator instance.
func NewBitbucketValidator() *BitbucketValidator {
	v := validator.New()

	// Register Bitbucket-specific custom validators
	v.RegisterValidation("bitbucket_username", validateBitbucketUsername)

	return &BitbucketValidator{validator: v}
}

// ValidatePayload validates a Bitbucket webhook payload with comprehensive rules.
func (bv *BitbucketValidator) ValidatePayload(payload models.BitbucketPayload) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "BitbucketPayload",
		Provider:  "bitbucket_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := bv.validator.Struct(payload); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatBitbucketValidationError(fieldError),
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
		warnings := ValidateBitbucketBusinessLogic(payload)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "bitbucket_validator",
		FieldCount:         countBitbucketStructFields(payload),
		RuleCount:          bv.getRuleCount(),
	}

	return result
}

// validateBitbucketUsername validates Bitbucket username format.
func validateBitbucketUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Bitbucket username rules:
	// - 1-30 characters
	// - alphanumeric characters, dashes, underscores
	// - cannot start or end with dash or underscore
	// - cannot have consecutive special characters
	if len(username) == 0 || len(username) > 30 {
		return false
	}

	// Check for valid characters and patterns
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9_-])*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`, username)
	if !matched {
		return false
	}

	// Check for consecutive special characters
	if strings.Contains(username, "__") || strings.Contains(username, "--") || strings.Contains(username, "_-") || strings.Contains(username, "-_") {
		return false
	}

	return true
}

// ValidateBitbucketBusinessLogic performs Bitbucket-specific business logic validation.
func ValidateBitbucketBusinessLogic(payload models.BitbucketPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for pull request specific validations
	if payload.PullRequest != nil {
		warnings = append(warnings, checkBitbucketPullRequestPatterns(*payload.PullRequest)...)
		warnings = append(warnings, checkBitbucketCollaborationPatterns(*payload.PullRequest)...)
		warnings = append(warnings, checkBitbucketSecurityConcerns(*payload.PullRequest)...)
	}

	// Check for push specific validations
	if payload.Push != nil {
		warnings = append(warnings, checkBitbucketPushPatterns(*payload.Push)...)
	}

	// Check repository health indicators
	warnings = append(warnings, checkBitbucketRepositoryHealth(payload.Repository)...)

	// Check for comment patterns
	if payload.Comment != nil {
		warnings = append(warnings, checkBitbucketCommentPatterns(*payload.Comment)...)
	}

	return warnings
}

// checkBitbucketPullRequestPatterns checks for pull request specific patterns.
func checkBitbucketPullRequestPatterns(pr models.BitbucketPullRequest) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for WIP indicators in title
	title := strings.ToLower(pr.Title)
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
				Value:      pr.Title,
				Suggestion: "Consider updating title when ready for review",
				Category:   "workflow",
			})
			break
		}
	}

	// Check for missing description
	if len(pr.Description) < 10 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.Description",
			Message:    "Pull request description is missing or too short",
			Code:       "MISSING_DESCRIPTION",
			Suggestion: "Add a detailed description explaining the changes, motivation, and testing approach",
			Category:   "documentation",
		})
	}

	// Check for large task count
	if pr.TaskCount > 20 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.TaskCount",
			Message:    fmt.Sprintf("High number of tasks in pull request: %d", pr.TaskCount),
			Code:       "HIGH_TASK_COUNT",
			Value:      pr.TaskCount,
			Suggestion: "Consider breaking down into smaller, focused pull requests",
			Category:   "maintainability",
		})
	}

	// Check for long-lived PR
	if time.Since(pr.CreatedOn).Hours() > 24*7 {
		warnings = append(warnings, models.ValidationWarning{
			Field: "PullRequest.CreatedOn",
			Message: fmt.Sprintf("Pull request is %d days old",
				int(time.Since(pr.CreatedOn).Hours()/24)),
			Code:       "STALE_PR",
			Suggestion: "Consider rebasing, updating, or closing if no longer relevant",
			Category:   "workflow",
		})
	}

	// Check for branch naming patterns
	sourceBranch := strings.ToLower(pr.Source.Name)
	if strings.HasPrefix(sourceBranch, "master") || strings.HasPrefix(sourceBranch, "main") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.Source.Name",
			Message:    "Pull request source branch appears to be a main branch",
			Code:       "MAIN_BRANCH_SOURCE",
			Suggestion: "Consider using feature branches instead of working directly on main branches",
			Category:   "workflow",
		})
	}

	return warnings
}

// checkBitbucketCollaborationPatterns checks for collaboration and workflow patterns.
func checkBitbucketCollaborationPatterns(pr models.BitbucketPullRequest) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for missing reviewers
	if len(pr.Reviewers) == 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.Reviewers",
			Message:    "No reviewers assigned to pull request",
			Code:       "NO_REVIEWERS",
			Suggestion: "Request appropriate reviewers for code quality and knowledge sharing",
			Category:   "workflow",
		})
	}

	// Check for self-approval patterns
	authorID := pr.Author.AccountID
	for _, reviewer := range pr.Reviewers {
		if reviewer.User.AccountID == authorID && reviewer.Approved {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "PullRequest.Reviewers",
				Message:    "Pull request author has approved their own changes",
				Code:       "SELF_APPROVAL",
				Suggestion: "Consider having other team members review and approve changes",
				Category:   "workflow",
			})
		}
	}

	// Check for insufficient approvals
	approvalCount := 0
	for _, reviewer := range pr.Reviewers {
		if reviewer.Approved {
			approvalCount++
		}
	}

	if approvalCount == 0 && pr.State == "OPEN" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "PullRequest.Reviewers",
			Message:    "Pull request has no approvals yet",
			Code:       "NO_APPROVALS",
			Suggestion: "Ensure adequate review and approval before merging",
			Category:   "workflow",
		})
	}

	return warnings
}

// checkBitbucketSecurityConcerns checks for potential security-related issues.
func checkBitbucketSecurityConcerns(pr models.BitbucketPullRequest) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check PR title and description for security keywords
	securityKeywords := []string{
		"password", "secret", "key", "token", "credential", "auth",
		"security", "vulnerability", "exploit", "backdoor", "hardcode",
		"api_key", "private_key", "access_token", "secret_key",
	}

	titleLower := strings.ToLower(pr.Title)
	descriptionLower := strings.ToLower(pr.Description)

	foundKeywords := []string{}
	for _, keyword := range securityKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(descriptionLower, keyword) {
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

	// Check for potentially dangerous file patterns
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

	return warnings
}

// checkBitbucketPushPatterns checks for push-specific patterns.
func checkBitbucketPushPatterns(push models.BitbucketPush) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for large number of changes
	if len(push.Changes) > 10 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Push.Changes",
			Message:    fmt.Sprintf("Large number of changes in push: %d", len(push.Changes)),
			Code:       "LARGE_PUSH",
			Value:      len(push.Changes),
			Suggestion: "Consider breaking large changes into smaller, focused commits",
			Category:   "maintainability",
		})
	}

	// Check for forced pushes
	for _, change := range push.Changes {
		if change.Forced {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Push.Changes.Forced",
				Message:    "Forced push detected",
				Code:       "FORCED_PUSH",
				Suggestion: "Forced pushes can cause issues for collaborators. Use with caution",
				Category:   "workflow",
			})
			break
		}
	}

	// Check for commits to main branches
	for _, change := range push.Changes {
		if change.New != nil {
			branchName := strings.ToLower(change.New.Name)
			if branchName == "master" || branchName == "main" {
				warnings = append(warnings, models.ValidationWarning{
					Field:      "Push.Changes.New.Name",
					Message:    fmt.Sprintf("Direct push to main branch: %s", change.New.Name),
					Code:       "MAIN_BRANCH_PUSH",
					Suggestion: "Consider using pull requests for changes to main branches",
					Category:   "workflow",
				})
			}
		}
	}

	return warnings
}

// checkBitbucketRepositoryHealth checks repository health indicators.
func checkBitbucketRepositoryHealth(repo models.BitbucketRepository) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check if repository is a fork
	if repo.Name != repo.FullName {
		parts := strings.Split(repo.FullName, "/")
		if len(parts) == 2 && parts[0] != repo.Owner.Username {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Repository.Fork",
				Message:    "Webhook payload is from a forked repository",
				Code:       "FORK_REPOSITORY",
				Suggestion: "Review external contributions carefully for security and quality",
				Category:   "security",
			})
		}
	}

	// Check repository size
	if repo.Size > 1000000 { // Size in KB, so this is ~1GB
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Repository.Size",
			Message:    fmt.Sprintf("Large repository size: %d KB", repo.Size),
			Code:       "LARGE_REPOSITORY",
			Value:      repo.Size,
			Suggestion: "Consider repository cleanup or splitting large repositories",
			Category:   "maintenance",
		})
	}

	// Check for old repositories without recent activity
	if time.Since(repo.UpdatedOn).Hours() > 24*30*6 { // 6 months
		warnings = append(warnings, models.ValidationWarning{
			Field: "Repository.UpdatedOn",
			Message: fmt.Sprintf("Repository hasn't been updated in %d days",
				int(time.Since(repo.UpdatedOn).Hours()/24)),
			Code:       "STALE_REPOSITORY",
			Suggestion: "Consider archiving inactive repositories or updating documentation",
			Category:   "maintenance",
		})
	}

	return warnings
}

// checkBitbucketCommentPatterns checks for comment-specific patterns.
func checkBitbucketCommentPatterns(comment models.BitbucketComment) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for potentially sensitive information in comments
	content := strings.ToLower(comment.Content.Raw)
	sensitivePatterns := []string{
		"password", "secret", "token", "key", "credential",
		"api_key", "private_key", "access_token",
	}

	foundPatterns := []string{}
	for _, pattern := range sensitivePatterns {
		if strings.Contains(content, pattern) {
			foundPatterns = append(foundPatterns, pattern)
		}
	}

	if len(foundPatterns) > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field: "Comment.Content.Raw",
			Message: fmt.Sprintf("Potentially sensitive content in comment: %s",
				strings.Join(foundPatterns, ", ")),
			Code:       "SENSITIVE_COMMENT",
			Suggestion: "Avoid including sensitive information in comments",
			Category:   "security",
		})
	}

	// Check for excessively long comments
	if len(comment.Content.Raw) > 5000 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Comment.Content.Raw",
			Message:    fmt.Sprintf("Very long comment: %d characters", len(comment.Content.Raw)),
			Code:       "LONG_COMMENT",
			Value:      len(comment.Content.Raw),
			Suggestion: "Consider breaking long comments into smaller parts or using external documentation",
			Category:   "maintainability",
		})
	}

	return warnings
}

// formatBitbucketValidationError formats validation errors with Bitbucket-specific context.
func formatBitbucketValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required for Bitbucket webhook validation", fe.Field())
	case "bitbucket_username":
		return fmt.Sprintf("Field '%s' must be a valid Bitbucket username (1-30 chars, alphanumeric with dashes, underscores)", fe.Field())
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

// countBitbucketStructFields counts the number of fields in a Bitbucket struct for metrics.
func countBitbucketStructFields(payload models.BitbucketPayload) int {
	// This is a simplified count - in practice, you might use reflection
	// to count all nested fields for more accurate metrics
	return 55 // Approximate field count for Bitbucket payload
}

// getRuleCount returns the number of validation rules applied.
func (bv *BitbucketValidator) getRuleCount() int {
	// Return approximate number of validation rules
	return 35
}
