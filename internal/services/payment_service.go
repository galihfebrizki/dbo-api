package services

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/internal/repositories"
	"github.com/galihfebrizki/dbo-api/internal/responses"
	"github.com/galihfebrizki/dbo-api/utils/gorm"

	"github.com/sirupsen/logrus"
)

type IPaymentService interface {
	PaymentProccessSend(ctx context.Context, orderId string) (int, responses.GenericResponse)
	PaymentProccessReceived(ctx context.Context, orderData models.Order) error
}

type PaymentService struct {
	PaymentRepository repositories.IPaymentRepository
	OrderRepository   repositories.IOrderRepository
}

func NewPaymentService(repository repositories.IPaymentRepository, orderRepository repositories.IOrderRepository) IPaymentService {
	return &PaymentService{
		PaymentRepository: repository,
		OrderRepository:   orderRepository,
	}
}

func (s *PaymentService) PaymentProccessSend(ctx context.Context, orderId string) (int, responses.GenericResponse) {
	order, err := s.OrderRepository.GetOrderByOrderId(ctx, orderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
		} else {
			return http.StatusOK, *responses.NewGenericResponse(1008, nil)
		}
	}

	if order.Id == "" {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error("Order not found")
		return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
	}

	order.OrderItem, err = s.OrderRepository.GetOrderItemByOrderId(ctx, orderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
		}
	}

	if order.Status != helper.StatusCreate {
		return http.StatusOK, *responses.NewGenericResponse(1012, nil)
	}

	err = s.OrderRepository.UpdateStatusOrder(ctx, helper.StatusReadyToPay, orderId)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1, nil)
	}

	currentTime := time.Now()

	err = s.OrderRepository.InsertLog(ctx, models.OrderLog{
		OrderId:     orderId,
		OrderStatus: helper.StatusReadyToPay,
		CreatedAt:   &currentTime,
		UpdatedAt:   &currentTime,
	})
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
	}

	order.Status = helper.StatusReadyToPay

	err = s.PaymentRepository.PublishOrderToPayment(ctx, order)
	if err != nil {
		return http.StatusInternalServerError, *responses.NewGenericResponse(-1018, nil)
	}

	return http.StatusOK, *responses.NewGenericResponse(0, order)
}

func (s *PaymentService) PaymentProccessReceived(ctx context.Context, orderData models.Order) error {

	// check orderId are exist
	if orderData.Id == "" {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Infof("payment not valid orderId : %s", orderData.Id)
		return nil
	}

	// check order
	order, err := s.OrderRepository.GetOrderByOrderId(ctx, orderData.Id)
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return nil
	}

	if order.Status != helper.StatusReadyToPay {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(responses.GetErrorCodeEN(1012))
	}

	// call 3rd party payment

	err = s.OrderRepository.UpdateStatusOrder(ctx, helper.StatusPaid, order.Id)
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return nil
	}

	currentTime := time.Now()

	err = s.OrderRepository.InsertLog(ctx, models.OrderLog{
		OrderId:     order.Id,
		OrderStatus: helper.StatusPaid,
		CreatedAt:   &currentTime,
		UpdatedAt:   &currentTime,
	})
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
	}

	return nil
}
