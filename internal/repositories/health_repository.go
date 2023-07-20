package repositories

import (
	"context"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/utils/gorm"
	"github.com/galihfebrizki/dbo-api/utils/redis"

	log "github.com/sirupsen/logrus"
)

type IHealthRepository interface {
	CheckDBMaster(ctx context.Context) bool
	CheckDBSlave(ctx context.Context) bool
	CheckRedis(ctx context.Context) bool
}

type HealthRepository struct {
	Master gorm.IGormMaster
	Slave  gorm.IGormSlave
	Redis  redis.Iredis
}

func NewHealthRepository(master gorm.IGormMaster, slave gorm.IGormSlave, redis redis.Iredis) IHealthRepository {
	return &HealthRepository{
		Master: master,
		Slave:  slave,
		Redis:  redis,
	}
}

// CheckDBMaster implements IHealthRepository
func (r *HealthRepository) CheckDBMaster(ctx context.Context) bool {
	err := r.Master.Ping(ctx)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return false
	}

	return true
}

// CheckDBSlave implements IHealthRepository
func (r *HealthRepository) CheckDBSlave(ctx context.Context) bool {
	err := r.Slave.Ping(ctx)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return false
	}

	return true
}

// CheckRedis implements IHealthRepository
func (r *HealthRepository) CheckRedis(ctx context.Context) bool {
	err := r.Redis.Ping(ctx)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return false
	}

	return true
}
