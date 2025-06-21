package models

import "basesource/internal/utils"

type User struct {
	BaseModel
	Name     string `json:"name" gorm:"type:varchar(255);not null"`
	Email    string `json:"email" gorm:"type:varchar(255);not null;unique"`
	Password string `json:"password" gorm:"type:text;not null"`
	Role     string `json:"role" gorm:"type:varchar(50);not null;default:'user'"`
}

func (User) TableName() string {
	return "users"
}

type UserLoginRequest struct {
	Email    string `json:"email" blinding:"required"`
	Password string `json:"password" blinding:"required"`
}

type UserRegisterRequest struct {
	Email    *string `json:"email" binding:"required"`
	Name     *string `json:"name" binding:"required"`
	Password *string `json:"password" binding:"required"`
}

type UserResponse struct {
	Meta *utils.MetaData `json:"meta"`
	Data *User           `json:"data"`
}
