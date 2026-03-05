package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
	"strconv"
)

func CreateFeed(c *gin.Context) {
	var req forms.CreateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	feed, err := dao.CreateFeed(c, &req)
	if err != nil {
		global.Lg.Error("创建动态失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, feed)
	return
}
func QueryFeedsList(c *gin.Context) {
	var req forms.QueryFeedListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, list, err := dao.QueryFeedList("", 0, &req)
	if err != nil {
		global.Lg.Error("查询动态列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, &forms.QueryFeedListResponse{
		List:  list,
		Total: total,
	})
	return
}

func QueryFeedInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, enum.NewErr(enum.ParamErr, "参数错误"))
		return
	}

	feed, err := dao.QueryFeedInfo(id)
	if err != nil {
		global.Lg.Error("查询动态失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	//增加访问量
	dao.FeedAddView(id)

	response.Success(c, feed)
	return
}

func UpdateFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, enum.NewErr(enum.ParamErr, "参数错误"))
		return
	}

	var req forms.UpdateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	feed, err := dao.UpdateFeed(id, c.GetString("user_id"), &req)
	if err != nil {
		global.Lg.Error("更新动态失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, feed)
	return
}

func DeleteFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, enum.NewErr(enum.ParamErr, "参数错误"))
		return
	}

	if err := dao.DeleteFeed(id, c.GetString("user_id")); err != nil {
		global.Lg.Error("删除动态失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, nil)
	return
}

func QueryCoachFeedsList(c *gin.Context) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")

	var req forms.QueryFeedListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, list, err := dao.QueryFeedList(userId, userType, &req) //根据教练ID查询(coachId, &req)
	if err != nil {
		global.Lg.Error("查询教练动态列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, &forms.QueryFeedListResponse{
		List:  list,
		Total: total,
	})
	return
}

func QueryClubFeedsList(c *gin.Context) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")

	var req forms.QueryFeedListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, list, err := dao.QueryFeedList(userId, userType, &req) //根据俱乐部ID查询(clubId, &req)
	if err != nil {
		global.Lg.Error("查询俱乐部动态列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, &forms.QueryFeedListResponse{
		List:  list,
		Total: total,
	})
	return
}
