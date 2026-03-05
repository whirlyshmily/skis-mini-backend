package controller

import (
	"skis-admin-backend/dao"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func QueryOrdersCourseInfo(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	orderCourse, err := dao.QueryOrderCourseInfo(c, "", orderCourseId)
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, orderCourse)
	return
}

func QueryCoachOrderCourses(c *gin.Context) {
	var req forms.QueryOrderCoursesListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, orderCourses, err := dao.NewOrdersCoursesDao(c, global.DB).QueryCoachOrderCourses(c, &req)
	if err != nil {
		global.Lg.Error("查询订单课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, forms.QueryCoachOrderCoursesListResponse{
		Total: total,
		List:  orderCourses,
	})
	return
}
