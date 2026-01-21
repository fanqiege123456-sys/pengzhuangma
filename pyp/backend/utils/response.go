package utils

import (
	"crypto/rand"
	"fmt"
	"time"
)

// 响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 成功响应
func Success(data interface{}) Response {
	return Response{
		Code: SuccessCode,
		Msg:  GetErrorMessage(SuccessCode),
		Data: data,
	}
}

// SuccessWithMsg 带自定义消息的成功响应
func SuccessWithMsg(data interface{}, msg string) Response {
	return Response{
		Code: SuccessCode,
		Msg:  msg,
		Data: data,
	}
}

// Error 错误响应（兼容旧版）
func Error(code int, msg string) Response {
	return ErrorWithMsg(code, msg)
}



// 生成订单号
func GenerateOrderNo() string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	return fmt.Sprintf("%d%x", timestamp, randomBytes)
}

// 分页结构
type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type PageData struct {
	List       interface{} `json:"list"`
	Pagination Pagination  `json:"pagination"`
}
