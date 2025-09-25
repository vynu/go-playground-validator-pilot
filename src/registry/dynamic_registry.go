// Package registry provides dynamic model registration with file system monitoring
// This file now delegates to the unified registry system to eliminate code duplication
package registry

import (
	"context"
	"log"
	"net/http"
)

// DynamicModelRegistry is now a thin wrapper around UnifiedRegistry for backward compatibility
type DynamicModelRegistry struct {
	*UnifiedRegistry
}

// NewDynamicModelRegistry creates a new dynamic registry (delegates to unified)
func NewDynamicModelRegistry(unifiedRegistry *UnifiedRegistry, modelsPath, validationsPath string) *DynamicModelRegistry {
	return &DynamicModelRegistry{
		UnifiedRegistry: unifiedRegistry,
	}
}

// GetRegisteredModelsWithDetails returns detailed model information (delegates to unified)
func (dmr *DynamicModelRegistry) GetRegisteredModelsWithDetails() map[string]interface{} {
	return dmr.UnifiedRegistry.GetRegisteredModelsWithDetails()
}

// GetAllModels returns all registered models (delegates to unified)
func (dmr *DynamicModelRegistry) GetAllModels() map[ModelType]*ModelInfo {
	return dmr.UnifiedRegistry.GetAllModels()
}

// Global dynamic registry instance
var globalDynamicRegistry *DynamicModelRegistry

// GetGlobalDynamicRegistry returns the global dynamic model registry instance
func GetGlobalDynamicRegistry() *DynamicModelRegistry {
	if globalDynamicRegistry == nil {
		unifiedRegistry := GetGlobalRegistry()
		globalDynamicRegistry = NewDynamicModelRegistry(unifiedRegistry, "models", "validations")
	}
	return globalDynamicRegistry
}

// StartDynamicRegistration starts the unified registration system
func StartDynamicRegistration(ctx context.Context, mux *http.ServeMux) error {
	log.Println("🔄 Starting unified automatic registration system...")
	return StartRegistration(ctx, mux)
}