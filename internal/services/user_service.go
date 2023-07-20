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
	"github.com/galihfebrizki/dbo-api/middleware"
	"github.com/galihfebrizki/dbo-api/utils/gorm"
	utils "github.com/galihfebrizki/dbo-api/utils/snowflake"

	"github.com/sirupsen/logrus"
)

type IUserService interface {
	UserSessionValidation(ctx context.Context, username, password string) (int, responses.GenericResponse)
	GetLoginData(ctx context.Context, userId string) (int, responses.GenericResponse)
	GetUserByUserId(ctx context.Context, userId string) (int, responses.GenericResponse)
	GetListUser(ctx context.Context, page int, rowPerPage int) (int, responses.GenericResponse)
	CreateUser(ctx context.Context, user models.User) (int, responses.GenericResponse)
	UpdateUser(ctx context.Context, user models.User) (int, responses.GenericResponse)
	DeleteUser(ctx context.Context, userId string) (int, responses.GenericResponse)
	SearchUser(ctx context.Context, querySearch string) (int, responses.GenericResponse)
	IsSuperUser(ctx context.Context, userId string) bool
}

type UserService struct {
	UserRepository  repositories.IUserRepository
	OrderRepository repositories.IOrderRepository
}

func NewUserService(repository repositories.IUserRepository, orderRepository repositories.IOrderRepository) IUserService {
	return &UserService{
		UserRepository:  repository,
		OrderRepository: orderRepository,
	}
}

func (s *UserService) UserSessionValidation(ctx context.Context, username, password string) (int, responses.GenericResponse) {

	user, err := s.UserRepository.GetUserByUsernamePassword(ctx, username, password)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusUnauthorized, *responses.NewGenericResponse(1004, nil)
		}
	}

	token, expirationTime, err := middleware.GenerateJWTToken(user.Username, user.Id)
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return http.StatusInternalServerError, *responses.NewGenericResponse(1005, nil)
	}

	ok := s.UserRepository.CreateSessionUser(ctx, user.Id, token)
	if !ok {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error("Failed to create session")
		return http.StatusInternalServerError, *responses.NewGenericResponse(1006, nil)
	}

	return http.StatusOK, *responses.NewGenericResponse(0, models.LoginData{
		Token:     token,
		ExpiresAt: expirationTime,
	})
}

func (s *UserService) GetLoginData(ctx context.Context, userId string) (int, responses.GenericResponse) {

	user, err := s.UserRepository.GetUserSession(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(1007, nil)
		}
	}

	return http.StatusOK, *responses.NewGenericResponse(0, user)
}

func (s *UserService) GetUserByUserId(ctx context.Context, userId string) (int, responses.GenericResponse) {
	user, err := s.UserRepository.GetUserByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(1007, nil)
		}
	}

	return http.StatusOK, *responses.NewGenericResponse(0, user)
}

func (s *UserService) GetListUser(ctx context.Context, page int, rowPerPage int) (int, responses.GenericResponse) {
	user, count, err := s.UserRepository.GetUserPagination(ctx, page, rowPerPage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusOK, *responses.NewGenericResponse(1007, nil)
		}
	}

	return http.StatusOK, *responses.NewGenericResponse(0, responses.DataPaginationResponse{
		DataPage: user,
		Count:    count,
	})
}

func (s *UserService) CreateUser(ctx context.Context, user models.User) (int, responses.GenericResponse) {

	currentTime := time.Now()

	user.Id = utils.GenerateSnowflakeUser()
	user.CustomerData.UserId = user.Id
	user.Password = helper.MD5(user.Password)
	user.CreatedAt = &currentTime
	user.CustomerData.CreatedAt = &currentTime

	err := s.UserRepository.CreateUser(ctx, user)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	return http.StatusCreated, *responses.NewGenericResponse(0, user)
}

func (s *UserService) UpdateUser(ctx context.Context, user models.User) (int, responses.GenericResponse) {

	currentTime := time.Now()

	user.Password = helper.MD5(user.Password)
	user.UpdatedAt = &currentTime
	user.CustomerData.UpdatedAt = &currentTime

	err := s.UserRepository.UpdateUser(ctx, user)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	return http.StatusCreated, *responses.NewGenericResponse(0, user)
}

func (s *UserService) DeleteUser(ctx context.Context, userId string) (int, responses.GenericResponse) {

	order, err := s.OrderRepository.GetOrderByUserId(ctx, userId)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	if len(order) > 0 {
		return http.StatusOK, *responses.NewGenericResponse(1009, nil)
	}

	err = s.UserRepository.DeleteUser(ctx, userId)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	return http.StatusOK, *responses.NewGenericResponse(0, nil)
}

func (s *UserService) SearchUser(ctx context.Context, querySearch string) (int, responses.GenericResponse) {

	users, err := s.UserRepository.SearchUser(ctx, querySearch)
	if err != nil {
		return http.StatusOK, *responses.NewGenericResponse(1008, nil)
	}

	return http.StatusOK, *responses.NewGenericResponse(0, users)
}

func (s *UserService) IsSuperUser(ctx context.Context, userId string) bool {
	user, err := s.UserRepository.GetUserByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
			return false
		}
	}

	return user.Level > 0
}
