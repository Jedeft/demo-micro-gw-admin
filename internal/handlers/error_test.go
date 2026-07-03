package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jedeft/xuanwu/xerr"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newErrorContext 构造一个带 Debug 开关的 echo.Context，返回 context 与 recorder。
func newErrorContext(method string, debug bool) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Debug = debug
	req := httptest.NewRequest(method, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestErrorHandler_BizErr(t *testing.T) {
	bizErr := xerr.Get().NewErr(1, "biz error msg")
	c, rec := newErrorContext(http.MethodGet, false)

	ErrorHandler(bizErr, c)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp xerr.Err
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, bizErr.Code, resp.Code)
	assert.Equal(t, "biz error msg", resp.Msg)
}

func TestErrorHandler_HTTPError(t *testing.T) {
	httpErr := echo.NewHTTPError(http.StatusBadRequest, "bad request")
	c, rec := newErrorContext(http.MethodGet, false)

	ErrorHandler(httpErr, c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp xerr.Err
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, uint32(http.StatusBadRequest), resp.Code)
	assert.Contains(t, resp.Msg, "bad request")
}

func TestErrorHandler_DefaultDebug(t *testing.T) {
	err := errors.New("some internal error")
	c, rec := newErrorContext(http.MethodGet, true)

	ErrorHandler(err, c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp xerr.Err
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Contains(t, resp.Msg, "some internal error")
}

func TestErrorHandler_DefaultNonDebug(t *testing.T) {
	err := errors.New("some internal error")
	c, rec := newErrorContext(http.MethodGet, false)

	ErrorHandler(err, c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp xerr.Err
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), resp.Msg)
}

func TestErrorHandler_HeadMethod(t *testing.T) {
	err := errors.New("some error")
	e := echo.New()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ErrorHandler(err, c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}

func TestErrorHandler_CommittedResponse(t *testing.T) {
	err := errors.New("some error")
	c, rec := newErrorContext(http.MethodGet, true)
	c.Response().Committed = true

	ErrorHandler(err, c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}
