package handlers

import (
	"emission/internal/models"
	services2 "emission/internal/services"
	"emission/internal/utils"
	ginext2 "emission/pkg/http/ginext"
	"emission/pkg/http/logger"
	"net/http"
)

type EmissionHandlers struct {
	service services2.EmissionInterface
	authSrv services2.AuthInterface
}

func NewEmissionHandlers(service services2.EmissionInterface, authSrv services2.AuthInterface) *EmissionHandlers {
	return &EmissionHandlers{service: service, authSrv: authSrv}
}

func (h *EmissionHandlers) ListEmission(r *ginext2.Request) (*ginext2.Response, error) {
	log := logger.WithCtx(r.GinCtx, "CreateMetadata")
	currentUser, err := h.authSrv.CurrentUser(r.GinCtx.Request)
	if err != nil {
		log.WithError(err).Error("Error when get current user")
		return nil, ginext2.NewError(http.StatusUnauthorized, utils.MessageError()[http.StatusUnauthorized])
	}

	req := models.EmissionListRequest{
		FactoryID: currentUser,
	}
	if err := r.GinCtx.BindQuery(&req); err != nil {
		return nil, ginext2.NewError(http.StatusBadRequest, utils.MessageError()[http.StatusBadRequest])
	}

	filter := &models.EmissionFilter{
		EmissionListRequest: req,
		Pager:               ginext2.NewPagerWithGinCtx(r.GinCtx),
	}

	result, err := h.service.GetListEmission(r.Context(), filter)
	if err != nil {
		return nil, err
	}

	resp := ginext2.NewResponseWithPager(http.StatusOK, result.Records, result.Filter.Pager)
	return resp, nil
}

func (h *EmissionHandlers) CreateEmission(r *ginext2.Request) (*ginext2.Response, error) {
	log := logger.WithCtx(r.GinCtx, "CreateMetadata")
	currentUser, err := h.authSrv.CurrentUser(r.GinCtx.Request)
	if err != nil {
		log.WithError(err).Error("Error when get current user")
		return nil, ginext2.NewError(http.StatusUnauthorized, utils.MessageError()[http.StatusUnauthorized])
	}
	req := models.EmissionCreateRequest{
		FactoryID: currentUser,
	}
	if err := r.GinCtx.ShouldBind(&req); err != nil {
		return nil, ginext2.NewError(http.StatusBadRequest, utils.MessageError()[http.StatusBadRequest])
	}

	rs, err := h.service.CreateEmission(r.GinCtx, req)
	if err != nil {
		return nil, err
	}

	return ginext2.NewResponseData(http.StatusCreated, rs), nil
}
