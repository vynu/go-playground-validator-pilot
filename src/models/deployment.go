// Package models contains Deployment webhook payload models with validation rules.
package models

import "time"

// DeploymentPayload represents a deployment webhook payload structure
type DeploymentPayload struct {
	// Basic deployment information
	ID          string `json:"id" validate:"required,min=1,max=50"`
	AppName     string `json:"app_name" validate:"required,min=2,max=100,deployment_name"`
	Environment string `json:"environment" validate:"required,oneof=development staging production"`
	Version     string `json:"version" validate:"required,semver"`
	Status      string `json:"status" validate:"required,oneof=pending running completed failed"`

	// Deployment details
	Branch     string    `json:"branch" validate:"required,min=1,max=200"`
	CommitHash string    `json:"commit_hash" validate:"required,len=40,hexadecimal"`
	DeployedBy string    `json:"deployed_by" validate:"required,email"`
	DeployedAt time.Time `json:"deployed_at" validate:"required"`
	Rollback   bool      `json:"rollback" validate:"boolean"`
}
