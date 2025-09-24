# Complete go-playground/validator Usage Guide

## Introduction

The go-playground/validator library is Go's most comprehensive struct and field validation package, offering extensive validation capabilities through a simple tag-based approach. This guide provides everything you need to implement robust validation in production Go applications.

## Quick Setup and Installation

### Current Version: v10.27.0 (July 2025)

```bash
go get github.com/go-playground/validator/v10
```

### Basic Setup (Recommended for New Projects)

```go
import "github.com/go-playground/validator/v10"

// Initialize with recommended options for v11+ compatibility
var validate = validator.New(validator.WithRequiredStructEnabled())

type User struct {
    Name  string `validate:"required,min=2,max=100"`
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=18,lte=120"`
}

func main() {
    user := User{Name: "John", Email: "john@example.com", Age: 25}
    
    if err := validate.Struct(user); err != nil {
        for _, err := range err.(validator.ValidationErrors) {
            fmt.Printf("Field: %s, Error: %s\n", err.Field(), err.Tag())
        }
    }
}
```

## Latest Release Features (v10+)

**Key updates in recent versions:**
- **v10.27.0**: Russian E.164 improvements, bug fixes
- **v10.26.0**: EIN validation, enhanced translations (Indonesian, German, Korean, Thai)
- **v10.25.0**: Support for omitting empty/zero values including nil pointers
- **v10.23.0**: Case-insensitive `oneofci` validator, cron validation
- **v10.19.0**: Optional private field validation via `WithPrivateFieldValidation()`

**Performance characteristics:**
- Field validation: **27.88 ns/op** (success), **121.3 ns/op** (failure)
- Complex struct validation: **576.0 ns/op** (success)
- Thread-safe singleton design with struct caching

## Complete Validation Tags Reference

### Basic Constraints
```go
type Basic struct {
    Required    string `validate:"required"`           // Must not be zero value
    Optional    string `validate:"omitempty,min=5"`    // Skip if empty
    ExactLength string `validate:"len=10"`             // Exact length
    Range       int    `validate:"min=1,max=100"`      // Range validation
    OneOf       string `validate:"oneof=red green blue"` // Must be one of values
    OneOfCI     string `validate:"oneofci=Red GREEN blue"` // Case-insensitive oneof
}
```

### String Validations
```go
type StringValidation struct {
    Email       string `validate:"email"`              // Valid email
    URL         string `validate:"url"`                // Valid URL
    UUID        string `validate:"uuid4"`              // Valid UUID v4
    Alpha       string `validate:"alpha"`              // Only letters
    Alphanumeric string `validate:"alphanum"`          // Letters and numbers
    Numeric     string `validate:"numeric"`            // Only numbers
    Hexadecimal string `validate:"hexadecimal"`        // Valid hex
    Contains    string `validate:"contains=@"`         // Contains substring
    StartsWith  string `validate:"startswith=Mr."`     // Starts with prefix
    EndsWith    string `validate:"endswith=.com"`      // Ends with suffix
}
```

### Network and Format Validations
```go
type NetworkValidation struct {
    IPv4     string `validate:"ipv4"`           // Valid IPv4
    IPv6     string `validate:"ipv6"`           // Valid IPv6
    CIDR     string `validate:"cidr"`           // Valid CIDR notation
    MAC      string `validate:"mac"`            // Valid MAC address
    TCP      string `validate:"tcp_addr"`       // Valid TCP address
    FQDN     string `validate:"fqdn"`           // Fully qualified domain
    Hostname string `validate:"hostname"`       // Valid hostname
}
```

### Financial and Identity
```go
type FinancialValidation struct {
    CreditCard string `validate:"credit_card"`    // Valid credit card
    SSN        string `validate:"ssn"`            // US Social Security Number
    EIN        string `validate:"ein"`            // US Employer ID Number
    Bitcoin    string `validate:"btc_addr"`       // Bitcoin address
    Ethereum   string `validate:"eth_addr"`       // Ethereum address
    BIC        string `validate:"bic"`            // SWIFT code
    Currency   string `validate:"iso4217"`        // Currency code (USD, EUR)
}
```

### Geographic and Locale
```go
type GeoValidation struct {
    Country   string  `validate:"iso3166_1_alpha2"`     // 2-letter country (US, CA)
    Timezone  string  `validate:"timezone"`             // Valid timezone
    Language  string  `validate:"bcp47_language_tag"`   // Language tag (en-US)
    Latitude  float64 `validate:"latitude"`             // Valid latitude
    Longitude float64 `validate:"longitude"`            // Valid longitude
    Postcode  string  `validate:"postcode_iso3166_alpha2=US"` // US postal code
}
```

### Cross-Field Validation
```go
type CrossFieldValidation struct {
    Password        string `validate:"required,min=8"`
    ConfirmPassword string `validate:"required,eqfield=Password"`    // Must equal Password
    StartDate       string `validate:"required"`
    EndDate         string `validate:"required,gtfield=StartDate"`   // Must be after StartDate
    MinPrice        int    `validate:"required"`
    MaxPrice        int    `validate:"required,gtefield=MinPrice"`   // Must be >= MinPrice
}
```

### Advanced Structure Validation
```go
type AdvancedValidation struct {
    Tags        []string           `validate:"required,min=1,dive,required,min=2"`  // Validate each element
    Settings    map[string]string  `validate:"required,dive,keys,required,endkeys,required"`  // Validate keys and values
    Nested      *NestedStruct      `validate:"required"`                            // Validate nested struct
    OptionalRef *NestedStruct      `validate:"omitempty"`                          // Validate if present
}

type NestedStruct struct {
    Value string `validate:"required"`
}
```

## Real-World Data Structure Examples

### GitHub API Validation

Complete GitHub API data structures with comprehensive validation:

```go
type PullRequest struct {
    ID          int64       `validate:"required,gt=0"`
    Number      int         `validate:"required,gt=0"`
    Title       string      `validate:"required,min=1,max=256"`
    Body        string      `validate:"omitempty,max=65536"`
    State       string      `validate:"required,oneof=open closed merged"`
    Draft       bool        `validate:"omitempty"`
    Author      User        `validate:"required"`
    Assignees   []User      `validate:"omitempty,dive"`
    Repository  Repository  `validate:"required"`
    HeadRef     Reference   `validate:"required"`
    BaseRef     Reference   `validate:"required"`
    Commits     []Commit    `validate:"required,min=1,dive"`
    Labels      []Label     `validate:"omitempty,dive"`
    Milestone   *Milestone  `validate:"omitempty"`
    CreatedAt   time.Time   `validate:"required"`
    UpdatedAt   time.Time   `validate:"required,gtefield=CreatedAt"`
    MergedAt    *time.Time  `validate:"omitempty"`
    ClosedAt    *time.Time  `validate:"omitempty"`
}

type User struct {
    ID        int64  `validate:"required,gt=0"`
    Login     string `validate:"required,min=1,max=39,alphanum"`
    Email     string `validate:"omitempty,email"`
    Name      string `validate:"omitempty,max=255"`
    Company   string `validate:"omitempty,max=255"`
    Location  string `validate:"omitempty,max=255"`
    Bio       string `validate:"omitempty,max=160"`
    AvatarURL string `validate:"omitempty,url"`
    HTMLURL   string `validate:"required,url"`
    Type      string `validate:"required,oneof=User Organization"`
    SiteAdmin bool   `validate:"omitempty"`
}

type Repository struct {
    ID              int64     `validate:"required,gt=0"`
    Name            string    `validate:"required,min=1,max=100"`
    FullName        string    `validate:"required,min=1,max=100"`
    Owner           User      `validate:"required"`
    Private         bool      `validate:"omitempty"`
    HTMLURL         string    `validate:"required,url"`
    Description     string    `validate:"omitempty,max=350"`
    Language        string    `validate:"omitempty,max=50"`
    DefaultBranch   string    `validate:"required,min=1,max=255"`
    Visibility      string    `validate:"required,oneof=public private internal"`
    Size            int       `validate:"gte=0"`
    StargazersCount int       `validate:"gte=0"`
    ForksCount      int       `validate:"gte=0"`
    CreatedAt       time.Time `validate:"required"`
    UpdatedAt       time.Time `validate:"required,gtefield=CreatedAt"`
    PushedAt        time.Time `validate:"required"`
}

type Commit struct {
    SHA       string    `validate:"required,len=40,hexadecimal"`
    Message   string    `validate:"required,min=1"`
    Author    CommitUser `validate:"required"`
    Committer CommitUser `validate:"required"`
    Timestamp time.Time  `validate:"required"`
    TreeSHA   string     `validate:"required,len=40,hexadecimal"`
    Parents   []string   `validate:"omitempty,dive,len=40,hexadecimal"`
    URL       string     `validate:"required,url"`
}

type CommitUser struct {
    Name  string `validate:"required,min=1"`
    Email string `validate:"required,email"`
    Date  time.Time `validate:"required"`
}
```

### Company Team Structure Validation

Enterprise-grade organizational data validation:

```go
type Employee struct {
    ID             uuid.UUID        `validate:"required,uuid4"`
    PersonalInfo   PersonalInfo     `validate:"required"`
    ContactInfo    ContactInfo      `validate:"required"`
    EmploymentInfo EmploymentInfo   `validate:"required"`
    Roles          []Role           `validate:"required,min=1,dive"`
    Departments    []uuid.UUID      `validate:"required,min=1,dive,uuid4"`
    Manager        *uuid.UUID       `validate:"omitempty,uuid4"`
    DirectReports  []uuid.UUID      `validate:"omitempty,dive,uuid4"`
    Projects       []uuid.UUID      `validate:"omitempty,dive,uuid4"`
    Skills         []string         `validate:"omitempty,dive,min=2,max=50"`
    Certifications []Certification  `validate:"omitempty,dive"`
    Performance    PerformanceInfo  `validate:"omitempty"`
    CreatedAt      time.Time        `validate:"required"`
    UpdatedAt      time.Time        `validate:"required,gtefield=CreatedAt"`
    Status         string           `validate:"required,oneof=active inactive terminated"`
}

type PersonalInfo struct {
    FirstName    string     `validate:"required,min=1,max=50,alpha"`
    LastName     string     `validate:"required,min=1,max=50,alpha"`
    MiddleName   string     `validate:"omitempty,max=50,alpha"`
    DateOfBirth  *time.Time `validate:"omitempty"`
    Gender       string     `validate:"omitempty,oneof=male female other prefer_not_to_say"`
    MaritalStatus string    `validate:"omitempty,oneof=single married divorced widowed"`
    Nationality  string     `validate:"omitempty,iso3166_1_alpha2"`
    Languages    []string   `validate:"omitempty,dive,bcp47_language_tag"`
}

type ContactInfo struct {
    WorkEmail     string  `validate:"required,email"`
    PersonalEmail string  `validate:"omitempty,email"`
    WorkPhone     string  `validate:"omitempty,e164"`
    MobilePhone   string  `validate:"omitempty,e164"`
    Address       Address `validate:"required"`
    EmergencyContact EmergencyContact `validate:"required"`
}

type Address struct {
    Street1    string `validate:"required,min=5,max=100"`
    Street2    string `validate:"omitempty,max=100"`
    City       string `validate:"required,min=2,max=50"`
    State      string `validate:"required,min=2,max=50"`
    PostalCode string `validate:"required,min=3,max=10"`
    Country    string `validate:"required,iso3166_1_alpha2"`
    Timezone   string `validate:"omitempty,timezone"`
}

type Department struct {
    ID          uuid.UUID `validate:"required,uuid4"`
    Name        string    `validate:"required,min=2,max=100"`
    Description string    `validate:"omitempty,max=500"`
    Code        string    `validate:"required,min=2,max=10,alphanum"`
    Manager     uuid.UUID `validate:"required,uuid4"`
    Budget      Budget    `validate:"omitempty"`
    Location    Address   `validate:"omitempty"`
    ParentDept  *uuid.UUID `validate:"omitempty,uuid4"`
    SubDepts    []uuid.UUID `validate:"omitempty,dive,uuid4"`
    Employees   []uuid.UUID `validate:"omitempty,dive,uuid4"`
    CostCenter  string    `validate:"omitempty,alphanum"`
    Status      string    `validate:"required,oneof=active inactive"`
    CreatedAt   time.Time `validate:"required"`
    UpdatedAt   time.Time `validate:"required,gtefield=CreatedAt"`
}

type Budget struct {
    Amount       float64 `validate:"required,gt=0"`
    Currency     string  `validate:"required,iso4217"`
    FiscalYear   int     `validate:"required,gte=2020,lte=2030"`
    Spent        float64 `validate:"gte=0,ltefield=Amount"`
    LastUpdated  time.Time `validate:"required"`
}
```

### CI/CD Pipeline Data Validation

Modern DevOps pipeline structure validation:

```go
type Pipeline struct {
    ID            uuid.UUID         `validate:"required,uuid4"`
    Name          string           `validate:"required,min=1,max=100"`
    Description   string           `validate:"omitempty,max=500"`
    Repository    Repository       `validate:"required"`
    Triggers      []Trigger        `validate:"required,min=1,dive"`
    Stages        []Stage          `validate:"required,min=1,dive"`
    Variables     map[string]string `validate:"omitempty,dive,keys,alphanum,endkeys,required"`
    Secrets       []string         `validate:"omitempty,dive,min=1"`
    Timeout       int              `validate:"gte=60,lte=86400"` // 1 min to 24 hours
    Retry         RetryPolicy      `validate:"omitempty"`
    Notifications []Notification   `validate:"omitempty,dive"`
    Schedule      CronSchedule     `validate:"omitempty"`
    Status        string           `validate:"required,oneof=active paused disabled"`
    CreatedAt     time.Time        `validate:"required"`
    UpdatedAt     time.Time        `validate:"required,gtefield=CreatedAt"`
    CreatedBy     uuid.UUID        `validate:"required,uuid4"`
}

type Stage struct {
    ID           string              `validate:"required,min=1,max=50,alphanum"`
    Name         string              `validate:"required,min=1,max=100"`
    Jobs         []Job               `validate:"required,min=1,dive"`
    Dependencies []string            `validate:"omitempty,dive,min=1,max=50"`
    Condition    string              `validate:"omitempty,oneof=always on_success on_failure manual"`
    Environment  string              `validate:"omitempty,oneof=development staging production"`
    Timeout      int                 `validate:"gte=60,lte=7200"` // 1 min to 2 hours
    Variables    map[string]string   `validate:"omitempty,dive,keys,alphanum,endkeys,required"`
    Artifacts    ArtifactConfig      `validate:"omitempty"`
}

type Job struct {
    ID          string            `validate:"required,min=1,max=50,alphanum"`
    Name        string            `validate:"required,min=1,max=100"`
    Image       string            `validate:"required,min=1"`
    Commands    []string          `validate:"required,min=1,dive,min=1"`
    Environment map[string]string `validate:"omitempty,dive,keys,alphanum,endkeys,required"`
    Resources   ResourceLimits    `validate:"omitempty"`
    HealthCheck HealthCheck       `validate:"omitempty"`
    Retry       RetryPolicy       `validate:"omitempty"`
    Timeout     int               `validate:"gte=30,lte=3600"` // 30 sec to 1 hour
}

type BuildProcess struct {
    ID              uuid.UUID        `validate:"required,uuid4"`
    PipelineID      uuid.UUID        `validate:"required,uuid4"`
    BuildNumber     int              `validate:"required,gt=0"`
    Branch          string           `validate:"required,min=1,max=255"`
    CommitSHA       string           `validate:"required,len=40,hexadecimal"`
    Status          string           `validate:"required,oneof=pending running success failed canceled"`
    StartTime       time.Time        `validate:"required"`
    EndTime         *time.Time       `validate:"omitempty,gtfield=StartTime"`
    Duration        int              `validate:"gte=0"` // seconds
    Artifacts       []Artifact       `validate:"omitempty,dive"`
    TestResults     TestResults      `validate:"omitempty"`
    SecurityScans   []SecurityScan   `validate:"omitempty,dive"`
    QualityGates    []QualityGate    `validate:"omitempty,dive"`
    Logs            []LogEntry       `validate:"omitempty,dive"`
    TriggeredBy     TriggerSource    `validate:"required"`
    Environment     BuildEnvironment `validate:"required"`
}

type DeploymentRecord struct {
    ID               uuid.UUID         `validate:"required,uuid4"`
    BuildID          uuid.UUID         `validate:"required,uuid4"`
    Environment      string            `validate:"required,oneof=development staging production"`
    Strategy         string            `validate:"required,oneof=rolling blue_green canary recreate"`
    Status           string            `validate:"required,oneof=pending deploying success failed rolled_back"`
    Version          string            `validate:"required,semver"`
    StartTime        time.Time         `validate:"required"`
    EndTime          *time.Time        `validate:"omitempty,gtfield=StartTime"`
    Duration         int               `validate:"gte=0"`
    RollbackData     *RollbackInfo     `validate:"omitempty"`
    HealthChecks     []HealthCheck     `validate:"required,min=1,dive"`
    ApprovalRequired bool              `validate:"omitempty"`
    Approvers        []uuid.UUID       `validate:"omitempty,dive,uuid4"`
    ApprovalTime     *time.Time        `validate:"omitempty"`
    Configuration    DeploymentConfig  `validate:"required"`
    Monitoring       MonitoringConfig  `validate:"required"`
    DeployedBy       uuid.UUID         `validate:"required,uuid4"`
}
```

## Performance Best Practices

### Singleton Pattern (Critical)

**✅ Correct approach:**
```go
// Global validator instance - caches struct metadata
var validate = validator.New(validator.WithRequiredStructEnabled())

func ValidateUser(user User) error {
    return validate.Struct(user)
}
```

**❌ Avoid - Creates new instance every call:**
```go
func ValidateUser(user User) error {
    validate := validator.New() // Loses caching benefits!
    return validate.Struct(user)
}
```

### High-Performance Validation Patterns

```go
// Object pooling for high-frequency validation
var userPool = sync.Pool{
    New: func() interface{} {
        return &User{}
    },
}

// Batch validation with goroutines
func ValidateConcurrently(items []User) []error {
    errorsChan := make(chan error, len(items))
    var wg sync.WaitGroup
    
    // Limit concurrent goroutines
    sem := make(chan struct{}, runtime.GOMAXPROCS(0))
    
    for _, item := range items {
        wg.Add(1)
        go func(u User) {
            defer wg.Done()
            sem <- struct{}{}        // Acquire
            defer func() { <-sem }() // Release
            
            if err := validate.Struct(u); err != nil {
                errorsChan <- err
            }
        }(item)
    }
    
    wg.Wait()
    close(errorsChan)
    
    var errors []error
    for err := range errorsChan {
        errors = append(errors, err)
    }
    return errors
}
```

### Memory Optimization

```go
// Pre-allocate error slices with known capacity
func ValidateBatch(items []User) []error {
    errors := make([]error, 0, len(items)) // Pre-allocate capacity
    for _, item := range items {
        if err := validate.Struct(item); err != nil {
            errors = append(errors, err)
        }
    }
    return errors
}

// Custom validators with precompiled regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func fastEmailValidator(fl validator.FieldLevel) bool {
    return emailRegex.MatchString(fl.Field().String())
}
```

## Framework Integration Examples

### Gin Framework Integration

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
    Password string `json:"password" validate:"required,min=8"`
}

// Generic validation wrapper using Go generics
func ValidateJSON[T any](c *gin.Context) (*T, error) {
    var req T
    if err := c.ShouldBindJSON(&req); err != nil {
        return nil, err
    }
    
    if err := validate.Struct(req); err != nil {
        return nil, err
    }
    
    return &req, nil
}

func createUser(c *gin.Context) {
    req, err := ValidateJSON[CreateUserRequest](c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "validation failed",
            "details": formatValidationError(err),
        })
        return
    }
    
    // Process validated request
    c.JSON(http.StatusCreated, gin.H{
        "message": "User created successfully",
        "user": req,
    })
}

func formatValidationError(err error) map[string]string {
    errors := make(map[string]string)
    
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            errors[e.Field()] = fmt.Sprintf("Failed validation: %s", e.Tag())
        }
    }
    
    return errors
}

func main() {
    r := gin.Default()
    r.POST("/users", createUser)
    r.Run(":8080")
}
```

### Echo Framework Integration

```go
package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/go-playground/validator/v10"
)

type CustomValidator struct {
    validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
    return cv.validator.Struct(i)
}

type User struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=18,lte=120"`
}

func createUser(c echo.Context) error {
    u := new(User)
    if err := c.Bind(u); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    if err := c.Validate(u); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    return c.JSON(http.StatusCreated, u)
}

func main() {
    e := echo.New()
    e.Validator = &CustomValidator{validator: validator.New()}
    e.POST("/users", createUser)
    e.Logger.Fatal(e.Start(":8080"))
}
```

### GORM Database Integration

```go
type User struct {
    ID       uint   `gorm:"primarykey"`
    Name     string `gorm:"not null" validate:"required,min=2,max=50"`
    Email    string `gorm:"unique;not null" validate:"required,email"`
    Age      int    `validate:"gte=18,lte=120"`
    Profile  Profile `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Profile struct {
    ID     uint   `gorm:"primarykey"`
    UserID uint
    Bio    string `validate:"omitempty,max=500"`
    Phone  string `validate:"omitempty,e164"`
}

// GORM hooks with validation
func (u *User) BeforeCreate(tx *gorm.DB) error {
    return validate.Struct(u)
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
    return validate.Struct(u)
}

// Service layer with validation
type UserService struct {
    db *gorm.DB
}

func (s *UserService) CreateUser(user *User) error {
    // Validate before database operation
    if err := validate.Struct(user); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Additional business logic validation
    if s.EmailExists(user.Email) {
        return errors.New("email already exists")
    }
    
    return s.db.Create(user).Error
}

func (s *UserService) EmailExists(email string) bool {
    var count int64
    s.db.Model(&User{}).Where("email = ?", email).Count(&count)
    return count > 0
}
```

## Custom Validation Examples

### Field-Level Custom Validators

```go
// Custom validator for adult age verification
func validateAdultAge(fl validator.FieldLevel) bool {
    age := fl.Field().Int()
    return age >= 18 && age <= 120
}

// Custom password strength validator
func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    
    // At least 8 characters
    if len(password) < 8 {
        return false
    }
    
    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )
    
    for _, char := range password {
        switch {
        case 'A' <= char && char <= 'Z':
            hasUpper = true
        case 'a' <= char && char <= 'z':
            hasLower = true
        case '0' <= char && char <= '9':
            hasNumber = true
        case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
            hasSpecial = true
        }
    }
    
    return hasUpper && hasLower && hasNumber && hasSpecial
}

// Register custom validators
func init() {
    validate.RegisterValidation("adult_age", validateAdultAge)
    validate.RegisterValidation("strong_password", validateStrongPassword)
}

// Usage in structs
type Registration struct {
    Age      int    `validate:"adult_age"`
    Password string `validate:"strong_password"`
}
```

### Struct-Level Custom Validation

```go
// Complex business logic validation at struct level
func UserRegistrationValidation(sl validator.StructLevel) {
    user := sl.Current().Interface().(UserRegistration)
    
    // Business rule: Premium users must be adults
    if user.AccountType == "premium" && user.Age < 21 {
        sl.ReportError(user.Age, "Age", "Age", "premium_adult", "")
    }
    
    // Business rule: Corporate emails required for business accounts
    if user.AccountType == "business" && !strings.HasSuffix(user.Email, ".com") {
        sl.ReportError(user.Email, "Email", "Email", "corporate_email", "")
    }
    
    // Business rule: First and last names cannot be identical
    if user.FirstName == user.LastName {
        sl.ReportError(user.FirstName, "FirstName", "FirstName", "nefield", "LastName")
    }
}

// Register struct-level validation
func init() {
    validate.RegisterStructValidation(UserRegistrationValidation, UserRegistration{})
}

type UserRegistration struct {
    FirstName   string `validate:"required,min=2,max=50"`
    LastName    string `validate:"required,min=2,max=50"`
    Email       string `validate:"required,email"`
    Age         int    `validate:"required,gte=13"`
    AccountType string `validate:"required,oneof=basic premium business"`
}
```

## Advanced Error Handling

### Comprehensive Error Processing

```go
func ProcessValidationErrors(err error) map[string]interface{} {
    if err == nil {
        return nil
    }
    
    result := make(map[string]interface{})
    
    // Handle invalid validation error (structural issues)
    var invalidValidationError *validator.InvalidValidationError
    if errors.As(err, &invalidValidationError) {
        result["error"] = "Invalid validation configuration"
        result["details"] = err.Error()
        return result
    }
    
    // Handle validation errors (actual validation failures)
    var validationErrors validator.ValidationErrors
    if errors.As(err, &validationErrors) {
        fieldErrors := make(map[string]interface{})
        
        for _, err := range validationErrors {
            fieldErrors[err.Field()] = map[string]interface{}{
                "tag":    err.Tag(),
                "value":  err.Value(),
                "param":  err.Param(),
                "message": getCustomMessage(err),
            }
        }
        
        result["validation_errors"] = fieldErrors
        result["error_count"] = len(validationErrors)
        return result
    }
    
    // Handle other errors
    result["error"] = "Unknown validation error"
    result["details"] = err.Error()
    return result
}

func getCustomMessage(err validator.FieldError) string {
    switch err.Tag() {
    case "required":
        return fmt.Sprintf("%s is required", err.Field())
    case "email":
        return fmt.Sprintf("%s must be a valid email address", err.Field())
    case "min":
        return fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
    case "max":
        return fmt.Sprintf("%s must be no more than %s characters long", err.Field(), err.Param())
    default:
        return fmt.Sprintf("%s failed validation: %s", err.Field(), err.Tag())
    }
}
```

### Localization and Custom Messages

```go
import (
    "github.com/go-playground/locales/en"
    ut "github.com/go-playground/universal-translator"
    en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
    validate *validator.Validate
    trans    ut.Translator
)

func setupValidator() {
    validate = validator.New()
    
    // Setup translator
    english := en.New()
    uni := ut.New(english, english)
    trans, _ = uni.GetTranslator("en")
    
    // Register default translations
    en_translations.RegisterDefaultTranslations(validate, trans)
    
    // Register custom translations
    validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
        return ut.Add("required", "{0} is a required field", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("required", fe.Field())
        return t
    })
    
    validate.RegisterTranslation("strong_password", trans, func(ut ut.Translator) error {
        return ut.Add("strong_password", "{0} must contain at least one uppercase letter, one lowercase letter, one number, and one special character", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("strong_password", fe.Field())
        return t
    })
}

func translateValidationErrors(err error) []string {
    var messages []string
    
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            messages = append(messages, e.Translate(trans))
        }
    }
    
    return messages
}
```

## Great Expectations Comparison

### When to Choose go-playground/validator vs Great Expectations

**Choose go-playground/validator for:**
- **Web APIs and microservices** - Request/response validation
- **High-performance applications** - Sub-30ns validation times
- **Real-time systems** - Low-latency concurrent validation
- **Go applications** - Native Go integration and type safety
- **Simple data structures** - Struct and field validation
- **Development velocity** - Quick implementation with tags

**Choose Great Expectations for:**
- **Data engineering pipelines** - ETL and data processing workflows
- **Large dataset validation** - Multi-million row data quality checks
- **Data science workflows** - ML pipeline data validation
- **Business stakeholder communication** - Generated documentation and reports
- **Multi-source data** - Databases, files, cloud storage, streams
- **Regulatory compliance** - Audit trails and data governance

### Key Architectural Differences

| Aspect | go-playground/validator | Great Expectations |
|--------|------------------------|-------------------|
| **Scope** | Application-level validation | Data infrastructure validation |
| **Performance** | 30ns per validation | 100ms-10s+ per dataset |
| **Target Data** | Individual structs/objects | Entire datasets and tables |
| **Documentation** | Code-level error messages | Business-readable reports |
| **Setup Complexity** | Minutes (tag-based) | Days to weeks (configuration) |
| **Resource Usage** | Minimal overhead | High memory and CPU usage |
| **Integration** | Web frameworks, APIs | Data pipelines, ETL tools |

### Complementary Usage Pattern

Many organizations use both tools together:

```go
// API layer validation with go-playground/validator
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    
    // Fast API validation
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    if err := validate.Struct(req); err != nil {
        http.Error(w, "Validation failed", http.StatusBadRequest)
        return
    }
    
    // Process request...
}

// Meanwhile, Great Expectations validates the data pipeline
// that processes user data in batch operations:
# great_expectations/expectations/user_data_suite.json
{
  "expectations": [
    {
      "expectation_type": "expect_column_values_to_be_unique",
      "kwargs": {"column": "email"}
    },
    {
      "expectation_type": "expect_column_values_to_match_regex",
      "kwargs": {
        "column": "email",
        "regex": "^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$"
      }
    }
  ]
}
```

## Testing Patterns

### Comprehensive Unit Testing

```go
package main

import (
    "testing"
    "time"
    "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
)

func TestUserValidation(t *testing.T) {
    validate := validator.New()
    
    tests := []struct {
        name    string
        user    User
        wantErr bool
        errField string
        errTag   string
    }{
        {
            name: "valid user",
            user: User{
                Name:  "John Doe",
                Email: "john@example.com",
                Age:   25,
            },
            wantErr: false,
        },
        {
            name: "missing name",
            user: User{
                Email: "john@example.com",
                Age:   25,
            },
            wantErr:  true,
            errField: "Name",
            errTag:   "required",
        },
        {
            name: "invalid email",
            user: User{
                Name:  "John Doe",
                Email: "invalid-email",
                Age:   25,
            },
            wantErr:  true,
            errField: "Email",
            errTag:   "email",
        },
        {
            name: "age too young",
            user: User{
                Name:  "John Doe",
                Email: "john@example.com",
                Age:   17,
            },
            wantErr:  true,
            errField: "Age",
            errTag:   "gte",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validate.Struct(tt.user)
            
            if tt.wantErr {
                assert.Error(t, err)
                
                if validationErrors, ok := err.(validator.ValidationErrors); ok {
                    found := false
                    for _, e := range validationErrors {
                        if e.Field() == tt.errField && e.Tag() == tt.errTag {
                            found = true
                            break
                        }
                    }
                    assert.True(t, found, "Expected error for field %s with tag %s", tt.errField, tt.errTag)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

// Benchmark validation performance
func BenchmarkUserValidation(b *testing.B) {
    validate := validator.New()
    user := User{
        Name:  "John Doe",
        Email: "john@example.com",
        Age:   25,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = validate.Struct(user)
    }
}

// Test custom validators
func TestCustomValidators(t *testing.T) {
    validate := validator.New()
    validate.RegisterValidation("adult_age", validateAdultAge)
    
    type TestStruct struct {
        Age int `validate:"adult_age"`
    }
    
    // Test valid age
    err := validate.Struct(TestStruct{Age: 25})
    assert.NoError(t, err)
    
    // Test invalid age
    err = validate.Struct(TestStruct{Age: 16})
    assert.Error(t, err)
}
```

### Integration Testing with HTTP

```go
func TestCreateUserEndpoint(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := setupRouter()
    
    tests := []struct {
        name           string
        requestBody    string
        expectedStatus int
        checkResponse  func(t *testing.T, response *httptest.ResponseRecorder)
    }{
        {
            name: "valid user creation",
            requestBody: `{
                "name": "John Doe",
                "email": "john@example.com",
                "age": 25,
                "password": "SecurePass123!"
            }`,
            expectedStatus: http.StatusCreated,
            checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
                assert.Contains(t, response.Body.String(), "User created successfully")
            },
        },
        {
            name: "invalid email",
            requestBody: `{
                "name": "John Doe",
                "email": "invalid-email",
                "age": 25,
                "password": "SecurePass123!"
            }`,
            expectedStatus: http.StatusBadRequest,
            checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
                assert.Contains(t, response.Body.String(), "validation failed")
                assert.Contains(t, response.Body.String(), "Email")
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req, _ := http.NewRequest("POST", "/users", strings.NewReader(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            if tt.checkResponse != nil {
                tt.checkResponse(t, w)
            }
        })
    }
}
```

## Production Best Practices Summary

1. **Always use singleton pattern** - One validator instance per application
2. **Enable struct validation** - Use `WithRequiredStructEnabled()` for new projects
3. **Implement proper error handling** - Use `errors.As()` for type-safe error processing
4. **Optimize for performance** - Pre-allocate slices, use object pools for high-frequency validation
5. **Use custom validators wisely** - For domain-specific business rules
6. **Implement localization** - For user-facing applications
7. **Write comprehensive tests** - Unit tests and integration tests for validation logic
8. **Monitor performance** - Profile validation in production systems
9. **Cache validation results** - For identical payloads in appropriate scenarios
10. **Document validation rules** - Make business rules explicit in validation tags

## Conclusion

The go-playground/validator library provides comprehensive validation capabilities with excellent performance characteristics. By following the patterns and examples in this guide, you can implement robust validation for any Go application, from simple APIs to complex enterprise systems.

Key takeaways:
- **Sub-30ns validation performance** for successful validations
- **300+ built-in validation tags** covering most use cases
- **Thread-safe concurrent validation** with singleton pattern
- **Extensive customization options** for business-specific rules
- **Production-ready integration patterns** for all major Go frameworks

For data infrastructure and pipeline validation needs, consider complementing go-playground/validator with Great Expectations for comprehensive data quality management across your entire technology stack.