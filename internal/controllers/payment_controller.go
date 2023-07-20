package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/internal/responses"
	"github.com/galihfebrizki/dbo-api/internal/services"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type PaymentController struct {
	PaymentService services.IPaymentService
}

func NewPaymentController(service services.IPaymentService) *PaymentController {
	return &PaymentController{
		PaymentService: service,
	}
}

func (h *PaymentController) ConsumerPaymentProccess(ctx context.Context, message amqp.Delivery) bool {
	var orderData models.Order

	ctx = helper.SetRequestIDToContext(ctx, message.MessageId)
	logrus.WithField(helper.GetRequestIDContext(ctx)).Infof("Message inbound : %s", message.Body)

	err := json.Unmarshal([]byte(message.Body), &orderData)
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err.Error())
		return false
	}

	err = h.PaymentService.PaymentProccessReceived(ctx, orderData)
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err.Error())
		return false
	}

	return true
}

func (h *PaymentController) PaymentOrder(c *gin.Context) {
	var request models.PaymentOrder

	ctx := helper.GetGinContext(c)

	// Parse the JSON request body
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	c.JSON(h.PaymentService.PaymentProccessSend(ctx, request.OrderId))
}
