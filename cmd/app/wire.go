//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/galihfebrizki/dbo-api/internal/controllers"
	"github.com/galihfebrizki/dbo-api/internal/repositories"
	"github.com/galihfebrizki/dbo-api/internal/services"
	"github.com/galihfebrizki/dbo-api/utils/gorm"
	"github.com/galihfebrizki/dbo-api/utils/rabbitmq"
	"github.com/galihfebrizki/dbo-api/utils/redis"
)

var pkgSet = wire.NewSet(
	gorm.NewGormMasterConnectionPostgres,
	gorm.NewGormSlaveConnectionPostgres,
	redis.NewRedisConn,
	rabbitmq.NewRabbitMQConn,
)

var setHealth = wire.NewSet(
	repositories.NewHealthRepository,
	services.NewHealthService,
	controllers.NewHealthController,
)

var setPayment = wire.NewSet(
	repositories.NewPaymentRepository,
	services.NewPaymentService,
	controllers.NewPaymentController,
)

var setOrder = wire.NewSet(
	repositories.NewOrderRepository,
	services.NewOrderService,
	controllers.NewOrderController,
)

var setUser = wire.NewSet(
	repositories.NewUserRepository,
	services.NewUserService,
	controllers.NewUserController,
)

var setItem = wire.NewSet(
	repositories.NewItemRepository,
)

func InitializedServer(masterParam gorm.DBParamMasterConn, slaveParam gorm.DBParamSlaveConn, redisParam redis.RedisParam, mqParam rabbitmq.RabbitMQParam) *gin.Engine {
	wire.Build(
		pkgSet,
		setHealth,
		setOrder,
		setUser,
		setItem,
		setPayment,
		NewRouter,
	)
	return nil
}

func InitializedConsumer(masterParam gorm.DBParamMasterConn, slaveParam gorm.DBParamSlaveConn, redisParam redis.RedisParam, mqParam rabbitmq.RabbitMQParam) *AmqpController {
	wire.Build(
		pkgSet,
		setPayment,
		setOrder,
		NewAmqpConsumer,
	)
	return nil
}
