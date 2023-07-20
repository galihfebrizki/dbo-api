package services

import (
	"context"

	"github.com/galihfebrizki/dbo-api/config"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/internal/repositories"
)

type IHealthService interface {
	HealthCheck(ctx context.Context) models.SystemHealth
}

type HealthService struct {
	HealthRepository repositories.IHealthRepository
}

func NewHealthService(repository repositories.IHealthRepository) IHealthService {
	return &HealthService{
		HealthRepository: repository,
	}
}

// Health implements IHealthService
func (s *HealthService) HealthCheck(ctx context.Context) models.SystemHealth {
	return models.SystemHealth{
		Version: config.Get().Version,
		ServiceSupport: models.ServiceSupport{
			Master: s.HealthRepository.CheckDBMaster(ctx),
			Slave:  s.HealthRepository.CheckDBSlave(ctx),
			Redis:  s.HealthRepository.CheckRedis(ctx),
		},
	}
}
