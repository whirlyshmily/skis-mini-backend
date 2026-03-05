package controller

import (
	"path"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
	"skis-admin-backend/services"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetOssSign(c *gin.Context) {
	key := c.Query("key")

	key = path.Join("mini", time.Now().Format("20060102"), key)
	sign, err := services.GetOssSign(c, key)
	if err != nil {
		global.Lg.Error("get oss sign error", zap.Error(err))
		response.Err(c, err)
		return
	}

	m := make(map[string]interface{})
	m["sign"] = sign
	response.Success(c, m)
	return
}
