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

func QueryOrdersCoursesRecords(c *gin.Context) {
	orderCourseID := c.Param("order_course_id")
	//uid := c.GetString("uid")
	var req forms.QueryOrdersCoursesRecordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程记录列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	req.OrderCourseId = orderCourseID
	//req.Uid = uid

	total, list, err := dao.QueryOrdersCoursesRecords(&req)
	if err != nil {
		global.Lg.Error("查询订单课程记录列表错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, forms.BaseListResponse{
		List:  list,
		Total: total,
	})
	return
}

func QueryUserOrdersCoursesRecords(c *gin.Context) {
	uid := c.Param("uid") //优先取参数中的uid
	if uid == "" {
		uid = c.GetString("uid") //如果参数中没有uid，再取登录用户的uid
	}

	var req forms.QueryOrdersCoursesRecordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程记录列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	req.Uid = uid

	total, list, err := dao.QueryOrdersCoursesRecords(&req)
	if err != nil {
		global.Lg.Error("查询订单课程记录列表错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, forms.BaseListResponse{
		List:  list,
		Total: total,
	})
	return
}

func CreateOrdersCoursesRecord(c *gin.Context) {
	coachId := c.GetString("coach_id")
	orderCourseId := c.Param("order_course_id")
	var req forms.CreateOrdersCoursesRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("创建订单课程记录参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.CreateOrdersCoursesRecord(c, coachId, orderCourseId, &req)
	if err != nil {
		global.Lg.Error("创建订单课程记录错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}

func DeleteOrdersCoursesRecord(c *gin.Context) {
	uid := c.GetString("uid")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("更新订单课程记录参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err = dao.DeleteOrdersCoursesRecord(c, id, uid)
	if err != nil {
		global.Lg.Error("删除订单课程记录错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}
