package openplatform

import (
	"errors"
	"strings"
)

const (
	CodeSuccess       = 200
	CodeTokenInvalid  = 40001 // token 失效或无效，客户端应重新握手
	CodeQuotaExceeded = 40002 // 可用配额不足
)

var (
	ErrTokenInvalid  = errors.New("invalid token")
	ErrQuotaExceeded = errors.New("quota exceeded")
)

// BuildGatewayErrorResponse 非法/爆破类请求返回虚假成功（空 Data）；
// token 失效、配额不足返回专用错误码。
func BuildGatewayErrorResponse(err error) *GatewayResponse {
	if isTokenInvalid(err) {
		return &GatewayResponse{
			Code:    CodeTokenInvalid,
			Message: "token invalid or expired",
			Data:    nil,
		}
	}
	if isQuotaExceeded(err) {
		return &GatewayResponse{
			Code:    CodeQuotaExceeded,
			Message: "quota exceeded",
			Data:    nil,
		}
	}
	return DecoySuccessResponse()
}

func DecoySuccessResponse() *GatewayResponse {
	return &GatewayResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    nil,
	}
}

func isTokenInvalid(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrTokenInvalid) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return msg == "invalid token" || strings.Contains(msg, "token not found")
}

func isQuotaExceeded(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrQuotaExceeded)
}

func isDecoyResponse(resp *GatewayResponse) bool {
	return resp != nil && resp.Code == CodeSuccess && resp.Data == nil
}
