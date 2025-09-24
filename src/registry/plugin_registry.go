// Package registry provides plugin-style model registration capabilities.
package registry

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github-data-validator/models"
	"github-data-validator/validations"
)

// PluginConfig represents configuration for a plugin-based model
type PluginConfig struct {
	// Basic plugin information
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`

	// Model configuration
	ModelType string `json:"model_type"`
	ModelName string `json:"model_name"`
	Enabled   bool   `json:"enabled"`

	// File paths
	ModelFile     string `json:"model_file"`     // Path to Go model file
	ValidatorFile string `json:"validator_file"` // Path to Go validator file
	ConfigFile    string `json:"config_file"`    // Path to additional config

	// Registration metadata
	RegisteredAt string `json:"registered_at"`
	LastUpdated  string `json:"last_updated"`

	// Advanced configuration
	Priority      int               `json:"priority"`       // Loading priority (higher = first)
	Dependencies  []string          `json:"dependencies"`   // Other plugins this depends on
	ConflictsWith []string          `json:"conflicts_with"` // Plugins that conflict with this one
	Metadata      map[string]string `json:"metadata"`       // Additional metadata
}

// PluginManager manages plugin-based model registration
type PluginManager struct {
	registry    *ModelRegistry
	pluginDir   string
	plugins     map[string]*PluginConfig
	loadedOrder []string
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(registry *ModelRegistry, pluginDir string) *PluginManager {
	return &PluginManager{
		registry:  registry,
		pluginDir: pluginDir,
		plugins:   make(map[string]*PluginConfig),
	}
}

// LoadPlugins discovers and loads all plugins from the plugin directory
func (pm *PluginManager) LoadPlugins() error {
	log.Printf("🔌 Loading plugins from directory: %s", pm.pluginDir)

	// Ensure plugin directory exists
	if err := os.MkdirAll(pm.pluginDir, 0755); err != nil {
		return fmt.Errorf("creating plugin directory: %w", err)
	}

	// Discover plugin configuration files
	pluginFiles, err := filepath.Glob(filepath.Join(pm.pluginDir, "*.json"))
	if err != nil {
		return fmt.Errorf("discovering plugin files: %w", err)
	}

	if len(pluginFiles) == 0 {
		log.Printf("📋 No plugin configuration files found in %s", pm.pluginDir)
		return pm.createExamplePlugins()
	}

	// Load all plugin configurations
	for _, pluginFile := range pluginFiles {
		if err := pm.loadPluginConfig(pluginFile); err != nil {
			log.Printf("❌ Failed to load plugin from %s: %v", pluginFile, err)
			continue
		}
	}

	// Sort plugins by priority and dependencies
	if err := pm.resolveDependencies(); err != nil {
		log.Printf("⚠️  Dependency resolution issues: %v", err)
	}

	// Register plugins in order
	return pm.registerAllPlugins()
}

// loadPluginConfig loads a plugin configuration from file
func (pm *PluginManager) loadPluginConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading plugin config: %w", err)
	}

	var config PluginConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parsing plugin config: %w", err)
	}

	// Validate plugin configuration
	if err := pm.validatePluginConfig(&config); err != nil {
		return fmt.Errorf("validating plugin config: %w", err)
	}

	// Store plugin configuration
	pm.plugins[config.ModelType] = &config
	log.Printf("📄 Loaded plugin config: %s (v%s)", config.Name, config.Version)

	return nil
}

// validatePluginConfig validates a plugin configuration
func (pm *PluginManager) validatePluginConfig(config *PluginConfig) error {
	if config.Name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if config.ModelType == "" {
		return fmt.Errorf("model_type cannot be empty")
	}

	if config.Version == "" {
		config.Version = "1.0.0"
	}

	if config.Author == "" {
		config.Author = "Unknown"
	}

	// Set metadata defaults
	if config.Metadata == nil {
		config.Metadata = make(map[string]string)
	}

	// Update timestamps
	now := time.Now().Format(time.RFC3339)
	if config.RegisteredAt == "" {
		config.RegisteredAt = now
	}
	config.LastUpdated = now

	return nil
}

// resolveDependencies sorts plugins by priority and resolves dependencies
func (pm *PluginManager) resolveDependencies() error {
	// Simple priority-based sorting for now
	// TODO: Implement full dependency resolution

	type pluginPriority struct {
		modelType string
		priority  int
	}

	var priorities []pluginPriority
	for modelType, config := range pm.plugins {
		priorities = append(priorities, pluginPriority{
			modelType: modelType,
			priority:  config.Priority,
		})
	}

	// Sort by priority (higher first)
	for i := 0; i < len(priorities); i++ {
		for j := i + 1; j < len(priorities); j++ {
			if priorities[i].priority < priorities[j].priority {
				priorities[i], priorities[j] = priorities[j], priorities[i]
			}
		}
	}

	// Store loading order
	pm.loadedOrder = make([]string, len(priorities))
	for i, p := range priorities {
		pm.loadedOrder[i] = p.modelType
	}

	log.Printf("📊 Plugin loading order: %v", pm.loadedOrder)
	return nil
}

// registerAllPlugins registers all loaded plugins
func (pm *PluginManager) registerAllPlugins() error {
	registered := 0
	skipped := 0

	for _, modelType := range pm.loadedOrder {
		config := pm.plugins[modelType]

		if !config.Enabled {
			log.Printf("⏭️  Skipping disabled plugin: %s", config.Name)
			skipped++
			continue
		}

		// Check for conflicts
		if err := pm.checkConflicts(config); err != nil {
			log.Printf("⚠️  Plugin conflict for %s: %v", config.Name, err)
			skipped++
			continue
		}

		// Register the plugin
		if err := pm.registerPlugin(config); err != nil {
			log.Printf("❌ Failed to register plugin %s: %v", config.Name, err)
			skipped++
			continue
		}

		log.Printf("✅ Registered plugin: %s -> %s", config.ModelType, config.Name)
		registered++
	}

	log.Printf("🎉 Plugin registration complete: %d registered, %d skipped", registered, skipped)
	return nil
}

// checkConflicts checks if a plugin conflicts with already loaded plugins
func (pm *PluginManager) checkConflicts(config *PluginConfig) error {
	// Check if model type already exists
	if pm.registry.IsRegistered(ModelType(config.ModelType)) {
		return fmt.Errorf("model type %s already registered", config.ModelType)
	}

	// Check explicit conflicts
	for _, conflictType := range config.ConflictsWith {
		if pm.registry.IsRegistered(ModelType(conflictType)) {
			return fmt.Errorf("conflicts with already loaded plugin: %s", conflictType)
		}
	}

	return nil
}

// registerPlugin registers a single plugin with the registry
func (pm *PluginManager) registerPlugin(config *PluginConfig) error {
	// Create model info from plugin configuration
	modelInfo := &ModelInfo{
		Type:        ModelType(config.ModelType),
		Name:        config.ModelName,
		Description: config.Description,
		Version:     config.Version,
		Author:      config.Author,
		Tags:        config.Tags,
		CreatedAt:   config.RegisteredAt,
	}

	// Try to resolve model struct and validator
	if err := pm.resolvePluginComponents(config, modelInfo); err != nil {
		return fmt.Errorf("resolving plugin components: %w", err)
	}

	// Register with the main registry
	return pm.registry.RegisterModel(modelInfo)
}

// resolvePluginComponents resolves model struct and validator for a plugin
func (pm *PluginManager) resolvePluginComponents(config *PluginConfig, modelInfo *ModelInfo) error {
	// For now, use the same mapping as the automatic registration system
	// In a real implementation, you might use reflection to load from files

	// For now, use the same mapping as the automatic registration system
	// In a real implementation, you might use reflection to load from files

	// Create validator instance using helper
	validatorInstance := pm.createValidatorInstance(config.ModelType)
	if validatorInstance == nil {
		return fmt.Errorf("could not create validator instance for %s", config.ModelType)
	}

	// Get model struct type
	modelType := pm.getReflectType("models", config.ModelType)
	if modelType == nil {
		return fmt.Errorf("could not resolve model struct: %s", config.ModelType)
	}

	modelInfo.ModelStruct = modelType
	modelInfo.Validator = validatorInstance

	return nil
}

// createValidatorInstance creates a validator instance by name
func (pm *PluginManager) createValidatorInstance(validatorName string) ValidatorInterface {
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

// getReflectType attempts to get the reflect.Type for a given package and type name
func (pm *PluginManager) getReflectType(packageName, typeName string) reflect.Type {
	switch packageName {
	case "models":
		switch typeName {
		case "github":
			return reflect.TypeOf(models.GitHubPayload{})
		case "gitlab":
			return reflect.TypeOf(models.GitLabPayload{})
		case "bitbucket":
			return reflect.TypeOf(models.BitbucketPayload{})
		case "slack":
			return reflect.TypeOf(models.SlackMessagePayload{})
		case "api":
			return reflect.TypeOf(models.APIRequest{})
		case "database":
			return reflect.TypeOf(models.DatabaseQuery{})
		case "generic":
			return reflect.TypeOf(models.GenericPayload{})
		case "deployment":
			return reflect.TypeOf(models.DeploymentPayload{})
		case "incident":
			return reflect.TypeOf(models.IncidentPayload{})
		}
	}
	return nil
}

// createExamplePlugins creates example plugin configurations
func (pm *PluginManager) createExamplePlugins() error {
	log.Println("📝 Creating example plugin configurations...")

	examples := []PluginConfig{
		{
			Name:        "GitHub Integration Plugin",
			Version:     "1.2.0",
			Author:      "Core Team",
			Description: "GitHub webhook validation with advanced repository rules",
			Tags:        []string{"webhook", "github", "git", "collaboration", "plugin"},
			ModelType:   "github",
			ModelName:   "GitHub Webhook",
			Enabled:     true,
			Priority:    100,
			Metadata: map[string]string{
				"category": "version-control",
				"support":  "official",
			},
		},
		{
			Name:         "Deployment Pipeline Plugin",
			Version:      "2.0.0",
			Author:       "DevOps Team",
			Description:  "Advanced deployment webhook validation with CI/CD integration",
			Tags:         []string{"deployment", "webhook", "devops", "ci/cd", "plugin"},
			ModelType:    "deployment",
			ModelName:    "Deployment Webhook",
			Enabled:      true,
			Priority:     90,
			Dependencies: []string{"github"},
			Metadata: map[string]string{
				"category": "deployment",
				"support":  "official",
			},
		},
		{
			Name:          "Slack Integration Plugin",
			Version:       "1.1.0",
			Author:        "Communication Team",
			Description:   "Slack message validation with rich formatting support",
			Tags:          []string{"messaging", "slack", "communication", "plugin"},
			ModelType:     "slack",
			ModelName:     "Slack Message",
			Enabled:       false, // Disabled by default for demonstration
			Priority:      50,
			ConflictsWith: []string{"discord", "teams"},
			Metadata: map[string]string{
				"category": "communication",
				"support":  "community",
			},
		},
	}

	for i, example := range examples {
		filename := filepath.Join(pm.pluginDir, fmt.Sprintf("plugin_%02d_%s.json", i+1, example.ModelType))

		data, err := json.MarshalIndent(example, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling example plugin %s: %w", example.Name, err)
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("writing example plugin %s: %w", filename, err)
		}

		log.Printf("📄 Created example plugin: %s", filename)
	}

	log.Println("✨ Example plugin configurations created. Reload to register them.")
	return nil
}

// ExportPluginConfig exports current registered models as plugin configurations
func (pm *PluginManager) ExportPluginConfig(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	models := pm.registry.GetAllModels()
	exported := 0

	for modelType, modelInfo := range models {
		config := PluginConfig{
			Name:        modelInfo.Name + " Plugin",
			Version:     modelInfo.Version,
			Author:      modelInfo.Author,
			Description: modelInfo.Description + " (exported from registry)",
			Tags:        append(modelInfo.Tags, "exported", "plugin"),
			ModelType:   string(modelType),
			ModelName:   modelInfo.Name,
			Enabled:     true,
			Priority:    50,
			Metadata: map[string]string{
				"category":    "exported",
				"export_date": time.Now().Format(time.RFC3339),
				"source":      "registry",
			},
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("exported_%s.json", string(modelType)))
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			log.Printf("❌ Failed to marshal plugin config for %s: %v", modelType, err)
			continue
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			log.Printf("❌ Failed to write plugin config for %s: %v", modelType, err)
			continue
		}

		exported++
	}

	log.Printf("📦 Exported %d plugin configurations to %s", exported, outputDir)
	return nil
}

// GetPluginInfo returns information about a specific plugin
func (pm *PluginManager) GetPluginInfo(modelType string) (*PluginConfig, error) {
	config, exists := pm.plugins[modelType]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", modelType)
	}
	return config, nil
}

// ListPlugins returns a list of all loaded plugins
func (pm *PluginManager) ListPlugins() []PluginConfig {
	var plugins []PluginConfig
	for _, config := range pm.plugins {
		plugins = append(plugins, *config)
	}
	return plugins
}

// GetPluginStatistics returns statistics about loaded plugins
func (pm *PluginManager) GetPluginStatistics() map[string]interface{} {
	enabled := 0
	disabled := 0
	byAuthor := make(map[string]int)
	byPriority := make(map[int]int)

	for _, config := range pm.plugins {
		if config.Enabled {
			enabled++
		} else {
			disabled++
		}
		byAuthor[config.Author]++
		byPriority[config.Priority]++
	}

	return map[string]interface{}{
		"total_plugins":       len(pm.plugins),
		"enabled_plugins":     enabled,
		"disabled_plugins":    disabled,
		"plugins_by_author":   byAuthor,
		"plugins_by_priority": byPriority,
		"loading_order":       pm.loadedOrder,
	}
}

// EnablePlugin enables a plugin by model type
func (pm *PluginManager) EnablePlugin(modelType string) error {
	config, exists := pm.plugins[modelType]
	if !exists {
		return fmt.Errorf("plugin %s not found", modelType)
	}

	config.Enabled = true
	config.LastUpdated = time.Now().Format(time.RFC3339)

	// Re-register the plugin
	return pm.registerPlugin(config)
}

// DisablePlugin disables a plugin by model type
func (pm *PluginManager) DisablePlugin(modelType string) error {
	config, exists := pm.plugins[modelType]
	if !exists {
		return fmt.Errorf("plugin %s not found", modelType)
	}

	config.Enabled = false
	config.LastUpdated = time.Now().Format(time.RFC3339)

	// Unregister from registry
	return pm.registry.UnregisterModel(ModelType(modelType))
}

// RegisterHTTPEndpointsForPlugins registers HTTP endpoints for all plugins
func (pm *PluginManager) RegisterHTTPEndpointsForPlugins(mux *http.ServeMux) {
	log.Println("🔌 Registering HTTP endpoints for plugins...")

	pm.registry.RegisterHTTPEndpoints(mux)

	// Add plugin-specific endpoints
	mux.HandleFunc("GET /plugins", pm.handleListPlugins)
	mux.HandleFunc("GET /plugins/{type}", pm.handleGetPlugin)
	mux.HandleFunc("POST /plugins/{type}/enable", pm.handleEnablePlugin)
	mux.HandleFunc("POST /plugins/{type}/disable", pm.handleDisablePlugin)

	log.Println("✅ Plugin management endpoints registered")
}

// HTTP handlers for plugin management
func (pm *PluginManager) handleListPlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"plugins":    pm.ListPlugins(),
		"statistics": pm.GetPluginStatistics(),
	}

	json.NewEncoder(w).Encode(response)
}

func (pm *PluginManager) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	modelType := r.PathValue("type")

	config, err := pm.GetPluginInfo(modelType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (pm *PluginManager) handleEnablePlugin(w http.ResponseWriter, r *http.Request) {
	modelType := r.PathValue("type")

	if err := pm.EnablePlugin(modelType); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "enabled",
		"plugin":  modelType,
		"message": "Plugin enabled successfully",
	})
}

func (pm *PluginManager) handleDisablePlugin(w http.ResponseWriter, r *http.Request) {
	modelType := r.PathValue("type")

	if err := pm.DisablePlugin(modelType); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "disabled",
		"plugin":  modelType,
		"message": "Plugin disabled successfully",
	})
}
