package controllers

import (
	"net/http"
	"strconv"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/internal/responses"
	"github.com/galihfebrizki/dbo-api/internal/services"
	"github.com/galihfebrizki/dbo-api/middleware"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserService services.IUserService
}

func NewUserController(service services.IUserService) *UserController {
	return &UserController{
		UserService: service,
	}
}

func (h *UserController) Login(c *gin.Context) {
	var request models.AuthUser

	ctx := helper.GetGinContext(c)

	// Parse the JSON request body
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	c.JSON(h.UserService.UserSessionValidation(ctx, request.Username, request.Password))
}

func (h *UserController) GetLoginData(c *gin.Context) {

	ctx := helper.GetGinContext(c)

	token := middleware.GetToken(c)

	c.JSON(h.UserService.GetLoginData(ctx, token))
}

func (h *UserController) GetCustomerData(c *gin.Context) {
	ctx := helper.GetGinContext(c)

	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	userIdParam := c.Param("userId")
	if userIdParam == "" {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1003, nil))
		return
	}

	isSuperUser := h.UserService.IsSuperUser(ctx, userId.(string))
	if isSuperUser {
		c.JSON(h.UserService.GetUserByUserId(ctx, userIdParam))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}

func (h *UserController) GetSelfData(c *gin.Context) {
	ctx := helper.GetGinContext(c)

	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	c.JSON(h.UserService.GetUserByUserId(ctx, userId.(string)))
}

func (h *UserController) GetListCustomerData(c *gin.Context) {
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

	isSuperUser := h.UserService.IsSuperUser(ctx, userId.(string))
	if isSuperUser {
		c.JSON(h.UserService.GetListUser(ctx, page, size))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}

func (h *UserController) CreateUser(c *gin.Context) {
	var request models.User

	ctx := helper.GetGinContext(c)

	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	// Parse the JSON request body
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	isSuperUser := h.UserService.IsSuperUser(ctx, userId.(string))
	if isSuperUser {
		c.JSON(h.UserService.CreateUser(ctx, request))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}

func (h *UserController) UpdateUser(c *gin.Context) {
	var request models.User

	ctx := helper.GetGinContext(c)

	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	// Parse the JSON request body
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	isSuperUser := h.UserService.IsSuperUser(ctx, userId.(string))
	if isSuperUser {
		c.JSON(h.UserService.UpdateUser(ctx, request))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}

func (h *UserController) DeleteUser(c *gin.Context) {

	ctx := helper.GetGinContext(c)

	userId, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
		return
	}

	userIdParam := c.Query("id")
	if userIdParam == "" {
		c.JSON(http.StatusBadRequest, *responses.NewGenericResponse(1003, nil))
		return
	}

	isSuperUser := h.UserService.IsSuperUser(ctx, userId.(string))
	if isSuperUser {
		c.JSON(h.UserService.DeleteUser(ctx, userIdParam))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}

func (h *UserController) SearchUser(c *gin.Context) {

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
		c.JSON(h.UserService.SearchUser(ctx, querySearch))
	} else {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil))
	}
}
