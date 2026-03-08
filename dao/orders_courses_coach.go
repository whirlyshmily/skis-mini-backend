package dao

import (
	"context"
	"encoding/json"
	"math"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CoachConfirmOrderCourses(c *gin.Context, orderCourseId string, req *forms.CoachConfirmOrderCoursRequest) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	coachId := c.GetString("coach_id")
	order, orderCourse, err := GetOrderCourses(orderCourseId)
	if err != nil {
		return err
	}
	if order.UserID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "只能确认自己的订单")
	}
	if orderCourse.TeachState == model.TeachStateWaitAppointment {
		return enum.NewErr(enum.OrdersCoursesExitErr, "用户还没预约该课程")
	}

	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmUser {
		return enum.NewErr(enum.OrdersCoursesExitErr, "该课程已确认")
	}

	var bufferTimeIds []int64 //查出缓冲时间对应的教练的课程时间表ID
	bufferTimeCount := 0      //缓冲时间有几个30分钟
	if req.BufferTime != 0 {  //设置了缓冲时间
		bufferTimeCount = int(math.Ceil(float64(req.BufferTime) / 30))
		endTimeData := model.SkiResortsTeachTime{} //查出预约的结束时间
		err = global.DB.Model(&model.SkiResortsTeachTime{}).Where("id in ?", []int64(orderCourse.TeachTimeIDs)).
			Order("teach_end_time desc").Take(&endTimeData).Error
		if err != nil {
			err = enum.NewErr(enum.OrdersCoursesExitErr, "课程时间不存在")
			return
		}
		var bufferStartTimes []model.LocalTime
		buffeStartTi := time.Time(endTimeData.TeachEndTime)
		for i := 0; i < bufferTimeCount; i++ {
			bufferStartTimes = append(bufferStartTimes, model.LocalTime(buffeStartTi))
			buffeStartTi = buffeStartTi.Add(30 * time.Minute)
		}
		global.DB.Model(&model.SkiResortsTeachTime{}).
			Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0", userId, userType, orderCourse.SkiResortsID, bufferStartTimes, model.SkiTeachStateWaitAppointment).
			Pluck("id", &bufferTimeIds)
		if len(bufferTimeIds) != bufferTimeCount { //缓冲时间冲突
			err = enum.NewErr(enum.OrdersCoursesExitErr, "课后缓冲时间冲突")
			return
		}
	}
	isChangeStartTime, changeTimeIds, changeBufferTimeIds, startTi, err := getChangeStartTime(c, userId, req.TeachStartTime, userType, orderCourse, req.BufferTime)
	if err != nil {
		return err
	}
	insrtocsData := model.OrdersCoursesState{
		OrderCourseID: orderCourseId,
		UserID:        userId,
		UserType:      userType,
		Operate:       model.OperateCoachConfirmCourse,
		Remark:        model.OCSOperateStr[model.OperateCoachConfirmCourse],
		CoachID:       coachId,
		Process:       model.ProcessYes,
	}
	upOrdersCoursesData := map[string]interface{}{
		"teach_state":    model.TeachStateWaitClass,
		"teach_coach_id": coachId,
	}
	if isChangeStartTime { //教练要更改上课时间
		upOrdersCoursesData = map[string]interface{}{
			"teach_state":    model.TeachStateWaitUserConfirmCoachTime,
			"teach_coach_id": coachId,
		}
		insrtocsData = model.OrdersCoursesState{
			OrderCourseID:  orderCourseId,
			UserID:         userId,
			UserType:       userType,
			Operate:        model.OperateCoachChangeCourse,
			Remark:         model.OCSOperateStr[model.OperateCoachChangeCourse],
			CoachID:        coachId,
			TeachTimeIDs:   append(changeTimeIds, changeBufferTimeIds...),
			TeachStartTime: model.LocalTime(startTi),
			Process:        model.ProcessNo,
		}
	}

	if len(bufferTimeIds) > 0 { // 用户预约的时间加课后缓冲时间
		idsStr, _ := json.Marshal(append(orderCourse.TeachTimeIDs, bufferTimeIds...))
		upOrdersCoursesData["teach_time_ids"] = idsStr
		upOrdersCoursesData["teach_buffer_time"] = req.BufferTime
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		var timeIds []int64
		if len(bufferTimeIds) > 0 { //设置缓冲时间
			timeIds = append(bufferTimeIds, changeBufferTimeIds...)
			err = tx.Model(model.SkiResortsTeachTime{}).Where("id in ?", timeIds).
				Updates(map[string]interface{}{
					"teach_state": model.SkiTeachStateAfterClass,
				}).Error
		}

		if len(changeTimeIds) > 0 { //教练修改了上课时间，修改雪场对应的教学时间
			timeIds = append(timeIds, changeTimeIds...)
		}
		if len(timeIds) > 0 {
			err = SRTOrderCourses(tx, timeIds, orderCourseId, 0)
			if err != nil {
				global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
				return err
			}
		}

		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourseId).
			Updates(upOrdersCoursesData).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "确认课程失败")
		}

		err = tx.Model(model.OrdersCoursesState{}).Create(&insrtocsData).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "确认课程失败2")
		}

		err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate = ? and process = ?",
			orderCourseId, model.OperateUserAppointment, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		return err
	})
	if err != nil {
		return err
	}
	return nil
}
func BeforeCoachChangeOrderCourseTime(c *gin.Context, orderCourseId string) (resp *forms.CoachCanChangeTimeResp) {
	userId := c.GetString("user_id")
	order, orderCourse, err := GetOrderCourses(orderCourseId)
	resp = &forms.CoachCanChangeTimeResp{}
	resp.CanChangeTime = false

	if err != nil {
		resp.Reason = "课程不存在"
		return
	}
	if orderCourse.TeachState != model.TeachStateWaitClass && orderCourse.TeachState != model.TeachStateWaitCoachConfirmUser {
		resp.Reason = "待上课和待确认的课程才能修改上课时间"
		return
	}
	if order.UserType == model.UserTypeClub {
		resp.Reason = "俱乐部的课程只能联系俱乐部修改上课时间"
		return
	}
	if orderCourse.TeachCoachID == "" {
		if userId != order.UserID {
			resp.Reason = "只有上课教练才能修改课程时间"
			return
		}
	} else {
		if orderCourse.TeachCoachID != userId {
			resp.Reason = "只有上课教练才能修改课程时间"
			return
		}
	}
	resp.CanChangeTime = true
	resp.Reason = ""
	return
}

// 教练修改课程时间
func CoachChangeOrderCourseTime(c *gin.Context, orderCourseId string, req *forms.CoachChangeOrderCourseTimeRequest) (err error) {
	userId := c.GetString("user_id")
	order, orderCourse, err := GetOrderCourses(orderCourseId)
	if err != nil {
		return err
	}
	if orderCourse.TeachState != model.TeachStateWaitClass && orderCourse.TeachState != model.TeachStateWaitCoachConfirmUser {
		return enum.NewErr(enum.OrdersCoursesExitErr, "待上课和待确认的课程才能修改上课时间")
	}
	if order.UserType == model.UserTypeClub {
		return enum.NewErr(enum.OrdersCoursesExitErr, "俱乐部的课程只能联系俱乐部修改上课时间")
	}
	if orderCourse.TeachCoachID == "" {
		if userId != order.UserID {
			return enum.NewErr(enum.OrdersCoursesExitErr, "只有上课教练才能修改课程时间111")
		}
	} else {
		if orderCourse.TeachCoachID != userId {
			return enum.NewErr(enum.OrdersCoursesExitErr, "只有上课教练才能修改课程时间")
		}
	}
	_, changeTimeIds, changeBufferTimeIds, startTi, err := getChangeStartTime(c, userId, req.TeachStartTime, model.UserTypeCoach, orderCourse, orderCourse.TeachBufferTime)
	if err != nil {
		return err
	}
	userType := c.GetInt("user_type")

	operate := model.OperateCoachChangeCourse
	changeTeachState := model.TeachStateWaitUserConfirmCoachTime
	if orderCourse.TeachState == model.TeachStateWaitCoachConfirmUser {
		operate = model.OperateCoachChangeCourseTime
		changeTeachState = model.TeachStateWaitUserSecondConfirmTime
	}
	inspectorate := model.OrdersCoursesState{
		OrderCourseID:  orderCourseId,
		UserID:         userId,
		UserType:       userType,
		Operate:        operate,
		Remark:         model.OCSOperateStr[operate],
		TeachTimeIDs:   append(changeTimeIds, changeBufferTimeIds...),
		TeachStartTime: model.LocalTime(startTi),
		Process:        model.ProcessNo,
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		if len(changeBufferTimeIds) > 0 { //设置缓冲时间
			upData := map[string]interface{}{
				"teach_state": model.SkiTeachStateAfterClass,
			}
			err = tx.Model(model.SkiResortsTeachTime{}).Where("id in ?", changeBufferTimeIds).
				Updates(upData).Error
			changeTimeIds = append(changeTimeIds, changeBufferTimeIds...)
		}

		if len(changeTimeIds) > 0 { //教练修改了上课时间，修改雪场对应的教学时间
			err = SRTOrderCourses(tx, changeTimeIds, orderCourseId, 0)
			if err != nil {
				global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
				return err
			}
		}

		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourseId).
			Updates(map[string]interface{}{
				"teach_state": changeTeachState,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "确认课程失败")
		}

		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "确认课程失败2")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func getChangeStartTime(c *gin.Context, userId, changeStartTime string, userType int, orderCourse model.OrdersCourses,
	bufferTime int) (isChangeStartTime bool, changeTimeIds, changeBufferTimeIds []int64, startTi time.Time, err error) {
	isChangeStartTime = false
	//changeTimeIds 		预约时间对应的教练的课程时间表ID
	//changeBufferTimeIds 	缓冲时间对应的教练的课程时间表ID

	if changeStartTime == "" { //时间不为空，说明教练需要改变上课时间
		return
	}

	if changeStartTime[14:16] != "00" && changeStartTime[14:16] != "30" {
		err = enum.NewErr(enum.TeachTimeErr, "时间格式错误,"+changeStartTime[14:16]+"不是00或30")
		return
	}
	startTi, err = time.ParseInLocation("2006-01-02 15:04", changeStartTime[:16], time.Local)
	if err != nil {
		err = enum.NewErr(enum.OrdersCoursesExitErr, "时间格式错误"+changeStartTime)
		return
	}

	bufferTimeCount := 0 //缓冲时间有几个30分钟
	bufferTimeCount = int(math.Ceil(float64(bufferTime) / 30))

	if startTi.Equal(time.Time(orderCourse.TeachStartTime)) { //时间不一致，说明教练要修改上课时间
		return
	}

	year, month, day := time.Now().Add(24 * time.Hour).Date()
	nextDay := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	if nextDay.After(time.Time(orderCourse.TeachStartTime)) {
		err = enum.NewErr(enum.OrdersCoursesExitErr, "上课当天不能修改上课时间")
		return
	}
	isChangeStartTime = true //标记一下要修改上课时间

	teachCount := int(math.Ceil(float64(orderCourse.TeachTime) / 30)) //上课时间段30 分钟一个时间段，不到30分钟按30分钟计算
	var teachStartTimes []model.LocalTime                             //上课时间，半小时为时间段
	upStartTi := startTi
	for i := 0; i < teachCount; i++ {
		teachStartTimes = append(teachStartTimes, model.LocalTime(upStartTi))
		upStartTi = upStartTi.Add(30 * time.Minute)
	}
	global.DB.Model(&model.SkiResortsTeachTime{}).
		Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0", userId, userType, orderCourse.SkiResortsID, teachStartTimes, model.SkiTeachStateWaitAppointment).
		Pluck("id", &changeTimeIds)
	if len(changeTimeIds) != teachCount {
		err = enum.NewErr(enum.OrdersCoursesExitErr, "教练上课时间冲突，请重新调整上课时间")
		return
	}

	if bufferTimeCount != 0 { //课后缓冲时间，需要单独判断
		var changeTeachStartTimes []model.LocalTime
		for i := 0; i < bufferTimeCount; i++ {
			changeTeachStartTimes = append(changeTeachStartTimes, model.LocalTime(upStartTi))
			upStartTi = upStartTi.Add(30 * time.Minute)
		}
		global.DB.Model(&model.SkiResortsTeachTime{}).
			Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0",
				userId, userType, orderCourse.SkiResortsID, changeTeachStartTimes, model.SkiTeachStateWaitAppointment).
			Pluck("id", &changeBufferTimeIds)
		if len(changeBufferTimeIds) != bufferTimeCount {
			err = enum.NewErr(enum.OrdersCoursesExitErr, "调整后的课后缓冲时间不够，请调整上课时间")
			return
		}
	}
	return
}

func GetOrderCourses(orderCourseId string) (order model.Orders, orderCourse model.OrdersCourses, err error) {
	err = global.DB.Model(model.OrdersCourses{}).
		Where("order_course_id = ? and state = 0", orderCourseId).First(&orderCourse).Error
	if err != nil {
		err = enum.NewErr(enum.OrdersCoursesExitErr, "订单课程不存在")
		return
	}
	err = global.DB.Model(model.Orders{}).Preload("Goods").
		Where("order_id = ? and state = 0", orderCourse.OrderID).First(&order).Error
	if err != nil {
		err = enum.NewErr(enum.OrdersCoursesExitErr, "订单不存在")
	}
	return
}

func BeforeCoachCancelOrderCourses(c *gin.Context, orderCourseId string) (resp forms.BeforeCancelAppointmentCourseResp, err error) {
	userId := c.GetString("user_id")

	order, orderCourse, err := GetOrderCourses(orderCourseId)
	if err != nil {
		return
	}
	if order.UserID != userId {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "只能取消自己的课程")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState >= model.TeachStateFinish {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "课程已完成")
	}

	if order.Status != enum.OrderStatusPaid {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "订单未支付")
	}
	if orderCourse.TeachState == model.TeachStateWaitAppointment {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "课程未预约")
	}

	if time.Now().Add(2 * 24 * time.Hour).After(time.Time(orderCourse.TeachStartTime)) {
		return resp, enum.NewErr(model.OrderFaultCancel, "离上课时间，2个自然日内，只能有责取消预约")
	}

	if orderCourse.TeachState != model.TeachStateWaitClass {
		return resp, enum.NewErr(model.OrderFaultCancel, "待上课状态才能取消预约")
	}
	year, month, _ := time.Now().Date()
	thisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	monthOneDay := thisMonth.Format("2006-01-02")
	var cancelNum int64 //用户无责取消预约的次数
	err = global.DB.Model(model.OrdersCoursesState{}).
		Where("user_id  = ? and  operate = ? and state = 0", userId, model.OperateCoachCancelNoResponsibility).
		Where("created_at > ? ", monthOneDay).
		Order("id desc").Count(&cancelNum).Error
	if err != nil {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "未找到预约记录")
	}
	resp.Liability = 1
	if cancelNum >= model.OrderCourseCoachCancelNumber {
		resp.Liability = 3
	}
	return resp, nil
}
func CoachCancelOrderCourses(c *gin.Context, orderCourseId string) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")

	order, orderCourse, err := GetOrderCourses(orderCourseId)
	if err != nil {
		return err
	}
	if order.UserID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "只能取消自己的课程")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState >= model.TeachStateFinish {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已完成")
	}

	if order.Status != enum.OrderStatusPaid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单未支付")
	}
	if orderCourse.TeachState == model.TeachStateWaitAppointment {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程未预约")
	}
	operate := model.OperateCoachCancelBeforeCoachConfirm
	//如果教练还没确认，取消预约没有限制
	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmUser {
		if time.Now().Add(2 * 24 * time.Hour).After(time.Time(orderCourse.TeachStartTime)) {
			return enum.NewErr(model.OrderFaultCancel, "离上课时间，2个自然日内，只能有责取消预约")
		}
		if orderCourse.TeachState != model.TeachStateWaitClass {
			return enum.NewErr(model.OrderFaultCancel, "待上课状态才能取消预约")
		}
		year, month, _ := time.Now().Date()
		thisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		monthOneDay := thisMonth.Format("2006-01-02")
		var cancelNum int64 //用户无责取消预约的次数
		err = global.DB.Model(model.OrdersCoursesState{}).
			Where("user_id  = ? and  operate = ? and state = 0", userId, model.OperateCoachCancelNoResponsibility).
			Where("created_at > ? ", monthOneDay).
			Order("id desc").Count(&cancelNum).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "未找到预约记录")
		}
		if cancelNum >= model.OrderCourseCoachCancelNumber {
			return enum.NewErr(model.OrderFaultCancel, "您已用完本月无责取消机会，请联系客服")
		}
		operate = model.OperateCoachCancelNoResponsibility
	}

	//取消预约
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourseId).
			Updates(map[string]interface{}{
				"teach_start_time":  gorm.Expr("NULL"),
				"teach_state":       model.TeachStateWaitAppointment,
				"teach_buffer_time": 0,
				"teach_time_ids":    model.JSONIntArray{},
				"club_time_ids":     model.JSONIntArray{},
				"ski_resorts_id":    0,
			}).Error
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: 取消预约: %w", zap.Error(err), zap.Any("orderCourseId", orderCourseId))
			return enum.NewErr(enum.OrdersCoursesExitErr, "取消预约失败")
		}

		err = SRTOrderCourses(tx, orderCourse.TeachTimeIDs, orderCourseId, 1)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return err
		}

		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourseId,
			UserID:        userId,
			UserType:      userType,
			TeachTimeIDs:  model.JSONIntArray{},
			Operate:       operate,
			Remark:        model.OCSOperateStr[operate],
			Process:       model.ProcessYes,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			global.Lg.Error("OrdersCoursesDao:  取消预约: %w", zap.Error(err), zap.Any("inspectorate", inspectorate))
			return err
		}
		err = UpdateOrderTeachState(c, tx, &order, model.TeachStateWaitAppointment)
		if err != nil {
			return err
		}

		err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
			orderCourse.OrderCourseID, model.OperateUserAppointment, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		return err
	})
	return nil
}
func CoachTransferOrderCourses(c *gin.Context, orderCourseId string) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	order, orderCourse, err := GetOrderCourses(orderCourseId)
	if err != nil {
		return err
	}
	if order.UserID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单不是您的，请不要乱操作")
	}

	if order.TransferCoachID != "" {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单已转单，不能再次转单")
	}
	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState == model.TeachStateWaitCoachTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "用户已同意转单，请不要重复操作")
	}
	if orderCourse.TeachState != model.TeachStateWaitClass {
		return enum.NewErr(enum.OrdersCoursesExitErr, "待上课状态才能转单")
	}

	if order.Goods != nil && order.Goods.Pack != model.PackNo {
		return enum.NewErr(enum.OrdersCoursesExitErr, "打包课程不能转单")
	}

	if time.Now().Add(2 * 24 * time.Hour).After(time.Time(orderCourse.TeachStartTime)) {
		return enum.NewErr(model.OrderFaultCancel, "离上课时间，2个自然日内，不能转单")
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        userId,
			UserType:      userType,
			Operate:       model.OperateCoachTransferCourse,
			Remark:        model.OCSOperateStr[model.OperateCoachTransferCourse],
			Process:       model.ProcessNo,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			global.Lg.Error("转单记录添加失败", zap.Error(err))
			return enum.NewErr(enum.OrdersCoursesExitErr, "转单记录添加失败")
		}
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourse.OrderCourseID).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateWaitUserConfirmTransfer,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "转单失败")
		}
		return nil
	})
	if err != nil {
		global.Lg.Error("转单失败", zap.Error(err), zap.Any("orderCourseId", orderCourseId))
		return err
	}
	return err
}

func CoachCancelTransferOrderCourses(c *gin.Context, orderCourseId string) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	order, orderCourse, err := GetOrderCourses(orderCourseId)
	if err != nil {
		return err
	}
	if order.UserID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单不是您的，请不要乱操作")
	}
	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState != model.TeachStateWaitCoachTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "该课程不是转单状态，不能取消转单")
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
			orderCourse.OrderCourseID, model.OperateCoachTransferCourse, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		if err != nil {
			global.Lg.Error("取消转单记录修改失败", zap.Error(err))
			return enum.NewErr(enum.OrdersCoursesExitErr, "取消转单记录修改失败")
		}
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        userId,
			UserType:      userType,
			Operate:       model.OperateCoachCancelCourseTransfer,
			Remark:        model.OCSOperateStr[model.OperateCoachCancelCourseTransfer],
			Process:       model.ProcessYes,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			global.Lg.Error("取消转单记录添加失败", zap.Error(err))
			return enum.NewErr(enum.OrdersCoursesExitErr, "取消转单记录添加失败")
		}
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourse.OrderCourseID).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateWaitClass,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "取消转单失败")
		}
		return nil
	})
	if err != nil {
		global.Lg.Error("取消转单失败", zap.Error(err), zap.Any("orderCourseId", orderCourseId))
		return err
	}
	return err
}

func CoachTransferOrderToCoach(c *gin.Context, req *forms.CoachTransferOrderToCoachRequest) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")

	order, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return err
	}
	if order.UserID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单不是您的，请不要乱操作")
	}

	if order.TotalFee < req.TransferFee {
		return enum.NewErr(enum.OrdersCoursesExitErr, "转单金额不能大于订单金额")
	}
	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState == model.TeachStateWaitUserConfirmTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "先联系用户确认转单")
	}
	if orderCourse.TeachState != model.TeachStateWaitCoachTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "请先发起转单")
	}

	ordersCoursesState := model.OrdersCoursesState{}
	err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and state=0", orderCourse.OrderCourseID).
		Last(&ordersCoursesState).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "查询转单记录失败")
	}
	if ordersCoursesState.Operate == model.OperateCoachTransferToCoach && ordersCoursesState.Process == model.ProcessNo {
		return enum.NewErr(enum.OrdersCoursesExitErr, "已经发起转单请求，请联系教练处理")
	}

	var orderTagIds []int64
	for _, v := range orderCourse.CourseTags {
		orderTagIds = append(orderTagIds, v.TagID)
	}
	transferCoach, err := CheckCoachTag(req.CoachId, orderTagIds)
	if err != nil {
		return err
	}

	teachCount := int(math.Ceil(float64(orderCourse.TeachTime) / 30))
	var teachStartTimes []model.LocalTime

	upStartTi := time.Time(orderCourse.TeachStartTime)
	for i := 0; i < teachCount; i++ {
		teachStartTimes = append(teachStartTimes, model.LocalTime(upStartTi))
		upStartTi = upStartTi.Add(30 * time.Minute)
	}
	var ids []int64
	global.DB.Model(&model.SkiResortsTeachTime{}).
		Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0",
			transferCoach.CoachId, enum.UserTypeCoach, orderCourse.SkiResortsID, teachStartTimes, model.SkiTeachStateWaitAppointment).
		Pluck("id", &ids)
	if len(ids) != teachCount {
		return enum.NewErr(enum.OrdersCoursesExitErr, "教练时间冲突，请重新选择教练")
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        userId,
			UserType:      userType,
			CoachID:       transferCoach.CoachId,
			Operate:       model.OperateCoachTransferToCoach,
			Remark:        model.OCSOperateStr[model.OperateCoachTransferToCoach],
			TeachTimeIDs:  ids,
			TransferFee:   req.TransferFee,
			Process:       model.ProcessNo,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			global.Lg.Error("教练转单课程给其他教练记录添加失败", zap.Error(err))
			return enum.NewErr(enum.OrdersCoursesExitErr, "教练转单课程给其他教练记录添加失败")
		}

		// 修改课程状态为待确认转单
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateWaitConfirmTransfer,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态修改失败")
		}

		err = tx.Model(model.Orders{}).Where("order_id = ? and state=0", order.OrderID).
			Updates(map[string]interface{}{
				"transfer_coach_id": transferCoach.CoachId,
				"transfer_fee":      req.TransferFee,
			}).Error
		err = SRTOrderCourses(tx, orderCourse.TeachTimeIDs, orderCourse.OrderCourseID, 0)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return err
		}
		return nil
	})
	return err
}

func CoachReviewOrderFromCoach(c *gin.Context, req *forms.CoachReviewOrderFromCoachRequest) (err error) {
	userId := c.GetString("user_id")
	coachId := c.GetString("coach_id")
	//userId := "C20250823181716sjfsfji" //TODO: 临时测试
	userType := c.GetInt("user_type")

	order, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return err
	}

	if order.TransferCoachID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单不是您的，请不要乱操作")
	}
	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState == model.TeachStateWaitUserConfirmTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "用户还未确认转单")
	}
	if orderCourse.TeachState != model.TeachStateWaitConfirmTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "请让教练先发起转单")
	}
	//找出最新的操作记录，如果不是教练转单记录，则返回错误
	orderCourseState := model.OrdersCoursesState{}
	err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
		Last(&orderCourseState).Error
	if err != nil || orderCourseState.CoachID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "您没有该课程的转单记录")
	}
	if orderCourseState.Operate != model.OperateCoachTransferToCoach {
		return enum.NewErr(enum.OrdersCoursesExitErr, "请让教练先发起转单")
	}
	if orderCourseState.Process == model.ProcessYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "转单已处理")
	}
	if req.IsAgree { //教练同意接单
		var bufferTimeIds []int64 //查出缓冲时间对应的教练的课程时间表ID
		bufferTimeCount := 0      //缓冲时间有几个30分钟
		if req.BufferTime != 0 {  //设置了缓冲时间
			bufferTimeCount = int(math.Ceil(float64(req.BufferTime) / 30))
			endTimeData := model.SkiResortsTeachTime{} //查出预约的结束时间
			err = global.DB.Model(&model.SkiResortsTeachTime{}).Where("id in ?", []int64(orderCourseState.TeachTimeIDs)).
				Order("teach_end_time desc").Take(&endTimeData).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课程时间不存在")
			}
			var bufferStartTimes []model.LocalTime
			buffeStartTi := time.Time(endTimeData.TeachEndTime)
			for i := 0; i < bufferTimeCount; i++ {
				bufferStartTimes = append(bufferStartTimes, model.LocalTime(buffeStartTi))
				buffeStartTi = buffeStartTi.Add(30 * time.Minute)
			}
			global.DB.Model(&model.SkiResortsTeachTime{}).
				Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0",
					userId, userType, orderCourse.SkiResortsID, bufferStartTimes, model.SkiTeachStateWaitAppointment).
				Pluck("id", &bufferTimeIds)
			if len(bufferTimeIds) != bufferTimeCount { //缓冲时间冲突
				return enum.NewErr(enum.OrdersCoursesExitErr, "课后缓冲时间冲突")
			}
		}
		idsStr, err := json.Marshal(append(orderCourseState.TeachTimeIDs, bufferTimeIds...))
		if err != nil {
			global.Lg.Error("解析数据失败", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return enum.NewErr(enum.OrdersCoursesExitErr, "解析数据失败")
		}

		err = global.DB.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(model.OrdersCoursesState{}).Where("id = ? and operate=? and process=?",
				orderCourseState.ID, model.OperateCoachTransferToCoach, model.ProcessNo).
				Updates(map[string]interface{}{
					"process": model.ProcessYes,
				}).Error
			if err != nil {
				global.Lg.Error("取消转单记录修改失败", zap.Error(err))
				return enum.NewErr(enum.OrdersCoursesExitErr, "转单记录修改失败")
			}

			inspectorate := model.OrdersCoursesState{
				OrderCourseID: orderCourse.OrderCourseID,
				UserID:        userId,
				UserType:      userType,
				Operate:       model.OperateCoachAgreeTransferCourse,
				Remark:        model.OCSOperateStr[model.OperateCoachAgreeTransferCourse],
				Process:       model.ProcessYes,
			}
			err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "添加同意转单记录失败")
			}

			//教练同意转单，将课程状态改为待上课，教学时间改为当前教练的时间
			err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
				Updates(map[string]interface{}{
					"teach_state":    model.TeachStateWaitClassTransfer,
					"teach_coach_id": coachId,
					"teach_time_ids": string(idsStr),
				}).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态修改失败")
			}

			if len(bufferTimeIds) != 0 { //设置课后缓冲时间
				err = SRTOrderCourses(tx, bufferTimeIds, orderCourse.OrderCourseID, 0)
				if err != nil {
					global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
					return err
				}

				err = tx.Model(model.SkiResortsTeachTime{}).
					Where("id in (?)", bufferTimeIds).
					Updates(map[string]interface{}{
						"teach_state": model.SkiTeachStateAfterClass,
					}).Error
				if err != nil {
					return enum.NewErr(enum.OrdersCoursesExitErr, "课后缓冲时间锁定失败")
				}
			}

			//教练同意转单，将之前教练的时间释放
			err = SRTOrderCourses(tx, orderCourse.TeachTimeIDs, orderCourse.OrderCourseID, 1)
			if err != nil {
				global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
				return err
			}
			return nil
		})
	} else { //教练拒绝接单
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        userId,
			UserType:      userType,
			Operate:       model.OperateCoachDisagreeTransferCourse,
			Remark:        model.OCSOperateStr[model.OperateCoachDisagreeTransferCourse],
			Process:       model.ProcessYes,
		}
		err = CancelOrderFromCoachSql(c, inspectorate, orderCourseState)
		if err != nil {
			global.Lg.Error("CancelOrderFromCoachSql error", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return enum.NewErr(enum.OrdersCoursesExitErr, "转单失败")
		}
		err = global.DB.Transaction(func(tx *gorm.DB) error {
			// 修改课程状态为待确认转单
			err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
				Updates(map[string]interface{}{
					"teach_state": model.TeachStateWaitCoachTransfer,
				}).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "TeachStateWaitCoachTransfer 课程状态修改失败")
			}

			err = tx.Model(model.Orders{}).Where("order_id = ? and state=0", orderCourse.OrderID).
				Updates(map[string]interface{}{
					"transfer_coach_id": "",
					"transfer_fee":      0,
				}).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "订单教练ID修改失败")
			}
			return nil
		})

	}
	if err != nil {
		global.Lg.Error("CoachTransferOrderToCoach error", zap.Error(err), zap.Any("orderCourse", orderCourse))
		return enum.NewErr(enum.OrdersCoursesExitErr, "转单失败")
	}
	return nil
}

func CancelOrderFromCoachSql(c context.Context, inspectorate model.OrdersCoursesState, ocs model.OrdersCoursesState) (err error) {
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "添加拒绝转单记录失败")
		}

		err = tx.Model(model.OrdersCoursesState{}).Where("id = ? and process=?",
			ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		if err != nil {
			global.Lg.Error("取消转单记录修改失败", zap.Error(err))
			return enum.NewErr(enum.OrdersCoursesExitErr, "转单记录修改失败")
		}

		//教练拒绝转单，将教练的时间释放
		err = SRTOrderCourses(tx, ocs.TeachTimeIDs, inspectorate.OrderCourseID, 1)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", inspectorate))
			return err
		}
		return nil
	})
	return err
}

func CoachReviewOrderFromClub(c *gin.Context, req *forms.CoachReviewOrderFromClubRequest) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	coachId := c.GetString("coach_id")
	if userType != model.UserTypeCoach {
		return enum.NewErr(enum.OrdersCoursesExitErr, "您不是教练，无法进行该操作")
	}

	_, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return err
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState == model.TeachStateWaitUserConfirmClubTime {
		return enum.NewErr(enum.OrdersCoursesExitErr, "等待用户确认时间")
	}
	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmClub {
		return enum.NewErr(enum.OrdersCoursesExitErr, "该课程还未分配")
	}

	//找出最新的操作记录，如果不是教练转单记录，则返回错误
	orderCourseState := model.OrdersCoursesState{}
	err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
		Last(&orderCourseState).Error
	if err != nil || orderCourseState.CoachID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "您没有该课程的转单记录")
	}
	if orderCourseState.Operate != model.OperateClubAppointCoach {
		return enum.NewErr(enum.OrdersCoursesExitErr, "请先让俱乐部分配课程")
	}
	if orderCourseState.Process == model.ProcessYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "已处理")
	}
	if req.IsAgree { //教练同意俱乐部课程分配
		var bufferTimeIds []int64 //查出缓冲时间对应的教练的课程时间表ID
		bufferTimeCount := 0      //缓冲时间有几个30分钟
		if req.BufferTime != 0 {  //设置了缓冲时间
			bufferTimeCount = int(math.Ceil(float64(req.BufferTime) / 30))
			endTimeData := model.SkiResortsTeachTime{} //查出预约的结束时间
			err = global.DB.Model(&model.SkiResortsTeachTime{}).Where("id in ?", []int64(orderCourseState.TeachTimeIDs)).
				Order("teach_end_time desc").Take(&endTimeData).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课程时间不存在")
			}
			var bufferStartTimes []model.LocalTime
			buffeStartTi := time.Time(endTimeData.TeachEndTime)
			for i := 0; i < bufferTimeCount; i++ {
				bufferStartTimes = append(bufferStartTimes, model.LocalTime(buffeStartTi))
				buffeStartTi = buffeStartTi.Add(30 * time.Minute)
			}
			global.DB.Model(&model.SkiResortsTeachTime{}).
				Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0",
					userId, userType, orderCourse.SkiResortsID, bufferStartTimes, model.SkiTeachStateWaitAppointment).
				Pluck("id", &bufferTimeIds)
			if len(bufferTimeIds) != bufferTimeCount { //缓冲时间冲突
				return enum.NewErr(enum.OrdersCoursesExitErr, "课后缓冲时间冲突")
			}
		}
		idsStr, err := json.Marshal(append(orderCourseState.TeachTimeIDs, bufferTimeIds...))
		if err != nil {
			global.Lg.Error("解析数据失败", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return enum.NewErr(enum.OrdersCoursesExitErr, "解析数据失败")
		}

		err = global.DB.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(model.OrdersCoursesState{}).Where("id = ? and operate=? and process=?",
				orderCourseState.ID, model.OperateClubAppointCoach, model.ProcessNo).
				Updates(map[string]interface{}{
					"process": model.ProcessYes,
				}).Error
			if err != nil {
				global.Lg.Error("取消转单记录修改失败", zap.Error(err))
				return enum.NewErr(enum.OrdersCoursesExitErr, "转单记录修改失败")
			}

			inspectorate := model.OrdersCoursesState{
				OrderCourseID: orderCourse.OrderCourseID,
				UserID:        userId,
				UserType:      userType,
				CoachID:       coachId,
				Operate:       model.OperateCoachAgreeClubCourse,
				Remark:        model.OCSOperateStr[model.OperateCoachAgreeClubCourse],
				Process:       model.ProcessYes,
			}
			err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "添加同意转单记录失败")
			}

			//教练同意俱乐部安排，将课程状态改为待上课，教学时间改为当前教练的时间
			err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
				Updates(map[string]interface{}{
					"teach_state":       model.TeachStateWaitCoachClass,
					"teach_coach_id":    coachId,
					"teach_time_ids":    string(idsStr),
					"teach_buffer_time": req.BufferTime,
				}).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态修改失败")
			}

			if len(bufferTimeIds) != 0 { //设置课后缓冲时间
				err = tx.Model(model.SkiResortsTeachTime{}).
					Where("id in (?)", bufferTimeIds).
					Updates(map[string]interface{}{
						"teach_state": model.SkiTeachStateAfterClass,
					}).Error
				if err != nil {
					return enum.NewErr(enum.OrdersCoursesExitErr, "课后缓冲时间锁定失败")
				}

				err = SRTOrderCourses(tx, bufferTimeIds, orderCourse.OrderCourseID, 0)
				if err != nil {
					global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
					return err
				}
			}
			return nil
		})
	} else { //教练拒绝接单
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        userId,
			UserType:      userType,
			Operate:       model.OperateCoachDisagreeClubCourse,
			Remark:        model.OCSOperateStr[model.OperateCoachDisagreeClubCourse],
			Process:       model.ProcessYes,
		}
		err = CoachDisAgreeOrderFromClubSql(c, orderCourse, inspectorate, orderCourseState)
	}
	if err != nil {
		global.Lg.Error("CoachTransferOrderToCoach error", zap.Error(err), zap.Any("orderCourse", orderCourse))
		return enum.NewErr(enum.OrdersCoursesExitErr, "转单失败")
	}
	return nil
}

func CoachDisAgreeOrderFromClubSql(c context.Context, orderCourse model.OrdersCourses, inspectorate model.OrdersCoursesState, orderCourseState model.OrdersCoursesState) (err error) {
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCoursesState{}).Where("id = ? and operate=? and process=?",
			orderCourseState.ID, model.OperateClubAppointCoach, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		if err != nil {
			global.Lg.Error("取消转单记录修改失败", zap.Error(err))
			return enum.NewErr(enum.OrdersCoursesExitErr, "转单记录修改失败")
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "添加拒绝转单记录失败")
		}

		//教练拒绝转单，将教练的时间释放
		err = SRTOrderCourses(tx, orderCourseState.TeachTimeIDs, orderCourse.OrderCourseID, 1)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return err
		}
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", orderCourse.OrderCourseID).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateWaitClubConfirm,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态修改失败")
		}
		return nil
	})
	return err
}

func CoachReviewReplaceFromClub(c *gin.Context, req *forms.CoachReviewReplaceFromClubRequest) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	if userType != model.UserTypeCoach {
		return enum.NewErr(enum.OrdersCoursesExitErr, "您不是教练，无法核销")
	}

	_, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return err
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState == model.TeachStateCoachApplyTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "等待俱乐部分配课程")
	}
	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "该课程还未分配")
	}

	//找出最新的操作记录，如果不是教练转单记录，则返回错误
	orderCourseState := model.OrdersCoursesState{}
	err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
		Last(&orderCourseState).Error
	if err != nil || orderCourseState.CoachID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "您没有该课程的转单记录")
	}
	if orderCourseState.Operate != model.OperateClubTransferToCoach {
		return enum.NewErr(enum.OrdersCoursesExitErr, "请先让俱乐部先分配课程")
	}
	if orderCourseState.Process == model.ProcessYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "已处理")
	}
	if req.IsAgree { //教练同意俱乐部课程分配
		var bufferTimeIds []int64 //查出缓冲时间对应的教练的课程时间表ID
		bufferTimeCount := 0      //缓冲时间有几个30分钟
		if req.BufferTime != 0 {  //设置了缓冲时间
			bufferTimeCount = int(math.Ceil(float64(req.BufferTime) / 30))
			endTimeData := model.SkiResortsTeachTime{} //查出预约的结束时间
			err = global.DB.Model(&model.SkiResortsTeachTime{}).Where("id in ?", []int64(orderCourseState.TeachTimeIDs)).
				Order("teach_end_time desc").Take(&endTimeData).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课程时间不存在")
			}
			var bufferStartTimes []model.LocalTime
			buffeStartTi := time.Time(endTimeData.TeachEndTime)
			for i := 0; i < bufferTimeCount; i++ {
				bufferStartTimes = append(bufferStartTimes, model.LocalTime(buffeStartTi))
				buffeStartTi = buffeStartTi.Add(30 * time.Minute)
			}
			global.DB.Model(&model.SkiResortsTeachTime{}).
				Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0",
					userId, userType, orderCourse.SkiResortsID, bufferStartTimes, model.SkiTeachStateWaitAppointment).
				Pluck("id", &bufferTimeIds)
			if len(bufferTimeIds) != bufferTimeCount { //缓冲时间冲突
				return enum.NewErr(enum.OrdersCoursesExitErr, "课后缓冲时间冲突")
			}
		}
		idsStr, err := json.Marshal(append(orderCourseState.TeachTimeIDs, bufferTimeIds...))
		if err != nil {
			global.Lg.Error("解析数据失败", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return enum.NewErr(enum.OrdersCoursesExitErr, "解析数据失败")
		}

		err = global.DB.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate in ? and process=?",
				orderCourse.OrderCourseID, []int{model.OperateCoachApplyTransferCourse, model.OperateClubTransferToCoach}, model.ProcessNo).
				Updates(map[string]interface{}{
					"process": model.ProcessYes,
				}).Error
			if err != nil {
				global.Lg.Error("取消转单记录修改失败", zap.Error(err))
				return enum.NewErr(enum.OrdersCoursesExitErr, "转单记录修改失败")
			}

			inspectorate := model.OrdersCoursesState{
				OrderCourseID: orderCourse.OrderCourseID,
				UserID:        userId,
				UserType:      userType,
				Operate:       model.OperateCoachAgreeClubTransferToCoach,
				Remark:        model.OCSOperateStr[model.OperateCoachAgreeClubCourse],
				Process:       model.ProcessYes,
			}
			err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "添加同意转单记录失败")
			}

			//教练同意俱乐部安排，将课程状态改为待上课，教学时间改为当前教练的时间
			err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
				Updates(map[string]interface{}{
					"teach_state":    model.TeachStateWaitCoachClass,
					"teach_time_ids": string(idsStr),
				}).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态修改失败")
			}

			if len(bufferTimeIds) != 0 { //设置课后缓冲时间
				err = tx.Model(model.SkiResortsTeachTime{}).
					Where("id in (?)", bufferTimeIds).
					Updates(map[string]interface{}{
						"teach_state": model.SkiTeachStateAfterClass,
					}).Error
				if err != nil {
					return enum.NewErr(enum.OrdersCoursesExitErr, "课后缓冲时间锁定失败")
				}

				err = SRTOrderCourses(tx, bufferTimeIds, orderCourse.OrderCourseID, 0)
				if err != nil {
					global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
					return err
				}
			}

			//教练同意俱乐部课程分配，将之前接单的教练时间释放
			err = SRTOrderCourses(tx, orderCourse.TeachTimeIDs, orderCourse.OrderCourseID, 1)
			if err != nil {
				global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
				return err
			}
			return nil
		})
	} else { //教练拒绝接单
		lastOrderCourseState := model.OrdersCoursesState{}
		err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate in ? ",
			orderCourse.OrderCourseID, []int{model.OperateCoachAgreeClubCourse, model.OperateCoachAgreeClubTransferToCoach}).Last(&lastOrderCourseState).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "找不到之前的教练")
		}

		err = global.DB.Transaction(func(tx *gorm.DB) error {
			inspectorate := model.OrdersCoursesState{
				OrderCourseID: orderCourse.OrderCourseID,
				UserID:        userId,
				UserType:      userType,
				Operate:       model.OperateCoachDisagreeClubTransferToCoach,
				Remark:        model.OCSOperateStr[model.OperateCoachDisagreeClubTransferToCoach],
				Process:       model.ProcessYes,
			}
			err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "添加拒绝转单记录失败")
			}

			//教练拒绝转单，将教练的时间释放
			err = SRTOrderCourses(tx, orderCourseState.TeachTimeIDs, orderCourse.OrderCourseID, 1)
			if err != nil {
				global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
				return err
			}
			err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", req.OrderCourseId).
				Updates(map[string]interface{}{
					"teach_state":    model.TeachStateCoachApplyTransfer,
					"teach_coach_id": lastOrderCourseState.CoachID,
				}).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态修改失败")
			}

			err = tx.Model(model.SkiResortsTeachTime{}).
				Where("id in (?) and teach_state = ?", []int64(orderCourseState.TeachTimeIDs), model.SkiTeachStateAfterClass).
				Updates(map[string]interface{}{
					"teach_state": model.SkiTeachStateWaitAppointment,
				}).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课后缓冲时间释放失败")
			}
			err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate = ? and process = ?",
				orderCourse.OrderCourseID, model.OperateClubTransferToCoach, model.ProcessNo).
				Updates(map[string]interface{}{
					"process": model.ProcessYes,
				}).Error
			return err
		})
	}
	if err != nil {
		global.Lg.Error("CoachTransferOrderToCoach error", zap.Error(err), zap.Any("orderCourse", orderCourse))
		return enum.NewErr(enum.OrdersCoursesExitErr, "转单失败")
	}
	return nil
}

func CoachApplyTransferOrders(c *gin.Context, orderCourseId string) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	order, orderCourse, err := GetOrderCourses(orderCourseId)
	if err != nil {
		return err
	}
	if orderCourse.TeachCoachID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "不是您分配的课程")
	}
	if order.UserType != model.UserTypeClub {
		return enum.NewErr(enum.OrdersCoursesExitErr, "非俱乐部订单")
	}
	if orderCourse.TeachState == model.TeachStateCoachApplyTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "已经申请转单，请等待俱乐部确认")
	}
	if orderCourse.TeachState != model.TeachStateWaitCoachClass {
		return enum.NewErr(enum.OrdersCoursesExitErr, "待上课课程才能转单")
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourseId,
			UserID:        userId,
			UserType:      userType,
			Operate:       model.OperateCoachApplyTransferCourse,
			Remark:        model.OCSOperateStr[model.OperateCoachApplyTransferCourse],
			Process:       model.ProcessNo,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "添加拒绝转单记录失败")
		}

		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", orderCourseId).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateCoachApplyTransfer,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态修改失败")
		}
		return nil
	})
	return nil
}

func CoachCancelApplyTransferOrders(c *gin.Context, orderCourseId string) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	order, orderCourse, err := GetOrderCourses(orderCourseId)
	if err != nil {
		return err
	}
	if orderCourse.TeachCoachID != userId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "不是您分配的课程")
	}
	if order.UserType != model.UserTypeClub {
		return enum.NewErr(enum.OrdersCoursesExitErr, "非俱乐部订单")
	}
	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState != model.TeachStateCoachApplyTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "该课程不是教练申请转单状态，不能取消转单")
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
			orderCourse.OrderCourseID, model.OperateCoachApplyTransferCourse, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		if err != nil {
			global.Lg.Error("取消转单记录修改失败", zap.Error(err))
			return enum.NewErr(enum.OrdersCoursesExitErr, "取消转单记录修改失败")
		}
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        userId,
			UserType:      userType,
			Operate:       model.OperateCoachCancelApplyTransferCourse,
			Remark:        model.OCSOperateStr[model.OperateCoachCancelApplyTransferCourse],
			Process:       model.ProcessYes,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			global.Lg.Error("取消转单记录添加失败", zap.Error(err))
			return enum.NewErr(enum.OrdersCoursesExitErr, "取消转单记录添加失败")
		}
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourse.OrderCourseID).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateWaitCoachClass,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "取消转单失败")
		}
		return nil
	})
	if err != nil {
		global.Lg.Error("取消转单失败", zap.Error(err), zap.Any("orderCourseId", orderCourseId))
		return err
	}
	return err
}

func CoachVerifyCourses(c *gin.Context, checkCode string) (err error) {
	coachId := c.GetString("coach_id")
	orderCourse := model.OrdersCourses{}
	err = global.DB.Model(model.OrdersCourses{}).
		Where("check_code = ? and teach_coach_id=? and  state = 0", checkCode, coachId).First(&orderCourse).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "核销码不存在或教学教练不是您")
	}
	if orderCourse.TeachState == model.TeachStateWaitCoachConfirmUser {
		return enum.NewErr(enum.OrdersCoursesExitErr, "先确认课程")
	}
	if orderCourse.TeachState != model.TeachStateWaitClass && orderCourse.TeachState != model.TeachStateWaitCheck && orderCourse.TeachState != model.TeachStateWaitCoachClass {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态错误")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}

	order := model.Orders{}
	err = global.DB.Model(model.Orders{}).Preload("Goods").
		Where("order_id = ? and state = 0", orderCourse.OrderID).First(&order).Error

	insrtocsData := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        c.GetString("user_id"),
		UserType:      c.GetInt("user_type"),
		Operate:       model.OperateCoachVerifyCourse,
		Remark:        model.OCSOperateStr[model.OperateCoachVerifyCourse],
		CoachID:       coachId,
	}
	err = CompleteCourseSplitMoney(c, orderCourse, order, insrtocsData)

	if err != nil {
		global.Lg.Error("CompleteCourseSplitMoney error", zap.Error(err), zap.Any("orderCourse", orderCourse))
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程核销失败")
	}
	return err
}

func AdminTransferOrderToCoach(c *gin.Context, orderId string, req *forms.AdminTransferOrderToCoachRequest) (err error) {
	//转单给教练
	order, err := QueryOrderInfo("", orderId)
	if err != nil {
		global.Lg.Error("QueryOrderInfo failed", zap.Error(err))
		return err
	}

	if order.Pack == model.PackYes {
		global.Lg.Error("打包课不允许转单")
		return enum.NewErr(enum.OrderTransferErr, "打包课不允许转单")
	}

	if order.TeachState > model.PackTeachStateDoing { //订单未完成才能转单
		global.Lg.Error("只能待预约的课程转单", zap.Int("teach_state", order.TeachState))
		return enum.NewErr(enum.OrderTransferErr, "只能待预约的课程转单")
	}

	if len(order.OrdersCourses) > 1 {
		global.Lg.Error("只能单课程转单")
		return enum.NewErr(enum.OrderTransferErr, "只能单课程转单")
	}

	ocs := order.OrdersCourses[0]

	coach, err := CoachInfoByCoachId(req.CoachId)
	if err != nil {
		global.Lg.Error("CoachInfoByCoachId failed", zap.Error(err))
		return err
	}

	previousUserId := order.UserID
	previousUserType := order.UserType

	order.UserID = coach.CoachId
	order.UserType = model.UserTypeCoach
	order.TeachState = model.PackTeachStateWaitAppointment
	if err = global.DB.Transaction(func(tx *gorm.DB) error {
		//修改订单的教练或者额俱乐部
		err = tx.Model(&model.Orders{}).Where("order_id = ?", orderId).Save(order).Error
		if err != nil {
			global.Lg.Error("更新订单用户ID失败", zap.Error(err))
			return err
		}

		//插入转单记录
		_, err = CreateOrdersTransferRecords(c, tx, orderId, previousUserId, previousUserType, coach.CoachId, model.UserTypeCoach)
		if err != nil {
			global.Lg.Error("更新订单课程教练ID失败", zap.Error(err))
			return err
		}

		//释放教练时间
		if len(ocs.TeachTimeIDs) > 0 {
			//order_course_state状态流转
			ors := model.OrdersCoursesState{
				OrderCourseID: ocs.OrderCourseID,
				UserID:        "",
				UserType:      model.UserTypeOfficial,
				Operate:       model.OperateAdminOrderTransfer,
				CoachID:       req.CoachId,
				TransferFee:   order.TotalFee,
				Remark:        "管理台转单",
			}
			if err = CancelCoachCourseSql(c, tx, *order, ocs, ors); err != nil {
				global.Lg.Error("释放教练时间失败", zap.Error(err))
				return err
			}
		}
		return nil
	}); err != nil {
		global.Lg.Error("更新订单课程教练ID失败", zap.Error(err))
		return err
	}

	return nil
}
