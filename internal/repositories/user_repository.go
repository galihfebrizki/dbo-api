package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/galihfebrizki/dbo-api/config"
	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/utils/gorm"
	"github.com/galihfebrizki/dbo-api/utils/rabbitmq"
	"github.com/galihfebrizki/dbo-api/utils/redis"

	grm "gorm.io/gorm"

	"github.com/sirupsen/logrus"
)

type IUserRepository interface {
	GetUserByUserId(ctx context.Context, userId string) (models.User, error)
	GetUserByUsernamePassword(ctx context.Context, username, password string) (models.User, error)
	GetUserPagination(ctx context.Context, page int, rowPerPage int) ([]models.User, int, error)
	CreateSessionUser(ctx context.Context, userId, authSign string) bool
	CreateUser(ctx context.Context, user models.User) error
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUser(ctx context.Context, userId string) error
	SearchUser(ctx context.Context, querySearch string) ([]models.User, error)
	GetUserSession(ctx context.Context, userId string) (models.UserSession, error)
}

type UserRepository struct {
	Master   gorm.IGormMaster
	Slave    gorm.IGormSlave
	Redis    redis.Iredis
	Rabbitmq rabbitmq.IRabbitMQ
}

func NewUserRepository(master gorm.IGormMaster, slave gorm.IGormSlave, redis redis.Iredis, rabbitmq rabbitmq.IRabbitMQ) IUserRepository {
	return &UserRepository{
		Master:   master,
		Slave:    slave,
		Redis:    redis,
		Rabbitmq: rabbitmq,
	}
}

func (r *UserRepository) GetUserByUserId(ctx context.Context, userId string) (models.User, error) {
	var user models.User

	// get data from redis
	err := r.Redis.Get(ctx, "user_"+userId, &user)
	if err == nil {
		return user, nil
	}

	err = r.Slave.WithContext(ctx).
		Joins("CustomerData").
		Where("id = ?", userId).First(&user)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return models.User{}, err
	}

	// set data to redis
	err = r.Redis.Set(ctx, "user_"+userId, user, config.GetCacheTime())
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Errorf("set redis error: %s", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserPagination(ctx context.Context, page int, rowPerPage int) ([]models.User, int, error) {
	var (
		user  []models.User
		count int
	)

	offset := (page - 1) * rowPerPage

	err := r.Slave.WithContext(ctx).
		Joins("CustomerData").
		DB().Limit(rowPerPage).
		Offset(offset).Find(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return []models.User{}, 0, err
	}

	err = r.Slave.WithContext(ctx).
		Raw("SELECT count(id) as count FROM users", &count)

	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return []models.User{}, 0, err
	}

	return user, count, nil
}

func (r *UserRepository) GetUserByUsernamePassword(ctx context.Context, username, password string) (models.User, error) {
	var user models.User

	err := r.Slave.WithContext(ctx).
		Joins("CustomerData").
		Where("username = ? AND password = md5(?)", username, password).First(&user)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) error {

	tx := r.Master.WithContext(ctx).DB().Begin()
	err := tx.Exec(`
		INSERT INTO "users" ("id","username","password","full_name","status","level","created_at") VALUES (?,?,?,?,?,?,?)
	`, user.Id, user.Username, user.Password, user.FullName, user.Status, user.Level, user.CreatedAt).Error

	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	err = tx.Exec(`
	INSERT INTO "customer_data" ("user_id","dob","phone_number","gender","marital_status","address","district_address","city_address","province_address","postal_code","latitude_address","longitude_address","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
	`, user.CustomerData.UserId, user.CustomerData.Dob, user.CustomerData.PhoneNumber, user.CustomerData.Gender, user.CustomerData.MaritalStatus, user.CustomerData.Address, user.CustomerData.DistrictAddress, user.CustomerData.CityAddress,
		user.CustomerData.ProvinceAddress, user.CustomerData.PostalCode, user.CustomerData.LatitudeAddress, user.CustomerData.LongitudeAddress, user.CustomerData.CreatedAt).Error

	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	return tx.Commit().Error
}

func (r *UserRepository) UpdateUser(ctx context.Context, user models.User) error {

	// get data from redis
	err := r.Redis.Get(ctx, "user_"+user.Id, &user)
	if err == nil {
		// delete data from redis
		err = r.Redis.Del(ctx, "user_"+user.Id)
		if err != nil {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
	}

	tx := r.Master.WithContext(ctx).DB().Begin()
	err = tx.Where("id = ?", user.Id).Updates(&models.User{
		Username:  user.Username,
		Password:  user.Password,
		FullName:  user.FullName,
		Status:    user.Status,
		Level:     user.Level,
		UpdatedAt: user.UpdatedAt,
	}).Error
	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	err = tx.Where("user_id = ?", user.Id).Updates(&user.CustomerData).Error
	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	// set data to redis
	err = r.Redis.Set(ctx, "user_"+user.Id, user, config.GetCacheTime())
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Errorf("set redis error: %s", err)
	}

	return tx.Commit().Error
}

func (r *UserRepository) DeleteUser(ctx context.Context, userId string) error {

	tx := r.Master.WithContext(ctx).DB().Begin()
	err := tx.Exec("DELETE FROM users WHERE id = ?", userId).Error
	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	err = tx.Exec("DELETE FROM customer_data WHERE user_id = ?", userId).Error
	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	// get data from redis
	err = r.Redis.Get(ctx, "user_"+userId, &models.User{})
	if err == nil {
		// delete data from redis
		err = r.Redis.Del(ctx, "user_"+userId)
		if err != nil {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
	}

	return tx.Commit().Error
}

func (r *UserRepository) SearchUser(ctx context.Context, querySearch string) ([]models.User, error) {
	var users []models.User

	// get data from redis
	err := r.Redis.Get(ctx, "query_search_user_"+querySearch, &users)
	if err == nil {
		return users, nil
	}

	err = r.Slave.WithContext(ctx).
		Joins("CustomerData").
		Where("username ilike ? or full_name ilike ?", "%"+querySearch+"%", "%"+querySearch+"%").Find(&users)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return []models.User{}, err
	}

	// set data to redis
	err = r.Redis.Set(ctx, "query_search_user_"+querySearch, users, config.GetCacheTime())
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Errorf("set redis error: %s", err)
	}

	return users, nil
}

func (r *UserRepository) CreateSessionUser(ctx context.Context, userId, authSign string) bool {
	err := r.Master.WithContext(ctx).DB().Transaction(func(tx *grm.DB) error {

		currentTime := time.Now()

		// logout all session
		if err := tx.Where("logout_time is null And user_id = ?", userId).Updates(&models.UserSession{LogoutTime: &currentTime}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		// create new active session
		if err := tx.Create(&models.UserSession{
			UserId:     userId,
			Token:      authSign,
			LoginTime:  &currentTime,
			LogoutTime: nil,
		}).Error; err != nil {
			// return any error will rollback
			return err
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return false
	}

	return true
}

func (r *UserRepository) GetUserSession(ctx context.Context, token string) (models.UserSession, error) {
	var user models.UserSession

	err := r.Slave.WithContext(ctx).
		Where("token = ?", token).First(&user)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return models.UserSession{}, err
	}

	return user, nil
}
