package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/response"
	"strconv"
)

func QueryCoachesList(c *gin.Context) {
	var req forms.QueryCoachesListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	coaches, err := dao.QueryCoachesList(c, req)
	if err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, map[string]interface{}{
		"list": coaches,
	})
	return
}
func QueryMatchCoachesList(c *gin.Context) {
	var req forms.QueryMatchCoachesListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	coaches, err := dao.QueryMatchCoachesList(c, req)
	if err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, map[string]interface{}{
		"list": coaches,
	})
	return
}

func QueryCoachInfo(c *gin.Context) {
	coachId := c.Param("coach_id")
	coach, err := dao.QueryCoachInfo(coachId)
	if err != nil {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, coach)
	return
}

func EditCoachInfo(c *gin.Context) {
	var req forms.EditCoachInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.EditCoachInfo(c, req)
	if err != nil {
		global.Lg.Error("更新教练失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "更新教练信息成功")
	return
}
func QueryLoginCoachInfo(c *gin.Context) {
	coachId := c.GetString("coach_id")
	coach1, err := dao.QueryLoginCoachInfo(coachId)
	if err != nil {
		global.Lg.Error("QueryCoachInfo 查询教练信息失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, coach1)
	return
}

func CoachRemoveTag(c *gin.Context) {
	var req forms.CoachRemoveTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}

	err := dao.CoachRemoveTag(c, req)
	if err != nil {
		global.Lg.Error("CoachRemoveTag CoachRemoveTag 教练技能课程还存在", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "删除教练技能成功")
	return
}

func CoachAddTagReview(c *gin.Context) {
	var req forms.CoachAddTagReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}
	err := dao.NewCoachesTagsReviewDao(c, global.DB).CoachAddTagReview(c, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "提交教练技能申请成功")
	return
}

func CoachGetAllTags(c *gin.Context) {
	coachId := c.GetString("coach_id")
	if coachId == "" {
		response.Err(c, errors.New("请先申请成为教练"))
	}
	verified, err := strconv.Atoi(c.Query("verified"))
	if err != nil {
		response.Err(c, err)
		return
	}
	coachTags, err := dao.QueryCoachAllTags(coachId, verified)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, coachTags)
	return
}

func CoachGetTagReview(c *gin.Context) {
	var req forms.CoachGetTagReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}
	items, err := dao.NewCoachesTagsReviewDao(c, global.DB).CoachGetTagReview(c, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, items)
	return
}

func CoachEditSkiResorts(c *gin.Context) {
	var req forms.CoachEditSkiResortsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("修改教练滑雪场失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	NewCoachSkiResortsDao := dao.NewCoachSkiResortsDao(c, global.DB).CoachEditSkiResorts(c, req.SkiResortsIDs)
	if NewCoachSkiResortsDao != nil {
		global.Lg.Error("修改教练滑雪场失败", zap.Error(NewCoachSkiResortsDao))
		response.Err(c, NewCoachSkiResortsDao)
		return
	}
	response.Success(c, "修改教练滑雪场成功")
	return
}

func CoachGetSkiResorts(c *gin.Context) {
	coachId := c.Param("coach_id")
	if coachId == "" {
		response.Err(c, errors.New("教练ID不能为空"))
		return
	}
	var req forms.CoachGetSkiResortsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程记录列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	NewCoachSkiResortsDao := dao.NewCoachSkiResortsDao(c, global.DB).CoachGetSkiResorts(c, coachId, req.OrderCourseId)
	response.Success(c, NewCoachSkiResortsDao)
	return
}

func CoachGetSkiResortDate(c *gin.Context) {
	coachId := c.Param("coach_id")
	if coachId == "" {
		response.Err(c, errors.New("教练ID不能为空"))
		return
	}
	var req forms.CoachGetSkiResortDateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程记录列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	NewCoachSkiResortsDao := dao.NewCoachSkiResortsDao(c, global.DB).CoachGetSkiResortDate(c, coachId, &req)
	response.Success(c, NewCoachSkiResortsDao)
	return
}

func CoachGetSkiResortTime(c *gin.Context) {
	coachId := c.Param("coach_id")
	if coachId == "" {
		response.Err(c, errors.New("教练ID不能为空"))
		return
	}
	var req forms.CoachGetSkiResortTimeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程记录列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	NewCoachSkiResortsDao := dao.NewCoachSkiResortsDao(c, global.DB).CoachGetSkiResortTime(c, coachId, &req)
	response.Success(c, NewCoachSkiResortsDao)
	return
}

func CoachConfirmOrderCourses(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	var req forms.CoachConfirmOrderCoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("确认课程参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.CoachConfirmOrderCourses(c, orderCourseId, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "确认课程成功")
	return
}
func BeforeCoachChangeOrderCourseTime(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	data := dao.BeforeCoachChangeOrderCourseTime(c, orderCourseId)
	response.Success(c, data)
	return
}
func CoachChangeOrderCourseTime(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	var req forms.CoachChangeOrderCourseTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("确认课程参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.CoachChangeOrderCourseTime(c, orderCourseId, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "修改课程时间成功")
	return
}
func BeforeCoachCancelOrderCourses(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}

	resp, err := dao.BeforeCoachCancelOrderCourses(c, orderCourseId)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, resp)
	return
}
func CoachCancelOrderCourses(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	err := dao.CoachCancelOrderCourses(c, orderCourseId)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "取消课程成功")
	return
}

func CoachTransferOrderCourses(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	err := dao.CoachTransferOrderCourses(c, orderCourseId)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "转单申请提交成功，请联系用户确认")
	return
}
func CoachCancelTransferOrderCourses(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	err := dao.CoachCancelTransferOrderCourses(c, orderCourseId)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "取消转单申请成功")
	return

}
func CoachTransferOrderToCoach(c *gin.Context) {
	var req forms.CoachTransferOrderToCoachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("确认课程参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.CoachTransferOrderToCoach(c, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "已发起转单申请，请联系教练确认")
	return
}

func CoachReviewOrderFromCoach(c *gin.Context) {
	var req forms.CoachReviewOrderFromCoachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}
	err := dao.CoachReviewOrderFromCoach(c, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "转单审核成功，请联系教练确认")
	return
}

func CoachReviewOrderFromClub(c *gin.Context) {
	var req forms.CoachReviewOrderFromClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}
	err := dao.CoachReviewOrderFromClub(c, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "审核成功")
	return
}

func CoachReviewReplaceFromClub(c *gin.Context) {
	var req forms.CoachReviewReplaceFromClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}
	err := dao.CoachReviewReplaceFromClub(c, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "转单审核成功")
	return
}

func CoachApplyTransferOrders(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	err := dao.CoachApplyTransferOrders(c, orderCourseId)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "转单申请成功，请联系俱乐部确认")
	return
}
func CoachCancelApplyTransferOrders(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	err := dao.CoachCancelApplyTransferOrders(c, orderCourseId)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "取消转单申请成功")
	return

}

func CoachVerifyCourses(c *gin.Context) {
	checkCode := c.Param("check_code")
	if checkCode == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "核销码不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	err := dao.CoachVerifyCourses(c, checkCode)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "核销成功")
	return
}

func QueryApplyInfo(c *gin.Context) {
	uid := c.GetString("uid")
	applyInfo, err := dao.QueryApplyInfo(c, uid)
	if err == gorm.ErrRecordNotFound {
		response.Success(c, applyInfo)
		return
	}
	if err != nil {
		response.Err(c, err)
		return
	}

	response.Success(c, applyInfo)
	return
}

func ApplyCoach(c *gin.Context) {
	uid := c.GetString("uid")
	if uid == "" {
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不存在"))
		return
	}

	var req forms.CreateCoachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("创建教练参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	coach, err := dao.ApplyCoach(c, uid, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, coach)
	return

}

/*
	管理台转单给教练接口，整个订单转过去，需要修改的地方：

0.只能未预约的订单转单
1.订单表的user_id和user_type
2.插入转单记录
*/
func AdminTransferOrderToCoach(c *gin.Context) {
	var req forms.AdminTransferOrderToCoachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("确认课程参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.AdminTransferOrderToCoach(c, c.Param("order_id"), &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "已发起转单申请，请联系教练确认")
	return
}

func AdminMatchCoachesList(c *gin.Context) {
	coaches, err := dao.AdminQueryMatchCoachesList(c, c.Param("order_id"))
	if err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, map[string]interface{}{
		"list": coaches,
	})
	return
}

func QueryCoachReferralRecords(c *gin.Context) {
	coachId := c.GetString("coach_id")
	if coachId == "" {
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不存在"))
		return
	}
	records, err := dao.QueryReferralRecordsByReferralUserId(c, coachId, model.UserTypeCoach)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, records)
	return
}

func QueryCoachCourses(c *gin.Context) {
	coachId := c.GetString("coach_id")
	if coachId == "" {
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不存在"))
		return
	}

	courses, err := dao.QueryCoachCourses(c, coachId)
	if err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, courses)
	return
}

func QueryCoachComments(c *gin.Context) {
	coachId := c.Param("coach_id")
	if coachId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "教练ID不存在"))
		return
	}

	var req forms.ListQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询评论参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, comments, err := dao.QueryUserComments(c, coachId, model.UserTypeCoach, &req)
	if err != nil {
		global.Lg.Error("查询评论列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, map[string]interface{}{
		"total": total,
		"list":  comments,
	})
	return
}
