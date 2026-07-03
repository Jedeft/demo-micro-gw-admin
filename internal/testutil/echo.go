package testutil

import (
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

// NewJSONContext 构造一个携带 JSON body 的 echo.Context，并设置 path。
// target 为请求 URL，path 为 echo 路由路径，body 为 JSON 字符串。
func NewJSONContext(method, target, path, body string) echo.Context {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath(path)
	return c
}

// NewQueryContext 构造一个携带 query 参数的 echo.Context。
func NewQueryContext(method, target, path string, q url.Values) echo.Context {
	req := httptest.NewRequest(method, target, nil)
	if q != nil {
		req.URL.RawQuery = q.Encode()
	}
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath(path)
	return c
}

// NewRawContext 构造一个不携带 body 的 echo.Context（用于无入参接口或手控请求）。
func NewRawContext(method, target, path string) echo.Context {
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath(path)
	return c
}
