# Modular Validation System: Platform Integration Guide

## Overview

This guide provides step-by-step instructions for extending the **modular registry-based validation system** with new platform integrations and custom validation rules. You'll learn how to integrate new webhook platforms, create platform-specific validators, and implement business logic rules within the current modular architecture.

## Table of Contents

1. [Understanding the Flexible Validation Architecture](#understanding-the-flexible-validation-architecture)
2. [Adding a New Custom Model](#adding-a-new-custom-model)
3. [Creating Custom Validation Rules](#creating-custom-validation-rules)
4. [Implementing Business Logic Validators](#implementing-business-logic-validators)
5. [Configuring Validation Profiles](#configuring-validation-profiles)
6. [Adding Provider Support](#adding-provider-support)
7. [Creating Test Data and Tests](#creating-test-data-and-tests)
8. [Integration and Deployment](#integration-and-deployment)
9. [Advanced Customization](#advanced-customization)
10. [Best Practices and Troubleshooting](#best-practices-and-troubleshooting)

## Understanding the Modular Validation Architecture

### Core Components

The **modular registry-based validation system** consists of several interconnected components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Routes   │───▶│  Model Registry │───▶│   Platform      │
│   /validate/*   │    │   (ModelInfo)   │    │   Validators    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Swagger Docs   │    │ Validation      │    │   Business      │
│  (Dynamic)      │    │ Manager         │    │   Logic         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### File Structure

```
src/
├── main.go                      # HTTP server and routes
├── models/
│   ├── github.go               # GitHub webhook models
│   ├── gitlab.go               # GitLab webhook models
│   ├── bitbucket.go           # Bitbucket webhook models
│   └── slack.go                # Slack webhook models (example)
├── validations/
│   ├── github.go               # GitHub validation logic
│   ├── gitlab.go               # GitLab validation logic
│   ├── bitbucket.go           # Bitbucket validation logic
│   └── slack.go                # Slack validation logic (example)
├── registry/
│   ├── model_registry.go       # Model registration system
│   └── validation_manager.go   # Validation orchestration
├── utils.go                    # Utility functions
└── templates/                  # Templates for new platforms
```

### Current Platform Support

The system currently supports these webhook platforms:
- **GitHub**: GitHub webhook payloads (pull requests, issues, etc.)
- **GitLab**: GitLab merge request and issue payloads
- **Bitbucket**: Bitbucket pull request payloads
- **Slack**: Slack webhook payloads (message, events, commands)

## Adding a New Custom Model

### Example: Adding a Slack Message Model

Let's walk through adding a comprehensive Slack message validation model.

#### Step 1: Define the Model Structures

**File**: `src/models.go` - Add to existing file

```go
// Slack Message Model for webhook validation
type SlackMessagePayload struct {
    Token       string              `json:"token" validate:"required,min=10"`
    TeamID      string              `json:"team_id" validate:"required,alphanum"`
    TeamDomain  string              `json:"team_domain" validate:"required,hostname"`
    ChannelID   string              `json:"channel_id" validate:"required,alphanum"`
    ChannelName string              `json:"channel_name" validate:"required,min=1,max=80"`
    UserID      string              `json:"user_id" validate:"required,alphanum"`
    UserName    string              `json:"user_name" validate:"required,min=1,max=21,alphanum"`
    Command     string              `json:"command" validate:"required,startswith=/"`
    Text        string              `json:"text" validate:"omitempty,max=3000"`
    ResponseURL string              `json:"response_url" validate:"required,url"`
    TriggerID   string              `json:"trigger_id" validate:"required"`
    Timestamp   string              `json:"ts" validate:"required"`
    Message     SlackMessage        `json:"message" validate:"omitempty"`
    Event       SlackEvent          `json:"event" validate:"omitempty"`
    Challenge   string              `json:"challenge" validate:"omitempty"` // For URL verification
    Type        string              `json:"type" validate:"required,oneof=url_verification event_callback"`
}

type SlackMessage struct {
    Type        string              `json:"type" validate:"required,oneof=message"`
    Subtype     string              `json:"subtype" validate:"omitempty"`
    Text        string              `json:"text" validate:"required,max=40000"`
    User        string              `json:"user" validate:"required,alphanum"`
    Timestamp   string              `json:"ts" validate:"required"`
    ThreadTS    string              `json:"thread_ts" validate:"omitempty"`
    Channel     string              `json:"channel" validate:"required,alphanum"`
    Attachments []SlackAttachment   `json:"attachments" validate:"omitempty,dive"`
    Blocks      []SlackBlock        `json:"blocks" validate:"omitempty,dive"`
    Files       []SlackFile         `json:"files" validate:"omitempty,dive"`
    Reactions   []SlackReaction     `json:"reactions" validate:"omitempty,dive"`
    Replies     []SlackReply        `json:"replies" validate:"omitempty,dive"`
    IsStarred   bool                `json:"is_starred"`
    Pinned      bool                `json:"pinned_to"`
    Edited      *SlackEditInfo      `json:"edited" validate:"omitempty"`
}

type SlackEvent struct {
    Type           string    `json:"type" validate:"required,oneof=message reaction_added reaction_removed"`
    EventTimestamp string    `json:"event_ts" validate:"required"`
    User           string    `json:"user" validate:"required,alphanum"`
    Item           SlackItem `json:"item" validate:"omitempty"`
    Reaction       string    `json:"reaction" validate:"omitempty"`
}

type SlackAttachment struct {
    ID            int64             `json:"id" validate:"omitempty,gt=0"`
    Color         string            `json:"color" validate:"omitempty,hexcolor|oneof=good warning danger"`
    Fallback      string            `json:"fallback" validate:"omitempty,max=300"`
    AuthorName    string            `json:"author_name" validate:"omitempty,max=100"`
    AuthorLink    string            `json:"author_link" validate:"omitempty,url"`
    AuthorIcon    string            `json:"author_icon" validate:"omitempty,url"`
    Title         string            `json:"title" validate:"omitempty,max=300"`
    TitleLink     string            `json:"title_link" validate:"omitempty,url"`
    Text          string            `json:"text" validate:"omitempty,max=8000"`
    ImageURL      string            `json:"image_url" validate:"omitempty,url"`
    ThumbURL      string            `json:"thumb_url" validate:"omitempty,url"`
    Footer        string            `json:"footer" validate:"omitempty,max=300"`
    FooterIcon    string            `json:"footer_icon" validate:"omitempty,url"`
    Timestamp     int64             `json:"ts" validate:"omitempty,gt=0"`
    Fields        []SlackField      `json:"fields" validate:"omitempty,dive,max=10"`
    Actions       []SlackAction     `json:"actions" validate:"omitempty,dive,max=5"`
}

type SlackBlock struct {
    Type     string                 `json:"type" validate:"required,oneof=section divider image actions context"`
    Text     *SlackBlockText        `json:"text" validate:"omitempty"`
    Fields   []SlackBlockText       `json:"fields" validate:"omitempty,dive,max=10"`
    Elements []SlackBlockElement    `json:"elements" validate:"omitempty,dive,max=10"`
    BlockID  string                 `json:"block_id" validate:"omitempty,max=255"`
}

type SlackBlockText struct {
    Type  string `json:"type" validate:"required,oneof=plain_text mrkdwn"`
    Text  string `json:"text" validate:"required,max=3000"`
    Emoji bool   `json:"emoji"`
}

type SlackBlockElement struct {
    Type     string `json:"type" validate:"required,oneof=button static_select external_select"`
    Text     string `json:"text" validate:"omitempty,max=75"`
    Value    string `json:"value" validate:"omitempty,max=2000"`
    URL      string `json:"url" validate:"omitempty,url"`
    ActionID string `json:"action_id" validate:"omitempty,max=255"`
}

type SlackField struct {
    Title string `json:"title" validate:"required,max=300"`
    Value string `json:"value" validate:"required,max=2000"`
    Short bool   `json:"short"`
}

type SlackAction struct {
    Name    string            `json:"name" validate:"required,max=300"`
    Text    string            `json:"text" validate:"required,max=300"`
    Type    string            `json:"type" validate:"required,oneof=button select"`
    Value   string            `json:"value" validate:"omitempty,max=2000"`
    Style   string            `json:"style" validate:"omitempty,oneof=default primary danger"`
    URL     string            `json:"url" validate:"omitempty,url"`
    Options []SlackActionOption `json:"options" validate:"omitempty,dive,max=100"`
}

type SlackActionOption struct {
    Text  string `json:"text" validate:"required,max=300"`
    Value string `json:"value" validate:"required,max=2000"`
}

type SlackFile struct {
    ID                string `json:"id" validate:"required,alphanum"`
    Name              string `json:"name" validate:"required,max=255"`
    Title             string `json:"title" validate:"omitempty,max=255"`
    Mimetype          string `json:"mimetype" validate:"required"`
    Filetype          string `json:"filetype" validate:"required"`
    PrettyType        string `json:"pretty_type" validate:"omitempty"`
    User              string `json:"user" validate:"required,alphanum"`
    Mode              string `json:"mode" validate:"omitempty,oneof=hosted external snippet post"`
    Editable          bool   `json:"editable"`
    IsExternal        bool   `json:"is_external"`
    ExternalType      string `json:"external_type" validate:"omitempty"`
    Size              int64  `json:"size" validate:"gte=0"`
    URLPrivate        string `json:"url_private" validate:"omitempty,url"`
    URLPrivateDownload string `json:"url_private_download" validate:"omitempty,url"`
    Permalink         string `json:"permalink" validate:"omitempty,url"`
    EditLink          string `json:"edit_link" validate:"omitempty,url"`
    Preview           string `json:"preview" validate:"omitempty"`
    PreviewHighlight  string `json:"preview_highlight" validate:"omitempty"`
    Lines             int    `json:"lines" validate:"gte=0"`
    LinesMore         int    `json:"lines_more" validate:"gte=0"`
    IsPublic          bool   `json:"is_public"`
    PublicURLShared   bool   `json:"public_url_shared"`
    Channels          []string `json:"channels" validate:"omitempty,dive,alphanum"`
    Groups            []string `json:"groups" validate:"omitempty,dive,alphanum"`
    IMs               []string `json:"ims" validate:"omitempty,dive,alphanum"`
    InitialComment    SlackComment `json:"initial_comment" validate:"omitempty"`
    CommentsCount     int    `json:"comments_count" validate:"gte=0"`
}

type SlackComment struct {
    ID        string `json:"id" validate:"required,alphanum"`
    Timestamp string `json:"timestamp" validate:"required"`
    User      string `json:"user" validate:"required,alphanum"`
    Comment   string `json:"comment" validate:"required,max=8000"`
}

type SlackReaction struct {
    Name  string   `json:"name" validate:"required,max=100"`
    Count int      `json:"count" validate:"required,gt=0"`
    Users []string `json:"users" validate:"required,dive,alphanum"`
}

type SlackReply struct {
    User      string `json:"user" validate:"required,alphanum"`
    Timestamp string `json:"ts" validate:"required"`
}

type SlackEditInfo struct {
    User      string `json:"user" validate:"required,alphanum"`
    Timestamp string `json:"ts" validate:"required"`
}

type SlackItem struct {
    Type    string `json:"type" validate:"required,oneof=message file comment"`
    Channel string `json:"channel" validate:"required,alphanum"`
    TS      string `json:"ts" validate:"required"`
    File    string `json:"file" validate:"omitempty,alphanum"`
}
```

#### Step 2: Register the Model

**File**: `src/validation_registry.go` - Add to existing registrations

```go
func init() {
    // Register existing models...

    // Register new Slack model
    RegisterModel("SlackMessagePayload", SlackMessagePayload{})
    RegisterModel("SlackMessage", SlackMessage{})
    RegisterModel("SlackEvent", SlackEvent{})
    RegisterModel("SlackAttachment", SlackAttachment{})

    // Register custom validators for Slack
    registerSlackValidators()
}

func registerSlackValidators() {
    // Register custom validation functions
    validate.RegisterValidation("slack_timestamp", validateSlackTimestamp)
    validate.RegisterValidation("slack_channel_name", validateSlackChannelName)
    validate.RegisterValidation("slack_command", validateSlackCommand)
    validate.RegisterValidation("slack_token", validateSlackToken)
    validate.RegisterValidation("hexcolor", validateHexColor)
}
```

#### Step 3: Add Custom Validation Functions

**File**: `src/model_validators.go` - Add new validators

```go
// Custom validators for Slack models

// validateSlackTimestamp validates Slack timestamp format (Unix timestamp with microseconds)
func validateSlackTimestamp(fl validator.FieldLevel) bool {
    timestamp := fl.Field().String()

    // Slack timestamps are in format "1234567890.123456"
    parts := strings.Split(timestamp, ".")
    if len(parts) != 2 {
        return false
    }

    // Validate Unix timestamp part
    if _, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
        return false
    }

    // Validate microseconds part (should be 6 digits)
    if len(parts[1]) != 6 {
        return false
    }

    if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
        return false
    }

    return true
}

// validateSlackChannelName validates Slack channel naming conventions
func validateSlackChannelName(fl validator.FieldLevel) bool {
    channelName := fl.Field().String()

    // Channel names must be lowercase
    if channelName != strings.ToLower(channelName) {
        return false
    }

    // Must start with # for public channels or be alphanumeric for private
    if strings.HasPrefix(channelName, "#") {
        channelName = channelName[1:] // Remove # prefix
    }

    // Must be 1-80 characters, lowercase letters, numbers, hyphens, and underscores
    matched, _ := regexp.MatchString(`^[a-z0-9_-]{1,80}$`, channelName)
    return matched
}

// validateSlackCommand validates Slack slash command format
func validateSlackCommand(fl validator.FieldLevel) bool {
    command := fl.Field().String()

    // Must start with /
    if !strings.HasPrefix(command, "/") {
        return false
    }

    // Remove / and validate remaining
    commandName := command[1:]

    // Command names: 1-32 characters, lowercase letters, numbers, hyphens, underscores
    matched, _ := regexp.MatchString(`^[a-z0-9_-]{1,32}$`, commandName)
    return matched
}

// validateSlackToken validates Slack token format
func validateSlackToken(fl validator.FieldLevel) bool {
    token := fl.Field().String()

    // Slack tokens have specific prefixes and lengths
    patterns := []struct {
        prefix string
        length int
    }{
        {"xoxb-", 56},   // Bot User OAuth Access Token
        {"xoxp-", 72},   // User OAuth Access Token
        {"xapp-", 72},   // App-level token
        {"xoxs-", 56},   // Legacy token
    }

    for _, pattern := range patterns {
        if strings.HasPrefix(token, pattern.prefix) {
            return len(token) == pattern.length
        }
    }

    return false
}

// validateHexColor validates hex color codes
func validateHexColor(fl validator.FieldLevel) bool {
    color := fl.Field().String()

    // Remove # if present
    if strings.HasPrefix(color, "#") {
        color = color[1:]
    }

    // Must be 3 or 6 hex characters
    if len(color) != 3 && len(color) != 6 {
        return false
    }

    matched, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, color)
    return matched
}

// Business logic validators for Slack messages
func validateSlackMessageBusinessLogic(payload SlackMessagePayload) []ValidationWarning {
    var warnings []ValidationWarning

    // Check for excessive attachment count
    if len(payload.Message.Attachments) > 20 {
        warnings = append(warnings, ValidationWarning{
            Field:   "Message.Attachments",
            Message: fmt.Sprintf("Message has %d attachments, consider reducing for better performance", len(payload.Message.Attachments)),
            Code:    "EXCESSIVE_ATTACHMENTS",
        })
    }

    // Check for overly long message text
    if len(payload.Message.Text) > 4000 {
        warnings = append(warnings, ValidationWarning{
            Field:   "Message.Text",
            Message: "Message text is very long, consider using attachments or breaking into multiple messages",
            Code:    "LONG_MESSAGE_TEXT",
        })
    }

    // Check for suspicious token patterns (security)
    if strings.Contains(payload.Text, "xoxb-") || strings.Contains(payload.Text, "xoxp-") {
        warnings = append(warnings, ValidationWarning{
            Field:   "Text",
            Message: "Potential token exposure detected in message text",
            Code:    "POTENTIAL_TOKEN_EXPOSURE",
        })
    }

    // Check for excessive mentions
    mentionCount := strings.Count(payload.Message.Text, "<@")
    if mentionCount > 10 {
        warnings = append(warnings, ValidationWarning{
            Field:   "Message.Text",
            Message: fmt.Sprintf("Message contains %d user mentions, consider reducing to avoid notification spam", mentionCount),
            Code:    "EXCESSIVE_MENTIONS",
        })
    }

    // Check for file size limits
    for i, file := range payload.Message.Files {
        if file.Size > 1024*1024*100 { // 100MB
            warnings = append(warnings, ValidationWarning{
                Field:   fmt.Sprintf("Message.Files[%d].Size", i),
                Message: fmt.Sprintf("File '%s' is %d bytes, consider compressing large files", file.Name, file.Size),
                Code:    "LARGE_FILE_SIZE",
            })
        }
    }

    // Check for blocked file types
    blockedTypes := []string{"exe", "bat", "cmd", "scr", "pif", "com"}
    for i, file := range payload.Message.Files {
        for _, blockedType := range blockedTypes {
            if strings.EqualFold(file.Filetype, blockedType) {
                warnings = append(warnings, ValidationWarning{
                    Field:   fmt.Sprintf("Message.Files[%d].Filetype", i),
                    Message: fmt.Sprintf("File type '%s' may be blocked by security policies", file.Filetype),
                    Code:    "POTENTIALLY_BLOCKED_FILETYPE",
                })
                break
            }
        }
    }

    return warnings
}
```

#### Step 4: Add HTTP Routes

**File**: `src/flexible_server.go` - Add to route setup

```go
func setupFlexibleRoutes(engine *FlexibleValidationEngine) *http.ServeMux {
    mux := http.NewServeMux()

    // Existing routes...

    // Add new Slack-specific routes
    mux.HandleFunc("POST /validate/slack", withMiddleware(func(w http.ResponseWriter, r *http.Request) {
        handleModelValidation(w, r, engine, "SlackMessagePayload")
    }))

    mux.HandleFunc("POST /validate/slack/command", withMiddleware(func(w http.ResponseWriter, r *http.Request) {
        handleSlackCommandValidation(w, r, engine)
    }))

    mux.HandleFunc("POST /validate/slack/event", withMiddleware(func(w http.ResponseWriter, r *http.Request) {
        handleSlackEventValidation(w, r, engine)
    }))

    // Existing routes continue...

    return mux
}

// Specialized handler for Slack command validation
func handleSlackCommandValidation(w http.ResponseWriter, r *http.Request, engine *FlexibleValidationEngine) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        sendError(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Slack commands can be sent as form-encoded data
    var payload SlackMessagePayload

    contentType := r.Header.Get("Content-Type")
    if strings.Contains(contentType, "application/x-www-form-urlencoded") {
        // Parse form data and convert to JSON
        if err := r.ParseForm(); err != nil {
            sendError(w, "Failed to parse form data", http.StatusBadRequest)
            return
        }

        // Convert form values to struct
        payload = SlackMessagePayload{
            Token:       r.FormValue("token"),
            TeamID:      r.FormValue("team_id"),
            TeamDomain:  r.FormValue("team_domain"),
            ChannelID:   r.FormValue("channel_id"),
            ChannelName: r.FormValue("channel_name"),
            UserID:      r.FormValue("user_id"),
            UserName:    r.FormValue("user_name"),
            Command:     r.FormValue("command"),
            Text:        r.FormValue("text"),
            ResponseURL: r.FormValue("response_url"),
            TriggerID:   r.FormValue("trigger_id"),
            Type:        "command",
        }
    } else {
        // Parse as JSON
        if err := json.Unmarshal(body, &payload); err != nil {
            sendError(w, "Invalid JSON payload", http.StatusBadRequest)
            return
        }
    }

    // Validate using the flexible engine
    result := engine.ValidateModel(payload, "SlackMessagePayload", "default")

    // Add Slack-specific business logic
    if result.IsValid {
        warnings := validateSlackMessageBusinessLogic(payload)
        result.Warnings = append(result.Warnings, warnings...)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)
}

// Specialized handler for Slack event validation
func handleSlackEventValidation(w http.ResponseWriter, r *http.Request, engine *FlexibleValidationEngine) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        sendError(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    var payload SlackMessagePayload
    if err := json.Unmarshal(body, &payload); err != nil {
        sendError(w, "Invalid JSON payload", http.StatusBadRequest)
        return
    }

    // Handle Slack URL verification challenge
    if payload.Type == "url_verification" && payload.Challenge != "" {
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(payload.Challenge))
        return
    }

    // Validate the event payload
    result := engine.ValidateModel(payload, "SlackMessagePayload", "strict")

    // Add event-specific business logic
    if result.IsValid {
        warnings := validateSlackMessageBusinessLogic(payload)
        result.Warnings = append(result.Warnings, warnings...)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)
}
```

## Creating Custom Validation Rules

### Step 1: Define Custom Validators

**File**: `src/model_validators.go` - Add domain-specific validators

```go
// Example: Custom validator for email domains
func validateCorporateEmail(fl validator.FieldLevel) bool {
    email := fl.Field().String()

    allowedDomains := []string{
        "company.com",
        "subsidiary.com",
        "partner.org",
    }

    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return false
    }

    domain := strings.ToLower(parts[1])
    for _, allowedDomain := range allowedDomains {
        if domain == allowedDomain {
            return true
        }
    }

    return false
}

// Example: Custom validator for project codes
func validateProjectCode(fl validator.FieldLevel) bool {
    code := fl.Field().String()

    // Project codes: 3-letter department + 4-digit number + optional suffix
    // Examples: ENG-1234, MKT-5678-A, FIN-9999-BETA
    pattern := `^[A-Z]{3}-\d{4}(-[A-Z0-9]+)?$`
    matched, _ := regexp.MatchString(pattern, code)
    return matched
}

// Example: Custom validator for priority levels
func validatePriorityLevel(fl validator.FieldLevel) bool {
    priority := fl.Field().String()

    validPriorities := []string{
        "critical", "high", "medium", "low", "trivial",
        "p0", "p1", "p2", "p3", "p4",
    }

    priority = strings.ToLower(priority)
    for _, valid := range validPriorities {
        if priority == valid {
            return true
        }
    }

    return false
}

// Example: Custom validator for semantic version
func validateSemanticVersion(fl validator.FieldLevel) bool {
    version := fl.Field().String()

    // Semantic version: major.minor.patch with optional pre-release and build metadata
    // Examples: 1.0.0, 2.1.3-alpha.1, 1.0.0-beta.2+build.123
    pattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
    matched, _ := regexp.MatchString(pattern, version)
    return matched
}

// Register all custom validators
func registerCustomValidators() {
    validate.RegisterValidation("corporate_email", validateCorporateEmail)
    validate.RegisterValidation("project_code", validateProjectCode)
    validate.RegisterValidation("priority_level", validatePriorityLevel)
    validate.RegisterValidation("semver", validateSemanticVersion)
}
```

### Step 2: Create Cross-Field Validators

```go
// Example: Cross-field validation for date ranges
func validateDateRange(sl validator.StructLevel) {
    event := sl.Current().Interface().(Event)

    // End date must be after start date
    if !event.EndDate.IsZero() && !event.StartDate.IsZero() {
        if event.EndDate.Before(event.StartDate) {
            sl.ReportError(event.EndDate, "EndDate", "EndDate", "after_start_date", "")
        }
    }

    // Event duration cannot exceed maximum allowed
    if !event.EndDate.IsZero() && !event.StartDate.IsZero() {
        duration := event.EndDate.Sub(event.StartDate)
        maxDuration := 24 * time.Hour * 30 // 30 days

        if duration > maxDuration {
            sl.ReportError(event.EndDate, "EndDate", "EndDate", "max_duration", "")
        }
    }
}

// Example: Business logic validation for budget constraints
func validateBudgetConstraints(sl validator.StructLevel) {
    project := sl.Current().Interface().(Project)

    // Total allocated budget cannot exceed project budget
    var totalAllocated float64
    for _, allocation := range project.BudgetAllocations {
        totalAllocated += allocation.Amount
    }

    if totalAllocated > project.TotalBudget {
        sl.ReportError(project.BudgetAllocations, "BudgetAllocations", "BudgetAllocations", "exceeds_total_budget", "")
    }

    // Each allocation must have valid approval if over threshold
    approvalThreshold := 10000.0
    for i, allocation := range project.BudgetAllocations {
        if allocation.Amount > approvalThreshold && allocation.ApprovalID == "" {
            sl.ReportError(allocation.ApprovalID,
                fmt.Sprintf("BudgetAllocations[%d].ApprovalID", i),
                "ApprovalID", "requires_approval", "")
        }
    }
}

// Register struct-level validators
func registerStructValidators() {
    validate.RegisterStructValidation(validateDateRange, Event{})
    validate.RegisterStructValidation(validateBudgetConstraints, Project{})
}
```

## Implementing Business Logic Validators

### Step 1: Create Domain-Specific Business Rules

**File**: `src/model_validators.go` - Add business logic functions

```go
// Business logic validator for project management
func validateProjectBusinessLogic(project Project) []ValidationWarning {
    var warnings []ValidationWarning

    // Check for overallocated team members
    memberWorkload := make(map[string]float64)
    for _, allocation := range project.ResourceAllocations {
        memberWorkload[allocation.MemberID] += allocation.Percentage
    }

    for memberID, workload := range memberWorkload {
        if workload > 100.0 {
            warnings = append(warnings, ValidationWarning{
                Field:   "ResourceAllocations",
                Message: fmt.Sprintf("Team member %s is allocated %.1f%% (over 100%%)", memberID, workload),
                Code:    "OVERALLOCATED_RESOURCE",
            })
        }
    }

    // Check for unrealistic deadlines
    if !project.Deadline.IsZero() {
        timeUntilDeadline := time.Until(project.Deadline)
        estimatedDuration := time.Duration(project.EstimatedHours) * time.Hour

        if timeUntilDeadline < estimatedDuration/5 { // Less than 20% of estimated time
            warnings = append(warnings, ValidationWarning{
                Field:   "Deadline",
                Message: "Deadline may be unrealistic given estimated work hours",
                Code:    "UNREALISTIC_DEADLINE",
            })
        }
    }

    // Check for missing dependencies
    if len(project.Dependencies) == 0 && project.Complexity == "high" {
        warnings = append(warnings, ValidationWarning{
            Field:   "Dependencies",
            Message: "High complexity projects typically have dependencies",
            Code:    "MISSING_DEPENDENCIES",
        })
    }

    // Check for budget vs. estimated cost discrepancy
    if project.EstimatedCost > 0 && project.TotalBudget > 0 {
        discrepancy := math.Abs(project.EstimatedCost-project.TotalBudget) / project.TotalBudget
        if discrepancy > 0.2 { // More than 20% difference
            warnings = append(warnings, ValidationWarning{
                Field:   "EstimatedCost",
                Message: fmt.Sprintf("Estimated cost (%.2f) differs significantly from budget (%.2f)",
                    project.EstimatedCost, project.TotalBudget),
                Code:    "BUDGET_ESTIMATION_MISMATCH",
            })
        }
    }

    return warnings
}

// Business logic validator for user account management
func validateUserAccountBusinessLogic(user UserAccount) []ValidationWarning {
    var warnings []ValidationWarning

    // Check for suspicious account patterns
    if user.LoginCount == 0 && user.CreatedAt.Before(time.Now().AddDate(0, -1, 0)) {
        warnings = append(warnings, ValidationWarning{
            Field:   "LoginCount",
            Message: "Account created over a month ago but never used",
            Code:    "INACTIVE_ACCOUNT",
        })
    }

    // Check for excessive failed login attempts
    if user.FailedLoginAttempts > 10 {
        warnings = append(warnings, ValidationWarning{
            Field:   "FailedLoginAttempts",
            Message: fmt.Sprintf("Account has %d failed login attempts", user.FailedLoginAttempts),
            Code:    "EXCESSIVE_FAILED_LOGINS",
        })
    }

    // Check for password age
    if !user.PasswordChangedAt.IsZero() {
        passwordAge := time.Since(user.PasswordChangedAt)
        if passwordAge > 90*24*time.Hour { // 90 days
            warnings = append(warnings, ValidationWarning{
                Field:   "PasswordChangedAt",
                Message: "Password is older than 90 days, consider requiring update",
                Code:    "OLD_PASSWORD",
            })
        }
    }

    // Check for role escalation patterns
    if len(user.Roles) > 5 {
        warnings = append(warnings, ValidationWarning{
            Field:   "Roles",
            Message: fmt.Sprintf("User has %d roles, review for least privilege principle", len(user.Roles)),
            Code:    "EXCESSIVE_ROLES",
        })
    }

    // Check for admin roles without MFA
    hasAdminRole := false
    for _, role := range user.Roles {
        if strings.Contains(strings.ToLower(role), "admin") {
            hasAdminRole = true
            break
        }
    }

    if hasAdminRole && !user.MFAEnabled {
        warnings = append(warnings, ValidationWarning{
            Field:   "MFAEnabled",
            Message: "Admin accounts should have multi-factor authentication enabled",
            Code:    "ADMIN_WITHOUT_MFA",
        })
    }

    return warnings
}
```

### Step 2: Create Configurable Business Rules

**File**: `src/config/validation_rules.json` - Add business rule configurations

```json
{
  "models": {
    "SlackMessagePayload": {
      "description": "Slack message webhook validation",
      "fields": {
        "token": {
          "required": true,
          "type": "string",
          "validation": "required,slack_token",
          "description": "Slack authentication token"
        },
        "team_id": {
          "required": true,
          "type": "string",
          "validation": "required,alphanum",
          "description": "Slack team ID"
        },
        "channel_name": {
          "required": true,
          "type": "string",
          "validation": "required,slack_channel_name",
          "description": "Slack channel name"
        },
        "command": {
          "required": false,
          "type": "string",
          "validation": "omitempty,slack_command",
          "description": "Slash command if applicable"
        }
      },
      "business_rules": {
        "attachment_limit": {
          "enabled": true,
          "max_attachments": 20,
          "warning_message": "Consider reducing attachments for better performance"
        },
        "message_length_check": {
          "enabled": true,
          "max_length": 4000,
          "warning_message": "Long messages may be truncated"
        },
        "token_exposure_detection": {
          "enabled": true,
          "patterns": ["xoxb-", "xoxp-", "xapp-"],
          "warning_message": "Potential token exposure detected"
        },
        "mention_limit": {
          "enabled": true,
          "max_mentions": 10,
          "warning_message": "Excessive mentions may cause notification spam"
        },
        "file_size_check": {
          "enabled": true,
          "max_size_bytes": 104857600,
          "warning_message": "Large files may impact performance"
        },
        "blocked_file_types": {
          "enabled": true,
          "blocked_extensions": ["exe", "bat", "cmd", "scr", "pif", "com"],
          "warning_message": "File type may be blocked by security policies"
        }
      }
    },
    "Project": {
      "description": "Project management validation",
      "fields": {
        "name": {
          "required": true,
          "type": "string",
          "validation": "required,min=3,max=100",
          "description": "Project name"
        },
        "code": {
          "required": true,
          "type": "string",
          "validation": "required,project_code",
          "description": "Project code in format XXX-1234"
        },
        "priority": {
          "required": true,
          "type": "string",
          "validation": "required,priority_level",
          "description": "Project priority level"
        }
      },
      "business_rules": {
        "resource_allocation_check": {
          "enabled": true,
          "max_allocation_percent": 100,
          "warning_threshold_percent": 90,
          "warning_message": "Resource allocation approaching maximum"
        },
        "deadline_realism_check": {
          "enabled": true,
          "minimum_time_ratio": 0.2,
          "warning_message": "Deadline may be unrealistic"
        },
        "budget_variance_check": {
          "enabled": true,
          "max_variance_percent": 20,
          "warning_message": "Budget and estimated cost have significant variance"
        },
        "dependency_check": {
          "enabled": true,
          "high_complexity_requires_dependencies": true,
          "warning_message": "High complexity projects typically have dependencies"
        }
      }
    }
  },
  "validation_profiles": {
    "slack_strict": {
      "description": "Strict validation for production Slack integrations",
      "enabled_rules": ["all"],
      "business_rules": {
        "attachment_limit": {"max_attachments": 10},
        "message_length_check": {"max_length": 2000},
        "mention_limit": {"max_mentions": 5}
      }
    },
    "slack_permissive": {
      "description": "Permissive validation for development Slack testing",
      "enabled_rules": ["required_only"],
      "business_rules": {
        "attachment_limit": {"max_attachments": 50},
        "message_length_check": {"max_length": 8000},
        "mention_limit": {"max_mentions": 25}
      }
    },
    "project_enterprise": {
      "description": "Enterprise project validation with strict governance",
      "enabled_rules": ["all"],
      "business_rules": {
        "resource_allocation_check": {"max_allocation_percent": 85},
        "deadline_realism_check": {"minimum_time_ratio": 0.3},
        "budget_variance_check": {"max_variance_percent": 10}
      }
    }
  },
  "global_settings": {
    "enable_business_logic": true,
    "enable_warnings": true,
    "enable_caching": true,
    "cache_ttl_seconds": 300,
    "enable_detailed_errors": true,
    "max_validation_time_ms": 1000,
    "enable_performance_metrics": true
  }
}
```

## Configuring Validation Profiles

### Step 1: Create Profile-Specific Configurations

**File**: `src/flexible_engine.go` - Enhance profile support

```go
// ValidationProfile defines different validation strictness levels
type ValidationProfile struct {
    Name                string                 `json:"name"`
    Description         string                 `json:"description"`
    EnabledRules        []string               `json:"enabled_rules"`
    BusinessRules       map[string]interface{} `json:"business_rules"`
    ErrorThreshold      int                    `json:"error_threshold"`
    WarningThreshold    int                    `json:"warning_threshold"`
    TimeoutMS           int                    `json:"timeout_ms"`
    EnableCaching       bool                   `json:"enable_caching"`
    EnableMetrics       bool                   `json:"enable_metrics"`
    CustomValidators    []string               `json:"custom_validators"`
    FailFast            bool                   `json:"fail_fast"`
    DetailedErrors      bool                   `json:"detailed_errors"`
}

// LoadValidationProfile loads a specific validation profile
func (engine *FlexibleValidationEngine) LoadValidationProfile(profileName string) (*ValidationProfile, error) {
    configData, err := os.ReadFile("config/validation_rules.json")
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var config struct {
        ValidationProfiles map[string]ValidationProfile `json:"validation_profiles"`
    }

    if err := json.Unmarshal(configData, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }

    profile, exists := config.ValidationProfiles[profileName]
    if !exists {
        return nil, fmt.Errorf("validation profile '%s' not found", profileName)
    }

    return &profile, nil
}

// ApplyValidationProfile applies profile-specific settings to validation
func (engine *FlexibleValidationEngine) ApplyValidationProfile(profile *ValidationProfile, modelName string) error {
    // Configure timeout
    if profile.TimeoutMS > 0 {
        engine.timeout = time.Duration(profile.TimeoutMS) * time.Millisecond
    }

    // Configure caching
    engine.enableCaching = profile.EnableCaching

    // Configure metrics collection
    engine.enableMetrics = profile.EnableMetrics

    // Apply custom validators
    for _, validatorName := range profile.CustomValidators {
        if err := engine.enableCustomValidator(validatorName); err != nil {
            return fmt.Errorf("failed to enable custom validator '%s': %w", validatorName, err)
        }
    }

    // Store profile for business logic application
    engine.currentProfile = profile

    return nil
}

// ValidateWithProfile validates a model using a specific profile
func (engine *FlexibleValidationEngine) ValidateWithProfile(data interface{}, modelName, profileName string) ValidationResult {
    start := time.Now()

    // Load validation profile
    profile, err := engine.LoadValidationProfile(profileName)
    if err != nil {
        return ValidationResult{
            IsValid:   false,
            Errors:    []ValidationError{{Message: fmt.Sprintf("Profile error: %v", err)}},
            ModelType: modelName,
            Timestamp: time.Now(),
        }
    }

    // Apply profile settings
    if err := engine.ApplyValidationProfile(profile, modelName); err != nil {
        return ValidationResult{
            IsValid:   false,
            Errors:    []ValidationError{{Message: fmt.Sprintf("Profile application error: %v", err)}},
            ModelType: modelName,
            Timestamp: time.Now(),
        }
    }

    // Create validation context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), engine.timeout)
    defer cancel()

    // Perform validation with context
    result := engine.validateWithContext(ctx, data, modelName, profile)

    // Add performance metrics if enabled
    if profile.EnableMetrics {
        result.PerformanceMetrics = &PerformanceMetrics{
            ValidationDuration: time.Since(start),
            CacheHits:         engine.cacheHits,
            CacheMisses:       engine.cacheMisses,
            MemoryUsage:       engine.getMemoryUsage(),
        }
    }

    return result
}
```

### Step 2: Create Dynamic Profile Loading

```go
// DynamicProfileLoader allows runtime profile modifications
type DynamicProfileLoader struct {
    profiles       map[string]*ValidationProfile
    watchedFiles   []string
    fileWatcher    *fsnotify.Watcher
    updateChannel  chan string
    mu             sync.RWMutex
}

// NewDynamicProfileLoader creates a new dynamic profile loader
func NewDynamicProfileLoader() *DynamicProfileLoader {
    watcher, _ := fsnotify.NewWatcher()

    loader := &DynamicProfileLoader{
        profiles:      make(map[string]*ValidationProfile),
        fileWatcher:   watcher,
        updateChannel: make(chan string, 10),
    }

    // Start watching for file changes
    go loader.watchFiles()

    return loader
}

// LoadProfilesFromFile loads all profiles from a configuration file
func (loader *DynamicProfileLoader) LoadProfilesFromFile(filename string) error {
    loader.mu.Lock()
    defer loader.mu.Unlock()

    configData, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("failed to read config file: %w", err)
    }

    var config struct {
        ValidationProfiles map[string]ValidationProfile `json:"validation_profiles"`
    }

    if err := json.Unmarshal(configData, &config); err != nil {
        return fmt.Errorf("failed to parse config file: %w", err)
    }

    // Update profiles
    for name, profile := range config.ValidationProfiles {
        loader.profiles[name] = &profile
    }

    // Add file to watch list
    if !contains(loader.watchedFiles, filename) {
        loader.watchedFiles = append(loader.watchedFiles, filename)
        loader.fileWatcher.Add(filename)
    }

    return nil
}

// GetProfile retrieves a validation profile by name
func (loader *DynamicProfileLoader) GetProfile(name string) (*ValidationProfile, error) {
    loader.mu.RLock()
    defer loader.mu.RUnlock()

    profile, exists := loader.profiles[name]
    if !exists {
        return nil, fmt.Errorf("profile '%s' not found", name)
    }

    // Return a copy to prevent modification
    profileCopy := *profile
    return &profileCopy, nil
}

// watchFiles monitors configuration files for changes
func (loader *DynamicProfileLoader) watchFiles() {
    for {
        select {
        case event, ok := <-loader.fileWatcher.Events:
            if !ok {
                return
            }

            if event.Op&fsnotify.Write == fsnotify.Write {
                log.Printf("Configuration file modified: %s", event.Name)

                // Reload the file
                if err := loader.LoadProfilesFromFile(event.Name); err != nil {
                    log.Printf("Failed to reload config file %s: %v", event.Name, err)
                } else {
                    loader.updateChannel <- event.Name
                }
            }

        case err, ok := <-loader.fileWatcher.Errors:
            if !ok {
                return
            }
            log.Printf("File watcher error: %v", err)
        }
    }
}

// GetUpdateChannel returns a channel that receives file update notifications
func (loader *DynamicProfileLoader) GetUpdateChannel() <-chan string {
    return loader.updateChannel
}

// CreateCustomProfile creates a new validation profile programmatically
func (loader *DynamicProfileLoader) CreateCustomProfile(name, description string, options ProfileOptions) error {
    loader.mu.Lock()
    defer loader.mu.Unlock()

    profile := &ValidationProfile{
        Name:             name,
        Description:      description,
        EnabledRules:     options.EnabledRules,
        BusinessRules:    options.BusinessRules,
        ErrorThreshold:   options.ErrorThreshold,
        WarningThreshold: options.WarningThreshold,
        TimeoutMS:        options.TimeoutMS,
        EnableCaching:    options.EnableCaching,
        EnableMetrics:    options.EnableMetrics,
        CustomValidators: options.CustomValidators,
        FailFast:         options.FailFast,
        DetailedErrors:   options.DetailedErrors,
    }

    loader.profiles[name] = profile
    return nil
}

// ProfileOptions defines options for creating custom profiles
type ProfileOptions struct {
    EnabledRules     []string               `json:"enabled_rules"`
    BusinessRules    map[string]interface{} `json:"business_rules"`
    ErrorThreshold   int                    `json:"error_threshold"`
    WarningThreshold int                    `json:"warning_threshold"`
    TimeoutMS        int                    `json:"timeout_ms"`
    EnableCaching    bool                   `json:"enable_caching"`
    EnableMetrics    bool                   `json:"enable_metrics"`
    CustomValidators []string               `json:"custom_validators"`
    FailFast         bool                   `json:"fail_fast"`
    DetailedErrors   bool                   `json:"detailed_errors"`
}
```

## Adding Provider Support

### Step 1: Create New Validation Provider

**File**: `src/interfaces.go` - Add new provider interface

```go
// CustomValidationProvider implements domain-specific validation
type CustomValidationProvider struct {
    name            string
    rules           map[string]ValidationRule
    businessRules   map[string]BusinessRule
    enabledFeatures []string
}

// NewCustomValidationProvider creates a new custom validation provider
func NewCustomValidationProvider(name string) *CustomValidationProvider {
    return &CustomValidationProvider{
        name:            name,
        rules:           make(map[string]ValidationRule),
        businessRules:   make(map[string]BusinessRule),
        enabledFeatures: []string{},
    }
}

// GetName returns the provider name
func (p *CustomValidationProvider) GetName() string {
    return p.name
}

// Validate performs validation using custom rules
func (p *CustomValidationProvider) Validate(data interface{}, modelType string) ValidationResult {
    start := time.Now()

    result := ValidationResult{
        IsValid:   true,
        Provider:  p.name,
        ModelType: modelType,
        Timestamp: time.Now(),
        Errors:    []ValidationError{},
        Warnings:  []ValidationWarning{},
    }

    // Apply structural validation rules
    if err := p.validateStructure(data, modelType); err != nil {
        result.IsValid = false
        result.Errors = append(result.Errors, ValidationError{
            Field:   "structure",
            Message: err.Error(),
            Code:    "STRUCTURE_VIOLATION",
        })
    }

    // Apply business validation rules
    warnings := p.validateBusinessRules(data, modelType)
    result.Warnings = append(result.Warnings, warnings...)

    // Add performance metrics
    result.PerformanceMetrics = &PerformanceMetrics{
        ValidationDuration: time.Since(start),
        Provider:          p.name,
    }

    return result
}

// AddValidationRule adds a custom validation rule
func (p *CustomValidationProvider) AddValidationRule(fieldPath string, rule ValidationRule) {
    p.rules[fieldPath] = rule
}

// AddBusinessRule adds a custom business rule
func (p *CustomValidationProvider) AddBusinessRule(ruleName string, rule BusinessRule) {
    p.businessRules[ruleName] = rule
}

// ValidationRule defines a field-level validation rule
type ValidationRule struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Validator   func(interface{}) bool `json:"-"`
    Parameters  map[string]interface{} `json:"parameters"`
    ErrorMessage string                `json:"error_message"`
    Required    bool                   `json:"required"`
}

// BusinessRule defines a business logic validation rule
type BusinessRule struct {
    Name        string                        `json:"name"`
    Description string                        `json:"description"`
    Validator   func(interface{}) []ValidationWarning `json:"-"`
    Enabled     bool                          `json:"enabled"`
    Severity    string                        `json:"severity"` // error, warning, info
    Parameters  map[string]interface{}        `json:"parameters"`
}

// validateStructure performs structural validation
func (p *CustomValidationProvider) validateStructure(data interface{}, modelType string) error {
    // Use reflection to validate structure against registered rules
    v := reflect.ValueOf(data)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    if v.Kind() != reflect.Struct {
        return fmt.Errorf("expected struct, got %T", data)
    }

    t := v.Type()
    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        fieldType := t.Field(i)
        fieldPath := fmt.Sprintf("%s.%s", modelType, fieldType.Name)

        // Check if we have a rule for this field
        if rule, exists := p.rules[fieldPath]; exists {
            if rule.Required && isZeroValue(field) {
                return fmt.Errorf("required field '%s' is missing", fieldType.Name)
            }

            if !isZeroValue(field) && rule.Validator != nil {
                if !rule.Validator(field.Interface()) {
                    return fmt.Errorf("field '%s' validation failed: %s", fieldType.Name, rule.ErrorMessage)
                }
            }
        }
    }

    return nil
}

// validateBusinessRules applies business logic rules
func (p *CustomValidationProvider) validateBusinessRules(data interface{}, modelType string) []ValidationWarning {
    var warnings []ValidationWarning

    for _, rule := range p.businessRules {
        if rule.Enabled && rule.Validator != nil {
            ruleWarnings := rule.Validator(data)
            warnings = append(warnings, ruleWarnings...)
        }
    }

    return warnings
}

// isZeroValue checks if a reflect.Value is the zero value for its type
func isZeroValue(v reflect.Value) bool {
    switch v.Kind() {
    case reflect.String:
        return v.String() == ""
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        return v.Int() == 0
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        return v.Uint() == 0
    case reflect.Float32, reflect.Float64:
        return v.Float() == 0
    case reflect.Bool:
        return !v.Bool()
    case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
        return v.IsNil()
    case reflect.Array, reflect.Struct:
        return v.Interface() == reflect.Zero(v.Type()).Interface()
    default:
        return false
    }
}
```

### Step 2: Register Provider with Engine

**File**: `src/flexible_engine.go` - Add provider registration

```go
// RegisterProvider adds a new validation provider to the engine
func (engine *FlexibleValidationEngine) RegisterProvider(provider ValidationProvider) error {
    engine.mu.Lock()
    defer engine.mu.Unlock()

    providerName := provider.GetName()

    // Check for duplicate provider names
    for _, existingProvider := range engine.providers {
        if existingProvider.GetName() == providerName {
            return fmt.Errorf("provider '%s' already registered", providerName)
        }
    }

    engine.providers = append(engine.providers, provider)
    log.Printf("Registered validation provider: %s", providerName)

    return nil
}

// CreateSlackValidationProvider creates a custom provider for Slack validation
func CreateSlackValidationProvider() *CustomValidationProvider {
    provider := NewCustomValidationProvider("slack_custom")

    // Add Slack-specific validation rules
    provider.AddValidationRule("SlackMessagePayload.Token", ValidationRule{
        Name:        "slack_token_format",
        Description: "Validates Slack token format and authenticity",
        Validator: func(value interface{}) bool {
            token, ok := value.(string)
            if !ok {
                return false
            }
            return validateSlackToken(validator.FieldLevel{})
        },
        Required:     true,
        ErrorMessage: "Invalid Slack token format",
    })

    provider.AddValidationRule("SlackMessagePayload.ChannelName", ValidationRule{
        Name:        "slack_channel_naming",
        Description: "Validates Slack channel naming conventions",
        Validator: func(value interface{}) bool {
            channelName, ok := value.(string)
            if !ok {
                return false
            }
            // Implement channel naming validation
            return len(channelName) > 0 && len(channelName) <= 80
        },
        Required:     true,
        ErrorMessage: "Invalid Slack channel name",
    })

    // Add Slack-specific business rules
    provider.AddBusinessRule("attachment_security", BusinessRule{
        Name:        "attachment_security_check",
        Description: "Validates attachment security and content",
        Validator: func(data interface{}) []ValidationWarning {
            var warnings []ValidationWarning

            if payload, ok := data.(SlackMessagePayload); ok {
                for i, attachment := range payload.Message.Attachments {
                    // Check for potentially malicious content
                    if strings.Contains(strings.ToLower(attachment.Text), "javascript:") ||
                       strings.Contains(strings.ToLower(attachment.Text), "<script") {
                        warnings = append(warnings, ValidationWarning{
                            Field:   fmt.Sprintf("Message.Attachments[%d].Text", i),
                            Message: "Attachment contains potentially unsafe content",
                            Code:    "UNSAFE_ATTACHMENT_CONTENT",
                        })
                    }

                    // Check for oversized attachments
                    if len(attachment.Text) > 8000 {
                        warnings = append(warnings, ValidationWarning{
                            Field:   fmt.Sprintf("Message.Attachments[%d].Text", i),
                            Message: "Attachment text is very long and may be truncated",
                            Code:    "OVERSIZED_ATTACHMENT",
                        })
                    }
                }
            }

            return warnings
        },
        Enabled:  true,
        Severity: "warning",
    })

    provider.AddBusinessRule("compliance_check", BusinessRule{
        Name:        "compliance_validation",
        Description: "Validates messages for compliance requirements",
        Validator: func(data interface{}) []ValidationWarning {
            var warnings []ValidationWarning

            if payload, ok := data.(SlackMessagePayload); ok {
                // Check for PII in message content
                piiPatterns := []string{
                    `\b\d{3}-\d{2}-\d{4}\b`,      // SSN pattern
                    `\b\d{4}\s?\d{4}\s?\d{4}\s?\d{4}\b`, // Credit card pattern
                    `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`, // Email pattern
                }

                for _, pattern := range piiPatterns {
                    if matched, _ := regexp.MatchString(pattern, payload.Message.Text); matched {
                        warnings = append(warnings, ValidationWarning{
                            Field:   "Message.Text",
                            Message: "Message may contain personally identifiable information (PII)",
                            Code:    "POTENTIAL_PII",
                        })
                        break
                    }
                }

                // Check for restricted keywords
                restrictedKeywords := []string{"confidential", "internal", "classified", "secret"}
                messageText := strings.ToLower(payload.Message.Text)

                for _, keyword := range restrictedKeywords {
                    if strings.Contains(messageText, keyword) {
                        warnings = append(warnings, ValidationWarning{
                            Field:   "Message.Text",
                            Message: fmt.Sprintf("Message contains restricted keyword: %s", keyword),
                            Code:    "RESTRICTED_CONTENT",
                        })
                    }
                }
            }

            return warnings
        },
        Enabled:  true,
        Severity: "error",
    })

    return provider
}

// Initialize and register custom providers
func (engine *FlexibleValidationEngine) initializeCustomProviders() error {
    // Register Slack custom provider
    slackProvider := CreateSlackValidationProvider()
    if err := engine.RegisterProvider(slackProvider); err != nil {
        return fmt.Errorf("failed to register Slack provider: %w", err)
    }

    // Register other custom providers as needed

    return nil
}
```

## Creating Test Data and Tests

### Step 1: Create Test Data Files

**File**: `test_data/slack_message_payload.json`

```json
{
  "token": "xoxb-1234567890123-1234567890123-abcdefghijklmnopqrstuvwx",
  "team_id": "T1234567890",
  "team_domain": "example-corp",
  "channel_id": "C1234567890",
  "channel_name": "general",
  "user_id": "U1234567890",
  "user_name": "johnsmith",
  "command": "/weather",
  "text": "New York",
  "response_url": "https://hooks.slack.com/commands/1234567890123/1234567890123/abcdefghijklmnopqrstuvwx",
  "trigger_id": "1234567890123.1234567890123.abcdefghijklmnopqrstuvwx",
  "ts": "1695902400.123456",
  "type": "event_callback",
  "message": {
    "type": "message",
    "text": "Hello team! Here's the weather update for today.",
    "user": "U1234567890",
    "ts": "1695902400.123456",
    "channel": "C1234567890",
    "attachments": [
      {
        "id": 1,
        "color": "good",
        "fallback": "Weather update for New York",
        "title": "New York Weather",
        "text": "Sunny, 72°F (22°C)\nHumidity: 45%\nWind: 5 mph NW",
        "footer": "Weather API",
        "ts": 1695902400,
        "fields": [
          {
            "title": "Temperature",
            "value": "72°F (22°C)",
            "short": true
          },
          {
            "title": "Condition",
            "value": "Sunny",
            "short": true
          }
        ]
      }
    ],
    "blocks": [
      {
        "type": "section",
        "text": {
          "type": "mrkdwn",
          "text": "*Weather Update*\nCurrent conditions for New York"
        }
      }
    ],
    "files": [],
    "reactions": [
      {
        "name": "sunny",
        "count": 3,
        "users": ["U1234567890", "U0987654321", "U1357924680"]
      }
    ],
    "replies": [],
    "is_starred": false,
    "pinned_to": false
  },
  "event": {
    "type": "message",
    "event_ts": "1695902400.123456",
    "user": "U1234567890"
  }
}
```

**File**: `test_data/slack_invalid_payload.json`

```json
{
  "token": "invalid-token-format",
  "team_id": "",
  "channel_name": "INVALID_CHANNEL_NAME_WITH_CAPS",
  "command": "invalid-command-without-slash",
  "text": "This message contains a potential token exposure: xoxb-test-token-123456789",
  "message": {
    "text": "This is a very long message that exceeds reasonable limits for testing purposes. " +
           "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
           "Repeated many times to make it very long...",
    "attachments": [
      {
        "text": "Attachment with <script>alert('xss')</script> content"
      }
    ],
    "files": [
      {
        "name": "malicious.exe",
        "filetype": "exe",
        "size": 209715200
      }
    ]
  }
}
```

**File**: `test_data/project_payload.json`

```json
{
  "id": "proj-12345",
  "name": "Customer Authentication System",
  "code": "ENG-2024",
  "description": "Implement OAuth 2.0 authentication system with multi-factor authentication support",
  "priority": "high",
  "status": "in_progress",
  "complexity": "high",
  "estimated_hours": 480,
  "estimated_cost": 75000.00,
  "total_budget": 80000.00,
  "start_date": "2025-09-01T00:00:00Z",
  "deadline": "2025-12-01T00:00:00Z",
  "created_at": "2025-08-15T10:00:00Z",
  "updated_at": "2025-09-21T14:30:00Z",
  "owner": {
    "id": "user-123",
    "name": "Jane Doe",
    "email": "jane.doe@company.com",
    "role": "Project Manager"
  },
  "team_members": [
    {
      "id": "user-456",
      "name": "John Smith",
      "email": "john.smith@company.com",
      "role": "Senior Developer"
    },
    {
      "id": "user-789",
      "name": "Alice Johnson",
      "email": "alice.johnson@company.com",
      "role": "UX Designer"
    }
  ],
  "resource_allocations": [
    {
      "member_id": "user-456",
      "percentage": 75.0,
      "start_date": "2025-09-01T00:00:00Z",
      "end_date": "2025-12-01T00:00:00Z"
    },
    {
      "member_id": "user-789",
      "percentage": 50.0,
      "start_date": "2025-09-15T00:00:00Z",
      "end_date": "2025-11-15T00:00:00Z"
    }
  ],
  "budget_allocations": [
    {
      "category": "development",
      "amount": 50000.00,
      "approval_id": "approval-12345"
    },
    {
      "category": "design",
      "amount": 15000.00,
      "approval_id": "approval-12346"
    },
    {
      "category": "testing",
      "amount": 10000.00,
      "approval_id": "approval-12347"
    }
  ],
  "dependencies": [
    {
      "id": "proj-11111",
      "name": "User Management System",
      "type": "prerequisite"
    }
  ],
  "milestones": [
    {
      "name": "Authentication Core Complete",
      "date": "2025-10-15T00:00:00Z",
      "status": "pending"
    },
    {
      "name": "MFA Integration Complete",
      "date": "2025-11-15T00:00:00Z",
      "status": "pending"
    }
  ],
  "tags": ["authentication", "security", "oauth", "mfa"],
  "risks": [
    {
      "description": "Third-party OAuth provider API changes",
      "probability": "medium",
      "impact": "high",
      "mitigation": "Monitor provider documentation and maintain fallback options"
    }
  ]
}
```

### Step 2: Create Comprehensive Tests

**File**: `src/slack_model_test.go`

```go
package main

import (
    "encoding/json"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSlackMessageValidation(t *testing.T) {
    // Initialize validation engine
    engine := NewFlexibleValidationEngine()

    // Register Slack custom provider
    slackProvider := CreateSlackValidationProvider()
    err := engine.RegisterProvider(slackProvider)
    require.NoError(t, err)

    tests := []struct {
        name           string
        payloadFile    string
        expectedValid  bool
        expectedErrors int
        expectedWarnings int
        checkFields    []string
    }{
        {
            name:           "valid_slack_message",
            payloadFile:    "../test_data/slack_message_payload.json",
            expectedValid:  true,
            expectedErrors: 0,
            expectedWarnings: 0,
            checkFields:    []string{"token", "channel_name", "message.text"},
        },
        {
            name:           "invalid_slack_message",
            payloadFile:    "../test_data/slack_invalid_payload.json",
            expectedValid:  false,
            expectedErrors: 3, // Invalid token, channel name, command
            expectedWarnings: 4, // Token exposure, long message, unsafe attachment, large file
            checkFields:    []string{"token", "channel_name", "command"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Load test data
            data, err := os.ReadFile(tt.payloadFile)
            require.NoError(t, err)

            var payload SlackMessagePayload
            err = json.Unmarshal(data, &payload)
            require.NoError(t, err)

            // Validate using flexible engine
            result := engine.ValidateModel(payload, "SlackMessagePayload", "default")

            // Check validation result
            assert.Equal(t, tt.expectedValid, result.IsValid)
            assert.Len(t, result.Errors, tt.expectedErrors)
            assert.Len(t, result.Warnings, tt.expectedWarnings)

            // Check specific fields if validation failed
            if !result.IsValid {
                errorFields := make(map[string]bool)
                for _, err := range result.Errors {
                    errorFields[err.Field] = true
                }

                for _, field := range tt.checkFields {
                    if tt.expectedValid {
                        assert.False(t, errorFields[field], "Field %s should not have errors", field)
                    }
                }
            }

            // Verify performance metrics
            assert.NotNil(t, result.PerformanceMetrics)
            assert.True(t, result.PerformanceMetrics.ValidationDuration > 0)

            // Verify timestamp
            assert.WithinDuration(t, time.Now(), result.Timestamp, time.Second)
        })
    }
}

func TestSlackCustomValidators(t *testing.T) {
    tests := []struct {
        name      string
        validator string
        value     interface{}
        expected  bool
    }{
        {
            name:      "valid_slack_token",
            validator: "slack_token",
            value:     "xoxb-1234567890123-1234567890123-abcdefghijklmnopqrstuvwx",
            expected:  true,
        },
        {
            name:      "invalid_slack_token",
            validator: "slack_token",
            value:     "invalid-token",
            expected:  false,
        },
        {
            name:      "valid_channel_name",
            validator: "slack_channel_name",
            value:     "general",
            expected:  true,
        },
        {
            name:      "invalid_channel_name_caps",
            validator: "slack_channel_name",
            value:     "GENERAL",
            expected:  false,
        },
        {
            name:      "valid_slack_command",
            validator: "slack_command",
            value:     "/weather",
            expected:  true,
        },
        {
            name:      "invalid_slack_command",
            validator: "slack_command",
            value:     "weather",
            expected:  false,
        },
        {
            name:      "valid_hex_color",
            validator: "hexcolor",
            value:     "#FF0000",
            expected:  true,
        },
        {
            name:      "valid_hex_color_short",
            validator: "hexcolor",
            value:     "F00",
            expected:  true,
        },
        {
            name:      "invalid_hex_color",
            validator: "hexcolor",
            value:     "red",
            expected:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test individual validators
            switch tt.validator {
            case "slack_token":
                // Create mock FieldLevel for testing
                // Result would depend on implementation
                // This is a simplified test structure
            case "slack_channel_name":
                // Test channel name validation
            case "slack_command":
                // Test command validation
            case "hexcolor":
                // Test hex color validation
            }

            // For actual implementation, you would call the validator function
            // and compare the result with tt.expected
        })
    }
}

func TestSlackBusinessLogicValidation(t *testing.T) {
    tests := []struct {
        name             string
        payload          SlackMessagePayload
        expectedWarnings int
        expectedCodes    []string
    }{
        {
            name: "excessive_attachments",
            payload: SlackMessagePayload{
                Message: SlackMessage{
                    Attachments: make([]SlackAttachment, 25), // More than 20
                },
            },
            expectedWarnings: 1,
            expectedCodes:    []string{"EXCESSIVE_ATTACHMENTS"},
        },
        {
            name: "long_message_text",
            payload: SlackMessagePayload{
                Message: SlackMessage{
                    Text: strings.Repeat("A", 5000), // More than 4000 characters
                },
            },
            expectedWarnings: 1,
            expectedCodes:    []string{"LONG_MESSAGE_TEXT"},
        },
        {
            name: "token_exposure",
            payload: SlackMessagePayload{
                Text: "Here's my token: xoxb-test-123",
            },
            expectedWarnings: 1,
            expectedCodes:    []string{"POTENTIAL_TOKEN_EXPOSURE"},
        },
        {
            name: "excessive_mentions",
            payload: SlackMessagePayload{
                Message: SlackMessage{
                    Text: "<@U123> <@U124> <@U125> <@U126> <@U127> <@U128> <@U129> <@U130> <@U131> <@U132> <@U133>", // 11 mentions
                },
            },
            expectedWarnings: 1,
            expectedCodes:    []string{"EXCESSIVE_MENTIONS"},
        },
        {
            name: "large_file_size",
            payload: SlackMessagePayload{
                Message: SlackMessage{
                    Files: []SlackFile{
                        {
                            Name: "large_file.zip",
                            Size: 150 * 1024 * 1024, // 150MB
                        },
                    },
                },
            },
            expectedWarnings: 1,
            expectedCodes:    []string{"LARGE_FILE_SIZE"},
        },
        {
            name: "blocked_file_type",
            payload: SlackMessagePayload{
                Message: SlackMessage{
                    Files: []SlackFile{
                        {
                            Name:     "malware.exe",
                            Filetype: "exe",
                        },
                    },
                },
            },
            expectedWarnings: 1,
            expectedCodes:    []string{"POTENTIALLY_BLOCKED_FILETYPE"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            warnings := validateSlackMessageBusinessLogic(tt.payload)

            assert.Len(t, warnings, tt.expectedWarnings)

            if tt.expectedWarnings > 0 {
                warningCodes := make([]string, len(warnings))
                for i, warning := range warnings {
                    warningCodes[i] = warning.Code
                }

                for _, expectedCode := range tt.expectedCodes {
                    assert.Contains(t, warningCodes, expectedCode)
                }
            }
        })
    }
}

func TestSlackValidationProfiles(t *testing.T) {
    engine := NewFlexibleValidationEngine()
    slackProvider := CreateSlackValidationProvider()
    err := engine.RegisterProvider(slackProvider)
    require.NoError(t, err)

    // Load test payload
    data, err := os.ReadFile("../test_data/slack_message_payload.json")
    require.NoError(t, err)

    var payload SlackMessagePayload
    err = json.Unmarshal(data, &payload)
    require.NoError(t, err)

    profiles := []string{"slack_strict", "slack_permissive", "default"}

    for _, profile := range profiles {
        t.Run(profile, func(t *testing.T) {
            result := engine.ValidateWithProfile(payload, "SlackMessagePayload", profile)

            // All profiles should accept valid data
            assert.True(t, result.IsValid)
            assert.NotNil(t, result.PerformanceMetrics)

            // Verify profile was applied
            assert.Equal(t, "SlackMessagePayload", result.ModelType)

            // Profile-specific checks would go here
            switch profile {
            case "slack_strict":
                // Strict profile might have more warnings
            case "slack_permissive":
                // Permissive profile might have fewer warnings
            }
        })
    }
}

func BenchmarkSlackValidation(b *testing.B) {
    engine := NewFlexibleValidationEngine()
    slackProvider := CreateSlackValidationProvider()
    engine.RegisterProvider(slackProvider)

    // Load test data
    data, _ := os.ReadFile("../test_data/slack_message_payload.json")
    var payload SlackMessagePayload
    json.Unmarshal(data, &payload)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        result := engine.ValidateModel(payload, "SlackMessagePayload", "default")
        if !result.IsValid {
            b.Fatalf("Validation failed: %v", result.Errors)
        }
    }
}
```

## Integration and Deployment

### Step 1: Update Main Server

**File**: `src/main.go` - Integrate new models

```go
func main() {
    // Register custom validators during initialization
    registerCustomValidators()
    registerStructValidators()

    // Initialize flexible validation engine with custom providers
    flexibleEngine := NewFlexibleValidationEngine()

    // Initialize and register all custom providers
    if err := flexibleEngine.initializeCustomProviders(); err != nil {
        log.Fatalf("Failed to initialize custom providers: %v", err)
    }

    // Start flexible server with enhanced capabilities
    go func() {
        log.Println("Starting flexible validation server on :8081")
        flexibleServer := &http.Server{
            Addr:    ":8081",
            Handler: setupFlexibleRoutes(flexibleEngine),
        }
        log.Fatal(flexibleServer.ListenAndServe())
    }()

    // Start original server for backward compatibility
    server := NewAPIServer()
    log.Fatal(server.Start(":8080"))
}
```

### Step 2: Update Configuration Loading

**File**: `src/flexible_engine.go` - Enhance configuration support

```go
// LoadConfiguration loads validation configuration from file
func (engine *FlexibleValidationEngine) LoadConfiguration(configPath string) error {
    configData, err := os.ReadFile(configPath)
    if err != nil {
        return fmt.Errorf("failed to read config file: %w", err)
    }

    var config ValidationConfig
    if err := json.Unmarshal(configData, &config); err != nil {
        return fmt.Errorf("failed to parse config file: %w", err)
    }

    // Apply global settings
    engine.applyGlobalSettings(config.GlobalSettings)

    // Load model configurations
    for modelName, modelConfig := range config.Models {
        if err := engine.registerModelConfiguration(modelName, modelConfig); err != nil {
            log.Printf("Warning: Failed to register model configuration for %s: %v", modelName, err)
        }
    }

    // Load validation profiles
    engine.profiles = config.ValidationProfiles

    return nil
}

// ValidationConfig represents the complete configuration structure
type ValidationConfig struct {
    Models             map[string]ModelConfig              `json:"models"`
    ValidationProfiles map[string]ValidationProfile        `json:"validation_profiles"`
    GlobalSettings     GlobalSettings                      `json:"global_settings"`
}

// ModelConfig represents configuration for a specific model
type ModelConfig struct {
    Description   string                        `json:"description"`
    Fields        map[string]FieldConfig        `json:"fields"`
    BusinessRules map[string]BusinessRuleConfig `json:"business_rules"`
}

// FieldConfig represents configuration for a model field
type FieldConfig struct {
    Required    bool   `json:"required"`
    Type        string `json:"type"`
    Validation  string `json:"validation"`
    Description string `json:"description"`
}

// BusinessRuleConfig represents configuration for business rules
type BusinessRuleConfig struct {
    Enabled        bool                   `json:"enabled"`
    Parameters     map[string]interface{} `json:"parameters"`
    WarningMessage string                 `json:"warning_message"`
}

// GlobalSettings represents global validation settings
type GlobalSettings struct {
    EnableBusinessLogic      bool `json:"enable_business_logic"`
    EnableWarnings          bool `json:"enable_warnings"`
    EnableCaching           bool `json:"enable_caching"`
    CacheTTLSeconds         int  `json:"cache_ttl_seconds"`
    EnableDetailedErrors    bool `json:"enable_detailed_errors"`
    MaxValidationTimeMS     int  `json:"max_validation_time_ms"`
    EnablePerformanceMetrics bool `json:"enable_performance_metrics"`
}
```

### Step 3: Create Deployment Scripts

**File**: `deploy_custom_models.sh`

```bash
#!/bin/bash

# Deployment script for custom model validation infrastructure

set -e

echo "🚀 Deploying Custom Model Validation Infrastructure"

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="$PROJECT_ROOT/src"
CONFIG_DIR="$PROJECT_ROOT/config"
TEST_DATA_DIR="$PROJECT_ROOT/test_data"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Step 1: Validate project structure
print_status "Validating project structure..."

required_files=(
    "$SRC_DIR/models.go"
    "$SRC_DIR/model_validators.go"
    "$SRC_DIR/validation_registry.go"
    "$SRC_DIR/flexible_engine.go"
    "$SRC_DIR/flexible_server.go"
    "$CONFIG_DIR/validation_rules.json"
)

for file in "${required_files[@]}"; do
    if [[ ! -f "$file" ]]; then
        print_error "Required file missing: $file"
        exit 1
    fi
done

print_success "Project structure validated"

# Step 2: Validate configuration files
print_status "Validating configuration files..."

if ! jq . "$CONFIG_DIR/validation_rules.json" > /dev/null 2>&1; then
    print_error "Invalid JSON in validation_rules.json"
    exit 1
fi

print_success "Configuration files validated"

# Step 3: Validate test data files
print_status "Validating test data files..."

find "$TEST_DATA_DIR" -name "*.json" -exec jq . {} \; > /dev/null || {
    print_error "Invalid JSON in test data files"
    exit 1
}

print_success "Test data files validated"

# Step 4: Run Go mod tidy and verify
print_status "Managing Go dependencies..."

cd "$SRC_DIR"

if ! go mod tidy; then
    print_error "Failed to tidy Go modules"
    exit 1
fi

if ! go mod verify; then
    print_error "Failed to verify Go modules"
    exit 1
fi

print_success "Go dependencies verified"

# Step 5: Run comprehensive tests
print_status "Running comprehensive tests..."

# Run all tests with coverage
if ! go test -v -coverprofile=coverage.out ./...; then
    print_error "Tests failed"
    exit 1
fi

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

print_success "All tests passed"

# Step 6: Build the applications
print_status "Building applications..."

# Build main server
if ! go build -o main-server .; then
    print_error "Failed to build main server"
    exit 1
fi

# Build validation tools
if ! go build -o validation-cli -tags=cli .; then
    print_warning "CLI build failed (optional)"
fi

print_success "Applications built successfully"

# Step 7: Run integration tests
print_status "Running integration tests..."

# Start servers in background for testing
./main-server &
MAIN_PID=$!

# Wait for servers to start
sleep 3

# Test endpoints
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    print_success "Main server health check passed"
else
    print_error "Main server health check failed"
    kill $MAIN_PID 2>/dev/null || true
    exit 1
fi

if curl -f http://localhost:8081/health > /dev/null 2>&1; then
    print_success "Flexible server health check passed"
else
    print_error "Flexible server health check failed"
    kill $MAIN_PID 2>/dev/null || true
    exit 1
fi

# Test new Slack endpoint
if curl -f -X POST http://localhost:8081/validate/slack \
   -H "Content-Type: application/json" \
   -d '{"token":"test","team_id":"T123","channel_name":"test","type":"event_callback"}' > /dev/null 2>&1; then
    print_success "Slack validation endpoint test passed"
else
    print_warning "Slack validation endpoint test failed (may need valid test data)"
fi

# Cleanup
kill $MAIN_PID 2>/dev/null || true

print_success "Integration tests completed"

# Step 8: Generate documentation
print_status "Generating documentation..."

# Create API documentation
if command -v swagger &> /dev/null; then
    swagger generate spec -o api-docs.json
    print_success "API documentation generated"
else
    print_warning "Swagger not found, skipping API documentation"
fi

# Step 9: Create deployment package
print_status "Creating deployment package..."

DEPLOYMENT_DIR="deployment_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$DEPLOYMENT_DIR"

# Copy binaries
cp main-server "$DEPLOYMENT_DIR/"
[[ -f validation-cli ]] && cp validation-cli "$DEPLOYMENT_DIR/"

# Copy configuration
cp -r "$CONFIG_DIR" "$DEPLOYMENT_DIR/"

# Copy documentation
cp -r docs "$DEPLOYMENT_DIR/" 2>/dev/null || true
[[ -f coverage.html ]] && cp coverage.html "$DEPLOYMENT_DIR/"
[[ -f api-docs.json ]] && cp api-docs.json "$DEPLOYMENT_DIR/"

# Create deployment script
cat > "$DEPLOYMENT_DIR/deploy.sh" << 'EOF'
#!/bin/bash
echo "Starting validation servers..."
./main-server &
echo "Servers started. Check logs for status."
EOF

chmod +x "$DEPLOYMENT_DIR/deploy.sh"

# Create archive
tar -czf "${DEPLOYMENT_DIR}.tar.gz" "$DEPLOYMENT_DIR"

print_success "Deployment package created: ${DEPLOYMENT_DIR}.tar.gz"

# Step 10: Final validation
print_status "Performing final validation..."

# Validate package contents
if tar -tzf "${DEPLOYMENT_DIR}.tar.gz" | grep -q "main-server"; then
    print_success "Deployment package validated"
else
    print_error "Deployment package validation failed"
    exit 1
fi

# Cleanup temporary directory
rm -rf "$DEPLOYMENT_DIR"

print_success "🎉 Deployment completed successfully!"
print_status "Deployment package: ${DEPLOYMENT_DIR}.tar.gz"
print_status "Coverage report: coverage.html"
print_status "Next steps:"
echo "  1. Upload deployment package to target environment"
echo "  2. Extract and run deploy.sh"
echo "  3. Configure monitoring and alerting"
echo "  4. Update documentation and team training"
```

## Best Practices and Troubleshooting

### Best Practices

#### 1. Model Design
- **Use descriptive field names** that clearly indicate purpose
- **Apply appropriate validation tags** based on data requirements
- **Include comprehensive documentation** in struct comments
- **Follow Go naming conventions** for consistency
- **Use pointer types for optional fields** to distinguish between zero values and nil

#### 2. Validation Rules
- **Create reusable validators** for common patterns
- **Use regex patterns carefully** - ensure they're not overly complex
- **Implement proper error messages** that help users understand issues
- **Test edge cases thoroughly** including boundary conditions
- **Document custom validator behavior** clearly

#### 3. Business Logic
- **Separate structural validation from business rules**
- **Make business rules configurable** when possible
- **Use warning levels appropriately** (error vs warning vs info)
- **Consider performance impact** of complex business rules
- **Implement gradual rollout** for new business rules

#### 4. Configuration Management
- **Use version control** for configuration files
- **Implement configuration validation** before deployment
- **Support environment-specific** configuration
- **Document configuration changes** and their impact
- **Test configuration changes** in staging environment first

#### 5. Testing Strategy
- **Create comprehensive test data** covering all scenarios
- **Test both positive and negative cases**
- **Include performance benchmarks** for validation functions
- **Implement integration tests** for complete workflows
- **Test configuration changes** before deployment

### Troubleshooting Guide

#### Common Issues and Solutions

| Issue | Symptoms | Solution |
|-------|----------|----------|
| **Validation always fails** | All payloads return validation errors | Check struct tags match JSON field names exactly |
| **Custom validator not working** | Custom validator never called | Ensure validator is registered in `init()` function |
| **Configuration not loading** | Default settings always used | Verify JSON syntax and file path accessibility |
| **Poor performance** | Slow validation response times | Check for complex regex patterns, optimize business rules |
| **Memory leaks** | Increasing memory usage over time | Review slice allocations, ensure proper cleanup |
| **Provider not found** | Provider selection fails | Verify provider registration order and naming |

#### Debugging Techniques

##### 1. Enable Debug Logging
```go
// Add debug logging to validation functions
func debugValidator(fl validator.FieldLevel) bool {
    value := fl.Field().Interface()
    log.Printf("DEBUG: Validating field %s with value %v", fl.FieldName(), value)

    result := actualValidation(value)
    log.Printf("DEBUG: Validation result for %s: %v", fl.FieldName(), result)

    return result
}
```

##### 2. Validation Tracing
```go
// Add tracing to track validation flow
type ValidationTracer struct {
    steps []string
}

func (t *ValidationTracer) AddStep(step string) {
    t.steps = append(t.steps, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), step))
}

func (t *ValidationTracer) GetTrace() []string {
    return t.steps
}
```

##### 3. Performance Profiling
```bash
# Profile validation performance
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.
go tool pprof cpu.prof
go tool pprof mem.prof
```

##### 4. Configuration Validation
```go
// Validate configuration file structure
func validateConfiguration(configPath string) error {
    configData, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }

    var config ValidationConfig
    if err := json.Unmarshal(configData, &config); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    // Validate required fields
    if config.GlobalSettings.MaxValidationTimeMS <= 0 {
        return errors.New("max_validation_time_ms must be positive")
    }

    // Validate each model configuration
    for modelName, modelConfig := range config.Models {
        if err := validateModelConfig(modelName, modelConfig); err != nil {
            return fmt.Errorf("invalid model config for %s: %w", modelName, err)
        }
    }

    return nil
}
```

#### Error Investigation Workflow

1. **Check Logs**: Review server logs for error patterns and stack traces
2. **Validate Input**: Ensure input data matches expected structure and format
3. **Test Validators**: Test individual validators with problematic data
4. **Review Configuration**: Verify configuration files are valid and accessible
5. **Check Registration**: Ensure all models and validators are properly registered
6. **Test Isolation**: Create minimal test cases to isolate the issue
7. **Performance Analysis**: Profile validation functions for performance issues

### Monitoring and Maintenance

#### Metrics to Track
- **Validation Success Rate**: Percentage of successful validations
- **Response Time Distribution**: P50, P95, P99 response times
- **Error Rate by Model Type**: Track which models fail most often
- **Business Rule Violations**: Monitor warning patterns
- **Resource Usage**: CPU and memory utilization
- **Configuration Changes**: Track when and what changes were made

#### Alerting Recommendations
- **High Error Rate**: Alert when validation failure rate exceeds threshold
- **Performance Degradation**: Alert on response time increases
- **Resource Exhaustion**: Alert on high CPU or memory usage
- **Configuration Errors**: Alert on configuration reload failures
- **Business Rule Violations**: Alert on unusual warning patterns

This comprehensive guide provides everything needed to successfully add custom models and validation rules to the flexible validation infrastructure. The modular design allows for easy extension while maintaining performance and reliability.