package cron

import (
	"context"
	"go.uber.org/zap"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"
)

type OrdersCoursesJob struct {
}

func (m OrdersCoursesJob) Run() {
	c := context.Background()
	// 处理 teach_state=10 的 OrderCourse 数据，重置为 0
	m.processTeachStateTenToZero(c)

	// 处理 teach_state=100 且是明天之前的数据，修改为 teach_state=300
	m.processTeachStateHundredToThreeHundred(c)

	// 处理 teach_state=310 且是 7 天前的数据，修改为 teach_state=400
	m.processTeachStateThreeHundredTenToFourHundred(c)
}

// processTeachStateTenToZero 处理 teach_state=10 的订单课程，将其重置为 0
func (m OrdersCoursesJob) processTeachStateTenToZero(c context.Context) {
	// 查询 teach_state=10 的订单课程
	today := time.Now().Format("2006-01-02")

	//将今天之前的待核销的课程状态改为已上课
	err := global.DB.Table("orders_courses").
		Where("teach_start_time < ?", today).
		Where(" is_check = ? and state = 0", model.IsCheckNo).
		Where("teach_state in ?", []model.TeachState{model.TeachStateWaitCheck}).
		Updates(map[string]interface{}{
			"teach_state": model.TeachStateAlreadyClass,
		}).Error
	if err != nil {
		global.Lg.Error("查询 teach_state=10 的订单课程失败", zap.Error(err))
	}

	err = global.DB.Model(model.OrdersCourses{}).
		Where("teach_state in ?", []model.TeachState{model.TeachStateWaitCoachConfirmUser, model.TeachStateWaitClubConfirm}).
		Where("teach_start_time < ?", today).
		Update("teach_state", model.TeachStateWaitAppointment).Error

	if err != nil {
		global.Lg.Error("查询 teach_state=10 的订单课程失败", zap.Error(err))
		return
	}
}

// processTeachStateHundredToThreeHundred 处理 teach_state=100 且是明天之前的订单课程，将其修改为 teach_state=300
func (m OrdersCoursesJob) processTeachStateHundredToThreeHundred(c context.Context) {
	// 获取明天的日期
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	// 查询 teach_state=100 且 teach_start_time 在明天之前的订单课程
	err := global.DB.Model(model.OrdersCourses{}).
		Where("teach_state in ?", []model.TeachState{model.TeachStateWaitClass, model.TeachStateWaitCoachClass, model.TeachStateWaitClassTransfer}).
		Where("teach_start_time < ?", tomorrow).
		Update("teach_state", model.TeachStateWaitCheck).Error

	if err != nil {
		global.Lg.Error("处理 teach_state=100 的订单课程失败", zap.Error(err))
		return
	}

	global.Lg.Info("成功处理 teach_state=100 的订单课程为 teach_state=300")
}

// processTeachStateThreeHundredTenToFourHundred 处理 teach_state=310 且是 7 天前的订单课程，将其修改为 teach_state=400
func (m OrdersCoursesJob) processTeachStateThreeHundredTenToFourHundred(c context.Context) {
	// 获取 7 天前的日期
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	// 查询 teach_state=310 且 teach_start_time 在 7 天前的订单课程
	err := global.DB.Model(model.OrdersCourses{}).
		Where("teach_state = ?", model.TeachStateAlreadyClass).
		Where("teach_start_time < ?", sevenDaysAgo).
		Update("teach_state", model.TeachStateFinish).Error

	if err != nil {
		global.Lg.Error("处理 teach_state=310 的订单课程失败", zap.Error(err))
		return
	}

	global.Lg.Info("成功处理 teach_state=310 的订单课程为 teach_state=400")
}
