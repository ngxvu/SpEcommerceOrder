package services

import (
	"context"
	"github.com/sirupsen/logrus"
	model "order/internal/models"
	"order/internal/repositories"
	pgGorm "order/internal/repositories/pg-gorm"
	jwt_user2 "order/pkg/core/jwt"
	"order/pkg/core/logger"
	"order/pkg/http/utils"
	"order/pkg/http/utils/app_errors"
	"order/pkg/http/utils/sync_ob"
)

type AuthUserService struct {
	repo      repositories.AuthUserRepoInterface
	newPgRepo pgGorm.PGInterface
}

type AuthUserServiceInterface interface {
	Login(ctx context.Context, request model.UserLoginRequest) (*jwt_user2.JWTUserDataResponse, error)
	Register(ctx context.Context, request model.UserRegisterRequest) (*jwt_user2.JWTUserDataResponse, error)
}

func NewAuthUserService(repo repositories.AuthUserRepoInterface, newRepo pgGorm.PGInterface) *AuthUserService {
	return &AuthUserService{
		repo:      repo,
		newPgRepo: newRepo,
	}
}

func (s *AuthUserService) Login(ctx context.Context, request model.UserLoginRequest) (*jwt_user2.JWTUserDataResponse, error) {
	log := logger.WithTag("AuthUserService|Login")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	getUser, err := s.repo.GetUser(ctx, map[string]interface{}{"email": request.Email}, tx)
	if err != nil {
		logger.LogError(log, err, "fail to get user by email")
		err = app_errors.AppError(app_errors.StatusNotFound, app_errors.StatusNotFound)
		return nil, err
	}

	isAuthenticated := utils.CheckPasswordHash(request.Password, getUser.Password)
	if !isAuthenticated {
		logger.LogError(log, nil, "password does not match")
		err = app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
		return nil, err
	}

	response, err := s.generateTokensAndCreateResponse(ctx, getUser, log)
	if err != nil {
		logger.LogError(log, err, "fail to generate tokens and create response")
		err = app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
		return nil, err
	}

	return response, nil
}

func (s *AuthUserService) Register(ctx context.Context, request model.UserRegisterRequest) (*jwt_user2.JWTUserDataResponse, error) {
	log := logger.WithTag("AuthUserService|Register")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	existingUser, err := s.repo.GetUser(ctx, map[string]interface{}{"email": request.Email}, tx)
	if err != nil {
		logger.LogError(log, err, "failed to get user by email")
		err = app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
		return nil, err
	}

	if existingUser != nil {
		logger.LogError(log, nil, "user already exists with this email")
		err = app_errors.AppError(app_errors.StatusConflict, app_errors.StatusConflict)
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(*request.Password)
	if err != nil {
		logger.LogError(log, err, "failed to hash password")
		err := app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
		return nil, err
	}

	ob := model.User{}

	sync_ob.Sync(request, &ob)
	ob.Password = hashedPassword

	getUser, err := s.repo.Register(ctx, &ob, tx)
	if err != nil {
		logger.LogError(log, err, "failed to register user")
		err = app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
		return nil, err
	}

	tx.Commit()

	response, err := s.generateTokensAndCreateResponse(ctx, getUser, log)
	if err != nil {
		logger.LogError(log, err, "failed to generate tokens and create response")
		err = app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
		return nil, err
	}

	return response, nil
}

func (s *AuthUserService) generateTokensAndCreateResponse(ctx context.Context, user *model.User, log *logrus.Entry) (*jwt_user2.JWTUserDataResponse, error) {

	accessTokenClaims, err := jwt_user2.GenerateJWTTokenUser(ctx, user.Role, "access")
	if err != nil {
		logger.LogError(log, err, "fail to generate access token")
		return nil, app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
	}

	refreshTokenClaims, err := jwt_user2.GenerateJWTTokenUser(ctx, user.Role, "refresh")
	if err != nil {
		logger.LogError(log, err, "fail to generate refresh token")
		return nil, app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
	}

	securityAuthenticatedUser := jwt_user2.SecAuthUserMapper(user, accessTokenClaims, refreshTokenClaims)
	if securityAuthenticatedUser == nil {
		err = app_errors.AppError(app_errors.StatusUnauthorized, app_errors.StatusUnauthorized)
		logger.LogError(log, err, "fail to map security authenticated user")
		return nil, err
	}

	response := jwt_user2.JWTUserDataResponse{
		Meta: utils.NewMetaData(ctx),
		Data: *securityAuthenticatedUser,
	}

	return &response, nil
}
