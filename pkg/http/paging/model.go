package paging

import (
	"context"
	"time"
)

type Param struct {
	Page       *int       `json:"page" form:"page"`
	PageSize   *int       `json:"page_size" form:"page_size"`
	CreateFrom *time.Time `json:"create_from" form:"create_from"`
	CreateTo   *time.Time `json:"create_to" form:"create_to"`
	Sort       *string    `json:"sort" form:"sort"`
	Search     *string    `json:"search" form:"search"`
}

type Filter struct {
	Param
	Pager *Pager
}

type BodyMeta struct {
	TraceID   string `json:"traceId"`
	Success   bool   `json:"success"`
	TotalRows int64  `json:"total"`
	Page      int    `json:"page"`
	PageSize  int    `json:"size"`
	PageCount int    `json:"pageCount"`
	CanNext   bool   `json:"canNext"`
	CanPre    bool   `json:"canPre"`
}

// GeneralBody defines a general response body
type GeneralBody struct {
	Meta  BodyMeta    `json:"meta,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
}

func NewBodyPaginated(ctx context.Context, data interface{}, pager *Pager) *GeneralBody {
	requestID, ok := ctx.Value("x-request-id").(string)
	if !ok {
		return nil
	}

	return &GeneralBody{
		Data: data,
		Meta: BodyMeta{
			TraceID:   requestID,
			Success:   true,
			TotalRows: pager.TotalRows,
			Page:      pager.GetPage(),
			PageSize:  pager.GetPageSize(),
			PageCount: pager.GetTotalPages(),
			CanNext:   pager.CanNext(),
			CanPre:    pager.CanPre(),
		},
	}
}
