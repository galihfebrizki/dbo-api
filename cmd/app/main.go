package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/galihfebrizki/dbo-api/config"
	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/utils/log"
	utils "github.com/galihfebrizki/dbo-api/utils/snowflake"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {
	config.InitConfig()
	utils.InitSnowflakeUser()
	utils.InitSnowflakeOrder()
	utils.InitSnowflakeOrderItem()
	log.InitLog(config.Get().Env, config.Get().LogLevel)
}

func main() {
	ctx := helper.SetRequestIDToContext(context.Background(), helper.GenerateRandomString(32))

	cfg := config.Get()

	signal.Notify(helper.ExitAMQP, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	defer close(helper.ExitAMQP)
	defer close(helper.ExitHTTP)
	defer close(helper.ExitConsumer)

	// init consumer
	consumer := InitializedConsumer(
		config.BuildMasterDBParam(),
		config.BuildSlaveDBParam(),
		config.BuildRedisParam(),
		config.BuildRabbitMQParam(),
	)

	// init server
	gin := InitializedServer(
		config.BuildMasterDBParam(),
		config.BuildSlaveDBParam(),
		config.BuildRedisParam(),
		config.BuildRabbitMQParam(),
	)

	startConsumer(ctx, consumer, cfg)

	startServer(ctx, gin, cfg)
}

func startServer(ctx context.Context, e *gin.Engine, cfg config.ConfigStructure) {
	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: e,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("shutting down the server")
		}
	}()

	// Wait for exit signal
	<-helper.ExitHTTP
	logrus.WithField(helper.GetRequestIDContext(ctx)).Infoln("Wait for http process done")

	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.ServerTimeout*int(time.Second)))
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal(err)
	}

	logrus.WithField(helper.GetRequestIDContext(ctx)).Infoln("http already exited")
}

func startConsumer(ctx context.Context, amqp *AmqpController, cfg config.ConfigStructure) {
	go amqp.StartConsumer(ctx, cfg)
}
