package utils

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SendErrResponse 发送错误响应
func SendErrResponse(ctx context.Context, c *app.RequestContext, code int, err error) {
	resp := Response{
		Code:    code,
		Message: err.Error(),
	}

	hlog.CtxErrorf(ctx, "Error response: %v", err)
	c.JSON(http.StatusOK, resp)
}

// SendSuccessResponse 发送成功响应
func SendSuccessResponse(ctx context.Context, c *app.RequestContext, code int, data interface{}) {
	resp := Response{
		Code:    0, // 0表示成功
		Message: "success",
		Data:    data,
	}

	c.JSON(http.StatusOK, resp)
}

// SendCustomResponse 发送自定义响应
func SendCustomResponse(ctx context.Context, c *app.RequestContext, httpCode int, respCode int, message string, data interface{}) {
	resp := Response{
		Code:    respCode,
		Message: message,
		Data:    data,
	}

	c.JSON(httpCode, resp)
}
