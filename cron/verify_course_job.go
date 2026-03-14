package cron

import (
	"context"
	"go.uber.org/zap"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

type VerifyCourseJob struct {
}

func (m VerifyCourseJob) Run() {
	var data []model.OrdersCourses
	global.DB.Model(model.OrdersCourses{}).
		Where("teach_start_time between now() - interval 10 day and now() - interval 7 day").
		Where(" is_check = ? and state = 0", model.IsCheckNo).
		Where("teach_state in ?", []model.TeachState{model.TeachStateAlreadyClass, model.TeachStateWaitCheck, model.TeachStateWaitCoachClass, model.TeachStateWaitClass, model.TeachStateWaitClassTransfer}).
		Find(&data)
	if len(data) == 0 {
		return
	}

	for _, orderCourse := range data {
		order := model.Orders{}
		err := global.DB.Model(model.Orders{}).Where("order_id = ? and state = 0", orderCourse.OrderID).First(&order).Error
		if err != nil {
			global.Lg.Error("没有找到订单", zap.Error(err), zap.Any("orderCourse", orderCourse))
			continue
		}

		if order.FrozenMoney == 1 {
			global.Lg.Error("订单已冻结，不能核销", zap.Error(err), zap.Any("orderCourse", orderCourse))
			continue
		}

		insrtocsData := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        enum.UserIdCron,
			UserType:      enum.UserTypeCron,
			Operate:       model.OperateCronVerifyCourse,
			Remark:        model.OCSOperateStr[model.OperateCronVerifyCourse],
			Process:       model.ProcessYes,
		}
		err = dao.CompleteCourseSplitMoney(context.Background(), orderCourse, order, insrtocsData)
		if err != nil {
			global.Lg.Error("拆分课程金额失败", zap.Error(err), zap.Any("orderCourse", orderCourse))
			continue
		}
	}
}
