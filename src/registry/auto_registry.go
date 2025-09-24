// Package registry provides automatic model discovery and registration capabilities.
package registry

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github-data-validator/models"
	"github-data-validator/validations"
)

// ModelConfig represents configuration for automatic model registration
type ModelConfig struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ModelStruct string   `json:"model_struct"`
	Validator   string   `json:"validator"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	Enabled     bool     `json:"enabled"`
}

// AutoRegistrationConfig contains configuration for automatic model discovery
type AutoRegistrationConfig struct {
	ModelsPath      string        `json:"models_path"`
	ValidationsPath string        `json:"validations_path"`
	ConfigPath      string        `json:"config_path"`
	CustomModels    []ModelConfig `json:"custom_models"`
	AutoDiscover    bool          `json:"auto_discover"`
}

// ModelDiscovery handles automatic model discovery and registration
type ModelDiscovery struct {
	config   *AutoRegistrationConfig
	registry *ModelRegistry
}

// NewModelDiscovery creates a new model discovery instance
func NewModelDiscovery(config *AutoRegistrationConfig, registry *ModelRegistry) *ModelDiscovery {
	return &ModelDiscovery{
		config:   config,
		registry: registry,
	}
}

// DiscoverAndRegisterModels automatically discovers and registers models
func (md *ModelDiscovery) DiscoverAndRegisterModels() error {
	log.Println("🔍 Starting automatic model discovery...")

	// Register models from configuration files
	if err := md.registerFromConfig(); err != nil {
		log.Printf("⚠️  Error registering from config: %v", err)
	}

	// Auto-discover models if enabled
	if md.config.AutoDiscover {
		if err := md.autoDiscoverModels(); err != nil {
			log.Printf("⚠️  Error auto-discovering models: %v", err)
		}
	}

	return nil
}

// registerFromConfig registers models from configuration files
func (md *ModelDiscovery) registerFromConfig() error {
	if md.config.ConfigPath == "" {
		return nil
	}

	// Load configuration from file
	configData, err := os.ReadFile(md.config.ConfigPath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	var config AutoRegistrationConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	// Register custom models from configuration
	for _, modelConfig := range config.CustomModels {
		if !modelConfig.Enabled {
			log.Printf("⏭️  Skipping disabled model: %s", modelConfig.Type)
			continue
		}

		if err := md.registerModelFromConfig(modelConfig); err != nil {
			log.Printf("❌ Failed to register model %s: %v", modelConfig.Type, err)
			continue
		}

		log.Printf("✅ Registered model from config: %s -> %s", modelConfig.Type, modelConfig.Name)
	}

	return nil
}

// registerModelFromConfig registers a single model from configuration
func (md *ModelDiscovery) registerModelFromConfig(config ModelConfig) error {
	// This is a simplified version - you would need to implement
	// reflection-based model and validator instantiation
	modelInfo := &ModelInfo{
		Type:        ModelType(config.Type),
		Name:        config.Name,
		Description: config.Description,
		Version:     config.Version,
		Author:      config.Author,
		Tags:        config.Tags,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	// Try to get model struct and validator using reflection
	if err := md.resolveModelComponents(config, modelInfo); err != nil {
		return fmt.Errorf("resolving model components: %w", err)
	}

	return md.registry.RegisterModel(modelInfo)
}

// autoDiscoverModels automatically discovers models by scanning source files
func (md *ModelDiscovery) autoDiscoverModels() error {
	log.Println("🔄 Auto-discovering models from source code...")

	// Discover model structs
	modelStructs, err := md.discoverModelStructs()
	if err != nil {
		return fmt.Errorf("discovering model structs: %w", err)
	}

	// Discover validators
	validators, err := md.discoverValidators()
	if err != nil {
		return fmt.Errorf("discovering validators: %w", err)
	}

	// Match models with validators and register
	return md.matchAndRegisterModels(modelStructs, validators)
}

// ModelStruct represents a discovered model structure
type ModelStruct struct {
	Name        string
	Package     string
	File        string
	ReflectType reflect.Type
}

// ValidatorInfo represents a discovered validator
type ValidatorInfo struct {
	Name        string
	Package     string
	File        string
	Constructor string // e.g., "NewDeploymentValidator"
}

// discoverModelStructs scans the models directory for struct definitions
func (md *ModelDiscovery) discoverModelStructs() ([]ModelStruct, error) {
	var modelStructs []ModelStruct

	modelsPath := md.config.ModelsPath
	if modelsPath == "" {
		modelsPath = "src/models"
	}

	err := filepath.WalkDir(modelsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		structs, err := md.parseGoFileForStructs(path)
		if err != nil {
			log.Printf("⚠️  Error parsing %s: %v", path, err)
			return nil
		}

		modelStructs = append(modelStructs, structs...)
		return nil
	})

	log.Printf("📊 Discovered %d model structs", len(modelStructs))
	return modelStructs, err
}

// discoverValidators scans the validations directory for validator definitions
func (md *ModelDiscovery) discoverValidators() ([]ValidatorInfo, error) {
	var validators []ValidatorInfo

	validationsPath := md.config.ValidationsPath
	if validationsPath == "" {
		validationsPath = "src/validations"
	}

	err := filepath.WalkDir(validationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		validatorInfos, err := md.parseGoFileForValidators(path)
		if err != nil {
			log.Printf("⚠️  Error parsing validators in %s: %v", path, err)
			return nil
		}

		validators = append(validators, validatorInfos...)
		return nil
	})

	log.Printf("📊 Discovered %d validators", len(validators))
	return validators, err
}

// parseGoFileForStructs parses a Go file and extracts struct definitions
func (md *ModelDiscovery) parseGoFileForStructs(filename string) ([]ModelStruct, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var structs []ModelStruct

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if _, ok := x.Type.(*ast.StructType); ok {
				// Found a struct - try to get its reflect.Type
				structName := x.Name.Name
				if strings.HasSuffix(structName, "Payload") || strings.HasSuffix(structName, "Model") {
					// Try to get the reflect type from the models package
					if modelType := md.getReflectType("models", structName); modelType != nil {
						structs = append(structs, ModelStruct{
							Name:        structName,
							Package:     node.Name.Name,
							File:        filename,
							ReflectType: modelType,
						})
					}
				}
			}
		}
		return true
	})

	return structs, nil
}

// parseGoFileForValidators parses a Go file and extracts validator definitions
func (md *ModelDiscovery) parseGoFileForValidators(filename string) ([]ValidatorInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var validators []ValidatorInfo

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// Look for New*Validator functions
			if x.Name != nil && strings.HasPrefix(x.Name.Name, "New") && strings.HasSuffix(x.Name.Name, "Validator") {
				validatorName := strings.TrimSuffix(strings.TrimPrefix(x.Name.Name, "New"), "Validator")
				validators = append(validators, ValidatorInfo{
					Name:        validatorName,
					Package:     node.Name.Name,
					File:        filename,
					Constructor: x.Name.Name,
				})
			}
		}
		return true
	})

	return validators, nil
}

// getReflectType attempts to get the reflect.Type for a given package and type name
func (md *ModelDiscovery) getReflectType(packageName, typeName string) reflect.Type {
	switch packageName {
	case "models":
		switch typeName {
		case "GitHubPayload":
			return reflect.TypeOf(models.GitHubPayload{})
		case "GitLabPayload":
			return reflect.TypeOf(models.GitLabPayload{})
		case "BitbucketPayload":
			return reflect.TypeOf(models.BitbucketPayload{})
		case "SlackMessagePayload":
			return reflect.TypeOf(models.SlackMessagePayload{})
		case "APIRequest":
			return reflect.TypeOf(models.APIRequest{})
		case "DatabaseQuery":
			return reflect.TypeOf(models.DatabaseQuery{})
		case "GenericPayload":
			return reflect.TypeOf(models.GenericPayload{})
		case "DeploymentPayload":
			return reflect.TypeOf(models.DeploymentPayload{})
		}
	}
	return nil
}

// matchAndRegisterModels matches discovered models with validators and registers them
func (md *ModelDiscovery) matchAndRegisterModels(modelStructs []ModelStruct, validators []ValidatorInfo) error {
	registered := 0

	for _, modelStruct := range modelStructs {
		// Find matching validator
		var matchingValidator *ValidatorInfo
		modelBaseName := md.extractBaseName(modelStruct.Name)

		for _, validator := range validators {
			if strings.EqualFold(validator.Name, modelBaseName) {
				matchingValidator = &validator
				break
			}
		}

		if matchingValidator == nil {
			log.Printf("⚠️  No matching validator found for model: %s", modelStruct.Name)
			continue
		}

		// Register the model
		if err := md.registerDiscoveredModel(modelStruct, *matchingValidator); err != nil {
			log.Printf("❌ Failed to register discovered model %s: %v", modelStruct.Name, err)
			continue
		}

		log.Printf("✅ Auto-registered model: %s -> %s", modelBaseName, modelStruct.Name)
		registered++
	}

	log.Printf("🎉 Successfully auto-registered %d models", registered)
	return nil
}

// extractBaseName extracts base name from model struct name (e.g., "DeploymentPayload" -> "Deployment")
func (md *ModelDiscovery) extractBaseName(structName string) string {
	name := structName
	if strings.HasSuffix(name, "Payload") {
		name = strings.TrimSuffix(name, "Payload")
	}
	if strings.HasSuffix(name, "Model") {
		name = strings.TrimSuffix(name, "Model")
	}
	return name
}

// registerDiscoveredModel registers a discovered model with its validator
func (md *ModelDiscovery) registerDiscoveredModel(modelStruct ModelStruct, validator ValidatorInfo) error {
	if modelStruct.ReflectType == nil {
		return fmt.Errorf("no reflect type available for %s", modelStruct.Name)
	}

	// Create validator instance using reflection or predefined mapping
	validatorInstance := md.createValidatorInstance(validator.Name)
	if validatorInstance == nil {
		return fmt.Errorf("could not create validator instance for %s", validator.Name)
	}

	modelType := ModelType(strings.ToLower(validator.Name))
	modelInfo := &ModelInfo{
		Type:        modelType,
		Name:        md.generateModelName(validator.Name),
		Description: md.generateModelDescription(validator.Name, modelStruct.Name),
		ModelStruct: modelStruct.ReflectType,
		Validator:   validatorInstance,
		Version:     "1.0.0",
		CreatedAt:   time.Now().Format(time.RFC3339),
		Author:      "Auto-Discovery",
		Tags:        md.generateModelTags(validator.Name),
	}

	return md.registry.RegisterModel(modelInfo)
}

// createValidatorInstance creates a validator instance by name
func (md *ModelDiscovery) createValidatorInstance(validatorName string) ValidatorInterface {
	switch strings.ToLower(validatorName) {
	case "github":
		return &GitHubValidatorWrapper{validator: validations.NewGitHubValidator()}
	case "gitlab":
		return &GitLabValidatorWrapper{validator: validations.NewGitLabValidator()}
	case "bitbucket":
		return &BitbucketValidatorWrapper{validator: validations.NewBitbucketValidator()}
	case "slack":
		return &SlackValidatorWrapper{validator: validations.NewSlackValidator()}
	case "api":
		return &APIValidatorWrapper{validator: validations.NewAPIValidator()}
	case "database":
		return &DatabaseValidatorWrapper{validator: validations.NewDatabaseValidator()}
	case "generic":
		return &GenericValidatorWrapper{validator: validations.NewGenericValidator()}
	case "deployment":
		return &DeploymentValidatorWrapper{validator: validations.NewDeploymentValidator()}
	default:
		return nil
	}
}

// generateModelName generates a human-readable model name
func (md *ModelDiscovery) generateModelName(baseName string) string {
	switch strings.ToLower(baseName) {
	case "github":
		return "GitHub Webhook"
	case "gitlab":
		return "GitLab Webhook"
	case "bitbucket":
		return "Bitbucket Webhook"
	case "slack":
		return "Slack Message"
	case "api":
		return "API Request/Response"
	case "database":
		return "Database Operations"
	case "generic":
		return "Generic Payload"
	case "deployment":
		return "Deployment Webhook"
	default:
		return baseName + " Validation"
	}
}

// generateModelDescription generates a model description
func (md *ModelDiscovery) generateModelDescription(baseName, structName string) string {
	return fmt.Sprintf("%s validation with comprehensive business rules (auto-discovered from %s)",
		md.generateModelName(baseName), structName)
}

// generateModelTags generates appropriate tags for a model
func (md *ModelDiscovery) generateModelTags(baseName string) []string {
	switch strings.ToLower(baseName) {
	case "github":
		return []string{"webhook", "github", "git", "collaboration"}
	case "gitlab":
		return []string{"webhook", "gitlab", "git", "collaboration"}
	case "bitbucket":
		return []string{"webhook", "bitbucket", "git", "collaboration"}
	case "slack":
		return []string{"messaging", "slack", "communication"}
	case "api":
		return []string{"api", "http", "rest", "web"}
	case "database":
		return []string{"database", "sql", "transaction", "query"}
	case "generic":
		return []string{"generic", "flexible", "json", "general"}
	case "deployment":
		return []string{"deployment", "webhook", "devops", "ci/cd"}
	default:
		return []string{"auto-discovered", strings.ToLower(baseName)}
	}
}

// resolveModelComponents resolves model struct and validator from configuration
func (md *ModelDiscovery) resolveModelComponents(config ModelConfig, modelInfo *ModelInfo) error {
	// Get model struct type
	modelType := md.getReflectType("models", config.ModelStruct)
	if modelType == nil {
		return fmt.Errorf("could not resolve model struct: %s", config.ModelStruct)
	}
	modelInfo.ModelStruct = modelType

	// Create validator instance
	validatorInstance := md.createValidatorInstance(strings.TrimSuffix(config.Validator, "Validator"))
	if validatorInstance == nil {
		return fmt.Errorf("could not create validator instance: %s", config.Validator)
	}
	modelInfo.Validator = validatorInstance

	return nil
}

// LoadAutoRegistrationConfig loads configuration from file
func LoadAutoRegistrationConfig(configPath string) (*AutoRegistrationConfig, error) {
	if configPath == "" {
		// Return default configuration
		return &AutoRegistrationConfig{
			ModelsPath:      "src/models",
			ValidationsPath: "src/validations",
			AutoDiscover:    true,
			CustomModels:    []ModelConfig{},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config AutoRegistrationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &config, nil
}

// EnableAutoRegistration enables automatic model registration on the global registry
func EnableAutoRegistration(configPath string) error {
	config, err := LoadAutoRegistrationConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading auto-registration config: %w", err)
	}

	registry := GetGlobalRegistry()
	discovery := NewModelDiscovery(config, registry)

	return discovery.DiscoverAndRegisterModels()
}
