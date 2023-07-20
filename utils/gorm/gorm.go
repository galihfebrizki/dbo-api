package gorm

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	OnConflict struct {
		UniqueColumn []string
		OnlyUpdate   []string
	}

	OnUpdate struct {
		SelectColumn []string
		Condition    interface{}
	}
	// aliases
	IGormMaster IGorm
	IGormSlave  IGorm
)

type IGorm interface {
	// main
	First(result interface{}, args ...interface{}) error
	Find(result interface{}, args ...interface{}) error
	Upsert(chunkSize int, data interface{}, onConflict OnConflict) error
	Create(data interface{}) error
	Update(data interface{}, onUpdate OnUpdate) error
	Raw(query string, result interface{}, args ...interface{}) error

	// clause
	Table(name string, args ...interface{}) IGorm
	Where(query interface{}, args ...interface{}) IGorm
	Joins(query string, args ...interface{}) IGorm
	Select(query string, args ...interface{}) IGorm
	Group(name string) IGorm
	Preload(name string, args ...interface{}) IGorm

	// sub
	Association() IGorm
	WithContext(ctx context.Context) IGorm
	Error() error
	RowsAffected() int64
	Ping(ctx context.Context) error
	Close() error
	DB() *gorm.DB
}

type Gorm struct {
	db             *gorm.DB
	replicaLagTime int
}

type DBParamMasterConn struct {
	Host       string
	Port       int
	UserName   string
	Password   string
	Name       string
	SQLDebug   bool
	Extras     string
	DbConnPool int
	DbLifeTime int
}

type DBParamSlaveConn struct {
	Host           string
	Port           int
	UserName       string
	Password       string
	Name           string
	SQLDebug       bool
	Extras         string
	DbConnPool     int
	DbLifeTime     int
	ReplicaLagTime int
}

func (g *Gorm) First(result interface{}, args ...interface{}) error {
	time.Sleep(time.Duration(g.replicaLagTime) * time.Millisecond)

	db := g.db.First(result, args...)
	if err := db.Error; err != nil {
		return err
	}

	return nil
}

func (g *Gorm) Find(result interface{}, args ...interface{}) error {
	time.Sleep(time.Duration(g.replicaLagTime) * time.Millisecond)

	db := g.db.Find(result, args...)
	if err := db.Error; err != nil {
		return err
	}

	return nil
}

func (g *Gorm) Upsert(chunkSize int, data interface{}, onConflict OnConflict) error {
	var (
		columns = make([]clause.Column, 0)
	)

	for _, col := range onConflict.UniqueColumn {
		columns = append(columns, clause.Column{
			Name: col,
		})
	}

	// sometimes can be changed
	db := g.db.Clauses(clause.OnConflict{
		Columns:   columns,
		DoUpdates: clause.AssignmentColumns(onConflict.OnlyUpdate),
	})

	if chunkSize > 0 {
		db = db.CreateInBatches(data, chunkSize)
	} else {
		db = db.Create(data)
	}
	return db.Error
}

func (g *Gorm) Create(data interface{}) error {
	db := g.db.Create(data)
	return db.Error
}

func (g *Gorm) Update(data interface{}, onUpdate OnUpdate) error {
	db := g.db.Select(onUpdate.SelectColumn).Where(onUpdate.Condition).Updates(data)
	return db.Error
}

func (g *Gorm) Raw(query string, result interface{}, args ...interface{}) error {
	time.Sleep(time.Duration(g.replicaLagTime) * time.Millisecond)

	err := g.db.Raw(query, args...).Scan(result).Error
	if err != nil {
		return err
	}

	return nil
}

/*
========================================
Clause Func
========================================
*/

func (g *Gorm) Table(name string, args ...interface{}) IGorm {
	db := g.db.Table(name, args...)
	return &Gorm{
		db: db,
	}
}

func (g *Gorm) Where(query interface{}, args ...interface{}) IGorm {
	db := g.db.Where(query, args...)
	return &Gorm{
		db: db,
	}
}

func (g *Gorm) Joins(name string, args ...interface{}) IGorm {
	db := g.db.Joins(name, args...)
	return &Gorm{
		db: db,
	}
}

func (g *Gorm) Preload(name string, args ...interface{}) IGorm {
	db := g.db.Preload(name, args...)
	return &Gorm{
		db: db,
	}
}

func (g *Gorm) Select(name string, args ...interface{}) IGorm {
	db := g.db.Select(name, args...)
	return &Gorm{
		db: db,
	}
}

func (g *Gorm) Group(name string) IGorm {
	db := g.db.Group(name)
	return &Gorm{
		db: db,
	}
}

/*
========================================
Sub Func
========================================
*/

func (g *Gorm) Association() IGorm {
	db := g.db.Session(&gorm.Session{FullSaveAssociations: true})

	return &Gorm{
		db: db,
	}
}

func (g *Gorm) WithContext(ctx context.Context) IGorm {
	db := g.db.WithContext(ctx)
	return &Gorm{
		db: db,
	}
}

func (g *Gorm) Error() error {
	return g.db.Error
}

func (g *Gorm) RowsAffected() int64 {
	return g.db.RowsAffected
}

// DB implements IGorm
func (g *Gorm) DB() *gorm.DB {
	return g.db
}

// Ping implements IGorm
func (g *Gorm) Ping(ctx context.Context) error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Close implements IGorm
func (g *Gorm) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Close()
	if err != nil {
		return err
	}

	return nil
}
