// Package models contains Incident reporting payload models with validation rules.
package models

import "time"

// IncidentPayload represents an incident reporting payload structure for e2e testing
type IncidentPayload struct {
	ID          string    `json:"id" validate:"required,min=3,max=50"`
	Title       string    `json:"title" validate:"required,min=10,max=200"`
	Description string    `json:"description" validate:"required,min=20,max=1000"`
	Severity    string    `json:"severity" validate:"required,oneof=low medium high critical"`
	Status      string    `json:"status" validate:"required,oneof=open investigating resolved closed"`
	Priority    int       `json:"priority" validate:"required,min=1,max=5"`
	Category    string    `json:"category" validate:"required,oneof=bug feature security performance"`
	Environment string    `json:"environment" validate:"required,oneof=development staging production"`
	ReportedBy  string    `json:"reported_by" validate:"required,min=3,max=100"`
	AssignedTo  string    `json:"assigned_to,omitempty" validate:"omitempty,min=3,max=100"`
	ReportedAt  time.Time `json:"reported_at" validate:"required"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Tags        []string  `json:"tags,omitempty" validate:"omitempty,dive,min=2,max=20"`
	Impact      string    `json:"impact,omitempty" validate:"omitempty,oneof=low medium high critical"`
}
