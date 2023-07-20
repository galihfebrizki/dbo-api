package repositories

import (
	"context"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/utils/gorm"
	"github.com/galihfebrizki/dbo-api/utils/rabbitmq"
	"github.com/galihfebrizki/dbo-api/utils/redis"

	"github.com/sirupsen/logrus"
)

type IPaymentRepository interface {
	PublishOrderToPayment(ctx context.Context, orderMsg models.Order) error
}

type PaymentRepository struct {
	Master   gorm.IGormMaster
	Slave    gorm.IGormSlave
	Redis    redis.Iredis
	Rabbitmq rabbitmq.IRabbitMQ
}

func NewPaymentRepository(master gorm.IGormMaster, slave gorm.IGormSlave, redis redis.Iredis, rabbitmq rabbitmq.IRabbitMQ) IPaymentRepository {
	return &PaymentRepository{
		Master:   master,
		Slave:    slave,
		Redis:    redis,
		Rabbitmq: rabbitmq,
	}
}

func (r *PaymentRepository) PublishOrderToPayment(ctx context.Context, orderMsg models.Order) error {
	err := r.Rabbitmq.PublishMessage(ctx, helper.PaymentProccess, orderMsg)
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	return nil
}
