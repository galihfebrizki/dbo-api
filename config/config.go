package config

import (
	"os"
	"strconv"
	"time"

	"github.com/galihfebrizki/dbo-api/utils/gorm"
	"github.com/galihfebrizki/dbo-api/utils/rabbitmq"
	"github.com/galihfebrizki/dbo-api/utils/redis"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	DEVELOPMENTENV = "development"
)

type ConfigStructure struct {
	Name              string
	Env               string
	Port              string
	Version           string
	ServerIdleTimeout int
	ServerTimeout     int
	DbConnectionPool  int
	DbLifeTime        int
	DbSqlDebug        bool
	Secret            string
	LogLevel          int
	Snowflake         struct {
		Order     int64
		OrderItem int64
		User      int64
	}
	Cache struct {
		Redis struct {
			Host string
			Port string
		}
		CacheTime int
	}
	Database struct {
		Postgres struct {
			Read struct {
				Host           string
				Port           int
				UserName       string
				Password       string
				Name           string
				Extras         string
				ReplicaLagTime int
			}
			Write struct {
				Host     string
				Port     int
				UserName string
				Password string
				Name     string
				Extras   string
			}
		}
	}
	MessageBroker struct {
		RabbitMq struct {
			URL         string
			Host        string
			SampleQueue string
			Username    string
			Password    string
			Concurrency int
			BindingTime int
		}
	}
}

var cfg ConfigStructure

func InitConfig() {

	godotenv.Load()

	//config for go running
	cfg.Name = GetEnvString("NAME", "")
	cfg.Env = GetEnvString("ENV", DEVELOPMENTENV)
	cfg.Port = GetEnvString("PORT", ":8000")
	cfg.Version = GetEnvString("VERSION", "v0.0.0")
	cfg.DbConnectionPool = GetEnvInt("DB_CONN_POOL", 100)
	cfg.DbLifeTime = GetEnvInt("DB_LIFE_TIME", 10)
	cfg.DbSqlDebug = GetEnvBool("DB_SQL_DEBUG", true)
	cfg.Secret = GetEnvString("SECRET", "")

	// log
	cfg.LogLevel = GetEnvInt("LOG_LEVEL", 1)

	// snowflake
	cfg.Snowflake.Order = GetEnvInt64("SNOWFLAKE_ORDER_NODE", 1)
	cfg.Snowflake.User = GetEnvInt64("SNOWFLAKE_USER_NODE", 2)
	cfg.Snowflake.OrderItem = GetEnvInt64("SNOWFLAKE_ORDER_ITEM_NODE", 3)

	// redis
	cfg.Cache.Redis.Host = GetEnvString("REDIS_HOST", "localhost")
	cfg.Cache.Redis.Port = GetEnvString("REDIS_PORT", ":6379")
	cfg.Cache.CacheTime = GetEnvInt("CACHE_TIME", 5)

	// postgres read
	cfg.Database.Postgres.Read.Host = GetEnvString("DB_READ_HOST", "localhost")
	cfg.Database.Postgres.Read.Port = GetEnvInt("DB_READ_PORT", 5432)
	cfg.Database.Postgres.Read.UserName = GetEnvString("DB_READ_USERNAME", "dbo_admin")
	cfg.Database.Postgres.Read.Password = GetEnvString("DB_READ_PASSWORD", "dbo_admin")
	cfg.Database.Postgres.Read.Name = GetEnvString("DB_READ_NAME", "toko")
	cfg.Database.Postgres.Read.Extras = GetEnvString("DB_READ_EXTRAS", "sslmode=disable")
	cfg.Database.Postgres.Read.ReplicaLagTime = GetEnvInt("DB_READ_REPLICALAG_TIME", 1)

	// postgres write
	cfg.Database.Postgres.Write.Host = GetEnvString("DB_WRITE_HOST", "localhost")
	cfg.Database.Postgres.Write.Port = GetEnvInt("DB_WRITE_PORT", 5432)
	cfg.Database.Postgres.Write.UserName = GetEnvString("DB_WRITE_USERNAME", "dbo_admin")
	cfg.Database.Postgres.Write.Password = GetEnvString("DB_WRITE_PASSWORD", "dbo_admin")
	cfg.Database.Postgres.Write.Name = GetEnvString("DB_WRITE_NAME", "toko")
	cfg.Database.Postgres.Write.Extras = GetEnvString("DB_WRITE_EXTRAS", "sslmode=disable")

	//messagebroker
	cfg.MessageBroker.RabbitMq.URL = GetEnvString("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	cfg.MessageBroker.RabbitMq.Host = GetEnvString("RABBITMQ_HOST", "http://localhost:15672")
	cfg.MessageBroker.RabbitMq.SampleQueue = GetEnvString("RABBITMQ_SAMPLE_QUEUE", "queue")
	cfg.MessageBroker.RabbitMq.Username = GetEnvString("RABBITMQ_USERNAME", "guest")
	cfg.MessageBroker.RabbitMq.Password = GetEnvString("RABBITMQ_PASSWORD", "guest")
	cfg.MessageBroker.RabbitMq.Concurrency = GetEnvInt("RABBITMQ_CONSUMER_CONCURENCY", 1)
	cfg.MessageBroker.RabbitMq.BindingTime = GetEnvInt("RABBITMQ_CONSUMER_BINDING_TIME", 60)

	if cfg.Env == DEVELOPMENTENV {
		log.Infof("start development mode with config: %+v\n", cfg)
	}
}

func Get() ConfigStructure {
	return cfg
}

func Set(key string, value string) error {
	err := os.Setenv(key, value)
	if err != nil {
		log.Printf("error when update environment variable service with error: %s", err.Error())
		return err
	}

	return err
}

func GetEnvString(key string, dflt string) string {
	value := os.Getenv(key)
	if value == "" {
		return dflt
	}
	return value
}

func GetEnvInt(key string, dflt int) int {
	value := os.Getenv(key)
	i, err := strconv.ParseInt(value, 10, 64)
	if value == "" && err != nil {
		return dflt
	}
	return int(i)
}

func GetEnvInt64(key string, dflt int64) int64 {
	value := os.Getenv(key)
	i, err := strconv.ParseInt(value, 10, 64)
	if value == "" && err != nil {
		return dflt
	}
	return i
}

func GetEnvFloat(key string, dflt float64) float64 {
	value := os.Getenv(key)
	i, err := strconv.ParseFloat(value, 64)
	if value == "" && err != nil {
		return dflt
	}
	return i
}

func GetEnvBool(key string, dflt bool) bool {
	value := os.Getenv(key)
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return result
}

func GetCacheTime() time.Duration {
	return time.Duration(Get().Cache.CacheTime) * time.Minute
}

func BuildMasterDBParam() gorm.DBParamMasterConn {
	return gorm.DBParamMasterConn{
		Host:       cfg.Database.Postgres.Write.Host,
		Port:       cfg.Database.Postgres.Write.Port,
		UserName:   cfg.Database.Postgres.Write.UserName,
		Password:   cfg.Database.Postgres.Write.Password,
		Name:       cfg.Database.Postgres.Write.Name,
		DbConnPool: cfg.DbConnectionPool,
		DbLifeTime: cfg.DbLifeTime,
		SQLDebug:   cfg.DbSqlDebug,
	}
}

func BuildSlaveDBParam() gorm.DBParamSlaveConn {
	return gorm.DBParamSlaveConn{
		Host:           cfg.Database.Postgres.Read.Host,
		Port:           cfg.Database.Postgres.Read.Port,
		UserName:       cfg.Database.Postgres.Read.UserName,
		Password:       cfg.Database.Postgres.Read.Password,
		Name:           cfg.Database.Postgres.Read.Name,
		Extras:         cfg.Database.Postgres.Read.Extras,
		DbConnPool:     cfg.DbConnectionPool,
		DbLifeTime:     cfg.DbLifeTime,
		SQLDebug:       cfg.DbSqlDebug,
		ReplicaLagTime: cfg.Database.Postgres.Read.ReplicaLagTime,
	}
}

func BuildRedisParam() redis.RedisParam {
	return redis.RedisParam{
		Address: cfg.Cache.Redis.Host + cfg.Cache.Redis.Port,
	}
}

func BuildRabbitMQParam() rabbitmq.RabbitMQParam {
	return rabbitmq.RabbitMQParam{
		Url:         cfg.MessageBroker.RabbitMq.URL,
		Concurrency: cfg.MessageBroker.RabbitMq.Concurrency,
	}
}
