package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	model "order/internal/models"
	repo "order/internal/repositories/pg-gorm"
	"order/internal/services"
	"order/pkg/http/utils/app_errors"
)

type OrderHandler struct {
	newRepo      repo.PGInterface
	orderService *services.OrderService
}

func NewOrderHandler(newRepo repo.PGInterface, orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{newRepo: newRepo, orderService: orderService}
}

func (o *OrderHandler) CreateOrder(ctx *gin.Context) {
	var requestCreateOrder model.CreateOrderRequest

	err := ctx.ShouldBindJSON(&requestCreateOrder)
	if err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	context := ctx.Request.Context()

	response, err := o.orderService.CreateOrder(context, requestCreateOrder)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
