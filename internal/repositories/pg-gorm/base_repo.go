package pg_gorm

import (
	"basesource/pkg/http/utils"
	"context"
	"gorm.io/gorm"
)

type RepoPG struct {
	db    *gorm.DB
	debug bool
}

func NewPGRepo(db *gorm.DB) PGInterface {
	return &RepoPG{db: db}
}

type PGInterface interface {
	GetRepo() *gorm.DB
	DBWithTimeout(ctx context.Context) (*gorm.DB, context.CancelFunc)
}

func (r *RepoPG) GetRepo() *gorm.DB {
	return r.db
}

func (r *RepoPG) DBWithTimeout(ctx context.Context) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, utils.GeneralQueryTimeout)
	return r.db.WithContext(ctx), cancel
}
