package ginext

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type testErrorHandlerSuite struct {
	suite.Suite
	engine *gin.Engine
}

func (ts *testErrorHandlerSuite) SetupSuite() {
	ts.engine = gin.New()
	ts.engine.Use(CreateErrorHandler())

	ts.engine.GET("/no-error", func(c *gin.Context) {
		c.Status(200)
	})

	ts.engine.GET("/panic-string", func(c *gin.Context) {
		panic("test panic string")
	})

	ts.engine.GET("/panic-api-error", func(c *gin.Context) {
		panic(NewError(400, "simple error message"))
	})

	ts.engine.GET("/add-error-stack", func(c *gin.Context) {
		_ = c.Error(NewError(400, "add error to stack"))
	})

	ts.engine.POST("/post-failed-unmarshal", func(c *gin.Context) {
		req := struct {
			Code uint `json:"code"`
		}{}
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(200, nil)
	})
}

func (ts *testErrorHandlerSuite) doRequest(url string, body ...string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", url, nil)
	if strings.HasPrefix(url, "/post-") {
		req = httptest.NewRequest("POST", url, strings.NewReader(body[0]))
	}
	ts.engine.ServeHTTP(w, req)

	return w
}

func (ts *testErrorHandlerSuite) TestErrorHandler() {
	cases := []struct {
		name     string
		path     string
		body     string
		wantCode int
		wantBody string
	}{
		{
			name:     "NoError",
			path:     "/no-error",
			wantCode: 200,
			wantBody: "",
		},
		{
			name:     "PanicString",
			path:     "/panic-string",
			wantCode: 500,
			wantBody: `{"error":{"detail":"unexpected error: test panic string","status":0,"metadata":null}}`,
		},
		{
			name:     "PanicError",
			path:     "/panic-api-error",
			wantCode: 400,
			wantBody: `{"error":{"detail":"simple error message","status":0,"metadata":null}}`,
		},
		{
			name:     "AddErrorStack",
			path:     "/add-error-stack",
			wantCode: 400,
			wantBody: `{"error":{"detail":"add error to stack","status":0,"metadata":null}}`,
		},
		{
			name:     "UnmarshalError",
			path:     "/post-failed-unmarshal",
			body:     `{"code": "1"}`, // post code as string instead of number to make error
			wantCode: 400,
			wantBody: fmt.Sprintf(`{"error":{"code":"invalid type %s, requires %s"}}`, "`string`", "`uint`"),
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			r := ts.doRequest(tc.path, tc.body)
			ts.Equal(tc.wantCode, r.Code)
			ts.Equal(tc.wantBody, r.Body.String())
		})
	}
}

func (ts *testErrorHandlerSuite) TestPushErrorStack() {
}

func TestErrorHandler(t *testing.T) {
	suite.Run(t, new(testErrorHandlerSuite))
}
