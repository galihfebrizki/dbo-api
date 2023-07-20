package services

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/internal/repositories"
	"github.com/galihfebrizki/dbo-api/internal/responses"
	"github.com/galihfebrizki/dbo-api/utils/gorm"
	utils "github.com/galihfebrizki/dbo-api/utils/snowflake"

	"github.com/sirupsen/logrus"
)

type IOrderService interface {
	GetOrderByOrderId(ctx context.Context, orderId string, userId string) (int, responses.GenericResponse)
	GetListOrder(ctx context.Context, page int, rowPerPage int, dateFrom string, dateTo string) (int, responses.GenericResponse)
	CreateOrder(ctx context.Context, order models.CreateOrder) (int, responses.GenericResponse)
	UpdateOrder(ctx context.Context, order models.UpdateOrder) (int, responses.GenericResponse)
	DeleteOrder(ctx context.Context, orderId string) (int, responses.GenericResponse)
	SearchOrder(ctx context.Context, querySearch string) (int, responses.GenericResponse)
}

type OrderService struct {
	OrderRepository repositories.IOrderRepository
	ItemRepository  repositories.IItemRepository
	UserService     IUserService
}

func NewOrderService(repository repositories.IOrderRepository, itemRepository repositories.IItemRepository, userService IUserService) IOrderService {
	return &OrderService{
		OrderRepository: repository,
		ItemRepository:  itemRepository,
		UserService:     userService,
	}
}

func (s *OrderService) GetOrderByOrderId(ctx context.Context, orderId string, userId string) (int, responses.GenericResponse) {
	var order models.Order

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

	isSuperUser := s.UserService.IsSuperUser(ctx, userId)
	if !isSuperUser && order.UserId != userId {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error("Unauthorized User")
		return http.StatusInternalServerError, *responses.NewGenericResponse(1004, nil)
	}

	order.OrderItem, err = s.OrderRepository.GetOrderItemByOrderId(ctx, orderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
		}
	}

	return http.StatusOK, *responses.NewGenericResponse(0, order)
}

func (s *OrderService) GetListOrder(ctx context.Context, page int, rowPerPage int, dateFrom string, dateTo string) (int, responses.GenericResponse) {

	orders, count, err := s.OrderRepository.GetOrderPagination(ctx, page, rowPerPage, dateFrom, dateTo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(1007, nil)
		} else {
			return http.StatusOK, *responses.NewGenericResponse(1008, nil)
		}
	}

	wg := sync.WaitGroup{}

	for i := 0; i < len(orders); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			orderItem, err := s.OrderRepository.GetOrderItemByOrderId(ctx, orders[index].Id)
			if err != nil {
				logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
			}
			orders[index].OrderItem = orderItem
		}(i)
	}

	wg.Wait()

	return http.StatusOK, *responses.NewGenericResponse(0, responses.DataPaginationResponse{
		DataPage: orders,
		Count:    count,
	})
}

func (s *OrderService) CreateOrder(ctx context.Context, order models.CreateOrder) (int, responses.GenericResponse) {

	currentTime := time.Now()
	orderItem := []models.InsertOrderItem{}
	dataOrder := models.InsertOrder{
		Id:            utils.GenerateSnowflakeOrder(),
		UserId:        order.UserId,
		Status:        helper.StatusCreate,
		PaymentMethod: order.PaymentMethod,
		CreatedAt:     &currentTime,
	}

	for _, oi := range order.OrderItem {
		item, err := s.ItemRepository.GetItemByItemId(ctx, oi.ItemId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
			} else {
				return http.StatusOK, *responses.NewGenericResponse(1008, nil)
			}
		}

		orderItem = append(orderItem, models.InsertOrderItem{
			Id:        utils.GenerateSnowflakeOrderItem(),
			OrderId:   dataOrder.Id,
			ItemId:    oi.ItemId,
			Quantity:  oi.Quantity,
			ItemPrice: (item.Price * int64(oi.Quantity)),
			CreatedAt: &currentTime,
		})

		dataOrder.TotalAmount += (item.Price * int64(oi.Quantity))
		dataOrder.TotalQuantity += oi.Quantity
	}

	dataOrder.OrderItem = orderItem

	err := s.OrderRepository.CreateOrder(ctx, dataOrder)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	err = s.OrderRepository.InsertLog(ctx, models.OrderLog{
		OrderId:     dataOrder.Id,
		OrderStatus: dataOrder.Status,
		CreatedAt:   &currentTime,
		UpdatedAt:   &currentTime,
	})
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
	}

	return s.GetOrderByOrderId(ctx, dataOrder.Id, order.UserId)
}

func (s *OrderService) UpdateOrder(ctx context.Context, order models.UpdateOrder) (int, responses.GenericResponse) {

	beforeOrderItem, err := s.OrderRepository.GetOrderItemByOrderId(ctx, order.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
		}
	}

	beforeOrder, err := s.OrderRepository.GetOrderByOrderId(ctx, order.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
		}
	}

	if beforeOrder.Status > helper.StatusCreate {
		return http.StatusOK, *responses.NewGenericResponse(1011, nil)
	}

	currentTime := time.Now()
	orderItem := []models.InsertOrderItem{}
	dataOrder := models.InsertOrder{
		Id:            order.OrderId,
		UserId:        order.UserId,
		Status:        helper.StatusCreate,
		PaymentMethod: order.PaymentMethod,
		CreatedAt:     &currentTime,
	}

	if len(beforeOrderItem) > len(order.OrderItem) {
		// delete some previous order item
		for i := len(order.OrderItem); i < len(beforeOrderItem); i++ {
			s.OrderRepository.DeleteOrderItem(ctx, beforeOrderItem[i].Id)
		}
	}

	i := 0
	orderItemId := ""
	createdAt := &currentTime
	updatedAt := &currentTime

	for _, oi := range order.OrderItem {

		item, err := s.ItemRepository.GetItemByItemId(ctx, oi.ItemId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
			}
		}
		if i < len(beforeOrderItem) {
			orderItemId = beforeOrderItem[i].Id
			createdAt = nil
		} else {
			orderItemId = utils.GenerateSnowflakeOrderItem()
			updatedAt = nil
		}

		orderItem = append(orderItem, models.InsertOrderItem{
			Id:        orderItemId,
			OrderId:   dataOrder.Id,
			ItemId:    oi.ItemId,
			Quantity:  oi.Quantity,
			ItemPrice: (item.Price * int64(oi.Quantity)),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})

		dataOrder.TotalAmount += (item.Price * int64(oi.Quantity))
		dataOrder.TotalQuantity += oi.Quantity

		i++
	}

	dataOrder.OrderItem = orderItem

	err = s.OrderRepository.UpdateOrder(ctx, dataOrder)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	return s.GetOrderByOrderId(ctx, dataOrder.Id, order.UserId)
}

func (s *OrderService) DeleteOrder(ctx context.Context, orderId string) (int, responses.GenericResponse) {

	order, err := s.OrderRepository.GetOrderByOrderId(ctx, orderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(-1018, nil)
		} else {
			return http.StatusOK, *responses.NewGenericResponse(1008, nil)
		}
	}

	if order.Id == "" {
		return http.StatusOK, *responses.NewGenericResponse(1009, nil)
	}

	if order.Status > helper.StatusCreate {
		return http.StatusOK, *responses.NewGenericResponse(1010, nil)
	}

	err = s.OrderRepository.DeleteOrder(ctx, orderId)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	return http.StatusOK, *responses.NewGenericResponse(0, nil)
}

func (s *OrderService) SearchOrder(ctx context.Context, querySearch string) (int, responses.GenericResponse) {

	orders, err := s.OrderRepository.SearchOrder(ctx, querySearch)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < len(orders); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			orderItem, err := s.OrderRepository.GetOrderItemByOrderId(ctx, orders[index].Id)
			if err != nil {
				logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
			}
			orders[index].OrderItem = orderItem
		}(i)
	}

	wg.Wait()

	return http.StatusOK, *responses.NewGenericResponse(0, orders)
}
