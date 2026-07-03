package supports

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetErrMsg(t *testing.T) {
	tests := []struct {
		name    string
		errNum  int
		wantN   int
		wantMsg string
	}{
		{name: "UsernamePwdError", errNum: UsernamePwdError, wantN: UsernamePwdError, wantMsg: "账号/密码错误"},
		{name: "MissingNameError", errNum: MissingNameError, wantN: MissingNameError, wantMsg: "用户名缺失"},
		{name: "MissingUsernameError", errNum: MissingUsernameError, wantN: MissingUsernameError, wantMsg: "账号缺失"},
		{name: "MissingPwdError", errNum: MissingPwdError, wantN: MissingPwdError, wantMsg: "密码缺失"},
		{name: "MissingPhoneError", errNum: MissingPhoneError, wantN: MissingPhoneError, wantMsg: "电话缺失"},
		{name: "MissingPKError", errNum: MissingPKError, wantN: MissingPKError, wantMsg: "主键缺失"},
		{name: "InvalidCaptchaError", errNum: InvalidCaptchaError, wantN: InvalidCaptchaError, wantMsg: "验证码错误"},
		{name: "MissingTokenError", errNum: MissingTokenError, wantN: MissingTokenError, wantMsg: "token为空"},
		{name: "SamePasswordError", errNum: SamePasswordError, wantN: SamePasswordError, wantMsg: "新旧密码不能相同"},
		{name: "DeleteSelfError", errNum: DeleteSelfError, wantN: DeleteSelfError, wantMsg: "个人无法删除自己信息"},
		{name: "unknown error code", errNum: 9999, wantN: 9999, wantMsg: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN, gotMsg := GetErrMsg(tt.errNum)
			assert.Equal(t, tt.wantN, gotN)
			assert.Equal(t, tt.wantMsg, gotMsg)
		})
	}
}
