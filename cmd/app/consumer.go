package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/galihfebrizki/dbo-api/config"
	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/controllers"
	"github.com/galihfebrizki/dbo-api/utils/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type Consumer struct {
	Ctx       context.Context
	AppName   string
	QueueName string
	Worker    func(ctx context.Context, d amqp.Delivery) bool
	Requeue   bool
}

type AmqpController struct {
	// consumer conn
	Amqp rabbitmq.IRabbitMQ

	// controllers -> add if any controller exist
	PaymentController *controllers.PaymentController
}

func NewAmqpConsumer(
	iRabbitMq rabbitmq.IRabbitMQ,
	paymentController *controllers.PaymentController,
) *AmqpController {
	return &AmqpController{
		Amqp:              iRabbitMq,
		PaymentController: paymentController,
	}
}

func (h *AmqpController) StartConsumer(ctx context.Context, cfg config.ConfigStructure) {
	appName := cfg.Name

	// register consumer
	consumerList := []Consumer{
		{
			Ctx:       ctx,
			AppName:   appName,
			QueueName: helper.PaymentProccess,
			Worker:    h.PaymentController.ConsumerPaymentProccess,
			Requeue:   true,
		},
	}

	go h.BindConsumer(ctx, consumerList, cfg)
}

func (h *AmqpController) BindConsumer(ctx context.Context, consumers []Consumer, cfg config.ConfigStructure) {
	tickerBind := time.Tick(time.Duration(cfg.MessageBroker.RabbitMq.BindingTime) * time.Second)
	consumerTag := helper.GenerateRandomString(10)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic and continue running the program
				log.WithField(helper.GetRequestIDContext(ctx)).Errorf("goroutine panicked: %v", r)
			}
		}()

		for range tickerBind {
			for _, consumer := range consumers {
				// check consumer
				queueURL := fmt.Sprintf("%s/api/queues/%s/%s", cfg.MessageBroker.RabbitMq.Host, "%2F", consumer.QueueName)

				_, isAvailable, err := h.GetConsumer(queueURL, cfg.MessageBroker.RabbitMq.Username, cfg.MessageBroker.RabbitMq.Password, consumerTag)

				log.WithField(helper.GetRequestIDContext(ctx)).Debugf("Consumer Bind : Queue %s is %v", consumer.QueueName, isAvailable)
				if err != nil {
					log.WithField(helper.GetRequestIDContext(ctx)).Error(err.Error())
					break
				}

				if !isAvailable && err == nil {

					log.WithField(helper.GetRequestIDContext(ctx)).Debugf("Started Consumer Bind : Queue %s with tag %s", consumer.QueueName, consumerTag)
					// start consumer
					go func(consumer Consumer) {
						defer func() {
							if r := recover(); r != nil {
								// Log the panic and continue running the program
								log.WithField(helper.GetRequestIDContext(ctx)).Errorf("goroutine panicked: %v", r)
							}
						}()

						err := h.Amqp.ConsumeMessage(consumer.Ctx, consumer.AppName, consumer.QueueName, consumerTag, consumer.Worker, consumer.Requeue)
						if err != nil {
							log.WithField(helper.GetRequestIDContext(ctx)).Error(err.Error())
						}
					}(consumer)

				}
			}
		}
	}()

	// listen OS exit signal
	<-helper.ExitAMQP

	// trigger exit all consumer
	for i := 0; i < helper.ConsumerCount; i++ {
		helper.ExitConsumer <- true
	}

	// waiting for all worker concurrency done
	for i := 0; i < config.Get().MessageBroker.RabbitMq.Concurrency*helper.ConsumerCount; i++ {
		<-helper.ExitConcurrency
	}

	// close amqp connection
	h.Amqp.Close(ctx)

	// trigger exit for http
	helper.ExitHTTP <- true
}

func (h *AmqpController) GetConsumer(queueURL, username, password, consumerTag string) (int, bool, error) {
	// Create an HTTP client
	client := &http.Client{}

	consumerStatus := false

	fmt.Println("URL :", queueURL)

	// Create an HTTP request
	req, err := http.NewRequest("GET", queueURL, nil)
	if err != nil {
		return 0, consumerStatus, err
	}

	// Set the basic authentication credentials if required
	req.SetBasicAuth(username, password)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return 0, consumerStatus, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, consumerStatus, err
	}

	// Parse the JSON response
	type consumerDetail struct {
		ConsumerTag string `json:"consumer_tag"`
	}

	var queueInfo struct {
		Consumers       int              `json:"consumers"`
		ConsumerDetails []consumerDetail `json:"consumer_details"`
	}
	err = json.Unmarshal(body, &queueInfo)
	if err != nil {
		return 0, consumerStatus, err
	}

	for _, consumer := range queueInfo.ConsumerDetails {
		consumerUnique := strings.Split(consumer.ConsumerTag, "|")
		consumerUniqueLen := len(consumerUnique)
		consumerId := consumerUnique[consumerUniqueLen-1]

		if consumerId == consumerTag {
			consumerStatus = true
		}
	}

	return queueInfo.Consumers, consumerStatus, nil
}
