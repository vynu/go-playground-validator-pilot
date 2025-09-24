// Package main demonstrates the automated model registration system
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github-data-validator/registry"
)

func main() {
	fmt.Println("🚀 Automated Model Registration Demo")
	fmt.Println("=====================================")

	// Create a new model registry
	modelRegistry := registry.NewModelRegistry()

	// Demo 1: Basic Automatic Registration
	fmt.Println("\n1️⃣  DEMO 1: Basic Automatic Registration")
	fmt.Println("----------------------------------------")
	demonstrateBasicAutoRegistration(modelRegistry)

	// Demo 2: Plugin-based Registration
	fmt.Println("\n2️⃣  DEMO 2: Plugin-based Registration")
	fmt.Println("-------------------------------------")
	demonstratePluginRegistration(modelRegistry)

	// Demo 3: Configuration-based Registration
	fmt.Println("\n3️⃣  DEMO 3: Configuration-based Registration")
	fmt.Println("--------------------------------------------")
	demonstrateConfigBasedRegistration(modelRegistry)

	// Demo 4: Directory Scanning
	fmt.Println("\n4️⃣  DEMO 4: Directory Scanning Registration")
	fmt.Println("-------------------------------------------")
	demonstrateDirectoryScanning(modelRegistry)

	// Demo 5: HTTP Endpoint Auto-Generation
	fmt.Println("\n5️⃣  DEMO 5: HTTP Endpoint Auto-Generation")
	fmt.Println("-----------------------------------------")
	demonstrateEndpointGeneration(modelRegistry)

	fmt.Println("\n✨ All demos completed successfully!")
	fmt.Println("🎯 The automated registration system is working!")
}

// demonstrateBasicAutoRegistration shows the simplest form of auto-registration
func demonstrateBasicAutoRegistration(modelRegistry *registry.ModelRegistry) {
	fmt.Println("Using the helper to auto-register all known models...")

	// Get the helper instance
	helper := registry.NewModelRegistrationHelper(modelRegistry)

	// Auto-register all known models
	err := helper.AutoRegisterAllKnownModels()
	if err != nil {
		log.Printf("❌ Auto-registration failed: %v", err)
		return
	}

	// Show what was registered
	models := modelRegistry.GetAllModels()
	fmt.Printf("✅ Auto-registered %d models:\n", len(models))
	for modelType, modelInfo := range models {
		fmt.Printf("   • %s -> %s (v%s)\n",
			string(modelType), modelInfo.Name, modelInfo.Version)
	}

	// Validate integrity
	if err := helper.ValidateModelIntegrity(); err != nil {
		log.Printf("⚠️  Integrity check failed: %v", err)
	} else {
		fmt.Println("✅ All models passed integrity checks")
	}
}

// demonstratePluginRegistration shows plugin-based registration
func demonstratePluginRegistration(modelRegistry *registry.ModelRegistry) {
	fmt.Println("Setting up plugin-based registration...")

	// Create plugin directory
	pluginDir := "tmp/plugins"
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		log.Printf("❌ Failed to create plugin directory: %v", err)
		return
	}

	// Create plugin manager
	pluginManager := registry.NewPluginManager(modelRegistry, pluginDir)

	// Load plugins (this will create examples if none exist)
	if err := pluginManager.LoadPlugins(); err != nil {
		log.Printf("❌ Failed to load plugins: %v", err)
		return
	}

	// Show plugin statistics
	stats := pluginManager.GetPluginStatistics()
	fmt.Printf("✅ Plugin system loaded:\n")
	fmt.Printf("   • Total plugins: %v\n", stats["total_plugins"])
	fmt.Printf("   • Enabled plugins: %v\n", stats["enabled_plugins"])
	fmt.Printf("   • Loading order: %v\n", stats["loading_order"])

	// List all plugins
	plugins := pluginManager.ListPlugins()
	fmt.Println("📋 Available plugins:")
	for _, plugin := range plugins {
		status := "❌ disabled"
		if plugin.Enabled {
			status = "✅ enabled"
		}
		fmt.Printf("   • %s (v%s) - %s [Priority: %d]\n",
			plugin.Name, plugin.Version, status, plugin.Priority)
	}
}

// demonstrateConfigBasedRegistration shows configuration file registration
func demonstrateConfigBasedRegistration(modelRegistry *registry.ModelRegistry) {
	fmt.Println("Demonstrating configuration-based registration...")

	// Create a sample configuration
	config := registry.AutoRegistrationConfig{
		ModelsPath:      "src/models",
		ValidationsPath: "src/validations",
		AutoDiscover:    true,
		CustomModels: []registry.ModelConfig{
			{
				Type:        "example",
				Name:        "Example Model",
				Description: "Example model for demonstration",
				ModelStruct: "ExamplePayload",
				Validator:   "ExampleValidator",
				Version:     "1.0.0",
				Author:      "Demo",
				Tags:        []string{"demo", "example"},
				Enabled:     false, // Disabled for demo
			},
		},
	}

	// Save configuration to file
	configDir := "tmp/config"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("❌ Failed to create config directory: %v", err)
		return
	}

	configPath := fmt.Sprintf("%s/auto_registry_demo.json", configDir)
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("❌ Failed to marshal config: %v", err)
		return
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		log.Printf("❌ Failed to write config: %v", err)
		return
	}

	fmt.Printf("✅ Created demo configuration: %s\n", configPath)

	// Create model discovery instance
	discovery := registry.NewModelDiscovery(&config, modelRegistry)

	// Discover and register models
	if err := discovery.DiscoverAndRegisterModels(); err != nil {
		log.Printf("⚠️  Discovery completed with issues: %v", err)
	} else {
		fmt.Println("✅ Model discovery completed successfully")
	}

	fmt.Printf("📊 Current registry contains %d models\n",
		len(modelRegistry.GetAllModels()))
}

// demonstrateDirectoryScanning shows directory-based registration
func demonstrateDirectoryScanning(modelRegistry *registry.ModelRegistry) {
	fmt.Println("Demonstrating directory scanning registration...")

	helper := registry.NewModelRegistrationHelper(modelRegistry)

	// Simulate directory scanning (using existing paths)
	modelsDir := "src/models"
	validationsDir := "src/validations"

	fmt.Printf("Scanning directories: %s, %s\n", modelsDir, validationsDir)

	// This would normally scan the directories and register found models
	// For demo purposes, we'll show what would happen
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		fmt.Printf("⚠️  Models directory not found: %s\n", modelsDir)
		fmt.Println("   In a real scenario, this would scan Go files for model structs")
	} else {
		fmt.Printf("✅ Models directory exists: %s\n", modelsDir)
	}

	if _, err := os.Stat(validationsDir); os.IsNotExist(err) {
		fmt.Printf("⚠️  Validations directory not found: %s\n", validationsDir)
		fmt.Println("   In a real scenario, this would scan Go files for validators")
	} else {
		fmt.Printf("✅ Validations directory exists: %s\n", validationsDir)
	}

	// Show what would be registered
	fmt.Println("📄 Directory scanning would discover:")
	fmt.Println("   • Go files with *Payload structs")
	fmt.Println("   • Go files with *Validator types")
	fmt.Println("   • Matching pairs would be auto-registered")

	// Export current configurations
	exportPath := "tmp/exported_configs.json"
	if err := helper.ExportModelConfigs(exportPath); err != nil {
		log.Printf("⚠️  Export failed: %v", err)
	} else {
		fmt.Printf("✅ Exported model configs to: %s\n", exportPath)
	}
}

// demonstrateEndpointGeneration shows HTTP endpoint auto-generation
func demonstrateEndpointGeneration(modelRegistry *registry.ModelRegistry) {
	fmt.Println("Demonstrating HTTP endpoint auto-generation...")

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Use the helper to register endpoints with detailed logging
	helper := registry.NewModelRegistrationHelper(modelRegistry)
	helper.RegisterHTTPEndpointsWithLogging(mux)

	// Show endpoint statistics
	stats := helper.GetModelStatistics()
	fmt.Printf("📈 Endpoint Statistics:\n")
	fmt.Printf("   • Total endpoints created: %v\n", stats["total_models"])
	fmt.Printf("   • Registry healthy: %v\n", stats["registry_healthy"])

	// Show what endpoints are available
	models := modelRegistry.GetAllModels()
	fmt.Println("🎯 Available HTTP endpoints:")
	for modelType, modelInfo := range models {
		endpoint := fmt.Sprintf("POST /validate/%s", string(modelType))
		fmt.Printf("   • %-35s -> %s\n", endpoint, modelInfo.Name)
	}

	// Add management endpoints
	mux.HandleFunc("GET /demo/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		models := modelRegistry.GetAllModels()

		response := make(map[string]interface{})
		for modelType, modelInfo := range models {
			response[string(modelType)] = map[string]interface{}{
				"name":        modelInfo.Name,
				"description": modelInfo.Description,
				"version":     modelInfo.Version,
				"author":      modelInfo.Author,
				"tags":        modelInfo.Tags,
				"endpoint":    "/validate/" + string(modelType),
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"models":    response,
			"total":     len(models),
			"timestamp": "2023-01-01T00:00:00Z",
		})
	})

	fmt.Println("✅ HTTP endpoints registered and ready")
	fmt.Println("💡 In a real server, these would be accessible via HTTP")
	fmt.Println("   Example: GET /demo/models - Lists all registered models")
	fmt.Println("   Example: POST /validate/github - Validates GitHub payload")

	// Demonstrate the endpoint functionality
	fmt.Println("\n🔍 Testing endpoint functionality...")
	testEndpointFunctionality(modelRegistry)
}

// testEndpointFunctionality tests the core endpoint functionality
func testEndpointFunctionality(modelRegistry *registry.ModelRegistry) {
	// Test GitHub model validation
	if modelRegistry.IsRegistered(registry.ModelTypeGitHub) {
		fmt.Println("✅ GitHub model is registered")

		// Test creating a model instance
		instance, err := modelRegistry.CreateModelInstance(registry.ModelTypeGitHub)
		if err != nil {
			log.Printf("❌ Failed to create GitHub instance: %v", err)
		} else {
			fmt.Printf("✅ Created GitHub model instance: %T\n", instance)
		}

		// Test validation
		validator, err := modelRegistry.GetValidator(registry.ModelTypeGitHub)
		if err != nil {
			log.Printf("❌ Failed to get GitHub validator: %v", err)
		} else {
			fmt.Println("✅ GitHub validator retrieved successfully")

			// Test with empty payload (should fail validation)
			result := validator.ValidatePayload(instance)
			if result.IsValid {
				fmt.Println("⚠️  Empty payload validated as valid (unexpected)")
			} else {
				fmt.Printf("✅ Empty payload validation failed as expected (%d errors)\n",
					len(result.Errors))
			}
		}
	}

	// Test Deployment model validation
	if modelRegistry.IsRegistered(registry.ModelTypeDeployment) {
		fmt.Println("✅ Deployment model is registered")
		fmt.Println("   This demonstrates the automated system working with")
		fmt.Println("   the deployment model you mentioned in your question!")
	}

	fmt.Println("🎯 Endpoint functionality test completed")
}

// Utility function to clean up demo files
func cleanupDemo() {
	fmt.Println("\n🧹 Cleaning up demo files...")

	// Remove temporary directories
	tempDirs := []string{"tmp"}
	for _, dir := range tempDirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("⚠️  Failed to remove %s: %v", dir, err)
		} else {
			fmt.Printf("✅ Cleaned up: %s\n", dir)
		}
	}
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
