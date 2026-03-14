package dao

import (
	"context"
	"errors"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/services"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func QueryOrdersRefundInfo(ctx context.Context, uid, orderId string) (*model.OrdersRefund, error) {
	var ordersRefund model.OrdersRefund
	if err := global.DB.Model(&model.OrdersRefund{}).Where("user_id = ? AND order_id = ? and state = 0", uid, orderId).First(&ordersRefund).Error; err != nil {
		global.Lg.Error("QueryOrdersRefundInfo error", zap.Error(err), zap.String("order_id", orderId))
		return nil, err
	}
	return &ordersRefund, nil
}

func CreateOrdersRefund(ctx context.Context, db *gorm.DB, ordersRefund *model.OrdersRefund) error {
	if err := db.WithContext(ctx).Create(ordersRefund).Error; err != nil {
		global.Lg.Error("CreateOrdersRefund error", zap.Error(err), zap.Any("ordersRefund", ordersRefund))
		return err
	}
	return nil
}

func OrderRefund(ctx context.Context, uid, orderId string) error {
	orderCourse, err := QueryOrderInfo(uid, orderId)
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err), zap.String("uid", uid), zap.String("orderId", orderId))
		return err
	}

	var order model.Orders
	db := global.DB.Model(&model.Orders{}).Preload("OrdersCourses").
		Where("uid = ? and order_id = ? and state = 0", uid, orderId)
	if err := db.First(&order).Error; err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err), zap.String("uid", uid), zap.String("orderId", orderId))
		return err
	}
	if order.Status == enum.OrderStatusPending {
		err = ClosePendingOrder(ctx, order)
		return err
	}
	if order.Status == enum.OrderStatusRefundSuccess || order.Status == enum.OrderStatusRefundProcessing {
		global.Lg.Error("订单已申请退款", zap.Error(err), zap.String("uid", uid), zap.String("orderId", orderId))
		return enum.NewErr(enum.OrderCourseRefundExistErr, "订单已申请退款")
	}
	refundMoney, err := GetRefundMoney(order, order.OrdersCourses)
	if err != nil {
		global.Lg.Error("查询订单退款金额失败", zap.Error(err), zap.String("uid", uid), zap.String("orderId", orderId))
		return err
	}
	if refundMoney.Money == 0 && refundMoney.UsedPoints == 0 {
		global.Lg.Error("订单无需退款", zap.Error(err), zap.String("uid", uid), zap.String("orderId", orderId))
		return enum.NewErr(enum.OrderCourseNoRefundErr, "订单无需退款")
	}

	refundOrder, err := QueryOrdersRefundInfo(ctx, uid, orderId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Lg.Error("查询订单退款信息失败", zap.Error(err), zap.String("uid", uid), zap.String("order_id", orderId))
		return err
	}

	if refundOrder != nil {
		global.Lg.Error("订单已申请退款", zap.Error(err), zap.String("uid", uid), zap.String("order_id", orderId))
		return enum.NewErr(enum.OrderCourseRefundExistErr, "订单已申请退款")
	}

	//启动事务处理退款逻辑
	if err = global.DB.Transaction(func(tx *gorm.DB) error {
		refundId := GenerateId("RF")
		refundStatus := "PROCESSING"
		orderStatus := enum.OrderStatusRefundProcessing
		//插入订单退款信息
		refundOrder = &model.OrdersRefund{
			UserId:       uid,
			UserType:     model.UserTypeUser,
			OrderId:      order.OrderID,
			RefundId:     refundId,
			RefundMoney:  refundMoney.Money,
			RefundPoints: refundMoney.UsedPoints,
			RefundType:   model.RefundTypeUser,
		}
		if refundMoney.Money > 0 {

		} else {
			refundStatus = "SUCCESS"
			refundTime := time.Now()
			refundOrder.RefundTime = &refundTime
			refundOrder.Remark = "无需退款，只退积分"
			orderStatus = enum.OrderStatusRefundSuccess
		}

		tx.Model(model.Orders{}).Where("order_id = ?", orderId).Updates(
			map[string]interface{}{
				"status": orderStatus,
			})
		//插入订单退款信息
		refundOrder.Status = refundStatus
		if err = CreateOrdersRefund(ctx, tx, refundOrder); err != nil {
			global.Lg.Error("插入订单退款信息失败", zap.Error(err))
			return err
		}

		//不需要退钱，只退积分，则直接更新订单状态为退款成功，并释放教练时间和用户积分
		if orderStatus == enum.OrderStatusRefundSuccess {
			err = ReleaseCoachTime(ctx, order, tx)
			if err != nil {
				global.Lg.Error("释放教练时间失败", zap.Error(err))
				return err
			}
			// 如果使用了积分，需要退还积分
			err = ReleaseUsedPoints(ctx, order, refundMoney.UsedPoints, tx)
			if err != nil {
				return err
			}

			err = CloseOrder(ctx, order, tx)
			if err != nil {
				global.Lg.Error("关闭订单失败", zap.Error(err))
				return err
			}

			if err = NewGoodsDao(ctx, tx).AddGoodsCanceledCnt(order.GoodID, 1); err != nil {
				global.Lg.Error("AddGoodsCanceledCnt error", zap.Error(err))
				return err
			}
		} else {
			// 调用微信退款接口
			refundResp, err := services.RefundOrder(orderCourse.OrderID, order.TransactionId, refundId, refundMoney.Money, global.Config.Mch.OrderRefundNotifyUrl)
			if err != nil {
				global.Lg.Error("调用微信退款接口失败", zap.Error(err))
				return err
			}
			global.Lg.Info("调用微信退款接口成功", zap.Any("refundResp", refundResp))
			byeResp, err := refundResp.MarshalJSON()
			if err != nil {
				global.Lg.Error("refundResp.MarshalJSON error", zap.Error(err))
			}
			refundStatus = string(*refundResp.Status)
			tx.Model(model.OrdersRefund{}).Where("refund_id", refundId).Updates(map[string]interface{}{
				"remark": string(byeResp),
			})
		}

		global.Lg.Info("订单退款处理成功", zap.String("order_id", orderId))
		return nil
	}); err != nil {
		global.Lg.Error("订单退款处理失败", zap.Error(err))
		return err
	}
	return nil
}

// CloseOrder 关闭订单（要先释放教练教学时间再调用这个方法）
func CloseOrder(ctx context.Context, order model.Orders, tx *gorm.DB) error {
	err := tx.Model(model.Orders{}).Where("order_id = ?", order.OrderID).Updates(
		map[string]interface{}{
			"status":      enum.OrderStatusRefundSuccess,
			"teach_state": model.PackTeachStateCancel,
		}).Error
	if err != nil {
		global.Lg.Error("更新订单状态失败", zap.Error(err))
		return err
	}
	//把审核之前的课程状态改为取消
	err = tx.Model(model.OrdersCourses{}).Where("order_id = ? and state = 0 and teach_state < ?", order.OrderID, model.TeachStateWaitCheck).Updates(
		map[string]interface{}{
			"teach_state": model.TeachStateCancel,
		}).Error

	if err != nil {
		global.Lg.Error("更新课程状态失败", zap.Error(err))
		return err
	}
	return nil
}

// ClosePendingOrder 关闭待支付订单
func ClosePendingOrder(ctx context.Context, order model.Orders) error {
	if order.Status != enum.OrderStatusPending {
		return enum.NewErr(enum.OrderCourseRefundExistErr, "订单不是待支付状态")
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		err := CloseOrder(ctx, order, tx)
		if err != nil {
			global.Lg.Error("更新订单状态失败", zap.Error(err))
			return err
		}
		refundTime := time.Now()
		refundOrder := &model.OrdersRefund{
			UserId:       order.Uid,
			UserType:     model.UserTypeUser,
			OrderId:      order.OrderID,
			RefundId:     GenerateId("RF"),
			RefundMoney:  0,
			RefundPoints: order.UsedPoints,
			Remark:       "关闭订单",
			Status:       "SUCCESS",
			RefundTime:   &refundTime,
			RefundType:   model.RefundTypeUser,
		}
		if err = CreateOrdersRefund(ctx, tx, refundOrder); err != nil {
			global.Lg.Error("插入订单退款信息失败", zap.Error(err))
			return err
		}

		err = ReleaseUsedPoints(ctx, order, order.UsedPoints, tx)
		if err != nil {
			global.Lg.Error("释放用户积分失败", zap.Error(err))
		}
		return err
	}); err != nil {
		global.Lg.Error("关闭待支付订单失败", zap.Error(err))
		return err
	}
	return nil
}

// ReleaseUsedPoints 要把使用的积分释放掉
func ReleaseUsedPoints(ctx context.Context, order model.Orders, usedPoints int64, tx *gorm.DB) error {
	if usedPoints > 0 {
		user, err := QueryUserInfo(order.Uid)
		if err != nil {
			global.Lg.Error("查询用户信息失败", zap.Error(err))
			return err
		}

		// 创建积分退还记录
		_, err = CreatePointsRecord(tx, order.Uid, order.OrderID, model.ActionTypeOrderRefund, usedPoints, user.LeftPoints+usedPoints, "订单退款积分退还")
		if err != nil {
			global.Lg.Error("创建积分退还记录失败", zap.Error(err))
			return err
		}

		// 返还用户积分
		err = AddUserPoints(tx, order.Uid, usedPoints)
		if err != nil {
			global.Lg.Error("返还用户积分失败", zap.Error(err))
			return err
		}
	}
	return nil
}

// ReleaseCoachTime 要把待上课的教练时间释放掉
func ReleaseCoachTime(ctx context.Context, order model.Orders, tx *gorm.DB) (err error) {
	for _, ordersCourse := range order.OrdersCourses {
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: ordersCourse.OrderCourseID,
			UserID:        order.Uid,
			UserType:      enum.UserTypeUser,
			TeachTimeIDs:  model.JSONIntArray{},
			Operate:       model.OperateUserCancelCourse,
			Remark:        model.OCSOperateStr[model.OperateUserCancelCourse],
			Process:       model.ProcessYes,
		}
		if ordersCourse.TeachState == model.TeachStateWaitClass { //教练的课程待上课状态
			err = CancelCoachCourseSql(ctx, tx, order, ordersCourse, inspectorate)
			if err != nil {
				global.Lg.Error("取消课程失败", zap.Error(err))
				return err
			}
		}
		if ordersCourse.TeachState == model.TeachStateWaitCoachClass { //俱乐部课程待上课状态
			err = CancelClubCourseSql(ctx, tx, order, ordersCourse, inspectorate)
			if err != nil {
				global.Lg.Error("取消课程失败", zap.Error(err))
				return err
			}
		}
	}
	return nil
}

type RefundData struct {
	OrderId    string
	Uid        string
	UsedPoints int64 //要退还的积分
	Money      int64 // 该退款给用户的金额
	MoneyType  int   // 收入金额类型（资金类型53种）
	RefundData []CoachClubRefundData
}

type CoachClubRefundData struct { //退款
	UserId           string // 获得收入的用户（可能是教练也可能俱乐部）
	UserType         int    // 获得收入的用户类型
	OrderCourseId    string
	Money            int64 // 收入金额
	MoneyType        int   // 收入金额类型（资金类型53种）
	ServiceMoney     int64 // 收入金额需要缴纳的服务费
	ServiceMoneyType int   // 收入金额服务费类型（资金类型53种）
}

func GetRefundMoney(order model.Orders, ordersCourses []model.OrdersCourses) (refundMoney RefundData, err error) {
	refundMoney = RefundData{
		OrderId:    order.OrderID,
		Uid:        order.Uid,
		MoneyType:  model.UserIncomeRefundNoFault,
		RefundData: []CoachClubRefundData{},
	}
	if order.Status == model.OrderStatusRefundIng {
		return refundMoney, enum.NewErr(enum.OrderCourseFinishedOrCanceledErr, "订单正在退款中")
	}
	if order.Status == model.OrderStatusRefund {
		return refundMoney, enum.NewErr(enum.OrderCourseFinishedOrCanceledErr, "订单已退款")
	}
	//TODO 支付的钱和积分计算还不清楚
	var consumeFee int64            //消耗的金额
	if order.Pack == model.PackNo { //单次课
		if ordersCourses[0].TeachState != model.TeachStateWaitAppointment && ordersCourses[0].TeachState != model.TeachStateWaitClass && ordersCourses[0].TeachState != model.TeachStateWaitCoachClass && ordersCourses[0].TeachState != model.TeachStateFinish {
			return RefundData{}, enum.NewErr(enum.OrderCourseFinishedOrCanceledErr, "有进行中的课程，无法退款（单次课）")
		}
		isResponsibleCancel := IsResponsibleCancel(ordersCourses[0])
		if isResponsibleCancel { //有责取消
			refundMoney.MoneyType = model.UserIncomeOneCRefundFault
			one := CoachClubRefundData{
				OrderCourseId: ordersCourses[0].OrderCourseID,
				Money:         ordersCourses[0].FaultMoney,
			}
			serviceRatio := enum.ServiceRatio
			consumeFee = one.Money
			if order.UserType == model.UserTypeCoach { //教练课程
				one.UserType = model.UserTypeCoach
				one.UserId = order.UserID
				one.MoneyType = model.CoachIncomeOneCRefundFault
				if order.TransferCoachID != "" && order.TransferFee > 0 { //转课
					one.UserId = order.TransferCoachID
					one.MoneyType = model.CoachIncomeOneCTranferRefundFault
				}
				coach, err := QueryCoachInfo(order.UserID)
				if err == nil && coach != nil {
					serviceRatio = coach.ServiceRate
				}
			}
			if order.UserType == model.UserTypeClub { //俱乐部课程，有责取消费用给教练
				if ordersCourses[0].TeachCoachID != "" {
					one.UserId = ordersCourses[0].TeachCoachID
					one.UserType = model.UserTypeCoach
					one.MoneyType = model.CoachIncomeOneCRefundFault
				}
			}
			refundMoney.RefundData = append(refundMoney.RefundData, one)
			one.ServiceMoney = one.Money * int64(serviceRatio) / 100
		}
	} else { //打包课
		for _, ordersCourse := range ordersCourses {
			if ordersCourse.TeachState != model.TeachStateWaitAppointment && ordersCourse.TeachState != model.TeachStateWaitClass && ordersCourse.TeachState != model.TeachStateWaitCoachClass && ordersCourse.TeachState != model.TeachStateFinish {
				return RefundData{}, enum.NewErr(enum.OrderCourseFinishedOrCanceledErr, "有进行中的课程，无法退款")
			}
			if ordersCourse.TeachState != model.TeachStateFinish { //已核销的课程才需要补钱
				continue
			}
			serviceRatio := enum.ServiceRatio
			coach, err := QueryCoachInfo(ordersCourse.TeachCoachID)
			if err == nil && coach != nil {
				serviceRatio = coach.ServiceRate
			}
			coachMoney := (ordersCourse.TeachMoney + ordersCourse.AreaMoney) * int64(100-order.Discount) / 100
			consumeFee += ordersCourse.TeachMoney + ordersCourse.AreaMoney
			re := CoachClubRefundData{
				UserId:           ordersCourse.TeachCoachID, //打包课不能转单，所以上课教练和售卖教练是同一个
				UserType:         model.UserTypeCoach,
				Money:            coachMoney,
				MoneyType:        model.CoachIncomePackCRefundFinishToCoach,
				ServiceMoney:     coachMoney * int64(serviceRatio) / 100,
				ServiceMoneyType: model.CoachPayPackCRefundFinishToCoachService,
			}
			if order.UserType == model.UserTypeClub { //俱乐部课程
				re.MoneyType = model.CoachIncomePackCRefundFinishToClub
				re.ServiceMoneyType = model.CoachPayPackCRefundFinishToClubService

				consumeFee += ordersCourse.ClubMoney
				clubMoney := ordersCourse.ClubMoney * int64(100-order.Discount) / 100
				refundMoney.RefundData = append(refundMoney.RefundData, CoachClubRefundData{
					UserId:           order.UserID,
					UserType:         model.UserTypeClub,
					Money:            clubMoney,
					MoneyType:        model.ClubIncomePackCRefundFinishToClub,
					ServiceMoney:     clubMoney * enum.ServiceRatio / 100,
					ServiceMoneyType: model.ClubPayPackCRefundFinishToClubService,
				})
			}
			refundMoney.RefundData = append(refundMoney.RefundData, re)
		}
		if len(refundMoney.RefundData) == 0 {
			refundMoney.MoneyType = model.UserIncomePackCRefundNoFinish
		}
	}
	if consumeFee > order.PaidFee {
		refundMoney.Money = 0
		refundMoney.UsedPoints = order.PaidFee + order.UsedPoints - consumeFee
	} else {
		refundMoney.Money = order.PaidFee - consumeFee
		refundMoney.UsedPoints = order.UsedPoints
	}
	return
}

func RefundMoneyRecord(c context.Context, db *gorm.DB, refundData CoachClubRefundData, orderId string) (err error) {
	if refundData.Money <= 0 {
		return nil
	}
	record := model.MoneyRecords{
		UserID:        refundData.UserId,
		UserType:      refundData.UserType,
		Money:         refundData.Money,
		MoneyType:     refundData.MoneyType,
		IncomeType:    model.IncomeTypeIncome,
		RelationType:  model.RelationTypeOrder,
		RelationID:    orderId,
		OrderCourseID: refundData.OrderCourseId,
	}
	err = NewMoneyRecordsDao(c, db).Create(c, &record, db)
	if err != nil {
		return
	}

	if refundData.ServiceMoney > 0 { //平台服务费（谁的收入谁交服务费）
		record = model.MoneyRecords{
			UserID:        refundData.UserId,
			UserType:      refundData.UserType,
			Money:         refundData.ServiceMoney,
			MoneyType:     refundData.ServiceMoneyType,
			IncomeType:    model.IncomeTypePay,
			RelationType:  model.RelationTypeOrder,
			RelationID:    orderId,
			OrderCourseID: refundData.OrderCourseId,
		}
		err = NewMoneyRecordsDao(c, db).Create(c, &record, db)
		if err != nil {
			return
		}
	}
	return nil
}

// 判断是否是有责取消
func IsResponsibleCancel(orderCourse model.OrdersCourses) bool {
	if orderCourse.TeachState == model.TeachStateWaitAppointment { //待预约
		return false
	}

	if time.Now().Add(2 * 24 * time.Hour).After(time.Time(orderCourse.TeachStartTime)) { //2天内，只能有责取消
		return true
	}
	year, month, _ := time.Now().Date()
	thisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	monthOneDay := thisMonth.Format("2006-01-02")
	var cancelNum int64 //用户无责取消预约的次数
	err := global.DB.Model(model.OrdersCoursesState{}).
		Where("user_id  = ? and  operate = ? and state = 0", orderCourse.Uid, model.OperateUserCancelNoResponsibility).
		Where("created_at > ? ", monthOneDay).
		Order("id desc").Count(&cancelNum).Error
	if err != nil {
		return false
	}
	if cancelNum >= model.OrderCourseUserCancelNumber {
		return true
	}
	return false
}

func QueryOrdersRefundInfoByRefundId(ctx context.Context, refundId string) (*model.OrdersRefund, error) {
	var ordersRefund model.OrdersRefund
	if err := global.DB.Model(&model.OrdersRefund{}).Where("refund_id = ? and state = 0", refundId).First(&ordersRefund).Error; err != nil {
		global.Lg.Error("QueryOrdersRefundInfo error", zap.Error(err), zap.String("refund_id", refundId))
		return nil, err
	}
	return &ordersRefund, nil
}

func AdminQueryOrderRefundLimit(ctx context.Context, orderId string) (*forms.OrderRefundLimitResponse, error) {
	order, err := QueryOrderInfo("", orderId)
	if err != nil {
		global.Lg.Error("QueryOrderInfo failed", zap.String("order_id", orderId), zap.Error(err))
		return nil, err
	}

	if order.Status >= enum.OrderStatusRefundProcessing {
		global.Lg.Error("已申请退款，无法重复申请")
		return nil, enum.NewErr(enum.OrderCourseRefundExistErr, "已申请退款，无法重复申请")
	}

	if order.TeachState == int(model.TeachStateFinish) || order.TeachState == int(model.TeachStateCancel) {
		global.Lg.Error("已结束的课程无法申请退款")
		return nil, enum.NewErr(enum.OrderCourseFinishedOrCanceledErr, "已结束的课程无法申请退款")
	}

	refundOrder, err := QueryOrdersRefundInfo(ctx, "", orderId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Lg.Error("查询订单退款信息失败", zap.Error(err), zap.String("order_id", orderId))
		return nil, err
	}

	if refundOrder != nil {
		global.Lg.Error("已申请退款，无法重复申请")
		return nil, enum.NewErr(enum.OrderCourseRefundExistErr, "已申请退款，无法重复申请")
	}

	refundMaxMoney := order.TotalFee //默认是总金额
	var refundTeachCoach bool
	for _, v := range order.OrdersCourses {
		if v.TeachState == model.TeachStateFinish || v.TeachState == model.TeachStateCancel {
			refundMaxMoney -= (v.TeachMoney + v.AreaMoney) * int64(order.Discount) / 100 //要减去已完成的课程金额，order_course的价格是原价，管理台要按折扣价扣除
			continue
		}

		if v.TeachCoachID != order.UserID {
			refundTeachCoach = true
		}
	}

	if refundMaxMoney < 0 {
		refundMaxMoney = 0
	}

	return &forms.OrderRefundLimitResponse{
		RefundMaxMoney:   refundMaxMoney,
		RefundTeachCoach: refundTeachCoach,
	}, nil

}

func AdminOrderRefund(ctx context.Context, orderId string, req *forms.AdminOrderRefundRequest) error {
	order, err := QueryOrderInfo("", orderId)
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err), zap.String("orderId", orderId))
		return err
	}
	uid := order.Uid

	limit, err := AdminQueryOrderRefundLimit(ctx, orderId)
	if err != nil {
		global.Lg.Error("查询订单退款限制失败", zap.Error(err), zap.String("order_id", orderId))
		return err
	}

	if req.RefundType == enum.RefundTypeAll { //全额退款退积分
		return AdminOrderRefundAll(ctx, order, limit)
	}

	//下面是部分退款（不退积分）
	if req.RefundUserMoney+req.RefundTeachCoachMoney > limit.RefundMaxMoney {
		global.Lg.Error("退款金额超出限制")
		return enum.NewErr(enum.OrderStatusRefundFailed, "退款金额超出限制")
	}

	//部分退款流程，1.给用户退款，调用微信退款接口，2.如果给教学教练退款的话，给教学教练记账，3.剩余的钱给卖课的教练或者俱乐部，也走记账，所以一个订单可能有3条退款记录
	//启动事务处理退款逻辑
	if err = global.DB.Transaction(func(tx *gorm.DB) error {
		refundId := GenerateId("RF")
		refundStatus := "SUCCESS"
		orderStatus := enum.OrderStatusRefundProcessing
		//首先给用户退款
		refundOrderUser := &model.OrdersRefund{
			UserId:       uid,
			UserType:     model.UserTypeUser,
			OrderId:      order.OrderID,
			RefundId:     refundId,
			RefundMoney:  req.RefundUserMoney,
			RefundPoints: 0, //不退积分
			RefundType:   model.RefundTypeAdmin,
		}
		if refundOrderUser.RefundMoney == 0 { //用户不需要退款
			refundTime := time.Now()
			refundOrderUser.RefundTime = &refundTime
			refundOrderUser.Remark = "无需退款，只退积分"
			orderStatus = enum.OrderStatusRefundSuccess
		}

		tx.Model(model.Orders{}).Where("order_id = ?", orderId).Updates(
			map[string]interface{}{
				"status": orderStatus,
			})
		//插入订单退款信息
		refundOrderUser.Status = refundStatus
		if err = CreateOrdersRefund(ctx, tx, refundOrderUser); err != nil {
			global.Lg.Error("插入订单退款信息失败", zap.Error(err))
			return err
		}

		//如果给教学教练退款
		if req.RefundTeachCoachMoney > 0 {
			for _, v := range order.OrdersCourses {
				//已完成的课程忽略
				if v.TeachState == model.TeachStateFinish || v.TeachState == model.TeachStateCancel {
					continue
				}

				//未完成的课程退课
				refundOrderCoach := &model.OrdersRefund{
					UserId:       v.TeachCoachID,
					UserType:     model.UserTypeCoach,
					OrderId:      order.OrderID,
					RefundId:     GenerateId("RF"),
					RefundMoney:  req.RefundTeachCoachMoney,
					RefundPoints: 0, //不退积分
					RefundTime:   refundOrderUser.RefundTime,
					RefundType:   model.RefundTypeAdmin,
					Status:       refundOrderUser.Status,
					Remark:       "退款给教学教练",
				}
				if err = CreateOrdersRefund(ctx, tx, refundOrderCoach); err != nil {
					global.Lg.Error("插入订单退款信息失败", zap.Error(err))
					return err
				}

				teachCoach, err := CoachInfoByCoachIdWithLevel(v.TeachCoachID)
				if err != nil {
					global.Lg.Error("查询教练信息失败", zap.Error(err))
					continue
				}

				re := CoachClubRefundData{
					UserId:           v.TeachCoachID, //打包课不能转单，所以上课教练和售卖教练是同一个
					UserType:         model.UserTypeCoach,
					Money:            req.RefundTeachCoachMoney,
					MoneyType:        model.CoachIncomePackCRefundFinishToCoach,
					ServiceMoney:     req.RefundTeachCoachMoney * int64(teachCoach.ServiceRate) / 100,
					ServiceMoneyType: model.CoachPayPackCRefundFinishToCoachService,
				}
				if err = RefundMoneyRecord(ctx, tx, re, order.OrderID); err != nil {
					global.Lg.Error("插入退款记录失败", zap.Error(err))
					return err
				}

				//给教练加钱
				if err = AddCoachBalance(ctx, tx, v.TeachCoachID, re.Money-re.ServiceMoney); err != nil {
					global.Lg.Error("给教学教练加钱失败", zap.Error(err))
					return err
				}
			}
		}

		//要把待上课的教练时间释放掉
		if err = ReleaseCoachTime(ctx, *order, tx); err != nil {
			global.Lg.Error("释放教练时间失败", zap.Error(err))
			return err
		}

		//剩余的钱再给卖课教练或者俱乐部
		if limit.RefundMaxMoney > req.RefundUserMoney+req.RefundTeachCoachMoney { //用户和教学教练退款的钱加起来小于最大退款金额，剩余的钱给卖课教练或者俱乐部
			serviceRate := 15 //默认15%
			if order.UserType == model.UserTypeCoach {
				teachCoach, err := CoachInfoByCoachIdWithLevel(order.UserID)
				if err != nil {
					global.Lg.Error("查询教练信息失败", zap.Error(err))
					return err
				}
				serviceRate = teachCoach.ServiceRate
			} else if order.UserType == model.UserTypeClub {
				club, err := QueryClubInfoByClubId(order.UserID)
				if err != nil {
					global.Lg.Error("查询俱乐部信息失败", zap.Error(err))
					return err
				}
				serviceRate = club.ServiceRate
			}
			//给卖课教练或者俱乐部记账
			re := CoachClubRefundData{
				UserId:           order.UserID,
				UserType:         order.UserType,
				Money:            limit.RefundMaxMoney - req.RefundUserMoney - req.RefundTeachCoachMoney,
				MoneyType:        model.CoachIncomePackCRefundFinishToClub,
				ServiceMoney:     (limit.RefundMaxMoney - req.RefundUserMoney - req.RefundTeachCoachMoney) * int64(serviceRate) / 100,
				ServiceMoneyType: model.CoachPayPackCRefundFinishToClubService,
			}
			if err = RefundMoneyRecord(ctx, tx, re, order.OrderID); err != nil {
				global.Lg.Error("插入退款记录失败", zap.Error(err))
				return err
			}

			if order.UserType == model.UserTypeCoach {
				//给卖课教练加钱
				if err = AddCoachBalance(ctx, tx, order.UserID, re.Money-re.ServiceMoney); err != nil {
					global.Lg.Error("给卖课教练加钱失败", zap.Error(err))
					return err
				}
			} else if order.UserType == model.UserTypeClub {
				//给俱乐部加钱
				if err = AddClubBalance(ctx, tx, order.UserID, re.Money-re.ServiceMoney); err != nil {
					global.Lg.Error("给俱乐部加钱失败", zap.Error(err))
					return err
				}
			}
		}

		if refundOrderUser.RefundMoney > 0 {
			// 调用微信退款接口
			refundResp, err := services.RefundOrder(order.OrderID, order.TransactionId, refundId, refundOrderUser.RefundMoney, global.Config.Mch.OrderRefundNotifyUrl)
			if err != nil {
				global.Lg.Error("调用微信退款接口失败", zap.Error(err))
				return err
			}
			global.Lg.Debug("调用微信退款接口成功", zap.Any("refundResp", refundResp))
			byeResp, err := refundResp.MarshalJSON()
			if err != nil {
				global.Lg.Error("refundResp.MarshalJSON error", zap.Error(err))
			}
			refundStatus = string(*refundResp.Status)
			tx.Model(model.OrdersRefund{}).Where("refund_id", refundId).Updates(map[string]interface{}{
				"status": refundStatus,
				"remark": string(byeResp),
			})
		} else { //不需要回调，都在这里处理
			if order.UserType == model.UserTypeCoach {
				if err = AddCoachFinishedCourse(ctx, tx, order.UserID, 1); err != nil {
					global.Lg.Error("给教学教练加课程失败", zap.Error(err))
					return err
				}
			} else if order.UserType == model.UserTypeClub { //给俱乐部加课程
				if err = AddClubFinishedCourse(ctx, tx, order.UserID, 1); err != nil {
					global.Lg.Error("给俱乐部加课程失败", zap.Error(err))
					return err
				}
			}
			//统计订单统计
			if err = NewGoodsDao(ctx, tx).AddGoodsCanceledCnt(order.GoodID, 1); err != nil {
				global.Lg.Error("给商品加取消订单数失败", zap.Error(err))
				return err
			}
		}

		global.Lg.Info("订单退款处理成功", zap.String("order_id", orderId))
		return nil
	}); err != nil {
		global.Lg.Error("订单退款处理失败", zap.Error(err))
		return err
	}
	return nil
}

func AdminOrderRefundAll(ctx context.Context, order *model.Orders, limit *forms.OrderRefundLimitResponse) (err error) {
	//全额退款退积分，所有的钱都退给用户，这里不用区分是单次课还是打包课，统一处理
	//启动事务处理退款逻辑
	if err = global.DB.Transaction(func(tx *gorm.DB) error {
		refundId := GenerateId("RF")
		refundStatus := "PROCESSING"
		orderStatus := enum.OrderStatusRefundProcessing
		//插入订单退款信息
		refundOrder := &model.OrdersRefund{
			UserId:       order.Uid,
			UserType:     model.UserTypeUser,
			OrderId:      order.OrderID,
			RefundId:     refundId,
			RefundMoney:  limit.RefundMaxMoney,
			RefundPoints: order.UsedPoints,
			RefundType:   model.RefundTypeAdmin,
		}

		if refundOrder.RefundMoney == 0 {
			refundTime := time.Now()
			refundOrder.RefundTime = &refundTime
			orderStatus = enum.OrderStatusRefundSuccess
		}

		tx.Model(model.Orders{}).Where("order_id = ?", order.OrderID).Updates(
			map[string]interface{}{
				"status": orderStatus,
			})
		//插入订单退款信息
		refundOrder.Status = refundStatus
		if err = CreateOrdersRefund(ctx, tx, refundOrder); err != nil {
			global.Lg.Error("插入订单退款信息失败", zap.Error(err))
			return err
		}

		//不需要退钱，只退积分，则直接更新订单状态为退款成功，并释放教练时间和用户积分
		if orderStatus == enum.OrderStatusRefundSuccess {
			err = ReleaseCoachTime(ctx, *order, tx)
			if err != nil {
				global.Lg.Error("释放教练时间失败", zap.Error(err))
				return err
			}
			// 如果使用了积分，需要退还积分
			err = ReleaseUsedPoints(ctx, *order, refundOrder.RefundPoints, tx)
			if err != nil {
				return err
			}

			err = CloseOrder(ctx, *order, tx)
			if err != nil {
				global.Lg.Error("关闭订单失败", zap.Error(err))
				return err
			}

			if err = NewGoodsDao(ctx, tx).AddGoodsCanceledCnt(order.GoodID, 1); err != nil {
				global.Lg.Error("AddGoodsCanceledCnt error", zap.Error(err))
				return err
			}
		} else {
			// 调用微信退款接口
			refundResp, err := services.RefundOrder(order.OrderID, order.TransactionId, refundId, refundOrder.RefundMoney, global.Config.Mch.OrderRefundNotifyUrl)
			if err != nil {
				global.Lg.Error("调用微信退款接口失败", zap.Error(err))
				return err
			}
			global.Lg.Debug("调用微信退款接口成功", zap.Any("refundResp", refundResp))
			byeResp, err := refundResp.MarshalJSON()
			if err != nil {
				global.Lg.Error("refundResp.MarshalJSON error", zap.Error(err))
			}
			refundStatus = string(*refundResp.Status)
			tx.Model(model.OrdersRefund{}).Where("refund_id", refundId).Updates(map[string]interface{}{
				"status": refundStatus,
				"remark": string(byeResp),
			})
		}
		global.Lg.Info("订单退款处理成功", zap.String("order_id", order.OrderID))
		return nil
	}); err != nil {
		global.Lg.Error("订单退款处理失败", zap.Error(err))
		return err
	}

	return
}
