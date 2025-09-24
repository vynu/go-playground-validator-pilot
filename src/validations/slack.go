// Package validations contains Slack-specific validation logic and business rules.
// This module implements custom validators and business logic for Slack webhook payloads.
package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github-data-validator/models"
	"github.com/go-playground/validator/v10"
)

// SlackValidator provides Slack-specific validation functionality.
type SlackValidator struct {
	validator *validator.Validate
}

// NewSlackValidator creates a new Slack validator instance.
func NewSlackValidator() *SlackValidator {
	v := validator.New()

	// Register Slack-specific custom validators
	v.RegisterValidation("slack_token", validateSlackToken)
	v.RegisterValidation("slack_channel_name", validateSlackChannelName)
	v.RegisterValidation("slack_command", validateSlackCommand)
	v.RegisterValidation("slack_timestamp", validateSlackTimestamp)
	v.RegisterValidation("hexcolor", validateSlackHexColor)

	return &SlackValidator{validator: v}
}

// ValidatePayload validates a Slack webhook payload with comprehensive rules.
func (sv *SlackValidator) ValidatePayload(payload models.SlackMessagePayload) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "SlackMessagePayload",
		Provider:  "slack_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := sv.validator.Struct(payload); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatSlackValidationError(fieldError),
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
		warnings := ValidateSlackBusinessLogic(payload)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "slack_validator",
		FieldCount:         countSlackStructFields(payload),
		RuleCount:          sv.getRuleCount(),
	}

	return result
}

// validateSlackToken validates Slack token format and authenticity.
func validateSlackToken(fl validator.FieldLevel) bool {
	token := fl.Field().String()

	// Slack tokens have specific prefixes and lengths
	patterns := []struct {
		prefix string
		length int
	}{
		{"xoxb-", 56}, // Bot User OAuth Access Token
		{"xoxp-", 72}, // User OAuth Access Token
		{"xapp-", 72}, // App-level token
		{"xoxs-", 56}, // Legacy token
		{"xoxa-", 56}, // App access token
		{"xoxr-", 56}, // Refresh token
	}

	for _, pattern := range patterns {
		if strings.HasPrefix(token, pattern.prefix) {
			return len(token) == pattern.length
		}
	}

	return false
}

// validateSlackChannelName validates Slack channel naming conventions.
func validateSlackChannelName(fl validator.FieldLevel) bool {
	channelName := fl.Field().String()

	// Handle special channel types
	if strings.HasPrefix(channelName, "D") { // Direct message
		return regexp.MustCompile(`^D[A-Z0-9]{8,}$`).MatchString(channelName)
	}
	if strings.HasPrefix(channelName, "G") { // Group message
		return regexp.MustCompile(`^G[A-Z0-9]{8,}$`).MatchString(channelName)
	}

	// Public channel names
	if strings.HasPrefix(channelName, "#") {
		channelName = channelName[1:] // Remove # prefix
	}

	// Channel names must be lowercase, 1-80 characters, letters, numbers, hyphens, underscores
	return regexp.MustCompile(`^[a-z0-9_-]{1,80}$`).MatchString(channelName)
}

// validateSlackCommand validates Slack slash command format.
func validateSlackCommand(fl validator.FieldLevel) bool {
	command := fl.Field().String()

	// Must start with /
	if !strings.HasPrefix(command, "/") {
		return false
	}

	// Remove / and validate remaining
	commandName := command[1:]

	// Command names: 1-32 characters, lowercase letters, numbers, hyphens, underscores
	return regexp.MustCompile(`^[a-z0-9_-]{1,32}$`).MatchString(commandName)
}

// validateSlackTimestamp validates Slack timestamp format (Unix timestamp with microseconds).
func validateSlackTimestamp(fl validator.FieldLevel) bool {
	timestamp := fl.Field().String()

	// Slack timestamps are in format "1234567890.123456"
	parts := strings.Split(timestamp, ".")
	if len(parts) != 2 {
		return false
	}

	// Validate Unix timestamp part (must be reasonable)
	if len(parts[0]) < 10 || len(parts[0]) > 11 {
		return false
	}

	// Validate microseconds part (should be 6 digits)
	if len(parts[1]) != 6 {
		return false
	}

	// Both parts should be numeric
	return regexp.MustCompile(`^[0-9]+$`).MatchString(parts[0]) &&
		regexp.MustCompile(`^[0-9]+$`).MatchString(parts[1])
}

// validateSlackHexColor validates hexadecimal color codes for Slack.
func validateSlackHexColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()

	// Remove # if present
	if strings.HasPrefix(color, "#") {
		color = color[1:]
	}

	// Must be 3 or 6 hex characters
	if len(color) != 3 && len(color) != 6 {
		return false
	}

	return regexp.MustCompile(`^[0-9a-fA-F]+$`).MatchString(color)
}

// ValidateSlackBusinessLogic performs Slack-specific business logic validation.
func ValidateSlackBusinessLogic(payload models.SlackMessagePayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for excessive attachments
	warnings = append(warnings, checkExcessiveAttachments(payload)...)

	// Check for long message content
	warnings = append(warnings, checkLongMessageContent(payload)...)

	// Check for potential token exposure
	warnings = append(warnings, checkTokenExposure(payload)...)

	// Check for excessive mentions
	warnings = append(warnings, checkExcessiveMentions(payload)...)

	// Check for file security concerns
	warnings = append(warnings, checkFileSecurityConcerns(payload)...)

	// Check for spam patterns
	warnings = append(warnings, checkSpamPatterns(payload)...)

	// Check for compliance issues
	warnings = append(warnings, checkComplianceConcerns(payload)...)

	return warnings
}

// checkExcessiveAttachments checks for too many attachments in a message.
func checkExcessiveAttachments(payload models.SlackMessagePayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	attachmentCount := len(payload.Message.Attachments)
	blockCount := len(payload.Message.Blocks)

	if attachmentCount > 20 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Message.Attachments",
			Message:    fmt.Sprintf("Message has %d attachments, which may impact performance", attachmentCount),
			Code:       "EXCESSIVE_ATTACHMENTS",
			Value:      attachmentCount,
			Suggestion: "Consider reducing attachments or using blocks for better performance",
			Category:   "performance",
		})
	}

	if blockCount > 50 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Message.Blocks",
			Message:    fmt.Sprintf("Message has %d blocks, which may impact rendering", blockCount),
			Code:       "EXCESSIVE_BLOCKS",
			Value:      blockCount,
			Suggestion: "Consider simplifying the message structure for better user experience",
			Category:   "performance",
		})
	}

	return warnings
}

// checkLongMessageContent checks for overly long message content.
func checkLongMessageContent(payload models.SlackMessagePayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	messageLength := len(payload.Message.Text)
	if messageLength > 4000 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Message.Text",
			Message:    fmt.Sprintf("Message text is %d characters long, may be truncated", messageLength),
			Code:       "LONG_MESSAGE_TEXT",
			Value:      messageLength,
			Suggestion: "Consider breaking long messages into multiple parts or using attachments",
			Category:   "usability",
		})
	}

	// Check individual attachment text lengths
	for i, attachment := range payload.Message.Attachments {
		if len(attachment.Text) > 8000 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      fmt.Sprintf("Message.Attachments[%d].Text", i),
				Message:    fmt.Sprintf("Attachment text is %d characters, may be truncated", len(attachment.Text)),
				Code:       "LONG_ATTACHMENT_TEXT",
				Value:      len(attachment.Text),
				Suggestion: "Consider using multiple attachments or external links for long content",
				Category:   "usability",
			})
		}
	}

	return warnings
}

// checkTokenExposure checks for potential token exposure in message content.
func checkTokenExposure(payload models.SlackMessagePayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for token patterns in various text fields
	tokenPatterns := []string{"xoxb-", "xoxp-", "xapp-", "xoxs-", "xoxa-", "xoxr-"}

	textFields := []struct {
		field string
		text  string
	}{
		{"Text", payload.Text},
		{"Message.Text", payload.Message.Text},
	}

	// Add attachment texts
	for i, attachment := range payload.Message.Attachments {
		textFields = append(textFields, struct {
			field string
			text  string
		}{
			field: fmt.Sprintf("Message.Attachments[%d].Text", i),
			text:  attachment.Text,
		})
	}

	for _, textField := range textFields {
		for _, pattern := range tokenPatterns {
			if strings.Contains(textField.text, pattern) {
				warnings = append(warnings, models.ValidationWarning{
					Field:      textField.field,
					Message:    "Potential Slack token exposure detected in message content",
					Code:       "POTENTIAL_TOKEN_EXPOSURE",
					Suggestion: "Remove any sensitive tokens from message content",
					Category:   "security",
				})
				break
			}
		}
	}

	return warnings
}

// checkExcessiveMentions checks for too many user mentions.
func checkExcessiveMentions(payload models.SlackMessagePayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	messageText := payload.Message.Text
	mentionCount := strings.Count(messageText, "<@")
	channelMentionCount := strings.Count(messageText, "<!channel>") +
		strings.Count(messageText, "<!here>") +
		strings.Count(messageText, "<!everyone>")

	if mentionCount > 10 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Message.Text",
			Message:    fmt.Sprintf("Message contains %d user mentions, may cause notification spam", mentionCount),
			Code:       "EXCESSIVE_MENTIONS",
			Value:      mentionCount,
			Suggestion: "Consider reducing mentions or using DMs for multiple individual notifications",
			Category:   "etiquette",
		})
	}

	if channelMentionCount > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Message.Text",
			Message:    "Message contains channel-wide mentions (@channel, @here, @everyone)",
			Code:       "CHANNEL_WIDE_MENTION",
			Value:      channelMentionCount,
			Suggestion: "Use channel-wide mentions sparingly to avoid notification fatigue",
			Category:   "etiquette",
		})
	}

	return warnings
}

// checkFileSecurityConcerns checks for potentially problematic file attachments.
func checkFileSecurityConcerns(payload models.SlackMessagePayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for potentially dangerous file types
	dangerousExtensions := []string{"exe", "bat", "cmd", "scr", "pif", "com", "vbs", "js", "jar"}

	// Check for very large files
	maxSafeSize := int64(100 * 1024 * 1024) // 100MB

	for i, file := range payload.Message.Files {
		// Check file type
		for _, ext := range dangerousExtensions {
			if strings.EqualFold(file.Filetype, ext) {
				warnings = append(warnings, models.ValidationWarning{
					Field:      fmt.Sprintf("Message.Files[%d].Filetype", i),
					Message:    fmt.Sprintf("File type '%s' may be blocked by security policies", file.Filetype),
					Code:       "POTENTIALLY_DANGEROUS_FILETYPE",
					Value:      file.Filetype,
					Suggestion: "Consider alternative file formats or compression",
					Category:   "security",
				})
				break
			}
		}

		// Check file size
		if file.Size > maxSafeSize {
			warnings = append(warnings, models.ValidationWarning{
				Field:      fmt.Sprintf("Message.Files[%d].Size", i),
				Message:    fmt.Sprintf("File '%s' is %d bytes, may impact performance", file.Name, file.Size),
				Code:       "LARGE_FILE_SIZE",
				Value:      file.Size,
				Suggestion: "Consider compressing large files or using external file sharing",
				Category:   "performance",
			})
		}

		// Check for potentially sensitive file names
		sensitivePatterns := []string{"password", "secret", "key", "token", "credential", "private"}
		fileNameLower := strings.ToLower(file.Name)

		for _, pattern := range sensitivePatterns {
			if strings.Contains(fileNameLower, pattern) {
				warnings = append(warnings, models.ValidationWarning{
					Field:      fmt.Sprintf("Message.Files[%d].Name", i),
					Message:    "File name suggests it may contain sensitive information",
					Code:       "SENSITIVE_FILENAME",
					Value:      file.Name,
					Suggestion: "Ensure file doesn't contain sensitive data before sharing",
					Category:   "security",
				})
				break
			}
		}
	}

	return warnings
}

// checkSpamPatterns checks for potential spam indicators.
func checkSpamPatterns(payload models.SlackMessagePayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	messageText := strings.ToLower(payload.Message.Text)

	// Check for excessive capitalization
	if len(payload.Message.Text) > 10 {
		upperCount := 0
		for _, r := range payload.Message.Text {
			if r >= 'A' && r <= 'Z' {
				upperCount++
			}
		}
		upperRatio := float64(upperCount) / float64(len(payload.Message.Text))

		if upperRatio > 0.7 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Message.Text",
				Message:    "Message contains excessive capitalization which may be perceived as shouting",
				Code:       "EXCESSIVE_CAPS",
				Value:      upperRatio,
				Suggestion: "Consider using normal capitalization for better readability",
				Category:   "etiquette",
			})
		}
	}

	// Check for repeated characters or words
	if regexp.MustCompile(`(.)\1{4,}`).MatchString(messageText) {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Message.Text",
			Message:    "Message contains repeated characters which may indicate spam",
			Code:       "REPEATED_CHARACTERS",
			Suggestion: "Remove excessive character repetition",
			Category:   "spam",
		})
	}

	// Check for promotional keywords
	promotionalKeywords := []string{"buy now", "limited time", "act fast", "click here", "free money", "winner"}
	foundKeywords := []string{}

	for _, keyword := range promotionalKeywords {
		if strings.Contains(messageText, keyword) {
			foundKeywords = append(foundKeywords, keyword)
		}
	}

	if len(foundKeywords) > 2 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Message.Text",
			Message:    fmt.Sprintf("Message contains promotional keywords: %s", strings.Join(foundKeywords, ", ")),
			Code:       "PROMOTIONAL_CONTENT",
			Suggestion: "Ensure message complies with workspace guidelines for promotional content",
			Category:   "compliance",
		})
	}

	return warnings
}

// checkComplianceConcerns checks for compliance-related issues.
func checkComplianceConcerns(payload models.SlackMessagePayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	messageText := strings.ToLower(payload.Message.Text)

	// Check for PII patterns
	piiPatterns := []struct {
		pattern string
		name    string
	}{
		{`\b\d{3}-\d{2}-\d{4}\b`, "SSN"},
		{`\b\d{4}\s?\d{4}\s?\d{4}\s?\d{4}\b`, "Credit Card"},
		{`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`, "Email"},
		{`\b\d{3}-\d{3}-\d{4}\b`, "Phone Number"},
	}

	for _, pii := range piiPatterns {
		if matched, _ := regexp.MatchString(pii.pattern, payload.Message.Text); matched {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Message.Text",
				Message:    fmt.Sprintf("Message may contain %s which is personally identifiable information (PII)", pii.name),
				Code:       "POTENTIAL_PII",
				Suggestion: "Review message for PII and consider using DMs or secure channels",
				Category:   "compliance",
			})
		}
	}

	// Check for restricted content keywords
	restrictedKeywords := []string{"confidential", "internal only", "classified", "proprietary", "nda"}
	foundRestricted := []string{}

	for _, keyword := range restrictedKeywords {
		if strings.Contains(messageText, keyword) {
			foundRestricted = append(foundRestricted, keyword)
		}
	}

	if len(foundRestricted) > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Message.Text",
			Message:    fmt.Sprintf("Message contains restricted content indicators: %s", strings.Join(foundRestricted, ", ")),
			Code:       "RESTRICTED_CONTENT",
			Suggestion: "Ensure appropriate channel permissions and data handling policies",
			Category:   "compliance",
		})
	}

	return warnings
}

// formatSlackValidationError formats validation errors with Slack-specific context.
func formatSlackValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required for Slack webhook validation", fe.Field())
	case "slack_token":
		return fmt.Sprintf("Field '%s' must be a valid Slack token format", fe.Field())
	case "slack_channel_name":
		return fmt.Sprintf("Field '%s' must be a valid Slack channel name", fe.Field())
	case "slack_command":
		return fmt.Sprintf("Field '%s' must be a valid Slack slash command", fe.Field())
	case "slack_timestamp":
		return fmt.Sprintf("Field '%s' must be a valid Slack timestamp format", fe.Field())
	case "hexcolor":
		return fmt.Sprintf("Field '%s' must be a valid hex color or predefined Slack color", fe.Field())
	case "alphanum":
		return fmt.Sprintf("Field '%s' must contain only alphanumeric characters", fe.Field())
	case "hostname":
		return fmt.Sprintf("Field '%s' must be a valid hostname", fe.Field())
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL format", fe.Field())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", fe.Field(), fe.Param())
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("Field '%s' failed validation: %s", fe.Field(), fe.Tag())
	}
}

// countSlackStructFields counts the number of fields in a Slack struct for metrics.
func countSlackStructFields(payload models.SlackMessagePayload) int {
	// Simplified count - in practice, use reflection for accurate count
	return 45 // Approximate field count for Slack payload
}

// getRuleCount returns the number of validation rules applied.
func (sv *SlackValidator) getRuleCount() int {
	return 30 // Approximate number of validation rules
}
