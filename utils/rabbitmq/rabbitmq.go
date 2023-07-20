package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/galihfebrizki/dbo-api/helper"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type IRabbitMQ interface {
	Connect(ctx context.Context, url string) error
	ConsumeMessage(ctx context.Context, appName, queueName, consumerTag string, worker func(ctx context.Context, d amqp.Delivery) bool, requeue bool) error
	PublishMessage(ctx context.Context, queue string, body interface{}) error
	PublishMessageToDeathLetter(ctx context.Context, queue string, body interface{}, ttl int) error
	Close(ctx context.Context) error
}

type RabbitMQ struct {
	url         string
	concurrency int
	connection  *amqp.Connection
	channel     *amqp.Channel
	mutex       sync.Mutex
	once        sync.Once
}

type RabbitMQParam struct {
	Url         string
	Concurrency int
}

func NewRabbitMQConn(param RabbitMQParam) IRabbitMQ {
	conn, err := amqp.Dial(param.Url)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ Datasource. URL: %s, Error: %s", param.Url, err.Error())
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel : %s", err.Error())
	}

	return &RabbitMQ{
		url:         param.Url,
		concurrency: param.Concurrency,
		connection:  conn,
		channel:     channel,
		mutex:       sync.Mutex{},
		once:        sync.Once{},
	}
}

func (mq *RabbitMQ) Connect(ctx context.Context, url string) error {
	var err error

	log.WithField(helper.GetRequestIDContext(ctx)).Info("create a connection to url : ", url)

	mq.connection, err = amqp.Dial(url)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Info("failed to create a connection to url : ", url)
		return err
	}

	mq.channel, err = mq.connection.Channel()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Info("create a channel connection to url : ", url)
		return err
	}

	return nil
}

func (mq *RabbitMQ) ConsumeMessage(ctx context.Context, appName, queueName, consumerTag string, worker func(ctx context.Context, d amqp.Delivery) bool, requeue bool) error {

	mq.once.Do(func() {
		mq.mutex.Lock()
		helper.ConsumerCount++
		mq.mutex.Unlock()
	})

	if mq.connection == nil || mq.connection.IsClosed() {
		err := mq.Connect(ctx, mq.url)
		if err != nil {
			log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to reconnect a connection")
		}
	}

	//QueueDeclare declares a queue to hold messages and deliver to consumers. Declaring creates a queue if it doesn't already exist, or ensures that an existing queue matches the same parameters.
	queue, err := mq.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to declare a queue")
		return err
	}

	//Qos controls how many messages or how many bytes the server will try to keep on the network for consumers before receiving delivery acks. The intent of Qos is to make sure the network buffers stay full between the server and client.
	prefetchCount := mq.concurrency
	err = mq.channel.Qos(
		prefetchCount, // prefetch count
		0,             // prefetch size
		false,         // global
	)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to set QOS")
		return err
	}

	consumerTag = fmt.Sprintf("%s|%s|%s", appName, queueName, consumerTag)

	messages, err := mq.channel.Consume(
		queue.Name,  // queue
		consumerTag, // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to register a consumer")
		return err
	}

	for i := 0; i < mq.concurrency; i++ {
		go func(_ctx context.Context, _messages <-chan amqp.Delivery, _done chan bool) {
			for message := range _messages {
				log.WithField(helper.GetRequestIDContext(_ctx)).Infof("consumer %s consume message id %s with body : %v", message.ConsumerTag, message.MessageId, string(message.Body))
				if worker(_ctx, message) {
					log.WithField(helper.GetRequestIDContext(_ctx)).Infof("consumer %s ack message id %s with body : %v", message.ConsumerTag, message.MessageId, string(message.Body))
					message.Ack(false)
				} else {
					log.WithField(helper.GetRequestIDContext(_ctx)).Infof("consumer %s nack message id %s with body : %v", message.ConsumerTag, message.MessageId, string(message.Body))
					message.Nack(false, requeue)
				}
			}
			_done <- true
		}(ctx, messages, helper.ExitConcurrency)
	}

	log.WithField(helper.GetRequestIDContext(ctx)).Infof("Consumer %s already started", consumerTag)

	// Wait for exit signal
	<-helper.ExitConsumer
	log.WithField(helper.GetRequestIDContext(ctx)).Infoln("Got exit signal")

	// Stop receiving message from queue
	err = mq.channel.Cancel(consumerTag, false)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to cancel the consumer")
		return err
	}

	log.WithField(helper.GetRequestIDContext(ctx)).Infoln("Stopped receiving message from queue")

	return nil
}

func (mq *RabbitMQ) PublishMessage(ctx context.Context, queue string, body interface{}) error {

	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if mq.connection == nil || mq.connection.IsClosed() {
		err := mq.Connect(ctx, mq.url)
		if err != nil {
			log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to reconnect a connection")
		}
	}

	// parse data
	dataParse, err := json.Marshal(body)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Errorf("error while parse data : %s", err.Error())
	}

	q, err := mq.channel.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Errorf("error while declare queue : %s", err.Error())
		return err
	}

	_, msgId := helper.GetRequestIDContext(ctx)

	// publish data
	err = mq.channel.PublishWithContext(
		ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(dataParse),
			MessageId:    msgId.(string),
		})

	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Errorf("error while publish data : %s", err.Error())
		return err
	}

	log.WithField(helper.GetRequestIDContext(ctx)).Infof("success send message : %s", body)

	return nil
}

func (mq *RabbitMQ) PublishMessageToDeathLetter(ctx context.Context, queue string, body interface{}, ttl int) error {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if mq.connection == nil || mq.connection.IsClosed() {
		err := mq.Connect(ctx, mq.url)
		if err != nil {
			log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to reconnect a connection")
		}
	}
	// parse data
	dataParse, err := json.Marshal(body)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Errorf("error while parse data : %s", err.Error())
	}
	_, msgId := helper.GetRequestIDContext(ctx)

	exchangeName := fmt.Sprintf("%s.%s", queue, "retry")
	err = mq.channel.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Errorf("error while declaring exchange : %s", err.Error())
		return err
	}

	err = mq.channel.QueueBind(
		queue,
		queue,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Errorf("error while bind queue : %s", err.Error())
		return err
	}

	queueDeathLetter := fmt.Sprintf("%s.%ds.%s", queue, ttl/1000, "retry")
	argsDeathLetter := make(amqp.Table)
	argsDeathLetter["x-dead-letter-exchange"] = exchangeName
	argsDeathLetter["x-dead-letter-routing-key"] = queue
	argsDeathLetter["x-message-ttl"] = ttl

	_, err = mq.channel.QueueDeclare(
		queueDeathLetter,
		true,
		false,
		false,
		false,
		argsDeathLetter)

	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Errorf("error while declaring queue : %s", err.Error())
		return err
	}

	// publish data
	err = mq.channel.PublishWithContext(
		ctx,
		"",               // exchange
		queueDeathLetter, // routing key
		false,            // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(dataParse),
			MessageId:    msgId.(string),
		})

	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Errorf("error while publish data : %s", err.Error())
		return err
	}

	log.WithField(helper.GetRequestIDContext(ctx)).Infof("success send message : %s", body)

	return nil
}

// Close implements IRabbitMQ
func (mq *RabbitMQ) Close(ctx context.Context) error {
	log.WithField(helper.GetRequestIDContext(ctx)).Info("close a connection")

	// Close queue channel and connection
	err := mq.channel.Close()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to close the channel")
		return err
	}

	err = mq.connection.Close()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Error("failed to close the connection")
		return err
	}

	return nil
}
