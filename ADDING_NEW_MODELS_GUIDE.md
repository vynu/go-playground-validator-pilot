# Adding New Models and Validations Guide

## 📚 Overview

This comprehensive guide walks you through the complete process of adding new models and validations to the Go Playground Data Validator project. The system is designed to be **model-agnostic** and **auto-discovering**, making it easy to add new models without breaking existing functionality.

## 🏗️ Architecture Overview

The validator uses an automatic model discovery system that:
- ✅ **Automatically discovers** new models on server startup
- ✅ **Registers HTTP endpoints** for each model (`/validate/{model_name}`)
- ✅ **Zero configuration** required - just add files and restart
- ✅ **Thread-safe** validation with concurrent request support

## 📝 Step-by-Step Guide to Adding New Models

### Step 1: Create the Model Structure

Create a new file `src/models/your_model.go`:

```go
package models

import (
    "time"
)

// UserProfilePayload represents a user profile validation model
type UserProfilePayload struct {
    // Basic Information
    ID          string    `json:"id" validate:"required,min=3,max=50"`
    Username    string    `json:"username" validate:"required,min=3,max=30,alphanum"`
    Email       string    `json:"email" validate:"required,email"`
    FirstName   string    `json:"first_name" validate:"required,min=1,max=50"`
    LastName    string    `json:"last_name" validate:"required,min=1,max=50"`

    // Profile Details
    Age         int       `json:"age" validate:"required,min=13,max=120"`
    Bio         string    `json:"bio" validate:"max=500"`
    Avatar      string    `json:"avatar" validate:"omitempty,url"`
    Website     string    `json:"website" validate:"omitempty,url"`

    // Settings
    IsPublic    bool      `json:"is_public"`
    IsVerified  bool      `json:"is_verified"`
    Role        string    `json:"role" validate:"required,oneof=user admin moderator"`
    Status      string    `json:"status" validate:"required,oneof=active inactive suspended"`

    // Metadata
    CreatedAt   time.Time `json:"created_at" validate:"required"`
    UpdatedAt   time.Time `json:"updated_at" validate:"required"`
    LastLoginAt *time.Time `json:"last_login_at,omitempty"`

    // Arrays and nested data
    Tags        []string  `json:"tags" validate:"dive,min=1,max=20"`
    Preferences map[string]interface{} `json:"preferences,omitempty"`
}
```

### Step 2: Create the Validator

Create a new file `src/validations/user_profile.go`:

```go
package validations

import (
    "fmt"
    "regexp"
    "strings"
    "time"

    "github.com/go-playground/validator/v10"
    "goplayground-data-validator/models"
)

// UserProfileValidator handles validation for UserProfile payloads
type UserProfileValidator struct {
    validator *validator.Validate
}

// NewUserProfileValidator creates a new UserProfile validator instance
func NewUserProfileValidator() *UserProfileValidator {
    v := validator.New()

    // Register custom validation rules
    v.RegisterValidation("username_format", validateUsernameFormat)
    v.RegisterValidation("safe_bio", validateSafeBio)

    return &UserProfileValidator{validator: v}
}

// ValidatePayload validates a UserProfile payload and returns structured results
func (uv *UserProfileValidator) ValidatePayload(payload models.UserProfilePayload) models.ValidationResult {
    // Initialize result structure
    result := models.ValidationResult{
        IsValid:   true,
        ModelType: "user_profile",
        Provider:  "go-playground",
        Errors:    []models.ValidationError{},
        Warnings:  []models.ValidationWarning{},
    }

    // Perform struct validation using go-playground/validator
    if err := uv.validator.Struct(payload); err != nil {
        result.IsValid = false
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            for _, ve := range validationErrors {
                result.Errors = append(result.Errors, models.ValidationError{
                    Field:   strings.ToLower(ve.Field()),
                    Message: uv.getCustomErrorMessage(ve),
                    Code:    uv.getErrorCode(ve.Tag()),
                    Value:   fmt.Sprintf("%v", ve.Value()),
                })
            }
        }
    }

    // Apply custom validations only if basic validation passed
    if result.IsValid {
        // Custom Validation 1: Username uniqueness check (simulated)
        if err := uv.validateUsernameUniqueness(payload.Username); err != nil {
            result.IsValid = false
            result.Errors = append(result.Errors, models.ValidationError{
                Field:   "username",
                Message: err.Error(),
                Code:    "USERNAME_TAKEN",
                Value:   payload.Username,
            })
        }

        // Custom Validation 2: Age-Role consistency
        if err := uv.validateAgeRoleConsistency(payload.Age, payload.Role); err != nil {
            result.IsValid = false
            result.Errors = append(result.Errors, models.ValidationError{
                Field:   "age",
                Message: err.Error(),
                Code:    "AGE_ROLE_MISMATCH",
                Value:   fmt.Sprintf("age=%d, role=%s", payload.Age, payload.Role),
            })
        }
    }

    // Add business logic warnings (even if validation failed)
    result.Warnings = uv.validateBusinessLogic(payload)

    return result
}

// Custom validation functions
func validateUsernameFormat(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    // Username must start with letter, contain only alphanumeric and underscores
    matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, username)
    return matched
}

func validateSafeBio(fl validator.FieldLevel) bool {
    bio := fl.Field().String()
    // Check for potentially unsafe content (basic example)
    unsafeWords := []string{"script", "javascript", "onclick"}
    bioLower := strings.ToLower(bio)
    for _, word := range unsafeWords {
        if strings.Contains(bioLower, word) {
            return false
        }
    }
    return true
}

// validateUsernameUniqueness simulates checking username uniqueness
func (uv *UserProfileValidator) validateUsernameUniqueness(username string) error {
    // In real implementation, this would check against database
    reservedUsernames := []string{"admin", "root", "system", "api", "www"}
    usernameLower := strings.ToLower(username)

    for _, reserved := range reservedUsernames {
        if usernameLower == reserved {
            return fmt.Errorf("username '%s' is reserved and cannot be used", username)
        }
    }
    return nil
}

// validateAgeRoleConsistency ensures age is appropriate for role
func (uv *UserProfileValidator) validateAgeRoleConsistency(age int, role string) error {
    if role == "admin" && age < 18 {
        return fmt.Errorf("admin role requires minimum age of 18, got age %d", age)
    }
    if role == "moderator" && age < 16 {
        return fmt.Errorf("moderator role requires minimum age of 16, got age %d", age)
    }
    return nil
}

// validateBusinessLogic performs user profile business validation checks
func (uv *UserProfileValidator) validateBusinessLogic(payload models.UserProfilePayload) []models.ValidationWarning {
    var warnings []models.ValidationWarning

    // Warning: Public profile without bio
    if payload.IsPublic && payload.Bio == "" {
        warnings = append(warnings, models.ValidationWarning{
            Field:      "bio",
            Message:    "Public profiles should have a bio for better user experience",
            Code:       "PUBLIC_PROFILE_NO_BIO",
            Suggestion: "Add a brief bio to help others understand your profile",
        })
    }

    // Warning: Admin without verification
    if payload.Role == "admin" && !payload.IsVerified {
        warnings = append(warnings, models.ValidationWarning{
            Field:      "is_verified",
            Message:    "Admin users should be verified for security purposes",
            Code:       "ADMIN_NOT_VERIFIED",
            Suggestion: "Complete the verification process for admin accounts",
        })
    }

    // Warning: Long inactive account
    if payload.LastLoginAt != nil {
        daysSinceLogin := time.Since(*payload.LastLoginAt).Hours() / 24
        if daysSinceLogin > 90 {
            warnings = append(warnings, models.ValidationWarning{
                Field:      "last_login_at",
                Message:    fmt.Sprintf("Account inactive for %.0f days", daysSinceLogin),
                Code:       "LONG_INACTIVE_ACCOUNT",
                Suggestion: "Consider account reactivation or cleanup procedures",
            })
        }
    }

    // Warning: Too many tags
    if len(payload.Tags) > 10 {
        warnings = append(warnings, models.ValidationWarning{
            Field:      "tags",
            Message:    fmt.Sprintf("Profile has %d tags, consider reducing for better performance", len(payload.Tags)),
            Code:       "TOO_MANY_TAGS",
            Suggestion: "Keep tags under 10 for optimal performance",
        })
    }

    return warnings
}

// getCustomErrorMessage provides friendly error messages
func (uv *UserProfileValidator) getCustomErrorMessage(ve validator.FieldError) string {
    switch ve.Tag() {
    case "required":
        return fmt.Sprintf("%s is required", strings.Title(ve.Field()))
    case "email":
        return "Must be a valid email address"
    case "min":
        return fmt.Sprintf("Must be at least %s characters/value", ve.Param())
    case "max":
        return fmt.Sprintf("Must be at most %s characters/value", ve.Param())
    case "oneof":
        return fmt.Sprintf("Must be one of: %s", ve.Param())
    case "alphanum":
        return "Must contain only letters and numbers"
    case "url":
        return "Must be a valid URL"
    case "username_format":
        return "Username must start with a letter and contain only letters, numbers, and underscores"
    case "safe_bio":
        return "Bio contains potentially unsafe content"
    default:
        return fmt.Sprintf("Validation failed for %s", ve.Field())
    }
}

// getErrorCode returns appropriate error codes
func (uv *UserProfileValidator) getErrorCode(tag string) string {
    switch tag {
    case "required":
        return "REQUIRED_FIELD_MISSING"
    case "email":
        return "INVALID_EMAIL_FORMAT"
    case "min":
        return "VALUE_TOO_SHORT"
    case "max":
        return "VALUE_TOO_LONG"
    case "oneof":
        return "INVALID_ENUM_VALUE"
    case "alphanum":
        return "INVALID_ALPHANUMERIC"
    case "url":
        return "INVALID_URL_FORMAT"
    case "username_format":
        return "INVALID_USERNAME_FORMAT"
    case "safe_bio":
        return "UNSAFE_CONTENT_DETECTED"
    default:
        return "VALIDATION_FAILED"
    }
}
```

### Step 3: Create Test Data

Create valid test data file `test_data/valid/user_profile.json`:

```json
{
    "id": "user_12345",
    "username": "john_doe",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "age": 28,
    "bio": "Software developer passionate about clean code and great user experiences. Love hiking and photography in my free time.",
    "avatar": "https://example.com/avatars/john_doe.jpg",
    "website": "https://johndoe.dev",
    "is_public": true,
    "is_verified": true,
    "role": "user",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-09-27T14:22:00Z",
    "last_login_at": "2024-09-26T09:15:00Z",
    "tags": ["developer", "golang", "photography", "hiking"],
    "preferences": {
        "theme": "dark",
        "notifications": true,
        "privacy_level": "medium"
    }
}
```

Create invalid test data file `test_data/invalid/user_profile.json`:

```json
{
    "id": "usr",
    "username": "123invalid",
    "email": "not-an-email",
    "first_name": "",
    "age": 12,
    "role": "admin",
    "status": "invalid_status",
    "created_at": "invalid-date",
    "tags": ["", "a", "very_long_tag_that_exceeds_limit"]
}
```

### Step 4: Build and Test

#### Build the Validator
```bash
cd src
go build -o ../validator main.go
```

#### Test with E2E Suite
```bash
./e2e_test_suite.sh
```

The E2E suite will automatically:
- ✅ Discover your new `user_profile` model
- ✅ Register the `/validate/user_profile` endpoint
- ✅ Test with your valid and invalid JSON data
- ✅ Report validation results

## 🔧 Manual Testing with cURL

### Test Valid Payload
```bash
# Start the validator server (if not already running)
PORT=8086 ./validator &

# Test valid user profile
curl -X POST http://localhost:8086/validate/user_profile \
  -H "Content-Type: application/json" \
  -d @test_data/valid/user_profile.json | jq .
```

**Expected Response:**
```json
{
  "is_valid": true,
  "model_type": "user_profile",
  "provider": "go-playground",
  "errors": [],
  "warnings": []
}
```

### Test Invalid Payload
```bash
curl -X POST http://localhost:8086/validate/user_profile \
  -H "Content-Type: application/json" \
  -d @test_data/invalid/user_profile.json | jq .
```

**Expected Response:**
```json
{
  "is_valid": false,
  "model_type": "user_profile",
  "provider": "go-playground",
  "errors": [
    {
      "field": "id",
      "message": "Must be at least 3 characters/value",
      "code": "VALUE_TOO_SHORT",
      "value": "usr"
    },
    {
      "field": "username",
      "message": "Username must start with a letter and contain only letters, numbers, and underscores",
      "code": "INVALID_USERNAME_FORMAT",
      "value": "123invalid"
    },
    {
      "field": "email",
      "message": "Must be a valid email address",
      "code": "INVALID_EMAIL_FORMAT",
      "value": "not-an-email"
    }
  ],
  "warnings": []
}
```

### Test Business Logic Warnings
```bash
# Test admin user without verification
curl -X POST http://localhost:8086/validate/user_profile \
  -H "Content-Type: application/json" \
  -d '{
    "id": "admin_001",
    "username": "admin_user",
    "email": "admin@example.com",
    "first_name": "Admin",
    "last_name": "User",
    "age": 25,
    "bio": "",
    "is_public": true,
    "is_verified": false,
    "role": "admin",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-09-27T14:22:00Z",
    "tags": ["admin"]
  }' | jq .
```

**Expected Response with Warnings:**
```json
{
  "is_valid": true,
  "model_type": "user_profile",
  "provider": "go-playground",
  "errors": [],
  "warnings": [
    {
      "field": "bio",
      "message": "Public profiles should have a bio for better user experience",
      "code": "PUBLIC_PROFILE_NO_BIO",
      "suggestion": "Add a brief bio to help others understand your profile"
    },
    {
      "field": "is_verified",
      "message": "Admin users should be verified for security purposes",
      "code": "ADMIN_NOT_VERIFIED",
      "suggestion": "Complete the verification process for admin accounts"
    }
  ]
}
```

### Test Generic Validation Endpoint
```bash
# Using the generic endpoint with model_type
curl -X POST http://localhost:8086/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "user_profile",
    "payload": {
      "id": "user_123",
      "username": "test_user",
      "email": "test@example.com",
      "first_name": "Test",
      "last_name": "User",
      "age": 25,
      "role": "user",
      "status": "active",
      "created_at": "2024-09-27T10:30:00Z",
      "updated_at": "2024-09-27T14:22:00Z"
    }
  }' | jq .
```

## 🧪 Unit Testing Your New Model

Refer to the [UNIT_TESTING_GUIDE.md](./UNIT_TESTING_GUIDE.md) for comprehensive unit testing instructions.

### Quick Unit Test Example

Create `src/models/user_profile_test.go`:

```go
package models

import (
    "encoding/json"
    "testing"
    "time"
)

func TestUserProfilePayload_Validation(t *testing.T) {
    tests := []struct {
        name        string
        payload     UserProfilePayload
        expectValid bool
    }{
        {
            name: "valid user profile",
            payload: UserProfilePayload{
                ID:        "user_123",
                Username:  "test_user",
                Email:     "test@example.com",
                FirstName: "Test",
                LastName:  "User",
                Age:       25,
                Role:      "user",
                Status:    "active",
                CreatedAt: time.Now(),
                UpdatedAt: time.Now(),
            },
            expectValid: true,
        },
        {
            name: "invalid email",
            payload: UserProfilePayload{
                ID:       "user_123",
                Username: "test_user",
                Email:    "invalid-email",
                Age:      25,
            },
            expectValid: false,
        },
    }

    // Run tests...
}
```

### Run Unit Tests
```bash
# Test your specific model
cd src
go test -v ./models/ -run TestUserProfile

# Test your validator
go test -v ./validations/ -run TestUserProfile

# Run all tests with coverage
go test -v ./... -cover
```

## 📊 E2E Testing Integration

The E2E test suite automatically integrates your new model:

### What the E2E Suite Tests
1. **Model Discovery**: Verifies your model is auto-registered
2. **Endpoint Creation**: Tests `/validate/user_profile` endpoint exists
3. **Valid Payload**: Tests with `test_data/valid/user_profile.json`
4. **Invalid Payload**: Tests with `test_data/invalid/user_profile.json`
5. **HTTP Methods**: Ensures proper HTTP method handling
6. **JSON Format**: Validates response format

### E2E Test Output Example
```
🎯 Phase 4: Existing Model Validation Testing
==============================================
ℹ️  Testing all discovered models with available test data...
🧪 Testing valid payload for user_profile
✅ valid payload validation passed for user_profile
🧪 Testing invalid payload for user_profile
✅ invalid payload validation passed for user_profile
```

## 🎯 Advanced Custom Validations

### Adding Complex Custom Validators

```go
// Register complex custom validator
func NewUserProfileValidator() *UserProfileValidator {
    v := validator.New()

    // Custom validator for strong passwords
    v.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
        password := fl.Field().String()
        return len(password) >= 8 &&
               regexp.MustCompile(`[A-Z]`).MatchString(password) &&
               regexp.MustCompile(`[a-z]`).MatchString(password) &&
               regexp.MustCompile(`[0-9]`).MatchString(password) &&
               regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(password)
    })

    // Custom validator for social media handles
    v.RegisterValidation("social_handle", func(fl validator.FieldLevel) bool {
        handle := fl.Field().String()
        return regexp.MustCompile(`^@[a-zA-Z0-9_]+$`).MatchString(handle)
    })

    return &UserProfileValidator{validator: v}
}
```

### Cross-Field Validation

```go
// ValidatePayload with cross-field validation
func (uv *UserProfileValidator) ValidatePayload(payload models.UserProfilePayload) models.ValidationResult {
    // ... existing validation code ...

    // Cross-field validation: UpdatedAt must be after CreatedAt
    if payload.UpdatedAt.Before(payload.CreatedAt) {
        result.IsValid = false
        result.Errors = append(result.Errors, models.ValidationError{
            Field:   "updated_at",
            Message: "Updated date must be after created date",
            Code:    "INVALID_DATE_ORDER",
            Value:   payload.UpdatedAt.Format(time.RFC3339),
        })
    }

    return result
}
```

## 🚀 Best Practices

### ✅ Model Design Best Practices

1. **Use Clear Field Names**
   ```go
   // Good
   FirstName string `json:"first_name"`
   LastName  string `json:"last_name"`

   // Avoid
   FName string `json:"fname"`
   LName string `json:"lname"`
   ```

2. **Include Comprehensive Validation Tags**
   ```go
   Email string `json:"email" validate:"required,email,max=255"`
   Age   int    `json:"age" validate:"required,min=13,max=120"`
   ```

3. **Use Appropriate Data Types**
   ```go
   // Use time.Time for dates
   CreatedAt time.Time `json:"created_at"`

   // Use pointers for optional dates
   LastLoginAt *time.Time `json:"last_login_at,omitempty"`

   // Use enums with oneof validation
   Status string `json:"status" validate:"oneof=active inactive suspended"`
   ```

4. **Document Complex Fields**
   ```go
   // Preferences stores user-specific configuration as key-value pairs
   // Supported keys: theme, notifications, privacy_level, language
   Preferences map[string]interface{} `json:"preferences,omitempty"`
   ```

### ✅ Validation Best Practices

1. **Layer Your Validations**
   ```go
   // Layer 1: Struct validation (go-playground/validator)
   if err := uv.validator.Struct(payload); err != nil {
       // Handle basic validation errors
   }

   // Layer 2: Custom field validation
   if result.IsValid {
       // Apply custom validations
   }

   // Layer 3: Business logic warnings
   result.Warnings = uv.validateBusinessLogic(payload)
   ```

2. **Provide Helpful Error Messages**
   ```go
   func (uv *UserProfileValidator) getCustomErrorMessage(ve validator.FieldError) string {
       switch ve.Tag() {
       case "email":
           return "Please provide a valid email address (e.g., user@example.com)"
       case "min":
           return fmt.Sprintf("Must be at least %s characters long", ve.Param())
       default:
           return "Please check this field and try again"
       }
   }
   ```

3. **Use Meaningful Error Codes**
   ```go
   // Good: Specific, actionable error codes
   "USERNAME_ALREADY_EXISTS"
   "INVALID_EMAIL_FORMAT"
   "AGE_BELOW_MINIMUM"

   // Avoid: Generic error codes
   "ERROR_001"
   "VALIDATION_FAILED"
   "INVALID_INPUT"
   ```

### ✅ Testing Best Practices

1. **Comprehensive Test Data**
   ```bash
   test_data/
   ├── valid/
   │   ├── user_profile.json          # Standard valid case
   │   ├── user_profile_minimal.json  # Minimal required fields
   │   └── user_profile_admin.json    # Admin role specific
   ├── invalid/
   │   ├── user_profile.json          # Multiple validation errors
   │   ├── user_profile_email.json    # Email-specific errors
   │   └── user_profile_age.json      # Age-specific errors
   └── examples/
       └── user_profile_sample.json   # Complete example
   ```

2. **Test Edge Cases**
   ```go
   // Test boundary values
   Age: 13,  // Minimum age
   Age: 120, // Maximum age

   // Test empty and null values
   Bio: "",
   LastLoginAt: nil,

   // Test maximum lengths
   Username: strings.Repeat("a", 30), // Max username length
   ```

3. **Test Business Logic Scenarios**
   ```go
   // Test role-specific validations
   {Role: "admin", Age: 17, ExpectError: true},     // Admin too young
   {Role: "admin", IsVerified: false, ExpectWarning: true}, // Admin not verified
   ```

### ✅ Code Organization Best Practices

1. **Follow Naming Conventions**
   ```go
   // Model file: user_profile.go
   // Model struct: UserProfilePayload
   // Validator file: user_profile.go
   // Validator struct: UserProfileValidator
   // Constructor: NewUserProfileValidator
   ```

2. **Keep Validators Focused**
   ```go
   // One validator per model
   type UserProfileValidator struct {
       validator *validator.Validate
   }

   // Separate concerns
   func (uv *UserProfileValidator) ValidatePayload(payload models.UserProfilePayload) models.ValidationResult
   func (uv *UserProfileValidator) validateBusinessLogic(payload models.UserProfilePayload) []models.ValidationWarning
   func (uv *UserProfileValidator) getCustomErrorMessage(ve validator.FieldError) string
   ```

3. **Use Constants for Magic Values**
   ```go
   const (
       MinUserAge = 13
       MaxUserAge = 120
       MaxUsernameLength = 30
       MaxBioLength = 500
   )
   ```

## 🔧 Troubleshooting Common Issues

### Model Not Discovered
**Problem**: Your model doesn't appear in `/models` endpoint

**Solutions**:
1. Check naming convention: `{ModelName}Payload` struct
2. Ensure the model file is in `src/models/` directory
3. Verify the validator file is in `src/validations/` directory
4. Check constructor naming: `New{ModelName}Validator`
5. Restart the server to trigger re-discovery

### Validation Not Working
**Problem**: Custom validations are not being applied

**Solutions**:
1. Check `ValidatePayload` method signature matches interface
2. Ensure custom validators are registered in constructor
3. Verify error codes and messages are being set correctly
4. Check that struct validation is running before custom validation

### Test Data Not Found
**Problem**: E2E tests show "No test data found for model"

**Solutions**:
1. Ensure JSON files are in correct directories:
   - `test_data/valid/{model_name}.json`
   - `test_data/invalid/{model_name}.json`
2. Check JSON syntax is valid
3. Verify model name matches exactly (case-sensitive)

### Build Errors
**Problem**: Compilation errors when building

**Solutions**:
1. Run `go mod tidy` to resolve dependencies
2. Check import paths are correct
3. Ensure all required fields have proper struct tags
4. Verify go syntax is correct

## 📈 Performance Considerations

### Efficient Validation Design

1. **Pre-compile Regular Expressions**
   ```go
   var (
       usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
       bioSafetyRegex = regexp.MustCompile(`(?i)(script|javascript|onclick)`)
   )

   func validateUsernameFormat(fl validator.FieldLevel) bool {
       return usernameRegex.MatchString(fl.Field().String())
   }
   ```

2. **Minimize Database Calls**
   ```go
   // Cache frequently checked values
   var reservedUsernamesMap = map[string]bool{
       "admin": true,
       "root":  true,
       "api":   true,
   }

   func (uv *UserProfileValidator) validateUsernameUniqueness(username string) error {
       if reservedUsernamesMap[strings.ToLower(username)] {
           return fmt.Errorf("username '%s' is reserved", username)
       }
       // Only check database if not in reserved list
       return uv.checkDatabaseForUsername(username)
   }
   ```

3. **Use Efficient Data Structures**
   ```go
   // Use slices for small datasets
   allowedRoles := []string{"user", "admin", "moderator"}

   // Use maps for larger datasets or frequent lookups
   allowedRolesMap := map[string]bool{
       "user":      true,
       "admin":     true,
       "moderator": true,
   }
   ```

## 🎉 Conclusion

Adding new models to the Go Playground Data Validator is designed to be:
- ✅ **Simple**: Just add 2 files and restart
- ✅ **Automatic**: Zero configuration required
- ✅ **Safe**: Model-agnostic core tests prevent breakage
- ✅ **Testable**: Comprehensive testing at multiple levels
- ✅ **Maintainable**: Clear patterns and best practices

By following this guide, you can confidently add new models and validations without breaking existing functionality while maintaining high code quality and comprehensive test coverage.

**Happy Coding! 🚀**