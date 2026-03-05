package controller

import (
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/response"
	"skis-admin-backend/services"
	"skis-admin-backend/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Login(c *gin.Context) {
	var req forms.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("登录参数错误", zap.Error(err))
		response.Err(c, enum.NewErr(enum.ParamErr, "参数错误"))
		return
	}

	// 调用微信接口验证code并获取用户信息
	wechatResp, err := services.VerifyWechatCode(global.Config.UserMiniProgram.AppId, global.Config.UserMiniProgram.Secret, req.Code)
	if err != nil {
		global.Lg.Error("微信登录验证失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.WeChatLoginErr, "微信登录验证失败"))
		return
	}

	// 查询或创建用户
	user, err := dao.GetOrCreateUser(c, wechatResp.OpenID, req.UserInfo, req.ReferralCode)
	if err != nil {
		global.Lg.Error("创建或获取用户失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.DBErr, "登录失败"))
		return
	}

	// 生成token
	token := utils.CreateToken(c, user.Uid, user.OpenId, user.UnionId, model.UserTypeUser)
	global.Lg.Debug("登录成功", zap.String("Uid", user.Uid), zap.String("token", token))

	response.Success(c, &forms.LoginResponse{
		Token: token,
		Users: user,
	})

	return
}

func UserActive(c *gin.Context) {
	// 更新用户活跃状态
	if err := dao.UserActive(c); err != nil {
		global.Lg.Error("更新用户活跃状态失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.DBErr, "更新用户活跃状态失败"))
		return
	}

	response.Success(c, nil)
	return
}

// 写一个获取用户手机号的接口
func UpdateUserPhone(c *gin.Context) {
	uid := c.GetString("uid")
	if uid == "" {
		global.Lg.Error("用户ID不能为空")
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不能为空"))
		return
	}

	global.Lg.Info("uid", zap.Any("uid", uid))

	var req forms.UpdateUserPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("获取手机号参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	// 调用微信接口验证code并获取用户信息
	wechatResp, err := services.GetUserPhoneInfo(c, req.Code)
	if err != nil {
		global.Lg.Error("微信登录验证失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.WeChatLoginErr, "微信登录验证失败"))
		return
	}

	global.Lg.Info("req", zap.Any("req", req), zap.Any("wechatResp", wechatResp))

	// 更新用户手机号
	updatedUser, err := dao.UpdateUserPhone(uid, wechatResp.PhoneInfo.PhoneNumber)
	if err != nil {
		global.Lg.Error("更新用户手机号失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.DBErr, "更新手机号失败"))
		return
	}

	response.Success(c, updatedUser)
	return
}

func QueryUserInfo(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		uid = c.GetString("uid")
	}

	user, err := dao.QueryUserInfo(uid)
	if err != nil {
		global.Lg.Error("查询用户信息失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.DBErr, "查询用户信息失败"))
		return
	}
	response.Success(c, user)
	return
}

func UpdateUserInfo(c *gin.Context) {
	var req forms.UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	user, err := dao.UpdateUserInfo(c, c.GetString("uid"), &req)
	if err != nil {
		global.Lg.Error("更新用户信息失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, user)
	return
}

func AppointmentCourse(c *gin.Context) {
	var req forms.AppointmentCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.NewOrdersCoursesDao(c, global.DB).AppointmentCourse(c, &req)
	if err != nil {
		global.Lg.Error("预约课程失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "预约课程成功")
	return
}

func BeforeCancelAppointmentCourse(c *gin.Context) {
	var req forms.CancelAppointmentCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	resp, err := dao.NewOrdersCoursesDao(c, global.DB).BeforeCancelAppointmentCourse(c, &req)
	if err != nil {
		global.Lg.Error("取消预约课程失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, resp)
	return
}
func CancelAppointmentCourse(c *gin.Context) {
	var req forms.CancelAppointmentCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.NewOrdersCoursesDao(c, global.DB).CancelAppointmentCourse(c, &req)
	if err != nil {
		global.Lg.Error("取消预约课程失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "取消预约课程成功")
	return
}

func VerifyCourse(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	err := dao.UserVerifyCourses(c, orderCourseId)
	if err != nil {
		global.Lg.Error("核销课程失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "验证课程成功")
	return
}

func ReviewTeachTime(c *gin.Context) {
	var req forms.ReviewTeachTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.ReviewTeachTime(c, &req)
	if err != nil {
		global.Lg.Error("确认课程失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "成功")
	return
}

func ReviewCoachTransferOrder(c *gin.Context) {
	var req forms.ReviewCoachTransferOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.ReviewCoachTransferOrder(c, &req)
	if err != nil {
		global.Lg.Error("确认课程失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "成功")
	return
}
