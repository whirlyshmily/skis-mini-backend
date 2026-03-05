package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"skis-admin-backend/dao"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
)

func QueryCoachesLevelsList(ctx *gin.Context) {
	levels, err := dao.QueryCoachesLevelsList()
	if err != nil {
		global.Lg.Error("查询课程等级列表失败", zap.Error(err))
		response.Err(ctx, err)
		return
	}
	response.Success(ctx, levels)
	return
}
