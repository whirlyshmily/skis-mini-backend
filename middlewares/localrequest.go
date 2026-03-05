package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"skis-admin-backend/enum"
	"skis-admin-backend/response"
	"strings"
)

func LocalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := GetClientIP(c.Request)

		// 判断是否为 localhost 或 loopback 地址
		if !isLocalhost(clientIP) {
			response.Err(c, enum.NewErr(enum.BadRequestErr, "非 localhost 或 loopback 地址，不允许访问"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetClientIP(r *http.Request) string {
	// 优先检查 X-Forwarded-For 头部
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	// 检查 X-Real-Ip 头部
	realIP := r.Header.Get("X-Real-Ip")
	if realIP != "" {
		return realIP
	}

	// 从 RemoteAddr 提取 IP，并去掉端口
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

func isLocalhost(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1" || ip == "[::1]" || strings.EqualFold(ip, "localhost")
}
