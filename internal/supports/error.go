package supports

// 错误编号递增： 0-9999，开发在这里自行定义当前服务错误码
const (
	// UsernamePwdError 账号/密码错误
	UsernamePwdError = iota + 1
	// MissingNameError 用户名缺失
	MissingNameError
	// MissingUsernameError 账号缺失
	MissingUsernameError
	// MissingPwdError 密码缺失
	MissingPwdError
	// MissingPhone 电话缺失
	MissingPhoneError
	// MissingPKError 主键Key缺失
	MissingPKError
	// InvalidCaptchaError 验证码错误
	InvalidCaptchaError
	// MissingTokenError token为空
	MissingTokenError
	// SamePasswordError 新旧密码不能相同
	SamePasswordError
	// DeleteSelfError 个人无法删除自己信息
	DeleteSelfError
)

// 错误编号字典
var errDic = map[int]string{
	UsernamePwdError:     "账号/密码错误",
	MissingNameError:     "用户名缺失",
	MissingUsernameError: "账号缺失",
	MissingPwdError:      "密码缺失",
	MissingPhoneError:    "电话缺失",
	MissingPKError:       "主键缺失",
	InvalidCaptchaError:  "验证码错误",
	MissingTokenError:    "token为空",
	SamePasswordError:    "新旧密码不能相同",
	DeleteSelfError:      "个人无法删除自己信息",
}

// GetErrMsg 获取错误信息
func GetErrMsg(errNum int) (int, string) {
	return errNum, errDic[errNum]
}
