package controller

import (
	"skis-admin-backend/dao"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateSkiResort(c *gin.Context) {
	var req forms.CreateSkiResortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	skiResort, err := dao.CreateSkiResort(&req)
	if err != nil {
		global.Lg.Error("创建滑雪场失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, skiResort)
	return
}

func QuerySkiResortsList(c *gin.Context) {
	var req forms.QuerySkiResortsListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, skiResorts, err := dao.QuerySkiResortsList(&req)
	if err != nil {
		global.Lg.Error("查询滑雪场列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, &forms.QuerySkiResortsListResponse{
		List:  skiResorts,
		Total: total,
	})
	return
}

func QuerySkiResortInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	skiResort, err := dao.QuerySkiResortInfo(id)
	if err != nil {
		global.Lg.Error("查询滑雪场信息失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, skiResort)
	return
}

func UpdateSkiResort(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	var req forms.UpdateSkiResortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	skiResort, err := dao.UpdateSkiResort(id, &req)
	if err != nil {
		global.Lg.Error("更新滑雪场失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, skiResort)
	return
}

func DeleteSkiResort(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err = dao.DeleteSkiResort(id)
	if err != nil {
		global.Lg.Error("删除滑雪场失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}

func QuerySkiResortTeachTimeList(c *gin.Context) {
	var req forms.QuerySkiResortTeachTimeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	teachTimes, err := dao.QuerySkiResortTeachTimeList(c, &req)
	if err != nil {
		global.Lg.Error("查询滑雪场教学时间失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, teachTimes)
	return
}

func QuerySkiResortTeachDateList(c *gin.Context) {
	var req forms.QuerySkiResortTeachDateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	teachTimes, err := dao.QuerySkiResortTeachDateList(c, &req)
	if err != nil {
		global.Lg.Error("查询滑雪场教学时间失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, teachTimes)
	return
}

// CreateSkiResortTeachTime 创建滑雪场教学时间
// @Summary 创建滑雪场教学时间
// @Description 接收JSON请求创建滑雪场教学时间记录
// @Tags 滑雪场管理
// @Accept json
// @Produce json
// @Param request body forms.CreateSkiResortTeachTimeRequest true "教学时间请求参数"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "创建失败"
// @Router /ski-resort/teach-time [post]
func CreateSkiResortTeachTime(c *gin.Context) {
	var req forms.CreateSkiResortTeachTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	if err := dao.CreateSkiResortTeachTime(c, &req); err != nil {
		global.Lg.Error("创建滑雪场教学时间失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}

func UpdateSkiResortTeachState(c *gin.Context) {
	var req forms.UpdateSkiResortTeachStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	if err := dao.UpdateSkiResortTeachState(c, &req); err != nil {
		global.Lg.Error("修改滑雪场教学状态失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}

func DeleteSkiResortTeachTime(c *gin.Context) {
	var req forms.DeleteSkiResortTeachTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	if err := dao.DeleteSkiResortTeachTime(c, &req); err != nil {
		global.Lg.Error("删除滑雪场教学时间失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}
func ScheduleEvent(c *gin.Context) {
	var req forms.ScheduleEventRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	userId := c.GetString("user_id")
	info, err := dao.ScheduleEvent(c, req.TeachDate, userId, req.SkiResortsId)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, info)
	return
}

func QuitClub(c *gin.Context) {
	userId := c.GetString("user_id")
	var data []*model.SkiResortsTeachTime
	//global.DB.Where("user_id = ? and teach_start_time >?", userId, time.Now().Format("2006-01-02 15:04:05")).Find(&data)
	global.DB.Select("date(created_at) as date").Where("user_id = ? and date(created_at) =?", userId, "2025-09-13").Find(&data)
	//db.Table("orders").Select("date(created_at) as date, sum(amount) as total").Group("date(created_at)").Having("sum(amount) > ?", 100).Scan(&results)

	response.Success(c, data)

	return

}
