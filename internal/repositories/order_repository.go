package repositories

import (
	"context"
	"errors"

	"github.com/galihfebrizki/dbo-api/config"
	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/models"
	"github.com/galihfebrizki/dbo-api/utils/gorm"
	"github.com/galihfebrizki/dbo-api/utils/rabbitmq"
	"github.com/galihfebrizki/dbo-api/utils/redis"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

type IOrderRepository interface {
	GetOrderByOrderId(ctx context.Context, orderId string) (models.Order, error)
	GetOrderItemByOrderId(ctx context.Context, orderId string) ([]models.OrderItem, error)
	GetOrderByUserId(ctx context.Context, userId string) ([]models.Order, error)
	GetOrderPagination(ctx context.Context, page int, rowPerPage int, dateFrom string, dateTo string) ([]models.Order, int, error)
	CreateOrder(ctx context.Context, order models.InsertOrder) error
	UpdateOrder(ctx context.Context, order models.InsertOrder) error
	DeleteOrder(ctx context.Context, orderId string) error
	DeleteOrderItem(ctx context.Context, orderItemId string) error
	SearchOrder(ctx context.Context, querySearch string) ([]models.Order, error)
	UpdateStatusOrder(ctx context.Context, status int, orderId string) error
	InsertLog(ctx context.Context, dataLog models.OrderLog) error
}

type OrderRepository struct {
	Master   gorm.IGormMaster
	Slave    gorm.IGormSlave
	Redis    redis.Iredis
	Rabbitmq rabbitmq.IRabbitMQ
}

func NewOrderRepository(master gorm.IGormMaster, slave gorm.IGormSlave, redis redis.Iredis, rabbitmq rabbitmq.IRabbitMQ) IOrderRepository {
	return &OrderRepository{
		Master:   master,
		Slave:    slave,
		Redis:    redis,
		Rabbitmq: rabbitmq,
	}
}

func (r *OrderRepository) GetOrderByOrderId(ctx context.Context, orderId string) (models.Order, error) {
	var order models.Order

	err := r.Slave.WithContext(ctx).
		Where(`"orders"."id" = ?`, orderId).First(&order)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return models.Order{}, err
	}

	return order, nil
}

func (r *OrderRepository) GetOrderItemByOrderId(ctx context.Context, orderId string) ([]models.OrderItem, error) {
	var order []models.OrderItem = make([]models.OrderItem, 0)

	err := r.Slave.WithContext(ctx).
		Select("order_items.*, items.item_name, items.sku").
		Joins(`LEFT JOIN items ON order_items."item_id" = items."id"`).
		Where("order_id = ?", orderId).Find(&order)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return order, err
	}

	return order, nil
}

func (r *OrderRepository) GetOrderByUserId(ctx context.Context, userId string) ([]models.Order, error) {
	var order []models.Order

	err := r.Slave.WithContext(ctx).
		Where(`"orders"."user_id" = ?`, userId).Find(&order)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return []models.Order{}, err
	}

	return order, nil
}

func (r *OrderRepository) GetOrderPagination(ctx context.Context, page int, rowPerPage int, dateFrom string, dateTo string) ([]models.Order, int, error) {
	var (
		orders []models.Order
		count  int
	)

	offset := (page - 1) * rowPerPage

	err := r.Slave.WithContext(ctx).
		DB().Where("created_at >= ? and created_at <= ?", dateFrom+" 00:00:00", dateTo+" 23:59:59").Limit(rowPerPage).
		Offset(offset).Find(&orders).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return []models.Order{}, 0, err
	}

	err = r.Slave.WithContext(ctx).
		Raw("SELECT count(id) as count FROM orders WHERE created_at >= ? and created_at <= ?", &count, dateFrom+" 00:00:00", dateTo+" 23:59:59")

	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return []models.Order{}, 0, err
	}

	return orders, count, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order models.InsertOrder) error {

	tx := r.Master.WithContext(ctx).DB().Begin()
	err := tx.Table("orders").Create(&order).Error

	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	err = tx.Table("order_items").Create(&order.OrderItem).Error

	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	return tx.Commit().Error
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, order models.InsertOrder) error {

	tx := r.Master.WithContext(ctx).DB().Begin()
	err := tx.Table("orders").Where("id = ?", order.Id).Updates(&order).Error
	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"item_id", "quantity", "item_price", "discount_amount", "updated_at"}),
	}).Table("order_items").Save(&order.OrderItem).Error
	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	return tx.Commit().Error
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, orderId string) error {

	tx := r.Master.WithContext(ctx).DB().Begin()
	err := tx.Exec("DELETE FROM orders WHERE id = ?", orderId).Error
	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", orderId).Error
	if err != nil {
		tx.Rollback()
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	return tx.Commit().Error
}

func (r *OrderRepository) DeleteOrderItem(ctx context.Context, orderItemId string) error {

	err := r.Master.WithContext(ctx).DB().Exec("DELETE FROM order_items WHERE id = ?", orderItemId).Error
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	return nil
}

func (r *OrderRepository) SearchOrder(ctx context.Context, querySearch string) ([]models.Order, error) {
	var orderId []models.Order

	// get data from redis
	err := r.Redis.Get(ctx, "query_search_order_"+querySearch, &orderId)
	if err == nil {
		return orderId, nil
	}

	err = r.Slave.WithContext(ctx).
		Raw(`SELECT distinct o.* FROM orders o JOIN order_items oi ON o.id = oi.order_id
			LEFT JOIN items i on i.id = oi.item_id
			WHERE i.item_name ilike ? or o.id = ?`, &orderId, "%"+querySearch+"%", querySearch)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Info(err)
		} else {
			logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		}
		return []models.Order{}, err
	}

	// set data to redis
	err = r.Redis.Set(ctx, "query_search_order_"+querySearch, orderId, config.GetCacheTime())
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Errorf("set redis error: %s", err)
	}

	return orderId, nil
}

func (r *OrderRepository) UpdateStatusOrder(ctx context.Context, status int, orderId string) error {
	err := r.Master.WithContext(ctx).DB().
		Exec(`UPDATE orders SET status = ? WHERE id = ?`, status, orderId).Error
	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	return nil
}

func (r *OrderRepository) InsertLog(ctx context.Context, dataLog models.OrderLog) error {

	err := r.Master.WithContext(ctx).DB().Table("order_logs").Create(&dataLog).Error

	if err != nil {
		logrus.WithField(helper.GetRequestIDContext(ctx)).Error(err)
		return err
	}

	return nil
}
