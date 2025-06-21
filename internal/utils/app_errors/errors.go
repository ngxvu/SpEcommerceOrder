package app_errors

import (
	"basesource/internal/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Package app_errors defines the domain app_errors used in the application.
const (
	StatusOK                  = "Successfully"
	StatusForbidden           = "Something when wrong, Your request has been rejected"
	StatusInternalServerError = "Internal server error"
	StatusBadRequest          = "Something when wrong with your request"
	StatusUnauthorized        = "IDMUnauthorized - Permission denied"
	StatusNotFound            = "Request not found - Check your input"
	StatusCreated             = "Created successfully"
	StatusGatewayTimeout      = "Gateway time out"
	StatusConflict            = "Your input has been conflict with another data"
	StatusTooManyRequests     = "Too many request"
	StatusValidationError     = "Validation has been failed"
)

// Response trả về cho APP FE khi có lỗi
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResponseError struct {
	ErrorResp ErrorResponse `json:"error"`
}

func (er *ResponseError) Error() string {
	return er.ErrorResp.Message
}

func AppError(err string, errType string) *ResponseError {
	return &ResponseError{
		ErrorResp: ErrorResponse{
			Code:    errType,
			Message: err,
		},
	}
}

type MetaResponse struct {
	TraceID string `json:"traceId"`
	Success bool   `json:"success"`
}

type MessagesResponse struct {
	Meta MetaResponse  `json:"meta"`
	Err  ErrorResponse `json:"error"`
}

func ErrorHandler(c *gin.Context) {
	// Execute request handlers and then handle any app_errors
	c.Next()
	errs := c.Errors

	if len(errs) > 0 {
		var err *ResponseError
		ok := errors.As(errs[0].Err, &err)
		if ok {
			meta := utils.NewMetaData(c.Request.Context())

			resp := MessagesResponse{
				Meta: MetaResponse{
					TraceID: meta.TraceID,
				},
				Err: ErrorResponse{
					Code:    err.ErrorResp.Code,
					Message: err.ErrorResp.Message,
				},
			}

			switch err.ErrorResp.Code {

			case StatusOK:
				c.JSON(http.StatusOK, resp)
				return
			case StatusBadRequest:
				c.JSON(http.StatusBadRequest, resp)
				return
			case StatusUnauthorized:
				c.JSON(http.StatusUnauthorized, resp)
				return
			case StatusForbidden:
				c.JSON(http.StatusForbidden, resp)
				return
			case StatusNotFound:
				c.JSON(http.StatusNotFound, resp)
				return
			case StatusConflict:
				c.JSON(http.StatusConflict, resp)
				return
			case StatusGatewayTimeout:
				c.JSON(http.StatusGatewayTimeout, resp)
				return
			case StatusTooManyRequests:
				c.JSON(http.StatusTooManyRequests, resp)
				return
			case StatusCreated:
				c.JSON(http.StatusCreated, resp)
				return
			case StatusInternalServerError:
				c.JSON(http.StatusInternalServerError, resp)
				return
			case StatusValidationError:
				c.JSON(http.StatusBadRequest, resp)
			default:
				c.JSON(http.StatusInternalServerError, resp)
				return
			}
		}
		return
	}
}
