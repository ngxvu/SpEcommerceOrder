package handlers

import (
	model "basesource/internal/models"
	pgGorm "basesource/internal/repo/pg-gorm"
	"basesource/internal/services"
	"basesource/internal/utils/app_errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthUserHandler struct {
	db              pgGorm.PGInterface
	authUserService services.AuthUserServiceInterface
}

func NewAuthUserHandler(pgRepo pgGorm.PGInterface,
	authUserService services.AuthUserServiceInterface) *AuthUserHandler {
	return &AuthUserHandler{db: pgRepo,
		authUserService: authUserService}
}

func (a *AuthUserHandler) Login(ctx *gin.Context) {

	var request model.UserLoginRequest

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	context := ctx.Request.Context()

	response, err := a.authUserService.Login(context, request)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (a *AuthUserHandler) Register(ctx *gin.Context) {

	var request model.UserRegisterRequest

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}
	context := ctx.Request.Context()

	register, err := a.authUserService.Register(context, request)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, register)
}
