package repositories

import (
	"context"
	"errors"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/utils/gorm"
	"github.com/galihfebrizki/dbo-api/utils/rabbitmq"
	"github.com/galihfebrizki/dbo-api/utils/redis"

	"github.com/sirupsen/logrus"
)

type IItemRepository interface {
	GetItemByItemId(ctx context.Context, itemId string) (models.Item, error)
}

type ItemRepository struct {
	Master   gorm.IGormMaster
	Slave    gorm.IGormSlave
	Redis    redis.Iredis
	Rabbitmq rabbitmq.IRabbitMQ
}

func NewItemRepository(master gorm.IGormMaster, slave gorm.IGormSlave, redis redis.Iredis, rabbitmq rabbitmq.IRabbitMQ) IItemRepository {
	return &ItemRepository{
		Master:   master,
		Slave:    slave,
		Redis:    redis,
		Rabbitmq: rabbitmq,
	}
}

func (r *ItemRepository) GetItemByItemId(ctx context.Context, itemId string) (models.Item, error) {
	var item models.Item

	err := r.Slave.WithContext(ctx).
		Where(`"items"."id" = ?`, itemId).First(&item)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return models.Item{}, err
	}

	return item, nil
}
