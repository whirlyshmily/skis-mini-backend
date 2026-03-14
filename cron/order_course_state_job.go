package cron

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

type OrdersCoursesStateJob struct {
}

func (m OrdersCoursesStateJob) Run() {
	c := context.Background()
	var data48 []model.OrdersCoursesState
	global.DB.Model(model.OrdersCoursesState{}).
		Where("created_at between now() - interval 30 day and now() - interval 2 day").
		Where("operate in ? and process = ?", []model.TeachState{model.OperateUserAgreeCoachTransferCourse, model.OperateCoachApplyTransferCourse}, model.ProcessNo).
		Find(&data48)

	for _, ocs := range data48 {
		_, orderCourse, err := dao.GetOrderCourses(ocs.OrderCourseID)
		if err != nil {
			global.Lg.Error("获取订单课程失败", zap.Error(err))
			continue
		}
		if ocs.Operate == model.OperateUserAgreeCoachTransferCourse { //用户同意教练转让课程，教练没确认课程，2天后自动取消
			cronCancelUserAgreeTransferOrder(c, orderCourse, ocs)
		}

		if ocs.Operate == model.OperateCoachApplyTransferCourse { //教练申请转让课程，用户没确认转让，2天后自动取消
			cronCancelCoachApplyTransferOrder(c, orderCourse, ocs)
		}
	}

	var data []model.OrdersCoursesState
	global.DB.Model(model.OrdersCoursesState{}).
		Where("created_at between now() - interval 2 day and now() - interval 1 day").
		Where(" process = ?", model.ProcessNo).
		Find(&data)
	for _, ocs := range data {
		OrdersCoursesStateJobProcessData(c, ocs)
	}

}

func OrdersCoursesStateJobProcessData(c context.Context, ocs model.OrdersCoursesState) {
	order, orderCourse, err := dao.GetOrderCourses(ocs.OrderCourseID)
	if err != nil {
		global.Lg.Error("获取订单课程失败", zap.Error(err))
		return
	}

	if ocs.Operate == model.OperateUserAppointment { //预约课程，教练或者俱乐部没确认课程，1天后自动取消
		cronCancelCourse(c, order, orderCourse, ocs)
	}
	if ocs.Operate == model.OperateCoachChangeCourseTime { //预约课程，教练没确认课程时修改了上课时间，需要学员确认，如果学员不确认，1天后自动取消
		cronCancelCoachChangeCourseTime(c, orderCourse, ocs)
	}
	if ocs.Operate == model.OperateClubAppointCoach { //俱乐部安排教练，教练没确认课程，1天后自动取消
		cronCancelOrderFromClub(c, orderCourse, ocs)
	}
	if ocs.Operate == model.OperateCoachChangeCourse { //教练修改上课时间，用户没确认课程，1天后自动取消
		cronCancelCoachChangeTeachTime(c, orderCourse, ocs)
	}
	if ocs.Operate == model.OperateClubChangeUserCourseTime { // 俱乐部修改用户课程时间，用户没确认课程，1天后自动取消
		cronCancelClubChangeTeachTime(c, order, orderCourse, ocs)
	}
	if ocs.Operate == model.OperateCoachTransferCourse { //教练转让课程，用户没确认课程，1天后自动取消
		cronCancelCoachTransferCourse(c, orderCourse)
	}
	if ocs.Operate == model.OperateCoachTransferToCoach { //教练转让课程给其他教练，用户没确认课程，1天后自动取消
		cronCancelCoachTransferOrder(c, orderCourse, ocs)
	}

	if ocs.Operate == model.OperateClubTransferToCoach { //俱乐部转移课程给教练，24小时后取消
		cronCancelClubTransferToCoach(c, orderCourse, ocs)
	}
}

// cronCancelCourse 定时任务取消课程函数
// 该函数用于处理定时取消课程的逻辑，根据订单课程的不同状态执行相应的取消操作
//
// 参数:
//
//	c: 上下文对象，用于传递请求作用域的数据和超时控制
//	order: 订单信息对象，包含订单的基本信息和用户类型
//	orderCourse: 订单课程信息对象，包含课程的具体状态和教学信息
//	ocs: 订单课程状态对象，用于更新处理状态
func cronCancelCourse(c context.Context, order model.Orders, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 检查课程状态，如果不是等待教练确认用户或等待俱乐部确认的状态，则直接更新处理状态为已完成
	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmUser && orderCourse.TeachState != model.TeachStateWaitClubConfirm {
		global.DB.Model(model.OrdersCoursesState{}).Where("id = ? and process=?",
			ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	// 构造订单课程状态记录，用于记录定时任务取消课程的操作日志
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		TeachTimeIDs:  model.JSONIntArray{},
		Operate:       model.OperateCronCancelCourse,
		Remark:        model.OCSOperateStr[model.OperateCronCancelCourse],
		Process:       model.ProcessYes,
	}

	var err error
	// 根据订单用户类型执行不同的取消课程SQL操作
	if order.UserType == enum.UserTypeCoach {
		err = global.DB.Transaction(func(tx *gorm.DB) error {
			return dao.CancelCoachCourseSql(c, tx, order, orderCourse, inspectorate)
		})
		if err != nil {
			global.Lg.Error("定时任务取消课程失败", zap.Error(err))
		}
	}
	if order.UserType == enum.UserTypeClub {
		err = global.DB.Transaction(func(tx *gorm.DB) error {
			return dao.CancelClubCourseSql(c, tx, order, orderCourse, inspectorate)
		})
		if err != nil {
			global.Lg.Error("定时任务取消课程失败", zap.Error(err))
		}
	}

	// 如果执行过程中出现错误，记录错误日志
	if err != nil {
		global.Lg.Error("定时任务取消课程失败", zap.Error(err))
	}
	return
}

// cronCancelCoachChangeCourseTime 定时任务取消教练修改课程时间
// 该函数用于处理教练修改课程时间的超时情况，如果用户未在规定时间内确认教练提出的时间变更，
// 系统将自动取消该时间变更请求
//
// 参数:
//
//	c: 上下文对象，用于传递请求作用域的数据和控制超时
//	orderCourse: 订单课程信息，包含当前课程的状态和详情
//	ocs: 订单课程状态信息，记录操作流程的状态
func cronCancelCoachChangeCourseTime(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 检查课程教学状态，如果不是等待用户确认教练时间的状态，则直接更新状态为已完成
	if orderCourse.TeachState != model.TeachStateWaitUserSecondConfirmTime {
		global.DB.Model(model.OrdersCoursesState{}).Where("id = ? and process=?",
			ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	// 构造定时任务的操作记录信息
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelCoachChangeCourseTime,
		Remark:        model.OCSOperateStr[model.OperateCronCancelCoachChangeCourseTime],
		Process:       model.ProcessYes,
	}

	// 执行用户拒绝教练教学时间的数据库操作
	err := dao.UserDisAgreeCoachTeachTimeSql(c, orderCourse, inspectorate, ocs)

	if err != nil {
		global.Lg.Error("定时任务取消教练修改课程时间失败", zap.Error(err))
	}
	return
}

// cronCancelOrderFromClub 定时任务取消俱乐部安排给教练的课程
// c: 上下文对象
// orderCourse: 订单课程信息
// ocs: 订单课程状态信息
func cronCancelOrderFromClub(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 如果教学状态不是等待教练确认俱乐部安排，则直接更新订单课程状态为已处理
	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmClub {
		global.DB.Model(model.OrdersCoursesState{}).Where("id = ? and process=?",
			ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	// 构造检查员操作记录
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelClubCourse,
		Remark:        model.OCSOperateStr[model.OperateCronCancelClubCourse],
		Process:       model.ProcessYes,
	}

	// 执行教练拒绝俱乐部课程安排的SQL操作
	err := dao.CoachDisAgreeOrderFromClubSql(c, orderCourse, inspectorate, ocs)

	if err != nil {
		global.Lg.Error("定时任务取消俱乐部安排给教练的课程失败", zap.Error(err))
	}
	return
}

// cronCancelCoachChangeTeachTime 定时任务取消教练修改课程时间
// 该函数用于处理教练修改课程时间的超时情况，如果用户未在规定时间内确认教练提出的时间变更，
// 系统将自动取消该时间变更请求
//
// 参数:
//
//	c: 上下文对象，用于传递请求作用域的数据和控制超时
//	orderCourse: 订单课程信息，包含当前课程的状态和详情
//	ocs: 订单课程状态信息，记录操作流程的状态
func cronCancelCoachChangeTeachTime(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 检查课程教学状态，如果不是等待用户确认教练时间的状态，则直接更新状态为已完成
	if orderCourse.TeachState != model.TeachStateWaitUserConfirmCoachTime {
		global.DB.Model(model.OrdersCoursesState{}).Where("id = ? and process=?",
			ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	// 构造定时任务的操作记录信息
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelCoachChangeCourse,
		Remark:        model.OCSOperateStr[model.OperateCronCancelCoachChangeCourse],
		Process:       model.ProcessYes,
	}

	// 执行用户拒绝教练教学时间的数据库操作
	err := dao.UserDisAgreeCoachTeachTimeSql(c, orderCourse, inspectorate, ocs)

	if err != nil {
		global.Lg.Error("定时任务取消教练修改课程时间失败", zap.Error(err))
	}
	return
}

// cronCancelClubChangeTeachTime 是一个定时任务函数，用于自动取消俱乐部申请的课程时间修改操作。
// 该函数会在满足条件时查找俱乐部和教练相关的未处理的时间修改申请，并执行取消操作。
//
// 参数说明：
// c: 上下文 context，用于控制请求的生命周期。
// order: 订单信息结构体，包含订单的基本信息。
// orderCourse: 订单课程信息结构体，包含具体课程的信息。
// ocs: 订单课程状态信息结构体，表示当前操作的状态信息。
func cronCancelClubChangeTeachTime(c context.Context, order model.Orders, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 如果当前用户是教练，则不执行任何操作，直接返回
	if ocs.UserType == enum.UserTypeCoach {
		return
	}

	if orderCourse.TeachState != model.TeachStateWaitUserConfirmClubTime {
		global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
			ocs.ID, model.OperateClubChangeUserCourseTime, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	// 查找俱乐部提交的修改课程时间申请（状态为未处理）
	ocsClubData := model.OrdersCoursesState{}
	err := global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and user_id = ? and operate=? and process = ? and state=0",
		orderCourse.OrderCourseID, order.UserID, model.OperateClubChangeUserCourseTime, model.ProcessNo).
		Last(&ocsClubData).Error
	if err != nil {
		global.Lg.Error("定时任务取消俱乐部修改课程时间失败", zap.Error(err))
		return
	}

	// 如果存在授课教练，也查找教练对应的修改课程时间申请（状态为未处理）
	ocsCoachData := model.OrdersCoursesState{}
	if orderCourse.TeachCoachID != "" {
		err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and user_id = ? and operate=? and process = ? and state=0",
			orderCourse.OrderCourseID, orderCourse.TeachCoachID, model.OperateClubChangeUserCourseTime, model.ProcessNo).
			Last(&ocsCoachData).Error
		if err != nil {
			global.Lg.Error("定时任务取消俱乐部修改课程时间失败", zap.Error(err))
			return
		}
	}

	// 构造系统自动取消操作的记录信息
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelClubChangeCourseTime,
		Remark:        model.OCSOperateStr[model.OperateCronCancelClubChangeCourseTime],
		Process:       model.ProcessYes,
	}

	// 调用数据库操作函数，执行取消俱乐部修改课程时间的操作
	err = dao.UserDisAgreeClubTeachTimeSql(c, orderCourse, inspectorate, ocsClubData.TeachTimeIDs, ocsCoachData.TeachTimeIDs)
	if err != nil {
		global.Lg.Error("定时任务取消俱乐部修改课程时间失败", zap.Error(err))
	}
	return
}

// cronCancelCoachTransferCourse 定时任务取消教练转让课程
// 该函数用于处理超时未处理的教练转让课程订单，自动取消转让操作
//
// 参数:
//
//	c: 上下文对象，用于传递超时和取消信号
//	orderCourse: 订单课程对象，包含需要取消转让的课程订单信息
func cronCancelCoachTransferCourse(c context.Context, orderCourse model.OrdersCourses) {
	if orderCourse.TeachState != model.TeachStateWaitUserConfirmTransfer {
		global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
			orderCourse.OrderCourseID, model.OperateCoachTransferCourse, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	// 构造审核记录对象，记录定时任务自动取消的操作
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelCoachTransferCourse,
		Remark:        model.OCSOperateStr[model.OperateCronCancelCoachTransferCourse],
		Process:       model.ProcessYes,
	}

	// 执行取消教练转让课程的数据库操作
	err := dao.CancelCoachTransferOrderSql(c, orderCourse, inspectorate)
	if err != nil {
		global.Lg.Error("定时任务取消教练转让课程失败", zap.Error(err))
	}
	return
}

func cronCancelCoachTransferOrder(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	if orderCourse.TeachState != model.TeachStateWaitConfirmTransfer {
		global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
			orderCourse.OrderCourseID, model.OperateCoachTransferToCoach, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelTransferCourseToCoach,
		Remark:        model.OCSOperateStr[model.OperateCronCancelTransferCourseToCoach],
		Process:       model.ProcessYes,
	}
	err := dao.CancelOrderFromCoachSql(c, inspectorate, ocs)
	if err != nil {
		global.Lg.Error("定时任务取消教练转让课程失败", zap.Error(err))
	}
	err = global.DB.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", orderCourse.OrderCourseID).
		Updates(map[string]interface{}{
			"teach_state": model.TeachStateWaitClass,
		}).Error
	return
}

func cronCancelClubTransferToCoach(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmTransfer {
		global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
			orderCourse.OrderCourseID, model.OperateCronCancelClubTransferToCoach, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelClubTransferToCoach,
		Remark:        model.OCSOperateStr[model.OperateCronCancelClubTransferToCoach],
		Process:       model.ProcessYes,
	}
	err := dao.CancelOrderFromCoachSql(c, inspectorate, ocs)
	if err != nil {
		global.Lg.Error("定时任务取消俱乐部转移课程给教练失败", zap.Error(err))
	}
	err = global.DB.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", orderCourse.OrderCourseID).
		Updates(map[string]interface{}{
			"teach_state": model.TeachStateCoachApplyTransfer,
		}).Error
	return
}

// cronCancelUserAgreeTransferOrder 定时取消用户同意转让课程
func cronCancelUserAgreeTransferOrder(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	if orderCourse.TeachState == model.TeachStateWaitCoachTransfer {
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        enum.UserIdCron,
			UserType:      enum.UserTypeCron,
			Operate:       model.OperateCronCancelTransferCourseToCoach,
			Remark:        model.OCSOperateStr[model.OperateCronCancelTransferCourseToCoach],
			Process:       model.ProcessYes,
		}

		//找出最新的操作记录，如果不是教练转单记录，则返回错误
		/*orderCourseState := model.OrdersCoursesState{}
		err := global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=? and state=0",
			orderCourse.OrderCourseID, model.OperateCoachTransferToCoach, model.ProcessNo).
			Last(&orderCourseState).Error
		if err != nil {
			global.Lg.Error("定时任务找不到教练转单记录", zap.Error(err))
			return
		}*/

		err := dao.CancelOrderFromCoachSql(c, inspectorate, ocs)
		if err != nil {
			global.Lg.Error("定时任务取消教练转让课程失败", zap.Error(err))
			return
		}
		err = global.DB.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", orderCourse.OrderCourseID).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateWaitClass,
			}).Error
	}

	global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate in ? and process=?",
		orderCourse.OrderCourseID, []int{model.OperateCoachTransferToCoach, model.OperateUserAgreeCoachTransferCourse}, model.ProcessNo).
		Updates(map[string]interface{}{
			"process": model.ProcessYes,
		})
	return
}

// cronCancelCoachApplyTransferOrder 定时任务取消教练申请转单课程
// 该函数用于处理教练申请转让课程的超时情况，如果用户未在规定时间内确认教练的转让申请，
// 系统将自动取消该转让申请
//
// 参数:
//
//	c: 上下文对象，用于传递超时和取消信号
//	orderCourse: 订单课程对象，包含需要取消转让申请的课程订单信息
//	ocs: 订单课程状态对象，记录操作流程的状态
func cronCancelCoachApplyTransferOrder(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 检查教学状态，如果不是等待教练确认转让的状态，则直接更新处理状态为已完成
	if orderCourse.TeachState != model.TeachStateCoachApplyTransfer {
		global.DB.Model(model.OrdersCoursesState{}).Where("id = ? and process=?",
			ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			})
		return
	}

	// 构造审核记录对象，记录定时任务自动取消的操作
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelCoachApplyTransferCourse,
		Remark:        model.OCSOperateStr[model.OperateCronCancelCoachApplyTransferCourse],
		Process:       model.ProcessYes,
	}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "插入记录失败")
		}

		err = tx.Model(model.OrdersCoursesState{}).Where("id = ?", ocs.ID).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "状态更新失败")
		}
		// 更新订单课程状态回到教练申请转单前的状态
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ? and state=0", orderCourse.OrderCourseID).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateWaitCoachClass,
			}).Error

		if err != nil {
			global.Lg.Error("更新订单课程状态失败", zap.Error(err))
		}
		return err
	})
	if err != nil {
		global.Lg.Error("成功取消教练申请转让课程", zap.Error(err))
	}

}
