package gormx

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Option func(db *gorm.DB) *gorm.DB

func applyOptions(db *gorm.DB, opts ...Option) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, len(opts))
	for i := range opts {
		scopes[i] = opts[i]
	}
	if len(scopes) == 0 {
		return db
	}
	return db.Scopes(scopes...)
}

func NoConflict(names ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		var columns []clause.Column
		if len(names) > 0 {
			columns = make([]clause.Column, len(names))
			for i := range names {
				columns[i] = clause.Column{
					Name: names[i],
				}
			}
		}
		return db.Clauses(clause.OnConflict{
			Columns:   columns,
			DoNothing: true,
		})
	}
}

func Pagination(page, size int) Option {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case size > 100:
			size = 100
		case size <= 0:
			size = 20
		}
		offset := (page - 1) * size
		return db.Offset(offset).Limit(size)
	}
}

func WithId(id int64) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id=?", id)
	}
}

func Wildcard() Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select("*")
	}
}
