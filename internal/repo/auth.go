package repo

import (
	"gorm.io/gorm"
	model "kimistore/internal/models"
)

type AuthUserRepository struct {
}

func NewAuthUserRepository() *AuthUserRepository {
	return &AuthUserRepository{}
}

type AuthUserRepoInterface interface {
	GetUser(userMap map[string]interface{}, tx *gorm.DB) (*model.User, error)
	Register(user *model.User, tx *gorm.DB) (*model.User, error)
}

func (a *AuthUserRepository) Register(user *model.User, tx *gorm.DB) (*model.User, error) {

	if err := tx.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (a *AuthUserRepository) GetUser(userMap map[string]interface{}, tx *gorm.DB) (*model.User, error) {

	var user model.User
	if err := tx.Where(userMap).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil

}
