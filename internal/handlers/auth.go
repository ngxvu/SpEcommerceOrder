package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"gorm.io/gorm"
	model "kimistore/internal/models"
	repo "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
	"kimistore/internal/utils/sync_ob"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/service/jwt_user"
	"net/http"
)

type AuthUserHandler struct {
	newRepo repo.PGInterface
}

func NewAuthUserHandler(pgRepo repo.PGInterface) *AuthUserHandler {
	return &AuthUserHandler{newRepo: pgRepo}
}

func (a *AuthUserHandler) Login(ctx *gin.Context) {

	log := logger.WithTag("AuthUserHandler|Login")

	tx := a.newRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := a.newRepo.DBWithTimeout(ctx)
	defer cancel()

	var request model.UserLoginRequest

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		err = app_errors.AppError("Invalid request", app_errors.StatusValidationError)
		logger.LogError(log, err, " fail to bind json")
		_ = ctx.Error(err)
		return
	}

	user, err := a.GetUser(map[string]interface{}{"email": request.Email}, ctx)
	if err != nil {
		logger.LogError(log, err, "fail to get user")
		_ = ctx.Error(err)
		return
	}

	isAuthenticated := utils.CheckPasswordHash(request.Password, user.Password)
	if !isAuthenticated {
		logger.LogError(log, err, "fail to check password")
		err := app_errors.AppError("Fail to Authorized", app_errors.StatusUnauthorized)
		_ = ctx.Error(err)
		return
	}

	accessTokenClaims, err := jwt_user.GenerateJWTTokenUser(ctx, user.Role, "access")
	if err != nil {
		logger.LogError(log, err, "fail to generate access token")
		err = app_errors.AppError("Fail to Authorized", app_errors.StatusUnauthorized)
		_ = ctx.Error(err)
		return
	}

	refreshTokenClaims, err := jwt_user.GenerateJWTTokenUser(ctx, user.Role, "refresh")
	if err != nil {
		logger.LogError(log, err, " fail to generate refresh token")
		_ = ctx.Error(err)
		return
	}

	securityAuthenticatedUser := jwt_user.SecAuthUserMapper(user, accessTokenClaims, refreshTokenClaims)

	response := jwt_user.JWTUserDataResponse{
		Meta: utils.NewMetaData(ctx),
		Data: *securityAuthenticatedUser,
	}

	tx.Commit()

	ctx.JSON(http.StatusOK, response)

}

func (a *AuthUserHandler) GetUser(userMap map[string]interface{}, ctx context.Context) (*model.User, error) {

	tx := a.newRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := a.newRepo.DBWithTimeout(ctx)
	defer cancel()

	var user model.User

	if err := tx.Where(userMap).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app_errors.AppError("User not found", app_errors.StatusNotFound)
		}
		err = app_errors.AppError("Failed to get user", app_errors.StatusInternalServerError)
		return nil, err
	}

	return &user, nil
}

func (a *AuthUserHandler) Register(ctx *gin.Context) {

	log := logger.WithTag("AuthUserHandler|RegisterUser")

	// Begin transaction
	tx := a.newRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := a.newRepo.DBWithTimeout(ctx)
	defer cancel()

	var request model.UserRegisterRequest

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		err = app_errors.AppError("Invalid request", app_errors.StatusValidationError)
		logger.LogError(log, err, " fail to bind json")
		_ = ctx.Error(err)
		return
	}

	existingOwnerCar, _ := a.GetUser(map[string]interface{}{"email": request.Email}, ctx)
	if existingOwnerCar != nil {
		err = app_errors.AppError("Email already exists", app_errors.StatusConflict)
		logger.LogError(log, err, " email already exists")
		_ = ctx.Error(err)
		return
	}

	hashedPassword, err := utils.HashPassword(*request.Password)
	if err != nil {
		err = app_errors.AppError("Failed to create user", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "fail to hash password")
		_ = ctx.Error(err)
		return
	}

	ob := model.User{}

	sync_ob.Sync(request, &ob)
	ob.Password = hashedPassword

	if err := tx.Create(&ob).Error; err != nil {
		err = app_errors.AppError("Failed to create user", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "fail to create ob")
		_ = ctx.Error(err)
		return
	}

	tx.Commit()

	// Return a clean response without sensitive data
	response := model.UserResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &ob,
	}
	ctx.JSON(http.StatusOK, response)
}
