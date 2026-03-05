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

func CoachJoinClubs(c *gin.Context) {
	var req forms.CoachJoinClubsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}

	err := dao.NewClubsCoachesDao(c, global.DB).CoachJoinClubs(c, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "加入俱乐部成功")
	return
}

func CoachQuitClubs(c *gin.Context) {
	var req forms.CoachQuitClubsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}
	err := dao.NewClubsCoachesDao(c, global.DB).CoachQuitClubs(c, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "退出俱乐部成功")
	return
}

func CoachClubsList(c *gin.Context) {
	var req forms.CoachClubsListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}
	list, err := dao.NewClubsCoachesDao(c, global.DB).CoachClubsList(c, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, gin.H{
		"list": list,
	})
	return
}
func ClubsCoachList(c *gin.Context) {
	clubId := c.GetString("club_id")
	if clubId == "" {
		response.Err(c, errors.New("请先申请成为俱乐部成员"))
		return
	}
	var req forms.ClubsCoachListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	list, err := dao.NewClubsCoachesDao(c, global.DB).ClubsCoachList(c, clubId, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, gin.H{
		"list": list,
	})
	return
}
