package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/response"
	"skis-admin-backend/services"
	"skis-admin-backend/utils"
)

func QueryClubsList(c *gin.Context) {
	var req forms.QueryClubListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, enum.NewErr(enum.ParamErr, "参数错误"))
		return
	}

	total, list, err := dao.QueryClubList(c, &req)
	if err != nil {
		global.Lg.Error("查询俱乐部列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, &forms.QueryClubListResponse{
		Total: total,
		List:  list,
	})

	return
}

func QueryClubInfo(c *gin.Context) {
	clubId := c.Param("club_id")
	if clubId == "" {
		clubId = c.GetString("club_id")
	}
	club, err := dao.QueryClubInfo(clubId)
	if err != nil {
		global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, club)
}

func ClubCheckCoach(c *gin.Context) {
	var req forms.ClubCheckCoachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	id := c.Param("id")
	if id == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeClub {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}

	err := dao.ClubCheckCoach(c, &req)
	if err != nil {
		global.Lg.Error("俱乐部审核教练失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}

func ClubChangeTeachTime(c *gin.Context) {
	var req forms.ClubChangeTeachTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeClub {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}

	err := dao.ClubChangeTeachTime(c, orderCourseId, &req)
	if err != nil {
		global.Lg.Error("俱乐部修改教练上课时间失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "修改上课时间成功，请联系用户确认")
	return
}

func ClubAppointmentCourse(c *gin.Context) {
	var req forms.ClubAppointmentCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeClub {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	err := dao.ClubAppointmentCourse(c, orderCourseId, &req)
	if err != nil {
		global.Lg.Error("俱乐部修改教练上课时间失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "安排成功，请联系教练确认")
	return
}

func ClubReplaceCoachCourse(c *gin.Context) {
	var req forms.ClubReplaceCoachCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "ID不能为空"))
		return
	}
	userType := c.GetInt("user_type")
	if userType != model.UserTypeClub {
		response.Err(c, enum.NewErr(enum.UserTypeError, "用户类型错误"))
		return
	}
	err := dao.ClubReplaceCoachCourse(c, orderCourseId, &req)
	if err != nil {
		global.Lg.Error("俱乐部更改教练失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "更改教练成功，请联系教练确认")
	return
}
func CoachJoinClub(c *gin.Context) {
	var req forms.CoachJoinClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.ClubCoachJoin(c, &req)
	if err != nil {
		global.Lg.Error("教练加入俱乐部失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}

func CoachQuitClub(c *gin.Context) {
	var req forms.CoachQuitClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.ClubCoachQuit(c, &req)
	if err != nil {
		global.Lg.Error("教练退出俱乐部失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}

func UpdateClub(c *gin.Context) {
	clubId := c.Param("club_id")

	var req forms.UpdateClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	club, err := dao.QueryClubInfoByClubId(clubId)
	if err != nil {
		global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	data := make(map[string]interface{})
	if req.Priority != nil {
		data["priority"] = *req.Priority
	}
	if req.Level != nil {
		data["level"] = *req.Level
	}
	if req.FrozenDeposit != nil {
		data["frozen_deposit"] = *req.FrozenDeposit
	}
	if req.Verified != nil {
		data["verified"] = *req.Verified
	}

	if req.Remark != nil {
		data["remark"] = *req.Remark
	}

	if err = dao.UpdateClub(clubId, data); err != nil {
		global.Lg.Error("更新俱乐部失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, club)
}

func DeleteClub(c *gin.Context) {
	clubId := c.Param("club_id")
	if clubId == "" {
		global.Lg.Error("参数错误")
		response.Err(c, enum.NewErr(enum.ParamErr, "俱乐部ID不能为空"))
		return
	}

	club, err := dao.QueryClubInfoByClubId(clubId)
	if err != nil {
		global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	if club.Verified == model.VerifiedVerified {
		global.Lg.Error("该俱乐部已通过审核，不能删除", zap.Any("club", club))
		response.Err(c, enum.NewErr(enum.ParamErr, "该俱乐部已通过审核，不能删除"))
		return
	}

	if err = dao.DeleteClub(clubId); err != nil {
		global.Lg.Error("删除俱乐部失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, nil)
}

func ClubsUserLogin(c *gin.Context) {
	var req forms.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("登录参数错误", zap.Error(err))
		response.Err(c, enum.NewErr(enum.ParamErr, "参数错误"))
		return
	}

	// 调用微信接口验证code并获取用户信息
	wechatResp, err := services.VerifyWechatCode(global.Config.ClubMiniProgram.AppId, global.Config.ClubMiniProgram.Secret, req.Code)
	if err != nil {
		global.Lg.Error("微信登录验证失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.WeChatLoginErr, "微信登录验证失败"))
		return
	}

	// 查询或创建用户
	user, err := dao.GetOrCreateClubsUser(wechatResp.OpenID, req.UserInfo)
	if err != nil {
		global.Lg.Error("创建或获取用户失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.DBErr, "登录失败"))
		return
	}

	// 生成token
	token := utils.CreateToken(c, user.Uid, user.OpenId, user.UnionId, model.UserTypeClub)

	response.Success(c, &forms.ClubsUserLoginResponse{
		Token:      token,
		ClubsUsers: user,
	})
	return
}
func ApplyClub(c *gin.Context) {
	var req forms.ApplyClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	club, err := dao.ApplyClub(c, &req)
	if err != nil {
		global.Lg.Error("申请俱乐部失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, club)
	return
}
func QueryApplyClubInfo(c *gin.Context) {
	uid := c.GetString("uid")
	applyInfo, err := dao.QueryApplyClubInfo(c, uid)
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

func QueryClubOrderCourses(c *gin.Context) {
	var req forms.QueryClubOrderCoursesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询俱乐部订单课程列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, orderCourses, err := dao.NewOrdersCoursesDao(c, global.DB).QueryClubOrderCourses(c, &req)
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

func UpdateClubsInfo(c *gin.Context) {
	var req forms.UpdateClubsInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.UpdateClubsInfo(c, &req)
	if err != nil {
		global.Lg.Error("更新俱乐部信息失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, nil)
	return
}

func QueryClubsCourses(c *gin.Context) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	if userId == "" || userType != model.UserTypeClub {
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不存在"))
		return
	}

	courses, err := dao.QueryClubsCourses(c, userId)
	if err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, courses)
	return
}

func QueryClubMatchCoachesList(c *gin.Context) {
	var req forms.QueryMatchCoachesListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询俱乐部匹配教练列表", zap.Error(err))
		response.Err(c, err)
		return
	}
	coaches, err := dao.QueryClubMatchCoachesList(c, req)
	if err != nil {
		global.Lg.Error("查询俱乐部匹配教练列表", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, map[string]interface{}{
		"list": coaches,
	})
	return
}

func QueryClubComments(c *gin.Context) {
	clubId := c.Param("club_id")
	if clubId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "俱乐部ID不存在"))
		return
	}

	var req forms.ListQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询评论参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, comments, err := dao.QueryUserComments(c, clubId, model.UserTypeClub, &req)
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
