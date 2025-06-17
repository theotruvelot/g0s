package database

import (
	"errors"
	"github.com/theotruvelot/g0s/internal/server/models"
	"gorm.io/gorm"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) GetAgentByID(agentID string) (*models.Agent, error) {
	agent := &models.Agent{}
	result := r.db.Where("id = ?", agentID).First(agent)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return agent, nil
}

func (r *AgentRepository) CreateAgent(agent *models.Agent) error {
	result := r.db.Create(agent)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *AgentRepository) UpdateAgent(agent *models.Agent) error {
	result := r.db.Save(agent)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *AgentRepository) DeleteAgent(agentID string) error {
	result := r.db.Delete(&models.Agent{}, agentID)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
