package handlers

import (
	"emission/internal/models"
	"emission/internal/services"
	"emission/internal/utils"
	ginext2 "emission/pkg/http/ginext"
	"net/http"
)

type FactoryHandlers struct {
	service services.FactoryInterface
}

func NewFactoryHandlers(service services.FactoryInterface) *FactoryHandlers {
	return &FactoryHandlers{service: service}
}

func (h *FactoryHandlers) CreateFactory(r *ginext2.Request) (*ginext2.Response, error) {
	req := models.FactoryRequest{}
	if err := r.GinCtx.ShouldBind(&req); err != nil {
		return nil, ginext2.NewError(http.StatusBadRequest, utils.MessageError()[http.StatusBadRequest])
	}

	rs, err := h.service.CreateFactory(r.GinCtx, req)
	if err != nil {
		return nil, err
	}

	return ginext2.NewResponseData(http.StatusCreated, rs), nil
}
