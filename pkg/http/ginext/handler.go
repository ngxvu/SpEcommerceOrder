package ginext

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Request struct {
	GinCtx *gin.Context
	ctx    context.Context
}

type Response struct {
	Code   int
	Header http.Header
	*GeneralBody
}

// NewResponse makes a new response with empty body
func NewResponse(code int) *Response {
	return &Response{
		Code:        code,
		GeneralBody: &GeneralBody{},
	}
}

// NewResponseData makes a new response with body data
func NewResponseData(code int, data interface{}) *Response {
	return &Response{
		Code:        code,
		GeneralBody: NewBody(data, nil),
	}
}

// NewResponseWithPager makes a new response with body data & pager
func NewResponseWithPager(code int, data interface{}, pager *Pager) *Response {
	return &Response{
		Code:        code,
		GeneralBody: NewBodyPaginated(data, pager),
	}
}

type Handler func(r *Request) (*Response, error)

// NewRequest creates a new handler request
func NewRequest(c *gin.Context) *Request {
	ctx := FromGinRequestContext(c)
	req := &Request{
		GinCtx: c,
		ctx:    ctx,
	}

	return req
}

func (r *Request) Context() context.Context {
	if r.ctx == nil {
		r.ctx = context.Background()
	}

	return r.ctx
}

func WrapHandler(handler Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			err  error
			resp *Response
		)

		defer func() {
			if err != nil {
				_ = c.Error(err)
				return
			}

			if resp == nil {
				return
			}

			for k, v := range resp.Header {
				for _, v_ := range v {
					c.Header(k, v_)
				}
			}

			if resp.Data != nil || resp.Error != nil {
				c.JSON(resp.Code, resp.GeneralBody)
			} else {
				c.Status(resp.Code)
			}
		}()

		req := NewRequest(c)
		resp, err = handler(req)
	}
}

// MustBind does a binding on v with income request data
// it'll panic if any invalid data (and by design, it should be recovered by error handler middleware)
func (r *Request) MustBind(v interface{}) {
	r.MustNoError(r.GinCtx.ShouldBind(v))
}

func (r *Request) MustBindUri(v interface{}) {
	r.MustNoError(r.GinCtx.ShouldBindUri(v))
}

// MustNoError makes a ASSERT on err variable, panic when it's not nil
// then it must be recovered by WrapHandler
func (r *Request) MustNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func (r *Request) Param(key string) string {
	return r.GinCtx.Param(key)
}

func (r *Request) Query(key string) string {
	return r.GinCtx.Query(key)
}
