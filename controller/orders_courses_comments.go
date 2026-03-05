package controller

import (
	"skis-admin-backend/dao"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func QueryOrdersCoursesComment(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")

	comments, err := dao.QueryOrdersCoursesComments(c, orderCourseId)
	if err != nil {
		global.Lg.Error("查询订单课程评论失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, comments)
	return
}

func CreateOrdersCoursesComment(c *gin.Context) {
	uid := c.GetString("uid")
	orderCourseId := c.Param("order_course_id")

	var req forms.OrdersCoursesComments
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("绑定JSON失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	comment, err := dao.CreateOrdersCoursesComment(c, uid, orderCourseId, &req)
	if err != nil {
		global.Lg.Error("创建订单课程评论失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, comment)
	return
}

func ReplyOrdersCoursesComment(c *gin.Context) {
	coachId := c.GetString("coach_id")
	orderCourseId := c.Param("order_course_id")
	id, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		global.Lg.Error("绑定ID失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	var req forms.OrdersCoursesComments
	if err = c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("绑定JSON失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	comment, err := dao.CreateOrdersCoursesCommentReply(c, id, orderCourseId, coachId, &req)
	if err != nil {
		global.Lg.Error("回复订单课程评论失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, comment)
	return
}
