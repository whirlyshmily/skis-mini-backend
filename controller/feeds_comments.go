package controller

import (
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func QueryFeedsCommentsList(c *gin.Context) {
	feedId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, enum.NewErr(enum.ParamErr, "feed_id 格式错误"))
		return
	}

	var req forms.QueryFeedsCommentsListRequest
	if err = c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	comments, err := dao.QueryFeedsCommentsList(c, feedId, &req)
	if err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, comments)
	return
}

func CreateFeedsComments(c *gin.Context) {
	feedId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, enum.NewErr(enum.ParamErr, "feed_id 格式错误"))
		return
	}

	var req *forms.CreateFeedsCommentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	comment, err := dao.CreateFeedsComment(c, feedId, req)
	if err != nil {
		global.Lg.Error("创建评论失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, comment)
	return
}

func QueryFeedsCommentInfo(c *gin.Context) {
	feedId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, enum.NewErr(enum.ParamErr, "feed_id 格式错误"))
		return
	}

	commentId, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		response.Err(c, enum.NewErr(enum.ParamErr, "id 格式错误"))
		return
	}

	comment, err := dao.QueryFeedsCommentInfo(c, feedId, commentId)
	if err != nil {
		global.Lg.Error("查询评论失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, comment)
	return

}

func UpdateFeedsComments(c *gin.Context) {
	feedId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, enum.NewErr(enum.ParamErr, "feed_id 格式错误"))
		return
	}

	commentId, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		response.Err(c, enum.NewErr(enum.ParamErr, "id 格式错误"))
		return
	}

	var req forms.CreateFeedsCommentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	comment, err := dao.UpdateFeedsComment(c, feedId, commentId, &req)
	if err != nil {
		global.Lg.Error("更新评论失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, comment)
	return
}

func DeleteFeedsComments(c *gin.Context) {
	feedId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, enum.NewErr(enum.ParamErr, "feed_id 格式错误"))
		return
	}

	commentId, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		response.Err(c, enum.NewErr(enum.ParamErr, "id 格式错误"))
		return
	}

	err = dao.DeleteFeedsComment(c, feedId, commentId)
	if err != nil {
		global.Lg.Error("删除评论失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}
