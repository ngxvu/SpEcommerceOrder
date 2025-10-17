package repo

import (
	model "basesource/internal/models"
	pgGorm "basesource/internal/repo/pg-gorm"
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type AuthUserRepository struct {
	db pgGorm.PGInterface
}

func NewAuthUserRepository(newPgRepo pgGorm.PGInterface) *AuthUserRepository {
	return &AuthUserRepository{db: newPgRepo}
}

type AuthUserRepoInterface interface {
	GetUser(ctx context.Context, userMap map[string]interface{}, tx *gorm.DB) (*model.User, error)
	Register(ctx context.Context, user *model.User, tx *gorm.DB) (*model.User, error)
}

func (a *AuthUserRepository) Register(ctx context.Context, user *model.User, tx *gorm.DB) (*model.User, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = a.db.DBWithTimeout(ctx)
		defer cancel()
	}

	if err := tx.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (a *AuthUserRepository) GetUser(ctx context.Context, userMap map[string]interface{}, tx *gorm.DB) (*model.User, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = a.db.DBWithTimeout(ctx)
		defer cancel()
	}

	var user model.User
	if err := tx.Where(userMap).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
