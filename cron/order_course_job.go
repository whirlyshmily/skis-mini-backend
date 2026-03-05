package cron

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"
)

type OrdersCoursesStateJob struct {
}

func (m OrdersCoursesStateJob) Run() {
	// 创建带超时的上下文，防止长时间阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 合并查询：一次性获取两天内和三天内的数据
	var data []model.OrdersCoursesState
	err := global.DB.Model(model.OrdersCoursesState{}).
		Where("(created_at BETWEEN NOW() - INTERVAL 2 DAY AND NOW() - INTERVAL 1 DAY AND process = ?) OR "+
			"(created_at BETWEEN NOW() - INTERVAL 3 DAY AND NOW() - INTERVAL 2 DAY AND operate = ? AND process = ?)",
			model.ProcessNo, model.OperateUserAgreeCoachTransferCourse, model.ProcessNo).
		Find(&data).Error

	if err != nil {
		global.Lg.Error("查询订单课程状态失败", zap.Error(err))
		return
	}

	// 分类处理数据
	for _, ocs := range data {
		// 处理两天内的数据
		if ocs.Process == model.ProcessNo &&
			ocs.CreatedAt.After(time.Now().AddDate(0, 0, -2)) &&
			ocs.CreatedAt.Before(time.Now().AddDate(0, 0, -1)) {
			OrdersCoursesStateJobProcessData(ctx, ocs)
		}

		// 处理三天内用户同意教练转让课程的数据
		if ocs.Operate == model.OperateUserAgreeCoachTransferCourse && ocs.Process == model.ProcessNo &&
			ocs.CreatedAt.After(time.Now().AddDate(0, 0, -3)) &&
			ocs.CreatedAt.Before(time.Now().AddDate(0, 0, -2)) {
			_, orderCourse, err := dao.GetOrderCourses(ocs.OrderCourseID)
			if err != nil {
				global.Lg.Error("获取订单课程失败", zap.Error(err), zap.String("order_course_id", ocs.OrderCourseID))
				continue
			}
			//用户同意教练转让课程，教练没确认课程，2天后自动取消
			cronCancelUserAgreeTransferOrder(ctx, orderCourse, ocs)
		}
	}

	// 使用结构化日志替代 fmt.Println
	global.Lg.Info("定时任务执行完成", zap.Int("processed_count", len(data)))
}
func OrdersCoursesStateJobProcessData(c context.Context, ocs model.OrdersCoursesState) {
	// 检查上下文是否已取消或超时
	if c.Err() != nil {
		global.Lg.Warn("上下文已取消或超时", zap.Error(c.Err()))
		return
	}

	// 获取订单课程信息
	order, orderCourse, err := dao.GetOrderCourses(ocs.OrderCourseID)
	if err != nil {
		global.Lg.Error("获取订单课程失败", zap.Error(err), zap.String("order_course_id", ocs.OrderCourseID))
		return
	}

	// 定义操作类型与处理函数的映射
	operationHandlers := map[int]func(context.Context, model.Orders, model.OrdersCourses, model.OrdersCoursesState){
		model.OperateUserAppointment: func(ctx context.Context, o model.Orders, oc model.OrdersCourses, state model.OrdersCoursesState) {
			cronCancelCourse(ctx, o, oc, state)
		},
		model.OperateClubAppointCoach: func(ctx context.Context, o model.Orders, oc model.OrdersCourses, state model.OrdersCoursesState) {
			cronCancelOrderFromClub(ctx, oc, state)
		},
		model.OperateCoachChangeCourse: func(ctx context.Context, o model.Orders, oc model.OrdersCourses, state model.OrdersCoursesState) {
			cronCancelCoachChangeTeachTime(ctx, oc, state)
		},
		model.OperateClubChangeUserCourseTime: func(ctx context.Context, o model.Orders, oc model.OrdersCourses, state model.OrdersCoursesState) {
			cronCancelClubChangeTeachTime(ctx, o, oc, state)
		},
		model.OperateCoachTransferCourse: func(ctx context.Context, o model.Orders, oc model.OrdersCourses, state model.OrdersCoursesState) {
			cronCancelCoachTransferCourse(ctx, oc)
		},
		model.OperateCoachTransferToCoach: func(ctx context.Context, o model.Orders, oc model.OrdersCourses, state model.OrdersCoursesState) {
			cronCancelCoachTransferOrder(ctx, oc, state)
		},
	}

	// 根据操作类型执行对应处理函数
	if handler, exists := operationHandlers[ocs.Operate]; exists {
		global.Lg.Info("开始处理订单课程状态任务", zap.Int("operate", ocs.Operate), zap.String("order_course_id", ocs.OrderCourseID))
		handler(c, order, orderCourse, ocs)
	} else {
		global.Lg.Warn("未知的操作类型", zap.Int("operate", ocs.Operate), zap.String("order_course_id", ocs.OrderCourseID))
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
	// 检查上下文是否已取消或超时
	if c.Err() != nil {
		global.Lg.Warn("上下文已取消或超时", zap.Error(c.Err()), zap.String("order_course_id", orderCourse.OrderCourseID))
		return
	}

	// 检查课程状态，如果不是等待教练确认用户或等待俱乐部确认的状态，则直接更新处理状态为已完成
	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmUser && orderCourse.TeachState != model.TeachStateWaitClubConfirm {
		err := global.DB.Model(model.OrdersCoursesState{}).
			Where("id = ? AND process = ?", ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{"process": model.ProcessYes}).Error
		if err != nil {
			global.Lg.Error("更新订单课程状态失败", zap.Error(err), zap.Uint("ocs_id", uint(ocs.ID)))
		}
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

	// 根据订单用户类型执行不同的取消课程SQL操作
	var cancelFunc func(*gorm.DB) error
	switch order.UserType {
	case enum.UserTypeCoach:
		cancelFunc = func(tx *gorm.DB) error {
			return dao.CancelCoachCourseSql(c, tx, order, orderCourse, inspectorate)
		}
	case enum.UserTypeClub:
		cancelFunc = func(tx *gorm.DB) error {
			return dao.CancelClubCourseSql(c, tx, order, orderCourse, inspectorate)
		}
	default:
		global.Lg.Error("未知的用户类型", zap.Int("user_type", order.UserType), zap.String("order_course_id", orderCourse.OrderCourseID))
		return
	}

	// 执行事务
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		return cancelFunc(tx)
	})

	// 记录错误日志
	if err != nil {
		global.Lg.Error("定时任务取消课程失败",
			zap.Error(err),
			zap.String("order_course_id", orderCourse.OrderCourseID),
			zap.Int("user_type", order.UserType))
	}
}

// cronCancelOrderFromClub 定时任务取消俱乐部安排给教练的课程
// c: 上下文对象
// orderCourse: 订单课程信息
// ocs: 订单课程状态信息
func cronCancelOrderFromClub(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 检查上下文是否已取消或超时
	if c.Err() != nil {
		global.Lg.Warn("上下文已取消或超时", zap.Error(c.Err()), zap.String("order_course_id", orderCourse.OrderCourseID))
		return
	}

	// 如果教学状态不是等待教练确认俱乐部安排，则直接更新订单课程状态为已处理
	if orderCourse.TeachState != model.TeachStateWaitCoachConfirmClub {
		err := global.DB.Model(model.OrdersCoursesState{}).
			Where("id = ? AND process = ?", ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{"process": model.ProcessYes}).Error
		if err != nil {
			global.Lg.Error("更新订单课程状态失败", zap.Error(err), zap.Uint("ocs_id", uint(ocs.ID)))
		}
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
		global.Lg.Error("定时任务取消俱乐部安排给教练的课程失败",
			zap.Error(err),
			zap.String("order_course_id", orderCourse.OrderCourseID),
			zap.Int("teach_state", int(orderCourse.TeachState)))
	}
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
	// 检查上下文是否已取消或超时
	if c.Err() != nil {
		global.Lg.Warn("上下文已取消或超时", zap.Error(c.Err()), zap.String("order_course_id", orderCourse.OrderCourseID))
		return
	}

	// 检查课程教学状态，如果不是等待用户确认教练时间的状态，则直接更新状态为已完成
	if orderCourse.TeachState != model.TeachStateWaitUserConfirmCoachTime {
		err := global.DB.Model(model.OrdersCoursesState{}).
			Where("id = ? AND process = ?", ocs.ID, model.ProcessNo).
			Updates(map[string]interface{}{"process": model.ProcessYes}).Error
		if err != nil {
			global.Lg.Error("更新订单课程状态失败", zap.Error(err), zap.Uint("ocs_id", uint(ocs.ID)))
		}
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
		global.Lg.Error("定时任务取消教练修改课程时间失败",
			zap.Error(err),
			zap.String("order_course_id", orderCourse.OrderCourseID),
			zap.Int("teach_state", int(orderCourse.TeachState)))
	}
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
	// 检查上下文是否已取消或超时
	if c.Err() != nil {
		global.Lg.Warn("上下文已取消或超时", zap.Error(c.Err()), zap.String("order_course_id", orderCourse.OrderCourseID))
		return
	}

	// 如果当前用户是教练，则不执行任何操作，直接返回
	if ocs.UserType == enum.UserTypeCoach {
		return
	}

	// 检查课程状态，如果不是等待用户确认俱乐部时间的状态，则直接更新状态为已完成
	if orderCourse.TeachState != model.TeachStateWaitUserConfirmClubTime {
		err := global.DB.Model(model.OrdersCoursesState{}).
			Where("order_course_id = ? AND operate = ? AND process = ?", ocs.ID, model.OperateClubChangeUserCourseTime, model.ProcessNo).
			Updates(map[string]interface{}{"process": model.ProcessYes}).Error
		if err != nil {
			global.Lg.Error("更新订单课程状态失败", zap.Error(err), zap.Uint("ocs_id", uint(ocs.ID)))
		}
		return
	}

	// 构造查询条件
	queryCondition := func(userID string) *gorm.DB {
		return global.DB.Model(model.OrdersCoursesState{}).
			Where("order_course_id = ? AND user_id = ? AND operate = ? AND process = ? AND state = 0",
				orderCourse.OrderCourseID, userID, model.OperateClubChangeUserCourseTime, model.ProcessNo)
	}

	// 查找俱乐部提交的修改课程时间申请（状态为未处理）
	ocsClubData := model.OrdersCoursesState{}
	err := queryCondition(order.UserID).Last(&ocsClubData).Error
	if err != nil {
		global.Lg.Error("定时任务取消俱乐部修改课程时间失败",
			zap.Error(err),
			zap.String("order_course_id", orderCourse.OrderCourseID),
			zap.String("user_id", order.UserID))
		return
	}

	// 如果存在授课教练，也查找教练对应的修改课程时间申请（状态为未处理）
	ocsCoachData := model.OrdersCoursesState{}
	if orderCourse.TeachCoachID != "" {
		err = queryCondition(orderCourse.TeachCoachID).Last(&ocsCoachData).Error
		if err != nil {
			global.Lg.Error("定时任务取消俱乐部修改课程时间失败",
				zap.Error(err),
				zap.String("order_course_id", orderCourse.OrderCourseID),
				zap.String("teach_coach_id", orderCourse.TeachCoachID))
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
		global.Lg.Error("定时任务取消俱乐部修改课程时间失败",
			zap.Error(err),
			zap.String("order_course_id", orderCourse.OrderCourseID))
	}
}

// cronCancelCoachTransferCourse 定时任务取消教练转让课程
// 该函数用于处理超时未处理的教练转让课程订单，自动取消转让操作
//
// 参数:
//
//	c: 上下文对象，用于传递超时和取消信号
//	orderCourse: 订单课程对象，包含需要取消转让的课程订单信息
func cronCancelCoachTransferCourse(c context.Context, orderCourse model.OrdersCourses) {
	// 检查上下文是否已取消或超时
	if c.Err() != nil {
		global.Lg.Warn("上下文已取消或超时", zap.Error(c.Err()), zap.String("order_course_id", orderCourse.OrderCourseID))
		return
	}

	// 检查课程状态，如果不是等待用户确认转让的状态，则直接更新状态为已完成
	if orderCourse.TeachState != model.TeachStateWaitUserConfirmTransfer {
		err := global.DB.Model(model.OrdersCoursesState{}).
			Where("order_course_id = ? AND operate = ? AND process = ?", orderCourse.OrderCourseID, model.OperateCoachTransferCourse, model.ProcessNo).
			Updates(map[string]interface{}{"process": model.ProcessYes}).Error
		if err != nil {
			global.Lg.Error("更新订单课程状态失败", zap.Error(err), zap.String("order_course_id", orderCourse.OrderCourseID))
		}
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
		global.Lg.Error("定时任务取消教练转让课程失败",
			zap.Error(err),
			zap.String("order_course_id", orderCourse.OrderCourseID),
			zap.Int("teach_state", int(orderCourse.TeachState)))
	}
}

func cronCancelCoachTransferOrder(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 检查上下文是否已取消或超时
	if c.Err() != nil {
		global.Lg.Warn("上下文已取消或超时", zap.Error(c.Err()), zap.String("order_course_id", orderCourse.OrderCourseID))
		return
	}

	// 检查课程状态，如果不是等待确认转让的状态，则直接更新状态为已完成
	if orderCourse.TeachState != model.TeachStateWaitConfirmTransfer {
		err := global.DB.Model(model.OrdersCoursesState{}).
			Where("order_course_id = ? AND operate = ? AND process = ?", orderCourse.OrderCourseID, model.OperateCoachTransferToCoach, model.ProcessNo).
			Updates(map[string]interface{}{"process": model.ProcessYes}).Error
		if err != nil {
			global.Lg.Error("更新订单课程状态失败", zap.Error(err), zap.String("order_course_id", orderCourse.OrderCourseID))
		}
		return
	}

	// 构造审核记录对象，记录定时任务自动取消的操作
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        enum.UserIdCron,
		UserType:      enum.UserTypeCron,
		Operate:       model.OperateCronCancelTransferCourseToCoach,
		Remark:        model.OCSOperateStr[model.OperateCronCancelTransferCourseToCoach],
		Process:       model.ProcessYes,
	}

	// 使用事务执行取消教练转让课程的操作
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 执行取消教练转让课程的SQL操作
		if err := dao.CancelOrderFromCoachSql(c, inspectorate, ocs); err != nil {
			return err
		}

		// 更新订单课程状态为等待上课
		if err := tx.Model(model.OrdersCourses{}).
			Where("order_course_id = ? AND state = 0", orderCourse.OrderCourseID).
			Updates(map[string]interface{}{"teach_state": model.TeachStateWaitClass}).Error; err != nil {
			return err
		}

		return nil
	})

	// 记录错误日志
	if err != nil {
		global.Lg.Error("定时任务取消教练转让课程失败",
			zap.Error(err),
			zap.String("order_course_id", orderCourse.OrderCourseID),
			zap.Int("teach_state", int(orderCourse.TeachState)))
	}
}

// cronCancelUserAgreeTransferOrder 定时取消用户同意转让课程
func cronCancelUserAgreeTransferOrder(c context.Context, orderCourse model.OrdersCourses, ocs model.OrdersCoursesState) {
	// 检查上下文是否已取消或超时
	if c.Err() != nil {
		global.Lg.Warn("上下文已取消或超时", zap.Error(c.Err()), zap.String("order_course_id", orderCourse.OrderCourseID))
		return
	}

	// 检查课程状态是否为等待教练转让
	if orderCourse.TeachState == model.TeachStateWaitCoachTransfer {
		// 构造审核记录对象，记录定时任务自动取消的操作
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourse.OrderCourseID,
			UserID:        enum.UserIdCron,
			UserType:      enum.UserTypeCron,
			Operate:       model.OperateCronCancelTransferCourseToCoach,
			Remark:        model.OCSOperateStr[model.OperateCronCancelTransferCourseToCoach],
			Process:       model.ProcessYes,
		}

		// 使用事务执行取消教练转让课程的操作
		err := global.DB.Transaction(func(tx *gorm.DB) error {
			// 执行取消教练转让课程的SQL操作
			if err := dao.CancelOrderFromCoachSql(c, inspectorate, ocs); err != nil {
				return err
			}

			// 更新订单课程状态为等待上课
			if err := tx.Model(model.OrdersCourses{}).
				Where("order_course_id = ? AND state = 0", orderCourse.OrderCourseID).
				Updates(map[string]interface{}{"teach_state": model.TeachStateWaitClass}).Error; err != nil {
				return err
			}

			return nil
		})

		// 记录错误日志
		if err != nil {
			global.Lg.Error("定时任务取消教练转让课程失败",
				zap.Error(err),
				zap.String("order_course_id", orderCourse.OrderCourseID),
				zap.Int("teach_state", int(orderCourse.TeachState)))
			return
		}
	}

	// 更新相关操作记录状态为已完成
	err := global.DB.Model(model.OrdersCoursesState{}).
		Where("order_course_id = ? AND operate IN ? AND process = ?", orderCourse.OrderCourseID,
			[]int{model.OperateCoachTransferToCoach, model.OperateUserAgreeCoachTransferCourse}, model.ProcessNo).
		Updates(map[string]interface{}{"process": model.ProcessYes}).Error

	// 记录更新操作的错误日志
	if err != nil {
		global.Lg.Error("更新订单课程状态失败",
			zap.Error(err),
			zap.String("order_course_id", orderCourse.OrderCourseID))
	}
}
