package utils

// 错误码定义
const (
	// 成功状态码
	SuccessCode = 200
	
	// 客户端错误 (4xx)
	BadRequestCode          = 400 // 请求参数错误
	UnauthorizedCode        = 401 // 未授权
	ForbiddenCode           = 403 // 禁止访问
	NotFoundCode            = 404 // 资源不存在
	MethodNotAllowedCode    = 405 // 方法不允许
	RequestTimeoutCode      = 408 // 请求超时
	ConflictCode            = 409 // 资源冲突
	TooManyRequestsCode     = 429 // 请求过于频繁
	
	// 服务器错误 (5xx)
	InternalServerErrorCode = 500 // 服务器内部错误
	BadGatewayCode          = 502 // 网关错误
	ServiceUnavailableCode  = 503 // 服务不可用
	GatewayTimeoutCode      = 504 // 网关超时
	
	// 业务错误 (6xx)
	ValidationErrorCode     = 600 // 验证失败
	DatabaseErrorCode       = 601 // 数据库错误
	RedisErrorCode          = 602 // Redis错误
	EmailErrorCode          = 603 // 邮件发送错误
	WechatErrorCode         = 604 // 微信API错误
	CoinsInsufficientCode   = 605 // 碰撞币不足
	MatchExpiredCode        = 606 // 匹配已过期
	NotAllowForceAddCode    = 607 // 不允许强制添加
	FriendExistsCode        = 608 // 好友已存在
)

// 错误信息映射
var ErrorMessages = map[int]string{
	// 成功
	SuccessCode: "操作成功",
	
	// 客户端错误
	BadRequestCode:          "请求参数错误",
	UnauthorizedCode:        "未授权，请登录",
	ForbiddenCode:           "禁止访问，权限不足",
	NotFoundCode:            "资源不存在",
	MethodNotAllowedCode:    "请求方法不允许",
	RequestTimeoutCode:      "请求超时",
	ConflictCode:            "资源冲突",
	TooManyRequestsCode:     "请求过于频繁，请稍后重试",
	
	// 服务器错误
	InternalServerErrorCode: "服务器内部错误",
	BadGatewayCode:          "网关错误",
	ServiceUnavailableCode:  "服务不可用",
	GatewayTimeoutCode:      "网关超时",
	
	// 业务错误
	ValidationErrorCode:     "验证失败",
	DatabaseErrorCode:       "数据库操作失败",
	RedisErrorCode:          "Redis操作失败",
	EmailErrorCode:          "邮件发送失败",
	WechatErrorCode:         "微信API调用失败",
	CoinsInsufficientCode:   "碰撞币不足",
	MatchExpiredCode:        "匹配已过期",
	NotAllowForceAddCode:    "对方不允许强制添加",
	FriendExistsCode:        "好友已存在",
}

// GetErrorMessage 获取错误信息
func GetErrorMessage(code int) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// ErrorWithMsg 带自定义错误信息的错误响应
func ErrorWithMsg(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

// ErrorWithDefaultMsg 使用默认错误信息的错误响应
func ErrorWithDefaultMsg(code int) Response {
	return Response{
		Code: code,
		Msg:  GetErrorMessage(code),
		Data: nil,
	}
}