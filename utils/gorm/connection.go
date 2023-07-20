package gorm

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewGormMasterConnectionMysql(params DBParamMasterConn) IGormMaster {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		params.UserName,
		params.Password,
		params.Host,
		params.Port,
		params.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(params.DbConnPool)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(params.DbConnPool)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(params.DbLifeTime))

	return &Gorm{
		db: db,
	}
}

func NewGormSlaveConnectionMysql(params DBParamSlaveConn) IGormSlave {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		params.UserName,
		params.Password,
		params.Host,
		params.Port,
		params.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(params.DbConnPool)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(params.DbConnPool)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(params.DbLifeTime))

	return &Gorm{
		db: db,
	}
}

func NewGormMasterConnectionPostgres(params DBParamMasterConn) IGormMaster {
	var (
		cfg = gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		params.Host,
		params.UserName,
		params.Password,
		params.Name,
		params.Port)

	if !params.SQLDebug {
		cfg.Logger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(dsn), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(params.DbConnPool)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(params.DbConnPool)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(params.DbLifeTime))

	return &Gorm{
		db: db,
	}
}

func NewGormSlaveConnectionPostgres(params DBParamSlaveConn) IGormSlave {
	var (
		cfg = gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		params.Host,
		params.UserName,
		params.Password,
		params.Name,
		params.Port)

	if !params.SQLDebug {
		cfg.Logger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(dsn), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(params.DbConnPool)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(params.DbConnPool)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(params.DbLifeTime))

	return &Gorm{
		db:             db,
		replicaLagTime: params.ReplicaLagTime,
	}
}
