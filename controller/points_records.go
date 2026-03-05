package controller

import (
	"skis-admin-backend/dao"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func QueryPointsRecordsList(c *gin.Context) {
	uid := c.GetString("uid")
	var req forms.QueryPointsRecordsListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.Lg.Error("QueryPointsRecordsList bind params error", zap.Error(err))
		response.Err(c, err)
		return
	}
	total, list, err := dao.QueryPointsRecordsList(uid, &req)
	if err != nil {
		global.Lg.Error("QueryPointsRecordsList error", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, &forms.QueryPointsRecordsListResponse{
		List:  list,
		Total: total,
	})

	return
}

func QueryPointsRecordInfo(c *gin.Context) {
	uid := c.GetString("uid")
	pointId := c.Param("point_id")

	record, err := dao.QueryPointsRecordInfo(uid, pointId)
	if err != nil {
		global.Lg.Error("QueryUserPointsRecordInfo error", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, record)
	return
}
