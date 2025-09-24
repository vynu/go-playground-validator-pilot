// Package main provides a code generator for creating new validation models using templates.
// This tool helps developers quickly scaffold new model types with comprehensive validation.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ModelConfig represents the configuration for generating a new model.
type ModelConfig struct {
	// Basic model information
	ServiceName    string `json:"service_name"`     // e.g., "Stripe", "Shopify", "Discord"
	DataType       string `json:"data_type"`        // e.g., "webhook", "API response", "message"
	EventType      string `json:"event_type"`       // e.g., "payment", "order", "notification"
	MainStructName string `json:"main_struct_name"` // e.g., "StripeWebhookPayload"
	ModelTypeName  string `json:"model_type_name"`  // e.g., "stripe", "shopify"
	ValidatorVar   string `json:"validator_var"`    // e.g., "sv", "shv"
	ProviderName   string `json:"provider_name"`    // e.g., "stripe_validator"
	EndpointPath   string `json:"endpoint_path"`    // e.g., "stripe", "shopify"
	VarName        string `json:"var_name"`         // e.g., "stripe", "shopify"

	// Validation configuration
	ValidTypes string `json:"valid_types"` // e.g., "payment invoice subscription"
	FieldCount int    `json:"field_count"` // Approximate number of fields
	RuleCount  int    `json:"rule_count"`  // Approximate number of validation rules

	// Metadata
	Version     string   `json:"version"`     // e.g., "1.0.0"
	Author      string   `json:"author"`      // e.g., "Your Name"
	Description string   `json:"description"` // Model description
	Tags        []string `json:"tags"`        // e.g., ["payment", "webhook", "api"]

	// Examples
	ExampleID     string `json:"example_id"`     // e.g., "evt_1234567890"
	ExampleType   string `json:"example_type"`   // e.g., "payment.succeeded"
	ExampleSource string `json:"example_source"` // e.g., "https://api.stripe.com"
	ExampleTag    string `json:"example_tag"`    // e.g., "production"

	// Field definitions
	CustomFields        []FieldConfig         `json:"custom_fields"`
	AdditionalStructs   []StructConfig        `json:"additional_structs"`
	CustomValidators    []ValidatorConfig     `json:"custom_validators"`
	BusinessLogicChecks []BusinessLogicConfig `json:"business_logic_checks"`

	// Documentation
	ValidationRulesDoc   []ValidationRuleDoc    `json:"validation_rules_doc"`
	BusinessLogicDoc     []BusinessLogicDoc     `json:"business_logic_doc"`
	CustomValidatorsDoc  []CustomValidatorDoc   `json:"custom_validators_doc"`
	ExampleBusinessLogic []ExampleBusinessLogic `json:"example_business_logic"`

	// Example data
	ExampleFields      []ExampleField `json:"example_fields"`
	ExamplePayload     []ExampleField `json:"example_payload"`
	ValidTestPayload   []ExampleField `json:"valid_test_payload"`
	InvalidTestPayload []ExampleField `json:"invalid_test_payload"`
	ExampleAPIPayload  []ExampleField `json:"example_api_payload"`
	ExampleCurlPayload []ExampleField `json:"example_curl_payload"`
	Examples           []string       `json:"examples"`
}

// FieldConfig represents a custom field definition.
type FieldConfig struct {
	Name            string `json:"name"`             // Go field name
	Type            string `json:"type"`             // Go type
	JSONTag         string `json:"json_tag"`         // JSON tag
	ValidationRules string `json:"validation_rules"` // Validation rules
}

// StructConfig represents an additional struct definition.
type StructConfig struct {
	Name        string        `json:"name"`        // Struct name
	Description string        `json:"description"` // Description
	Fields      []FieldConfig `json:"fields"`      // Fields
}

// ValidatorConfig represents a custom validator definition.
type ValidatorConfig struct {
	Tag             string `json:"tag"`              // Validator tag
	FunctionName    string `json:"function_name"`    // Function name
	Description     string `json:"description"`      // Description
	ValidationLogic string `json:"validation_logic"` // Implementation logic
	ReturnCondition string `json:"return_condition"` // Return condition
	ErrorMessage    string `json:"error_message"`    // Error message format
	Example         string `json:"example"`          // Usage example
	Pattern         string `json:"pattern"`          // Regex pattern if applicable
}

// BusinessLogicConfig represents a business logic check.
type BusinessLogicConfig struct {
	FunctionName   string `json:"function_name"`  // Function name
	Description    string `json:"description"`    // Description
	Implementation string `json:"implementation"` // Implementation code
}

// Documentation structures
type ValidationRuleDoc struct {
	Field          string `json:"field"`
	Description    string `json:"description"`
	ValidationRule string `json:"validation_rule"`
	Required       bool   `json:"required"`
	Example        string `json:"example"`
}

type BusinessLogicDoc struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	Severity       string `json:"severity"`
	Triggers       string `json:"triggers"`
	Recommendation string `json:"recommendation"`
}

type CustomValidatorDoc struct {
	Tag         string `json:"tag"`
	Description string `json:"description"`
	Parameter   string `json:"parameter"`
	Example     string `json:"example"`
}

type ExampleBusinessLogic struct {
	Name                  string `json:"name"`
	Description           string `json:"description"`
	FunctionName          string `json:"function_name"`
	ExampleImplementation string `json:"example_implementation"`
}

type ExampleField struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// DefaultModelConfig returns a default configuration that can be customized.
func DefaultModelConfig() ModelConfig {
	return ModelConfig{
		ServiceName:    "Example",
		DataType:       "webhook",
		EventType:      "notification",
		MainStructName: "ExampleWebhookPayload",
		ModelTypeName:  "example",
		ValidatorVar:   "ev",
		ProviderName:   "example_validator",
		EndpointPath:   "example",
		VarName:        "example",
		ValidTypes:     "notification event alert",
		FieldCount:     20,
		RuleCount:      15,
		Version:        "1.0.0",
		Author:         "Generated",
		Description:    "Example webhook payload validation with comprehensive business rules",
		Tags:           []string{"webhook", "example", "api"},
		ExampleID:      "evt_example_123",
		ExampleType:    "notification.sent",
		ExampleSource:  "https://api.example.com",
		ExampleTag:     "production",
		CustomFields: []FieldConfig{
			{
				Name:            "EventType",
				Type:            "string",
				JSONTag:         "event_type",
				ValidationRules: "required,oneof=notification event alert",
			},
			{
				Name:            "UserID",
				Type:            "string",
				JSONTag:         "user_id",
				ValidationRules: "required,min=1,max=255",
			},
		},
		CustomValidators: []ValidatorConfig{
			{
				Tag:             "example_id",
				FunctionName:    "validateExampleID",
				Description:     "Example ID format validation",
				ValidationLogic: `if len(value) == 0 || !strings.HasPrefix(value, "evt_") { return false }`,
				ReturnCondition: "len(value) >= 4 && strings.HasPrefix(value, \"evt_\")",
				ErrorMessage:    "must be a valid Example ID format (evt_xxxxx)",
				Example:         "evt_example_123",
				Pattern:         "^evt_[a-zA-Z0-9_]+$",
			},
		},
		BusinessLogicChecks: []BusinessLogicConfig{
			{
				FunctionName:   "checkExamplePatterns",
				Description:    "checks for Example-specific patterns and best practices",
				Implementation: `// Check for test events in production\nif strings.Contains(strings.ToLower(payload.EventType), "test") {\n\twarnings = append(warnings, models.ValidationWarning{\n\t\tField: "EventType",\n\t\tMessage: "Test event detected",\n\t\tCode: "TEST_EVENT",\n\t\tSuggestion: "Ensure test events are not processed in production",\n\t\tCategory: "workflow",\n\t})\n}`,
			},
		},
		ExamplePayload: []ExampleField{
			{Field: "ID", Value: `"evt_example_123"`},
			{Field: "Type", Value: `"notification"`},
			{Field: "EventType", Value: `"notification.sent"`},
			{Field: "UserID", Value: `"user_12345"`},
		},
	}
}

// Generator represents the model generator.
type Generator struct {
	templateDir string
	outputDir   string
	config      ModelConfig
}

// NewGenerator creates a new model generator.
func NewGenerator(templateDir, outputDir string) *Generator {
	return &Generator{
		templateDir: templateDir,
		outputDir:   outputDir,
	}
}

// LoadConfig loads configuration from a JSON file.
func (g *Generator) LoadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&g.config); err != nil {
		return fmt.Errorf("failed to decode config: %v", err)
	}

	return nil
}

// SaveConfig saves the current configuration to a JSON file.
func (g *Generator) SaveConfig(configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(g.config); err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	return nil
}

// InteractiveConfig guides the user through creating a configuration interactively.
func (g *Generator) InteractiveConfig() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("🚀 Model Generator - Interactive Configuration")
	fmt.Println("Press Enter to use default values shown in [brackets]")
	fmt.Println()

	// Basic information
	g.config.ServiceName = g.promptString(scanner, "Service name (e.g., Stripe, Shopify, Discord)", "Example")
	g.config.DataType = g.promptString(scanner, "Data type (e.g., webhook, API response, message)", "webhook")
	g.config.EventType = g.promptString(scanner, "Event type (e.g., payment, order, notification)", "notification")

	// Generate derived names
	g.config.MainStructName = g.config.ServiceName + strings.Title(g.config.DataType) + "Payload"
	g.config.ModelTypeName = strings.ToLower(g.config.ServiceName)
	g.config.ValidatorVar = strings.ToLower(g.config.ServiceName[:1]) + "v"
	g.config.ProviderName = strings.ToLower(g.config.ServiceName) + "_validator"
	g.config.EndpointPath = strings.ToLower(g.config.ServiceName)
	g.config.VarName = strings.ToLower(g.config.ServiceName)

	fmt.Printf("Generated struct name: %s\n", g.config.MainStructName)
	fmt.Printf("Generated model type: %s\n", g.config.ModelTypeName)

	g.config.Description = g.promptString(scanner, "Description",
		g.config.ServiceName+" "+g.config.DataType+" validation with comprehensive business rules")
	g.config.Author = g.promptString(scanner, "Author", "Generated")
	g.config.Version = g.promptString(scanner, "Version", "1.0.0")

	// Ask if they want to add custom fields
	if g.promptBool(scanner, "Add custom fields?", false) {
		g.config.CustomFields = g.promptFields(scanner)
	}

	// Ask if they want to add custom validators
	if g.promptBool(scanner, "Add custom validators?", false) {
		g.config.CustomValidators = g.promptValidators(scanner)
	}

	return nil
}

// promptString prompts for a string value with a default.
func (g *Generator) promptString(scanner *bufio.Scanner, prompt, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	if scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		if value == "" {
			return defaultValue
		}
		return value
	}
	return defaultValue
}

// promptBool prompts for a boolean value.
func (g *Generator) promptBool(scanner *bufio.Scanner, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)
	if scanner.Scan() {
		value := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if value == "" {
			return defaultValue
		}
		return value == "y" || value == "yes"
	}
	return defaultValue
}

// promptFields prompts for custom field definitions.
func (g *Generator) promptFields(scanner *bufio.Scanner) []FieldConfig {
	var fields []FieldConfig

	fmt.Println("\nAdding custom fields (enter empty name to finish):")
	for {
		name := g.promptString(scanner, "Field name", "")
		if name == "" {
			break
		}

		field := FieldConfig{
			Name:            name,
			Type:            g.promptString(scanner, "Field type", "string"),
			JSONTag:         g.promptString(scanner, "JSON tag", strings.ToLower(name)),
			ValidationRules: g.promptString(scanner, "Validation rules", "required,min=1"),
		}

		fields = append(fields, field)
		fmt.Printf("Added field: %s\n", name)
	}

	return fields
}

// promptValidators prompts for custom validator definitions.
func (g *Generator) promptValidators(scanner *bufio.Scanner) []ValidatorConfig {
	var validators []ValidatorConfig

	fmt.Println("\nAdding custom validators (enter empty tag to finish):")
	for {
		tag := g.promptString(scanner, "Validator tag", "")
		if tag == "" {
			break
		}

		validator := ValidatorConfig{
			Tag:          tag,
			FunctionName: "validate" + strings.Title(tag),
			Description:  g.promptString(scanner, "Description", tag+" validation"),
			ErrorMessage: g.promptString(scanner, "Error message", "must be a valid "+tag),
		}

		validators = append(validators, validator)
		fmt.Printf("Added validator: %s\n", tag)
	}

	return validators
}

// Generate generates all files based on the configuration.
func (g *Generator) Generate() error {
	// Create output directories
	if err := g.createDirectories(); err != nil {
		return err
	}

	// Generate files from templates
	templates := map[string]string{
		"model_template.go.tmpl":        filepath.Join("models", g.config.VarName+".go"),
		"validation_template.go.tmpl":   filepath.Join("validations", g.config.VarName+".go"),
		"registration_template.go.tmpl": "register_" + g.config.VarName + ".go",
		"README_template.md":            g.config.ServiceName + "_Integration_Guide.md",
	}

	for templateFile, outputFile := range templates {
		if err := g.generateFromTemplate(templateFile, outputFile); err != nil {
			return fmt.Errorf("failed to generate %s: %v", outputFile, err)
		}
		fmt.Printf("✅ Generated: %s\n", outputFile)
	}

	// Generate example configuration
	if err := g.generateExampleConfig(); err != nil {
		return fmt.Errorf("failed to generate example config: %v", err)
	}

	fmt.Println("\n🎉 Model generation completed successfully!")
	fmt.Printf("📁 Files generated in: %s\n", g.outputDir)

	return nil
}

// createDirectories creates the necessary output directories.
func (g *Generator) createDirectories() error {
	dirs := []string{
		g.outputDir,
		filepath.Join(g.outputDir, "models"),
		filepath.Join(g.outputDir, "validations"),
		filepath.Join(g.outputDir, "examples"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	return nil
}

// generateFromTemplate generates a file from a template.
func (g *Generator) generateFromTemplate(templateFile, outputFile string) error {
	// Read template
	templatePath := filepath.Join(g.templateDir, templateFile)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %v", templatePath, err)
	}

	// Create output file
	outputPath := filepath.Join(g.outputDir, outputFile)
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, g.config); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return nil
}

// generateExampleConfig generates example configuration files.
func (g *Generator) generateExampleConfig() error {
	// Server configuration example
	serverConfig := map[string]interface{}{
		"custom_models": []map[string]interface{}{
			{
				"type":        g.config.ModelTypeName,
				"name":        g.config.ServiceName + " " + g.config.DataType,
				"description": g.config.Description,
				"version":     g.config.Version,
				"author":      g.config.Author,
				"tags":        g.config.Tags,
			},
		},
	}

	configPath := filepath.Join(g.outputDir, "examples", "server_config.json")
	if err := g.writeJSON(configPath, serverConfig); err != nil {
		return err
	}

	// Example payloads
	validPayload := map[string]interface{}{
		"id":        g.config.ExampleID,
		"type":      g.config.ExampleType,
		"timestamp": "2023-01-01T00:00:00Z",
		"source":    g.config.ExampleSource,
	}

	for _, field := range g.config.ExamplePayload {
		// Parse the value (remove quotes for proper JSON)
		value := field.Value
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = value[1 : len(value)-1]
		}
		validPayload[field.Field] = value
	}

	validPath := filepath.Join(g.outputDir, "examples", g.config.VarName+"_valid.json")
	if err := g.writeJSON(validPath, validPayload); err != nil {
		return err
	}

	fmt.Printf("✅ Generated example config: examples/server_config.json\n")
	fmt.Printf("✅ Generated example payload: examples/%s_valid.json\n", g.config.VarName)

	return nil
}

// writeJSON writes data as formatted JSON to a file.
func (g *Generator) writeJSON(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// CLI interface
func main() {
	var (
		configPath    = flag.String("config", "", "Path to configuration JSON file")
		outputDir     = flag.String("output", "./generated", "Output directory for generated files")
		templateDir   = flag.String("templates", "./templates", "Directory containing templates")
		interactive   = flag.Bool("interactive", false, "Run in interactive mode")
		saveConfig    = flag.String("save-config", "", "Save configuration to file")
		defaultConfig = flag.Bool("default", false, "Generate default configuration file")
	)
	flag.Parse()

	generator := NewGenerator(*templateDir, *outputDir)

	// Generate default config
	if *defaultConfig {
		config := DefaultModelConfig()
		generator.config = config

		configFile := "model_config.json"
		if *saveConfig != "" {
			configFile = *saveConfig
		}

		if err := generator.SaveConfig(configFile); err != nil {
			log.Fatalf("Failed to save default config: %v", err)
		}

		fmt.Printf("✅ Default configuration saved to: %s\n", configFile)
		fmt.Println("📝 Edit the configuration file and run the generator again")
		return
	}

	// Load or create configuration
	if *interactive {
		generator.config = DefaultModelConfig()
		if err := generator.InteractiveConfig(); err != nil {
			log.Fatalf("Interactive configuration failed: %v", err)
		}

		if *saveConfig != "" {
			if err := generator.SaveConfig(*saveConfig); err != nil {
				log.Fatalf("Failed to save config: %v", err)
			}
			fmt.Printf("💾 Configuration saved to: %s\n", *saveConfig)
		}
	} else if *configPath != "" {
		if err := generator.LoadConfig(*configPath); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	} else {
		generator.config = DefaultModelConfig()
		fmt.Println("⚠️  Using default configuration. Use -config to specify a custom config file.")
	}

	// Generate files
	if err := generator.Generate(); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
