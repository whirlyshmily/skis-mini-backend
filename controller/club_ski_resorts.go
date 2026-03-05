package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"skis-admin-backend/dao"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
)

func ClubGetSkiResorts(c *gin.Context) {
	clubId := c.Param("club_id")
	if clubId == "" {
		response.Err(c, errors.New("俱乐部ID不能为空"))
		return
	}
	var req forms.CoachGetSkiResortsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程记录列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	NewCoachSkiResortsDao := dao.NewClubSkiResortsDao(c, global.DB).ClubGetSkiResorts(c, clubId, req.OrderCourseId)
	response.Success(c, NewCoachSkiResortsDao)
	return
}

func ClubGetSkiResortDate(c *gin.Context) {
	clubId := c.Param("club_id")
	if clubId == "" {
		response.Err(c, errors.New("俱乐部ID不能为空"))
		return
	}
	var req forms.CoachGetSkiResortDateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程记录列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	NewCoachSkiResortsDao := dao.NewClubSkiResortsDao(c, global.DB).ClubGetSkiResortDate(c, clubId, &req)
	response.Success(c, NewCoachSkiResortsDao)
	return
}

func ClubGetSkiResortTime(c *gin.Context) {
	clubId := c.Param("club_id")
	if clubId == "" {
		response.Err(c, errors.New("俱乐部ID不能为空"))
		return
	}
	var req forms.CoachGetSkiResortTimeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程记录列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	NewCoachSkiResortsDao := dao.NewClubSkiResortsDao(c, global.DB).ClubGetSkiResortTime(c, clubId, &req)
	response.Success(c, NewCoachSkiResortsDao)
	return
}

func ClubSkiResortTeachDateList(c *gin.Context) {
	var req forms.ClubSkiResortTeachDateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	teachTimes, err := dao.NewClubSkiResortsDao(c, global.DB).QuerySkiResortTeachDateList(c, &req)
	if err != nil {
		global.Lg.Error("查询俱乐部滑雪场教学时间失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, teachTimes)
	return
}
