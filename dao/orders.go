package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	wechatUtils "github.com/wechatpay-apiv3/wechatpay-go/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateOrder(c *gin.Context, uid, openId string, req *forms.CreateOrderRequest) (*forms.CreateOrderResp, error) {
	ctx := context.Background()
	//查询商品信息
	good, err := NewGoodsDao(context.Background(), global.DB).QueryGoodInfo(req.GoodId)
	if err != nil {
		global.Lg.Error("GetInfoByGoodId error", zap.Error(err))
		return nil, enum.NewErr(enum.GoodNoExistErr, "商品不存在")
	}
	if good.Stack == 0 {
		global.Lg.Error("goods stack error", zap.Error(err))
		return nil, enum.NewErr(enum.GoodNoStackErr, "商品已售罄")
	}
	if good.Pack == model.PackYes && len(good.GoodsCourses) == 0 {
		global.Lg.Error("goods courses error", zap.Error(err))
		return nil, enum.NewErr(enum.GoodNoCoursesErr, "商品课程不存在")
	}

	usePoints := req.UsePoints
	if good.Pack == model.PackYes { //打包课不能使用积分
		usePoints = 0
	} else {
		if good.PointsDeduct == 0 && req.UsePoints > 0 {
			global.Lg.Error("use credits error", zap.Error(err))
			return nil, enum.NewErr(enum.ParamErr, "未开启积分抵扣")
		}
	}
	totalFee := good.DiscountMoney
	user, err := QueryUserInfo(uid)
	if err != nil {
		global.Lg.Error("QueryUserInfo error", zap.Error(err))
		return nil, err
	}
	var leftPoints int64
	if usePoints > 0 {
		usePoints = totalFee
		//查询用户信息，查看积分
		if err != nil {
			global.Lg.Error("QueryUserInfo error", zap.Error(err))
			return nil, err
		}
		leftPoints = user.LeftPoints
		if usePoints > leftPoints {
			usePoints = leftPoints
		}
	}
	paidFee := totalFee - usePoints //积分1:1抵扣
	pointsFee := usePoints

	totalCourse := len(good.GoodsCourses)
	if totalCourse == 0 { //单次课没有goodCourse
		totalCourse = 1
	}

	orderId := GenerateId("KC")
	status := enum.OrderStatusPending
	//创建订单
	order := &model.Orders{
		Uid:        uid,
		UName:      req.UName,
		UPhone:     req.UPhone,
		GoodID:     req.GoodId,
		OrderID:    orderId,
		TotalFee:   totalFee,
		PaidFee:    paidFee,
		UsedPoints: usePoints,
		PointsFee:  pointsFee,
		UserID:     good.UserID,
		UserType:   good.UserType,
		TeachTime:  good.TeachTime,
		Pack:       good.Pack,
		Discount:   good.Discount,
		Status:     status,
		Progress:   fmt.Sprintf("%d/%d", 0, totalCourse),
		CreatedAt:  time.Now(),
	}

	//插入订单课程表
	var ordersCourses []model.OrdersCourses
	timestamp := time.Now().UnixNano()
	orderCourseState := model.StateDeleted
	if paidFee <= 0 {
		orderCourseState = model.StateNormal
		status = enum.OrderStatusPaid
	}
	if order.Pack == model.PackYes {
		for i, course := range good.GoodsCourses {
			da := model.OrdersCourses{
				OrderCourseID: orderId + fmt.Sprintf("-%1d", i+1),
				Uid:           order.Uid,
				OrderID:       order.OrderID,
				GoodID:        course.PackGoodID,
				CourseID:      course.PackGood.CourseID,
				CheckCode:     RandomStringBySeed(timestamp+int64(i), 7) + GenerateSecureRandomString(3),
				TeachTime:     good.TeachTime,
				Pack:          model.PackYes,
				TeachMoney:    course.PackGood.TeachMoney,
				ClubMoney:     course.PackGood.ClubMoney,
				AreaMoney:     course.PackGood.AreaMoney,
				FaultMoney:    course.PackGood.FaultMoney,
				PayMoney:      CalculateGoodsCourses(good, course),
				State:         orderCourseState, //生成订单的时候插入删除状态的记录
			}
			if paidFee <= 0 {
				da.State = model.StateNormal
			}
			if good.UserType == enum.UserTypeCoach { //如果订单是教练的，需要教学默认是这个教练
				da.TeachCoachID = good.UserID
			}
			ordersCourses = append(ordersCourses, da)
		}
	} else {
		ordersCourses = append(ordersCourses, model.OrdersCourses{
			OrderCourseID: orderId + fmt.Sprintf("-%1d", 1),
			Uid:           order.Uid,
			OrderID:       order.OrderID,
			GoodID:        order.GoodID,
			CourseID:      good.CourseID,
			CheckCode:     RandomStringBySeed(timestamp, 7) + GenerateSecureRandomString(3),
			TeachTime:     good.TeachTime,
			Pack:          model.PackNo,
			TeachMoney:    good.TeachMoney,
			ClubMoney:     good.ClubMoney,
			AreaMoney:     good.AreaMoney,
			FaultMoney:    good.FaultMoney,
			PayMoney:      totalFee,
			State:         orderCourseState, //生成订单的时候插入删除状态的记录
		})
	}

	var prePayResp *jsapi.PrepayWithRequestPaymentResponse
	//启动事务动事务
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		//调用微信接口生成预支付订单
		if paidFee > 0 {
			appId := c.GetString("app_id")
			prePayResp, err = services.GeneratePreOrder(appId, uid, openId, orderId, paidFee, good.Title, global.Config.Mch.OrderPayNotifyUrl)
			if err != nil {
				global.Lg.Error("GeneratePreOrder error", zap.Error(err))
				return err
			}
		} else {
			now := time.Now()
			order.Status = enum.OrderStatusPaid
			order.PayTime = &now //如果订单没有支付，则订单状态改为已支付
		}

		if usePoints > 0 {
			pointRecord, err := CreatePointsRecord(tx, uid, orderId, model.ActionTypeBuyCourseDeduct, usePoints, leftPoints-usePoints, "购买课程抵扣")
			if err != nil {
				global.Lg.Error("CreatePointsRecord error", zap.Error(err))
				return enum.NewErr(enum.OrdersCreateCreditsRecordErr, "创建积分记录失败")
			}
			global.Lg.Debug("CreatePointsRecord", zap.Any("pointRecord", pointRecord))

			//扣除用户积分
			err = SubUserPoints(tx, uid, usePoints)
			if err != nil {
				global.Lg.Error("UpdateUserCredits error", zap.Error(err))
				return err
			}
		}

		if err = tx.Model(model.Orders{}).Create(order).Error; err != nil {
			global.Lg.Error("CreateOrder error", zap.Error(err))
			return err
		}

		global.Lg.Debug("orders", zap.Any("ordersCourses", ordersCourses))
		if err = tx.Model(model.OrdersCourses{}).Create(&ordersCourses).Error; err != nil {
			global.Lg.Error("CreateOrdersCourses error", zap.Error(err), zap.Any("ordersCourses", ordersCourses))
			return err
		}

		if order.Status == enum.OrderStatusPaid { //如果已经支付了，增加订单统计
			//增加商品未完成数量
			if err = NewGoodsDao(ctx, tx).AddGoodsUnFinishedCnt(good.GoodID, 1); err != nil {
				global.Lg.Error("AddGoodsUnfinishedCnt error", zap.Error(err))
				return err
			}

			err = OrderStatistics(context.Background(), tx, 1, 0, order.CreatedAt)
			if err != nil {
				global.Lg.Error("CreateOrdersStatistics error", zap.Error(err))
				return err
			}
		}

		return nil
	})
	if err != nil {
		global.Lg.Error("CreateOrder error", zap.Error(err))
		return nil, err
	}

	return &forms.CreateOrderResp{
		OrderId:    order.OrderID,
		Status:     status,
		PrePayResp: prePayResp,
		LeftPoints: user.LeftPoints - pointsFee,
		CostPoints: pointsFee,
		GainPoints: order.PaidFee / 100,
	}, nil
}

// CalculateTotalFee 计算总价
func CalculateTotalFee(good *model.Goods) (int64, int64) {
	if good.Pack == model.PackNo { //单课程的总价 = 课程价格 + 场地价格
		return CalculateGoodsPackNoFee(good)
	} else {
		totalFee := int64(0)
		for _, v := range good.GoodsCourses { //套餐课程的总价 = 课程价格 * 折扣 + 场地价格
			totalFee += CalculateGoodsCourses(good, v)
		}
		return totalFee, totalFee * int64(good.Discount) / 100
	}
}

func CalculateGoodsPackNoFee(good *model.Goods) (int64, int64) {
	totalFee := good.TeachMoney + good.AreaMoney //课程价格
	if good.UserType == model.UserTypeClub {     //课程类型为俱乐部课程，需要加上clubMoney
		totalFee += good.ClubMoney
	}

	return totalFee, totalFee * int64(good.Discount) / 100
}

func CalculateGoodsCourses(good *model.Goods, courses *model.GoodsCourses) int64 {
	totalFee := courses.PackGood.TeachMoney + courses.PackGood.AreaMoney
	if good.UserType == model.UserTypeClub { //课程类型为俱乐部课程，需要加上clubMoney
		totalFee += courses.PackGood.ClubMoney
	}
	return totalFee * int64(good.Discount) / 100
}

func TestQueryOrdersList() (*model.Goods, error) {
	ordersCourses := []model.OrdersCourses{
		{
			OrderID:  "1",
			CourseID: "11243",
		},
		{
			OrderID:  "2",
			CourseID: "22354435",
		},
	}
	global.DB.Table("orders_courses").Create(ordersCourses)
	return nil, nil
	//list, err := NewGoodsDao(context.Background(), global.DB).Get(1)

	//return list, err
}
func QueryOrdersList(uid string, req *forms.QueryOrdersListRequest) (int64, []*model.Orders, error) {
	db := global.DB.Model(&model.Orders{}).Preload("Goods.CourseTags.Tag").Preload("Goods.GoodsCourses.PackGood.CourseTags.Tag").Where("uid = ? and status >= 1", uid).Where("state = 0")
	if len(req.TeachStates) > 0 {
		db = db.Where("teach_state in ?", req.TeachStates)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("QueryOrdersList error", zap.Error(err))
		return 0, nil, err
	}

	var orders []*model.Orders
	if err := db.Order("id desc").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&orders).Error; err != nil {
		global.Lg.Error("QueryOrdersList error", zap.Error(err))
		return 0, nil, err
	}

	for _, v := range orders {
		dealGoodTags(v.Goods)
	}

	return total, orders, nil
}

func QueryOrderInfo(uid, orderId string) (*model.Orders, error) {
	var order model.Orders
	db := global.DB.Model(&model.Orders{}).
		Preload("PointRecord", "action_type = ? and state = 0", model.ActionTypeBuyCourseDeduct).
		Preload("Refund").
		Preload("Goods.CourseTags", "state = 0").
		Preload("Goods.CourseTags.Tag", "state = 0").
		Preload("Goods.GoodsCourses", "state = 0").
		Preload("Goods.GoodsCourses.PackGood", "state = 0").
		Preload("Goods.GoodsCourses.PackGood.CourseTags", "state = 0").
		Preload("Goods.GoodsCourses.PackGood.CourseTags.Tag", "state = 0").
		Preload("OrdersCourses", "state = 0").Where("order_id = ? and state = 0", orderId)
	if uid != "" {
		db = db.Where("uid = ?", uid)
	}
	if err := db.First(&order).Error; err != nil {
		global.Lg.Error("QueryOrderInfo error", zap.Error(err))
		return nil, err
	}

	dealGoodTags(order.Goods)
	if order.PointRecord != nil {
		order.PointId = order.PointRecord.PointID
	}

	return &order, nil
}

func GetWxPayCallbackData(r *http.Request) (result payments.Transaction, err error) {
	wechatpayPublicKey, err := wechatUtils.LoadPublicKeyWithPath("./config/pub_key.pem")
	if err != nil {
		global.Lg.Error("加载公钥失败", zap.Error(err))
		return
	}
	// 初始化 notify.Handler
	handler, err := notify.NewRSANotifyHandler(global.Config.Mch.ApiKey, verifiers.NewSHA256WithRSAPubkeyVerifier(global.Config.Mch.PublicKeyId, *wechatpayPublicKey))
	if err != nil {
		global.Lg.Error("创建回调处理器失败", zap.Error(err))
		return
	}

	transaction := new(payments.Transaction)
	notifyReq, err := handler.ParseNotifyRequest(context.Background(), r, transaction)
	// 如果验签未通过，或者解密失败
	if err != nil {
		global.Lg.Error("回调验签失败", zap.Error(err))
		return
	}

	global.Lg.Info("回调成功", zap.Any("result", notifyReq.Resource.Plaintext))
	// 解析通知内容为支付结果
	if err = json.Unmarshal([]byte(notifyReq.Resource.Plaintext), &result); err != nil {
		global.Lg.Error("解析回调内容失败", zap.Error(err))
		return
	}

	//下面处理业务
	return result, nil
}
func OrderPayCallback(c *gin.Context) error {
	result, err := GetWxPayCallbackData(c.Request)
	if err != nil {
		return err
	}
	//这里只处理支付成功的情况
	//SUCCESS	支付成功	用户支付成功，资金已入账
	//REFUND	转入退款	交易已退款(全额或部分)
	//NOTPAY	未支付	订单已创建但未支付
	//CLOSED	已关闭	订单已关闭(商户或系统)
	//REVOKED	已撤销	付款码支付被用户撤销
	//USERPAYING	用户支付中	付款码支付用户已扫码但未确认
	//PAYERROR	支付失败	支付失败(余额不足等)
	if *result.TradeState != "SUCCESS" { //这里先只处理支付成功的情况
		global.Lg.Info("支付未成功", zap.Any("result", result))
		return nil
	}
	return OrderPayCallbackSql(c, *result.OutTradeNo, *result.TransactionId, *result.SuccessTime)
}
func OrderPayCallbackSql(c context.Context, orderId, transactionId, successTime string) error {
	order, err := QueryOrderInfo("", orderId)
	if err != nil {
		global.Lg.Error("获取充值记录失败", zap.Error(err))
		return nil //找不到，返回成功
	}

	if order.Status == enum.OrderStatusPaid {
		global.Lg.Info("订单已支付", zap.String("order_id", orderId))
		return nil
	}

	//开启事务
	return global.DB.Transaction(func(tx *gorm.DB) error {
		//更新订单表
		//20250910202200 转成时间类型 2025-09-10 20:22:00
		payTime, err := time.Parse(time.RFC3339, successTime)
		if err != nil {
			global.Lg.Error("时间转换失败", zap.Error(err), zap.String("out_trade_no", orderId), zap.String("success_time", successTime))
			payTime = time.Now()
		}
		record := model.Orders{
			TransactionId: transactionId,
			Status:        enum.OrderStatusPaid, //已支付
			PayTime:       &payTime,
		}
		if err = tx.Model(&model.Orders{}).Where("order_id = ?", orderId).Where("state = 0").Updates(record).Error; err != nil {
			global.Lg.Error("更新订单表失败", zap.Error(err))
			return err
		}

		// 更新订单课程表
		if err = tx.Model(&model.OrdersCourses{}).Where("order_id = ?", orderId).Where("state = 1").Update("state", 0).Error; err != nil {
			global.Lg.Error("更新订单课程表失败", zap.Error(err))
			return err
		}

		//插入资金流水记录
		mr := model.MoneyRecords{
			UserID:       order.Uid,
			UserType:     enum.UserTypeUser,
			Money:        order.PaidFee,
			MoneyType:    model.UserPayBuyCourse,
			IncomeType:   model.IncomeTypePay,
			RelationType: model.RelationTypeOrder,
			RelationID:   order.OrderID,
		}
		if err = NewMoneyRecordsDao(c, tx).Create(c, &mr, tx); err != nil {
			global.Lg.Error("插入资金流水记录失败", zap.Error(err))
			return err
		}

		//增加商品未完成数量
		if err = NewGoodsDao(c, tx).AddGoodsUnFinishedCnt(order.GoodID, 1); err != nil {
			global.Lg.Error("AddGoodsUnfinishedCnt error", zap.Error(err))
			return err
		}

		//增加订单统计
		err = OrderStatistics(context.Background(), tx, 1, order.PaidFee, payTime)
		if err != nil {
			global.Lg.Error("CreateOrdersStatistics error", zap.Error(err))
			return err
		}

		global.Lg.Debug("订单支付成功", zap.Any("order", order))
		return nil
	})

}

func UpdateOrder(ctx context.Context, db *gorm.DB, orderId string, data map[string]interface{}) error {
	if err := db.Model(&model.Orders{}).Where("order_id = ? and state = 0", orderId).Updates(data).Error; err != nil {
		global.Lg.Error("更新订单失败", zap.Error(err), zap.String("order_id", orderId))
		return err
	}
	return nil
}

func OrderRefundCallback(c *gin.Context) error {
	result, err := GetWxRefundCallbackData(c.Request)
	if err != nil {
		return err
	}

	global.Lg.Debug("退款回调", zap.Any("result", result))
	//这里只处理支付成功的情况
	//SUCCESS	支付成功	用户支付成功，资金已入账
	//REFUND	转入退款	交易已退款(全额或部分)
	//NOTPAY	未支付	订单已创建但未支付
	//CLOSED	已关闭	订单已关闭(商户或系统)
	//REVOKED	已撤销	付款码支付被用户撤销
	//USERPAYING	用户支付中	付款码支付用户已扫码但未确认
	//PAYERROR	支付失败	支付失败(余额不足等)
	if result.RefundStatus != "SUCCESS" { //这里先只处理支付成功的情况
		global.Lg.Info("支付未成功", zap.Any("result", result))
		return nil
	}

	err = OrdersRefundCallback(c, result)
	return err
}

func OrdersRefundCallback(c *gin.Context, result *RefundCallbackResult) error {
	refundId := result.OutRefundNo
	//查询退款信息
	refund, err := QueryOrdersRefundInfoByRefundId(c, refundId)
	if err != nil {
		global.Lg.Error("查询退款信息失败", zap.Error(err), zap.String("refund_id", refundId))
		return nil
	}

	order, err := QueryOrderInfo("", result.OutTradeNo)
	if err != nil {
		global.Lg.Error("获取充值记录失败", zap.Error(err))
		return nil //找不到，返回成功
	}
	refundMoney, err := GetRefundMoney(*order, order.OrdersCourses)
	if err != nil {
		global.Lg.Error("GetRefundMoney 获取退款金额失败", zap.Error(err))
		return err
	}

	//开启事务
	return global.DB.Transaction(func(tx *gorm.DB) error {
		for _, refundData := range refundMoney.RefundData {
			err = RefundMoneyRecord(c, tx, refundData, order.OrderID)
			if err != nil {
				global.Lg.Error("RefundMoneyRecord 插入退款记录失败", zap.Error(err))
				return err
			}
		}
		err = ReleaseCoachTime(c, *order, tx)
		if err != nil {
			global.Lg.Error("释放教练时间失败", zap.Error(err))
			return err
		}
		// 如果使用了积分，需要退还积分
		err = ReleaseUsedPoints(c, *order, refundMoney.UsedPoints, tx)
		if err != nil {
			global.Lg.Error("退还积分失败", zap.Error(err))
			return err
		}

		//插入资金流水记录
		mr := model.MoneyRecords{
			UserID:       order.Uid,
			UserType:     model.UserTypeUser,
			Money:        refund.RefundMoney,
			MoneyType:    model.UserIncomeRefundNoFault,
			IncomeType:   model.IncomeTypePay,
			RelationType: model.RelationTypeOrder,
			RelationID:   order.OrderID,
		}
		if err = NewMoneyRecordsDao(c, tx).Create(c, &mr, tx); err != nil {
			global.Lg.Error("插入资金流水记录失败", zap.Error(err))
			return err
		}

		//更新退款记录
		or := model.OrdersRefund{
			Status:              result.RefundStatus,
			RefundTime:          result.SuccessTime,
			RefundTransactionId: result.RefundId,
		}
		if err = tx.Model(&model.OrdersRefund{}).Where("refund_id = ?", refundId).Where("state = 0").Updates(or).Error; err != nil {
			global.Lg.Error("更新退款记录失败", zap.Error(err))
			return err
		}

		err = CloseOrder(c, *order, tx)
		if err != nil {
			global.Lg.Error("关闭订单失败", zap.Error(err))
			return err
		}

		//维护商品的退款数量
		err = NewGoodsDao(c, tx).AddGoodsCanceledCnt(order.GoodID, 1)
		if err != nil {
			global.Lg.Error("AddGoodsRefundCnt error", zap.Error(err))
			return err
		}

		global.Lg.Debug("订单支付成功", zap.Any("order", order))
		return nil
	})
}

func UpdateOrderTeachState(ctx context.Context, db *gorm.DB, order *model.Orders, teachState model.TeachState) error {
	packTeachState := getPackTeachState(teachState)

	if order.Pack == 0 || packTeachState == model.PackTeachStateDoing { //单次课，直接保持一致
		order.TeachState = packTeachState
		if order.TeachState == model.PackTeachStateFinish { //如果是已完成
			order.Progress = "1/1"
		}
		if err := db.Model(model.Orders{}).Where("order_id = ?", order.OrderID).Save(order).Error; err != nil {
			global.Lg.Error("更新订单教学状态失败", zap.Error(err), zap.String("order_id", order.OrderID))
			return err
		}

		if order.TeachState == model.PackTeachStateFinish {
			if order.UserType == model.UserTypeCoach {
				//更新教练信息
				if err := AddCoachFinishedCourse(ctx, db, order.UserID, 1); err != nil {
					global.Lg.Error("更新教练状态失败", zap.Error(err), zap.String("coach_id", order.UserID))
					return err
				}
			} else if order.UserType == model.UserTypeClub {
				if err := AddClubFinishedCourse(ctx, db, order.UserID, 1); err != nil {
					global.Lg.Error("更新用户状态失败", zap.Error(err), zap.String("user_id", order.UserID))
					return err
				}
			}
			//还要更新商品完成数量
			if err := NewGoodsDao(ctx, db).AddGoodsFinishedCnt(order.GoodID, 1); err != nil {
				global.Lg.Error("更新商品状态失败", zap.Error(err), zap.String("good_id", order.GoodID))
				return err
			}
		}

		return nil
	}

	//打包课的状态
	//待完成-一个都没预约
	//进行中-有一个子课程在进行中
	//已完成-全部课程都完成
	//查询ordersCourses表拿到所有的课程
	courses, err := NewOrdersCoursesDao(ctx, db).QueryGoodsCourses(ctx, order.GoodID)
	if err != nil {
		global.Lg.Error("查询商品课程失败", zap.Error(err), zap.String("good_id", order.GoodID))
		return err
	}

	teachStateMap := make(map[int]int)
	var progress string
	for _, course := range courses {
		tmpTeachState := getPackTeachState(course.TeachState)
		if _, ok := teachStateMap[tmpTeachState]; !ok {
			teachStateMap[tmpTeachState] = 0
		}
		teachStateMap[tmpTeachState]++
	}

	if teachState == model.TeachStateFinish {
		progress = fmt.Sprintf("%d/%d", teachStateMap[model.PackTeachStateFinish], len(courses))
	}

	//根据teachStateMap判断打包课的状态
	if len(teachStateMap) == 0 || teachStateMap[model.PackTeachStateWaitAppointment] == len(courses) {
		order.TeachState = model.PackTeachStateWaitAppointment
	} else if teachStateMap[model.PackTeachStateFinish] == len(courses) {
		order.TeachState = model.PackTeachStateFinish
	} else if teachStateMap[model.PackTeachStateCancel] > 0 {
		order.TeachState = model.PackTeachStateCancel
	} else {
		order.TeachState = model.PackTeachStateDoing
	}

	if progress != "" {
		order.Progress = progress
	}

	if err = db.Model(model.Orders{}).Where("order_id = ?", order.OrderID).Save(order).Error; err != nil {
		global.Lg.Error("更新订单教学状态失败", zap.Error(err), zap.String("order_id", order.OrderID))
		return err
	}

	if order.TeachState == model.PackTeachStateFinish {
		if order.UserType == model.UserTypeCoach {
			//更新教练信息
			if err = AddCoachFinishedCourse(ctx, db, order.UserID, 1); err != nil {
				global.Lg.Error("更新教练状态失败", zap.Error(err), zap.String("coach_id", order.UserID))
				return err
			}
		} else if order.UserType == model.UserTypeClub {
			if err = AddClubFinishedCourse(ctx, db, order.UserID, 1); err != nil {
				global.Lg.Error("更新用户状态失败", zap.Error(err), zap.String("user_id", order.UserID))
				return err
			}
		}
		//还要更新商品完成数量
		if err = NewGoodsDao(ctx, db).AddGoodsFinishedCnt(order.GoodID, 1); err != nil {
			global.Lg.Error("更新商品状态失败", zap.Error(err), zap.String("good_id", order.GoodID))
			return err
		}
	}

	return nil
}

func getPackTeachState(teachState model.TeachState) int {
	packTeachState := model.PackTeachStateWaitAppointment
	switch teachState {
	case model.TeachStateWaitAppointment:
		packTeachState = model.PackTeachStateWaitAppointment
	case model.TeachStateFinish:
		packTeachState = model.PackTeachStateFinish
	case model.TeachStateCancel:
		packTeachState = model.PackTeachStateCancel
	default:
		packTeachState = model.PackTeachStateDoing
	}
	return packTeachState
}

func GetWxRefundCallbackData(r *http.Request) (result *RefundCallbackResult, err error) {
	wechatpayPublicKey, err := wechatUtils.LoadPublicKeyWithPath("./config/pub_key.pem")
	if err != nil {
		global.Lg.Error("加载公钥失败", zap.Error(err))
		return
	}
	// 初始化 notify.Handler
	handler, err := notify.NewRSANotifyHandler(global.Config.Mch.ApiKey, verifiers.NewSHA256WithRSAPubkeyVerifier(global.Config.Mch.PublicKeyId, *wechatpayPublicKey))
	if err != nil {
		global.Lg.Error("创建回调处理器失败", zap.Error(err))
		return
	}

	// 注意：这里应该使用 refunddomestic.Refund 而不是 payments.Transaction
	refund := new(refunddomestic.Refund)
	notifyReq, err := handler.ParseNotifyRequest(context.Background(), r, refund)
	// 如果验签未通过，或者解密失败
	if err != nil {
		global.Lg.Error("回调验签失败", zap.Error(err))
		return
	}

	global.Lg.Info("退款回调成功", zap.Any("result", notifyReq.Resource.Plaintext))

	// 解析通知内容为退款结果
	if err = json.Unmarshal([]byte(notifyReq.Resource.Plaintext), &result); err != nil {
		global.Lg.Error("解析退款回调内容失败", zap.Error(err))
		return
	}

	return result, nil
}

type RefundCallbackResult struct {
	Mchid               string     `json:"mchid"`
	OutTradeNo          string     `json:"out_trade_no"`
	TransactionId       string     `json:"transaction_id"`
	OutRefundNo         string     `json:"out_refund_no"`
	RefundId            string     `json:"refund_id"`
	RefundStatus        string     `json:"refund_status"`
	SuccessTime         *time.Time `json:"success_time"`
	Amount              Amount     `json:"amount"`
	UserReceivedAccount string     `json:"user_received_account"`
}
type Amount struct {
	Total       int `json:"total"`
	Refund      int `json:"refund"`
	PayerTotal  int `json:"payer_total"`
	PayerRefund int `json:"payer_refund"`
}
