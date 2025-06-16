package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type AgentStatus string

const (
	AgentStatusPending  AgentStatus = "pending"
	AgentStatusActive   AgentStatus = "active"
	AgentStatusInactive AgentStatus = "inactive"
)

type AgentHealthStatus string

const (
	AgentHealthStatusUnknown   AgentHealthStatus = "unknown"
	AgentHealthStatusHealthy   AgentHealthStatus = "healthy"
	AgentHealthStatusUnhealthy AgentHealthStatus = "unhealthy"
	AgentHealthStatusDegraded  AgentHealthStatus = "degraded"
)

type Agent struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name            string    `gorm:"unique;not null"`
	Description     string
	Status          AgentStatus `gorm:"type:agent_status;default:'pending'"`
	LastSeen        time.Time
	LastIP          string
	EnrollmentToken string            `gorm:"unique;not null"`
	HealthStatus    AgentHealthStatus `gorm:"type:agent_health_status;default:'unknown'"`
	HealthLastCheck time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
