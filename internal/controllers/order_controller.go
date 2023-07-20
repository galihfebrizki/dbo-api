package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/internal/responses"
	"github.com/galihfebrizki/dbo-api/internal/services"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	OrderService services.IOrderService
	UserService  services.IUserService
}

func NewOrderController(service services.IOrderService, userService services.IUserService) *OrderController {
	return &OrderController{
		OrderService: service,
		UserService:  userService,
	}
}

func (h *OrderController) GetOrder(c *gin.Context) {
	ctx := helper.GetGinContext(c)

	orderId := c.Param("orderId")
	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	c.JSON(h.OrderService.GetOrderByOrderId(ctx, orderId, userId.(string)))
}

func (h *OrderController) GetListOrder(c *gin.Context) {
	ctx := helper.GetGinContext(c)

	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	if c.Query("page") == "" {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1003, nil))
		return
	}

	if c.Query("size") == "" {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1003, nil))
		return
	}

	if c.Query("date_from") == "" {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1003, nil))
		return
	}

	if c.Query("date_to") == "" {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1003, nil))
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	size, err := strconv.Atoi(c.Query("size"))
	if err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	layout := "2006-01-02"

	_, err = time.Parse(layout, c.Query("date_from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	_, err = time.Parse(layout, c.Query("date_to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	isSuperUser := h.UserService.IsSuperUser(ctx, userId.(string))
	if isSuperUser {
		c.JSON(h.OrderService.GetListOrder(ctx, page, size, c.Query("date_from"), c.Query("date_to")))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}

func (h *OrderController) CreateOrder(c *gin.Context) {
	var request models.CreateOrder

	ctx := helper.GetGinContext(c)

	// Parse the JSON request body
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	c.JSON(h.OrderService.CreateOrder(ctx, request))
}

func (h *OrderController) UpdateOrder(c *gin.Context) {
	var request models.UpdateOrder

	ctx := helper.GetGinContext(c)

	// Parse the JSON request body
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	c.JSON(h.OrderService.UpdateOrder(ctx, request))
}

func (h *OrderController) DeleteOrder(c *gin.Context) {

	ctx := helper.GetGinContext(c)

	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	orderIdParam := c.Query("id")
	if orderIdParam == "" {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	isSuperUser := h.UserService.IsSuperUser(ctx, userId.(string))
	if isSuperUser {
		c.JSON(h.OrderService.DeleteOrder(ctx, orderIdParam))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}

func (h *OrderController) SearchOrder(c *gin.Context) {

	ctx := helper.GetGinContext(c)

	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	querySearch := c.Query("query")
	if querySearch == "" {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	isSuperUser := h.UserService.IsSuperUser(ctx, userId.(string))
	if isSuperUser {
		c.JSON(h.OrderService.SearchOrder(ctx, querySearch))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}
