package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"skis-admin-backend/dao"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
)

func QueryCoursesList(c *gin.Context) {
	var req forms.QueryCoursesListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, courses, err := dao.QueryCoursesList(&req)
	if err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, &forms.QueryCoursesListResponse{
		List:  courses,
		Total: total,
	})
	return
}

func QueryCourseInfo(c *gin.Context) {
	course, err := dao.QueryCourseInfo(c.Param("course_id"))
	if err != nil {
		global.Lg.Error("查询课程详情失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, course)
	return
}
