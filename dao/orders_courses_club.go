package dao

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"
)

//START 俱乐部操作

func ClubChangeTeachTime(c *gin.Context, orderCourseId string, req *forms.ClubChangeTeachTimeRequest) (err error) {
	if req.TeachStartTime[14:16] != "00" && req.TeachStartTime[14:16] != "30" {
		return enum.NewErr(enum.TeachTimeErr, "时间格式错误,"+req.TeachStartTime[14:16]+"不是00或30")
	}
	startTi, err := time.ParseInLocation("2006-01-02 15:04", req.TeachStartTime[:16], time.Local)
	if err != nil {
		return enum.NewErr(enum.TeachTimeErr, "时间格式错误,"+req.TeachStartTime[:16])
	}
	if startTi.Before(time.Now()) {
		return enum.NewErr(enum.TeachTimeErr, "时间格式错误,"+req.TeachStartTime[:16]+"不能小于当前时间")
	}
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	if userType != model.UserTypeClub {
		return enum.NewErr(enum.ClubExitErr, "请先登录俱乐部账号")
	}
	var orderCourse model.OrdersCourses
	err = global.DB.Model(model.OrdersCourses{}).
		Where("order_course_id=? and state=0", orderCourseId).
		First(&orderCourse).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "预约课程不存在")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	order := model.Orders{}
	err = global.DB.Model(&order).Where("order_id=? and user_id=? and user_type=? and state=0", orderCourse.OrderID, userId, userType).First(&order).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "不是你的订单")
	}
	if order.Status != enum.OrderStatusPaid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单未支付")
	}

	if orderCourse.TeachState == model.TeachStateWaitCoachConfirmClub {
		return enum.NewErr(enum.OrdersCoursesExitErr, "等教练确认课程时间后再修改时间")
	}
	if orderCourse.TeachState == model.TeachStateWaitAppointment {
		return enum.NewErr(enum.OrdersCoursesExitErr, "等用户预约后再修改时间")
	}
	if orderCourse.TeachState == model.TeachStateWaitUserConfirmClubTime {
		return enum.NewErr(enum.OrdersCoursesExitErr, "已经修改时间，请联系用户确认时间")
	}
	if orderCourse.TeachState != model.TeachStateWaitClubConfirm && orderCourse.TeachState != model.TeachStateWaitCoachClass {
		return enum.NewErr(enum.OrdersCoursesExitErr, "该课程暂不能修改时间")
	}

	//检测时间是否冲突  START
	//上课时间段30 分钟一个时间段，不到30分钟按30分钟计算
	teachCount := int(math.Ceil(float64(orderCourse.TeachTime) / 30))
	var teachStartTimes []model.LocalTime
	upStartTi := startTi
	for i := 0; i < teachCount; i++ {
		teachStartTimes = append(teachStartTimes, model.LocalTime(upStartTi))
		upStartTi = upStartTi.Add(30 * time.Minute)
	}
	var clubTimeIds, coachTimeIds, bufferTimeIds []int64
	global.DB.Model(&model.SkiResortsTeachTime{}).
		Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0", userId, userType, orderCourse.SkiResortsID, teachStartTimes, model.SkiTeachStateWaitAppointment).
		Pluck("id", &clubTimeIds)
	if len(clubTimeIds) != teachCount {
		return enum.NewErr(enum.OrdersCoursesExitErr, "俱乐部时间冲突，请重新选择时间")
	}
	if orderCourse.TeachCoachID != "" { //已经有教练接单了
		global.DB.Model(&model.SkiResortsTeachTime{}).
			Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0",
				orderCourse.TeachCoachID, model.UserTypeCoach, orderCourse.SkiResortsID, teachStartTimes, model.SkiTeachStateWaitAppointment).
			Pluck("id", &coachTimeIds)
		if len(coachTimeIds) != teachCount {
			return enum.NewErr(enum.OrdersCoursesExitErr, "教练时间冲突，请重新选择时间")
		}

		bufferTimeCount := 0                  //缓冲时间有几个30分钟
		if orderCourse.TeachBufferTime != 0 { //设置了缓冲时间
			bufferTimeCount = int(math.Ceil(float64(orderCourse.TeachBufferTime) / 30))
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "课程时间不存在")
			}
			var bufferStartTimes []model.LocalTime
			buffeStartTi := upStartTi
			for i := 0; i < bufferTimeCount; i++ {
				bufferStartTimes = append(bufferStartTimes, model.LocalTime(buffeStartTi))
				buffeStartTi = buffeStartTi.Add(30 * time.Minute)
			}
			global.DB.Model(&model.SkiResortsTeachTime{}).
				Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0",
					orderCourse.TeachCoachID, model.UserTypeCoach, orderCourse.SkiResortsID, bufferStartTimes, model.SkiTeachStateWaitAppointment).
				Pluck("id", &bufferTimeIds)
			if len(bufferTimeIds) != bufferTimeCount { //缓冲时间冲突
				return enum.NewErr(enum.OrdersCoursesExitErr, "教练的课后缓冲时间冲突")
			}
		}
	}
	//检测时间是否冲突  END

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourseId).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateWaitUserConfirmClubTime,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "课程修改状态失败")
		}

		timeIds := clubTimeIds
		if orderCourse.TeachCoachID != "" {
			timeIds = append(timeIds, coachTimeIds...)
			timeIds = append(timeIds, bufferTimeIds...)
		}
		err = SRTOrderCourses(tx, timeIds, orderCourse.OrderCourseID, 0)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return err
		}
		insrtocsData := model.OrdersCoursesState{
			OrderCourseID:  orderCourseId,
			UserID:         userId,
			UserType:       userType,
			Operate:        model.OperateClubChangeUserCourseTime,
			Remark:         model.OCSOperateStr[model.OperateClubChangeUserCourseTime],
			TeachStartTime: model.LocalTime(startTi),
			TeachTimeIDs:   clubTimeIds,
			Process:        model.ProcessNo,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&insrtocsData).Error
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: 取消预约: %w", zap.Error(err), zap.Any("req", req), zap.Any("insrtocsData", insrtocsData))
			return enum.NewErr(enum.OrdersCoursesExitErr, "俱乐部插入记录失败")
		}
		if orderCourse.TeachCoachID != "" {
			if len(bufferTimeIds) != 0 {
				err = tx.Model(&model.SkiResortsTeachTime{}).Where("id in ?", bufferTimeIds).
					Updates(map[string]interface{}{
						"teach_state": model.SkiTeachStateAfterClass,
					}).Error
				if err != nil {
					return enum.NewErr(enum.OrdersCoursesExitErr, "更新教练课后缓冲时间失败")
				}
			}

			insrtocsData = model.OrdersCoursesState{
				OrderCourseID:  orderCourseId,
				UserID:         orderCourse.TeachCoachID,
				UserType:       model.UserTypeCoach,
				Operate:        model.OperateClubChangeUserCourseTime,
				Remark:         model.OCSOperateStr[model.OperateClubChangeUserCourseTime],
				TeachStartTime: model.LocalTime(startTi),
				TeachTimeIDs:   append(coachTimeIds, bufferTimeIds...),
				Process:        model.ProcessNo,
			}
			err = tx.Model(model.OrdersCoursesState{}).Create(&insrtocsData).Error
			if err != nil {
				global.Lg.Error("OrdersCoursesDao: 取消预约: %w", zap.Error(err), zap.Any("req", req), zap.Any("insrtocsData", insrtocsData))
				return err
			}
		}
		return nil
	})
	if err != nil {
		global.Lg.Error("OrdersCoursesDao: 确认课程: %w", zap.Error(err), zap.Any("req", req))
		return err
	}
	return nil
}

func ClubAppointmentCourse(c *gin.Context, orderCourseId string, req *forms.ClubAppointmentCourseRequest) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	var orderCourse model.OrdersCourses
	err = global.DB.Model(model.OrdersCourses{}).Preload("CourseTags", "state=0").
		Where("order_course_id=? and state=0", orderCourseId).
		First(&orderCourse).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "该课程不存在")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState == model.TeachStateWaitCoachConfirmClub {
		return enum.NewErr(enum.OrdersCoursesExitErr, "已经分配教练，请联系教练确认")
	}
	if orderCourse.TeachState != model.TeachStateWaitClubConfirm {
		return enum.NewErr(enum.OrdersCoursesExitErr, "当前状态不能分配教练")
	}
	order := model.Orders{}
	err = global.DB.Model(&order).
		Where("order_id=? and user_id=? and user_type=? and state=0", orderCourse.OrderID, userId, userType).First(&order).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "不是你的订单")
	}
	if order.Status != enum.OrderStatusPaid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单未支付")
	}

	var orderTagIds []int64
	for _, v := range orderCourse.CourseTags {
		orderTagIds = append(orderTagIds, v.TagID)
	}
	coach, err := CheckCoachTag(req.CoachId, orderTagIds)
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
			coach.CoachId, model.UserTypeCoach, orderCourse.SkiResortsID, teachStartTimes, model.SkiTeachStateWaitAppointment).
		Pluck("id", &ids)
	if len(ids) != teachCount {
		return enum.NewErr(enum.OrdersCoursesExitErr, "教练时间冲突，请重新预约")
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourseId).
			Updates(map[string]interface{}{
				"teach_state":    model.TeachStateWaitCoachConfirmClub,
				"teach_coach_id": coach.CoachId,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "课程修改状态失败")
		}
		err = SRTOrderCourses(tx, ids, orderCourse.OrderCourseID, 0)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return err
		}
		insrtocsData := model.OrdersCoursesState{
			OrderCourseID:  orderCourseId,
			UserID:         userId,
			UserType:       userType,
			Operate:        model.OperateClubAppointCoach,
			Remark:         model.OCSOperateStr[model.OperateClubAppointCoach],
			TeachStartTime: orderCourse.TeachStartTime,
			TeachTimeIDs:   ids,
			Process:        model.ProcessNo,
			CoachID:        coach.CoachId,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&insrtocsData).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "俱乐部插入记录失败")
		}
		return nil
	})
	if err != nil {
		global.Lg.Error("OrdersCoursesDao: 确认课程: %w", zap.Error(err), zap.Any("req", req))
	}

	return err
}

func ClubReplaceCoachCourse(c *gin.Context, orderCourseId string, req *forms.ClubReplaceCoachCourseRequest) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")

	var orderCourse model.OrdersCourses
	err = global.DB.Model(model.OrdersCourses{}).Preload("CourseTags", "state=0").
		Where("order_course_id=? and state=0", orderCourseId).
		First(&orderCourse).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "该课程不存在")
	}

	if orderCourse.TeachCoachID == req.CoachId {
		return enum.NewErr(enum.OrdersCoursesExitErr, "不能选择当前教练")
	}
	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}

	if orderCourse.TeachState == model.TeachStateWaitCoachConfirmTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "已经分配教练，请联系教练确认")
	}
	if orderCourse.TeachState != model.TeachStateCoachApplyTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "先让教练申请转课")
	}

	ordersCoursesState := model.OrdersCoursesState{}
	err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and  operate = ? and process = ?",
		orderCourseId, model.OperateCoachApplyTransferCourse, model.ProcessNo).
		Last(&ordersCoursesState).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "教练申请已过期")
	}
	if ordersCoursesState.CreatedAt.Before(time.Now().Add(-time.Hour * 24)) {
		return enum.NewErr(enum.OrdersCoursesExitErr, "教练申请已过期，请让教练重新申请")
	}

	order := model.Orders{}
	err = global.DB.Model(&order).
		Where("order_id=? and user_id=? and user_type=? and state=0", orderCourse.OrderID, userId, userType).First(&order).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "不是你的订单")
	}
	if order.Status != enum.OrderStatusPaid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单未支付")
	}

	if order.UserID != userId || order.UserType != userType {
		return enum.NewErr(enum.OrdersCoursesExitErr, "不是你的订单")
	}

	var orderTagIds []int64
	for _, v := range orderCourse.CourseTags {
		orderTagIds = append(orderTagIds, v.TagID)
	}
	coach, err := CheckCoachTag(req.CoachId, orderTagIds)
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
			coach.CoachId, model.UserTypeCoach, orderCourse.SkiResortsID, teachStartTimes, model.SkiTeachStateWaitAppointment).
		Pluck("id", &ids)
	if len(ids) != teachCount {
		return enum.NewErr(enum.OrdersCoursesExitErr, "教练时间冲突，请重新预约")
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourseId).
			Updates(map[string]interface{}{
				"teach_state":    model.TeachStateWaitCoachConfirmTransfer,
				"teach_coach_id": req.CoachId,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "课程修改状态失败")
		}
		err = SRTOrderCourses(tx, ids, orderCourseId, 0)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return err
		}
		insrtocsData := model.OrdersCoursesState{
			OrderCourseID:  orderCourseId,
			UserID:         userId,
			UserType:       userType,
			Operate:        model.OperateClubTransferToCoach,
			Remark:         model.OCSOperateStr[model.OperateClubTransferToCoach],
			TeachStartTime: orderCourse.TeachStartTime,
			TeachTimeIDs:   ids,
			Process:        model.ProcessNo,
			CoachID:        coach.CoachId,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&insrtocsData).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "俱乐部插入记录失败")
		}
		return nil
	})
	if err != nil {
		global.Lg.Error("OrdersCoursesDao: 确认课程: %w", zap.Error(err), zap.Any("req", req))
	}

	return err
}
