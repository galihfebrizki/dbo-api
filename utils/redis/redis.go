package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/galihfebrizki/dbo-api/helper"

	redis "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type Iredis interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	LPush(ctx context.Context, key string, value interface{}) error
	LPop(ctx context.Context, key string, dest interface{}) error
	RPush(ctx context.Context, key string, value interface{}) error
	RPop(ctx context.Context, key string, dest interface{}) error
	Llen(ctx context.Context, key string) (int64, error)
	LMove(ctx context.Context, source, dest, srcpos, destpos string) error
	LTrim(ctx context.Context, key string, start, stop int64) error
	Del(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

type Redis struct {
	redis *redis.Client
}

type RedisParam struct {
	Address string
}

func NewRedisConn(param RedisParam) Iredis {
	ctx := context.Background()

	redisServer := redis.NewClient(&redis.Options{
		Addr: param.Address,
	})

	status := redisServer.Ping(ctx)
	if status.Err() != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Fatal(status.Err().Error())
	}

	return &Redis{
		redis: redisServer,
	}
}

func (rdb *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	err = rdb.redis.Set(ctx, key, string(val), ttl).Err()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := rdb.redis.Get(ctx, key).Result()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	err = json.Unmarshal([]byte(val), &dest)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) LPush(ctx context.Context, key string, value interface{}) error {
	val, err := json.Marshal(value)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	err = rdb.redis.LPush(ctx, key, string(val)).Err()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) LPop(ctx context.Context, key string, dest interface{}) error {
	val, err := rdb.redis.LPop(ctx, key).Result()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	err = json.Unmarshal([]byte(val), &dest)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) RPush(ctx context.Context, key string, value interface{}) error {
	val, err := json.Marshal(value)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	err = rdb.redis.RPush(ctx, key, string(val)).Err()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) RPop(ctx context.Context, key string, dest interface{}) error {
	val, err := rdb.redis.RPop(ctx, key).Result()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	err = json.Unmarshal([]byte(val), &dest)
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) Llen(ctx context.Context, key string) (int64, error) {
	llen, err := rdb.redis.LLen(ctx, key).Result()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return 0, err
	}

	return llen, err
}

func (rdb *Redis) LMove(ctx context.Context, source, dest, srcpos, destpos string) error {
	err := rdb.redis.LMove(ctx, source, dest, srcpos, destpos).Err()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) LTrim(ctx context.Context, key string, start, stop int64) error {
	err := rdb.redis.LTrim(ctx, key, start, stop).Err()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) Del(ctx context.Context, key string) error {
	err := rdb.redis.Del(ctx, key).Err()
	if err != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(err.Error())
		return err
	}

	return err
}

func (rdb *Redis) Ping(ctx context.Context) error {
	status := rdb.redis.Ping(ctx)
	if status.Err() != nil {
		log.WithField(helper.GetRequestIDContext(ctx)).Debug(status.Err().Error())
		return status.Err()
	}

	return nil
}
