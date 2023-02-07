package gormx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	ErrNoRowsAffected = errors.New("no rows affected")
)

type Config struct {
	Dialector   gorm.Dialector
	MaxIdleConn int
	MaxOpenConn int
	MaxLifetime int64
	Debug       bool
}

type Gormx struct {
	cfg *Config
	db  *gorm.DB
}

func New(cfg *Config, opts ...gorm.Option) (*Gormx, error) {
	db, err := gorm.Open(cfg.Dialector, opts...)
	if err != nil {
		return nil, fmt.Errorf("open database connection failed, %w", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get origin db instance failed, %w", err)
	}

	if cfg.MaxIdleConn > 0 {
		sqlDb.SetMaxIdleConns(cfg.MaxIdleConn)
	}

	if cfg.MaxOpenConn > 0 {
		sqlDb.SetMaxOpenConns(cfg.MaxOpenConn)
	}

	if cfg.MaxLifetime > 0 {
		sqlDb.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	}

	return &Gormx{
		cfg: cfg,
		db:  db,
	}, nil
}

func (s *Gormx) DB() *gorm.DB {
	return s.db
}

func (s *Gormx) BuildOptions(opts ...Option) *gorm.DB {
	return s.buildWithOptions(opts...)
}

func (s *Gormx) Debug() *Gormx {
	return s.clone(s.db.Debug())
}

// WithContext 添加上下文，会新建 Gormx 对象
func (s *Gormx) WithContext(ctx context.Context) *Gormx {
	return s.clone(s.db.WithContext(ctx))
}

func (s *Gormx) Model(value interface{}) *Gormx {
	return s.clone(s.db.Model(value))
}

func (s *Gormx) WithConn(conn *gorm.DB) *Gormx {
	return s.clone(conn)
}

// Tx 开启事务
func (s *Gormx) Tx(fn func(tx *Gormx) error, opts ...*sql.TxOptions) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return fn(s.WithConn(tx))
	}, opts...)
}

func (s *Gormx) Insert(doc interface{}, opts ...Option) error {
	return s.buildWithOptions(opts...).Create(doc).Error
}

func (s *Gormx) Save(doc interface{}, opts ...Option) error {
	return s.buildWithOptions(opts...).Save(doc).Error
}

func (s *Gormx) FindOne(dest interface{}, opts ...Option) error {
	return s.buildWithOptions(opts...).First(dest).Error
}

func (s *Gormx) FindMany(dest interface{}, opts ...Option) error {
	return s.buildWithOptions(opts...).Find(dest).Error
}

func (s *Gormx) Pluck(column string, dest interface{}, opts ...Option) error {
	return s.buildWithOptions(opts...).Pluck(column, dest).Error
}

func (s *Gormx) Count(opts ...Option) (int64, error) {
	var total int64
	if err := s.buildWithOptions(opts...).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (s *Gormx) Exists(dest interface{}, opts ...Option) (bool, error) {
	var exists bool
	opts = append(opts, Wildcard())
	stmt := s.dryRun(opts...).Take(dest).Statement
	query := s.db.Raw(fmt.Sprintf("SELECT EXISTS(%s)", stmt.SQL.String()), stmt.Vars...)
	if err := query.Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Gormx) Updates(dest interface{}, opts ...Option) error {
	db := s.buildWithOptions(opts...).Updates(dest)
	if err := db.Error; err != nil {
		return err
	}
	if db.RowsAffected == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

func (s *Gormx) Update(column string, value interface{}, opts ...Option) error {
	db := s.buildWithOptions(opts...).Update(column, value)
	if err := db.Error; err != nil {
		return err
	}
	if db.RowsAffected == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

func (s *Gormx) Delete(dest interface{}, opts ...Option) error {
	return s.buildWithOptions(opts...).Delete(dest).Error
}

func (s *Gormx) Raw(sql string, values ...interface{}) *Gormx {
	return s.clone(s.db.Raw(sql, values...))
}

func (s *Gormx) Exec(sql string, values ...interface{}) error {
	return s.db.Exec(sql, values...).Error
}

func (s *Gormx) Scan(dest interface{}) error {
	return s.db.Scan(dest).Error
}

// ----------------------------------------------------------------------------------------------------------------------------

func (s *Gormx) dryRun(opts ...Option) *gorm.DB {
	return applyOptions(s.db.Session(&gorm.Session{DryRun: true}), opts...)
}

func (s *Gormx) buildWithOptions(opts ...Option) *gorm.DB {
	return applyOptions(s.db, opts...)
}

func (s *Gormx) clone(db *gorm.DB) *Gormx {
	return &Gormx{
		db: db,
	}
}
