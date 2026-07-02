package handlers

import (
	"github.com/Jedeft/xuanwu/xerr"

	"github.com/Jedeft/demo-micro-gw-admin/internal/supports"
)

// Uint32IDReq Uint32 id入参
type Uint32IDReq struct {
	ID uint32 `json:"id" query:"id"`
}

// StringIDReq string id入参
type StringIDReq struct {
	ID string `json:"id" query:"id"`
}

// RspRow 公共单个资源返回
type RspRow struct {
	Code uint32      `json:"code"`
	Msg  string      `json:"msg"`
	Row  interface{} `json:"row,omitempty"`
}

// RspRows 列表资源返回
type RspRows struct {
	Code  uint32      `json:"code"`
	Msg   string      `json:"msg"`
	Total int64       `json:"total"`
	Rows  interface{} `json:"rows,omitempty"`
}

// Valid 校验
func (r *Uint32IDReq) Valid() error {
	if r.ID == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPKError))
	}
	return nil
}

// Valid 校验
func (r *StringIDReq) Valid() error {
	if len(r.ID) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPKError))
	}
	return nil
}
