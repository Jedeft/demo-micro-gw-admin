package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Jedeft/xuanwu/log"
	"github.com/Jedeft/xuanwu/xerr"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ErrorHandler customize echo's HTTP error handler.
func ErrorHandler(err error, c echo.Context) {
	var (
		httpCode = http.StatusInternalServerError
		code     = uint32(http.StatusInternalServerError)
		msg      string
	)

	var bizErr *xerr.Err
	var httpErr *echo.HTTPError
	switch {
	case errors.As(err, &bizErr):
		// 我们自定的错误
		httpCode = http.StatusOK
		code = bizErr.Code
		msg = bizErr.Msg
		// 错误信息在response结构返回，tracer已收集相关信息
		log.Get().Bg().Error("biz err", zap.Error(err))
	case errors.As(err, &httpErr):
		// echo 框架的错误
		httpCode = httpErr.Code
		code = uint32(httpErr.Code) //nolint:gosec // G115: HTTP 状态码恒在 uint32 范围内
		msg = fmt.Sprintf("%v", httpErr.Message)
	default:
		if c.Echo().Debug {
			// 剩下的都是500 开了debug显示详细错误
			msg = err.Error()
			log.Get().Bg().Error("system err", zap.Error(err))
		} else {
			// 500 不开debug 用标准错误描述 以防泄漏信息
			msg = http.StatusText(httpCode)
			// 标准输出tracer无法捕获详细错误信息，此处进行log 日志上报
			log.Get().For(c.Request().Context()).Error("system err", zap.Error(err))
		}
	}

	// 判断 context 是否已经返回了
	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD {
			err = c.NoContent(httpCode)
		} else {
			err = c.JSON(httpCode, &xerr.Err{
				Code: code,
				Msg:  msg,
			})
		}
		if err != nil {
			c.Logger().Error(err.Error())
		}
	}
}
