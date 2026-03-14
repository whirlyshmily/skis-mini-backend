package dao

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

func CompleteCourseSplitMoney(c context.Context, orderCourse model.OrdersCourses, order model.Orders, insrtocsData model.OrdersCoursesState) (err error) {
	user, err := QueryUserInfo(orderCourse.Uid)
	if err != nil {
		global.Lg.Error("QueryUserInfo error", zap.Error(err), zap.Any("orderCourse", orderCourse))
		return err
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCourses{}).Where("order_course_id = ?", orderCourse.OrderCourseID).
			Updates(map[string]interface{}{
				"teach_state": model.TeachStateFinish,
				"is_check":    model.IsCheckYes,
			}).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "核销失败")
		}

		err = tx.Model(model.OrdersCoursesState{}).Create(&insrtocsData).Error
		if err != nil {
			global.Lg.Error("Create error", zap.Error(err), zap.Any("insrtocsData", insrtocsData))
			return err
		}

		err = UpdateOrderTeachState(c, tx, &order, model.TeachStateFinish)
		if err != nil {
			global.Lg.Error("UpdateOrderTeachState error", zap.Error(err), zap.Any("order", order))
			return err
		}

		//更新用户积分
		paidFee := order.PaidFee
		if order.Pack == model.PackYes { //打包课
			paidFee = (orderCourse.TeachMoney + orderCourse.AreaMoney + orderCourse.ClubMoney) * int64(order.Discount) / 100
		}
		user.AccumulatedPoints += paidFee / 100
		user.Level = GetPointsLevel(user.AccumulatedPoints)
		user.LeftPoints += paidFee / 100
		if err = tx.Model(user).Save(user).Error; err != nil {
			global.Lg.Error("UpdateUser error", zap.Error(err), zap.Any("user", user))
			return err
		}
		verifyData, err := VerifyCourses(c, tx, orderCourse, order)
		if err != nil {
			global.Lg.Error("VerifyCourses error", zap.Error(err), zap.Any("orderCourse", orderCourse), zap.Any("verifyData", verifyData))
			return err
		}
		return nil
	})
	if err != nil {
		global.Lg.Error("CompleteCourseSplitMoney error", zap.Error(err), zap.Any("orderCourse", orderCourse), zap.Any("order", order))
	}
	return err
}

type VerifyCourseData struct { //结算的时候，有多少角色分钱，就有多少个结构体
	Uid          string     // 用户ID
	OrderId      string     //订单ID
	OderCourseID string     //订单课程ID
	Seller       IncomeData //售卖课程的数据
	Transfer     IncomeData //转单数据（教练售卖的单次课）
	Teacher      IncomeData //教学数据（俱乐部售卖的课程）
}
type IncomeData struct { //销售数据
	UserId            string // 获得收入的用户（可能是教练也可能俱乐部）
	UserType          int    // 获得收入的用户类型
	Money             int64  // 收入金额
	MoneyType         int    // 收入金额类型（资金类型53种）
	ServiceMoney      int64  // 收入金额需要缴纳的服务费
	ServiceMoneyType  int    // 收入金额服务费类型（资金类型53种）
	ReferralMoney     int64  // 收入金额需要返给推荐人的金额（有服务费才给推荐人返）
	ReferralMoneyType int    // 收入金额推荐返佣类型（资金类型53种）
	ReferralUserId    string // 推荐人的用户ID
	ReferralUserType  int    // 推荐人的用户类型
}

func VerifyCourses(c context.Context, db *gorm.DB, orderCourse model.OrdersCourses, order model.Orders) (verifyData VerifyCourseData, err error) {
	if orderCourse.IsCheck == model.IsCheckYes {
		err = enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
		return
	}
	if order.FrozenMoney == model.FrozenYes {
		err = enum.NewErr(enum.OrderFrozenMoneyErr, "课程已冻结，请联系客服")
		return
	}

	if orderCourse.TeachState != model.TeachStateWaitClass && orderCourse.TeachState != model.TeachStateWaitCoachClass && orderCourse.TeachState < model.TeachStateWaitCheck {
		err = enum.NewErr(enum.OrdersCoursesExitErr, "课程未完成")
		return
	}
	if order.OrderID != "" && orderCourse.OrderID != order.OrderID {
		err = db.Model(&model.Orders{}).Where("order_id = ?", orderCourse.OrderID).First(&order).Error
		if err != nil {
			return
		}
	}
	verifyData = VerifyCourseData{
		Uid:          order.Uid,
		OrderId:      order.OrderID,
		OderCourseID: orderCourse.OrderCourseID,
		Seller: IncomeData{
			UserId:   order.UserID,
			UserType: order.UserType,
		},
		Transfer: IncomeData{},
		Teacher:  IncomeData{},
	}
	if verifyData.Seller.UserType == model.UserTypeClub { //完成课程（俱乐部）
		verifyData.Seller.MoneyType = model.ClubIncomeFinishCToClub
		verifyData.Seller.ServiceMoneyType = model.ClubPayFinishCToClubService
		verifyData.Teacher.MoneyType = model.CoachIncomeFinishCToClub
		verifyData.Teacher.ServiceMoneyType = model.CoachPayFinishCToClubService
		verifyData.Teacher.ReferralMoneyType = model.CoachIncomeFinishCToClubReferralBonus
	}
	if verifyData.Seller.UserType == model.UserTypeCoach { //完成课程（教练）
		verifyData.Seller.MoneyType = model.CoachIncomeFinishCToCoach
		verifyData.Seller.ServiceMoneyType = model.CoachPayFinishCToCoachService
		verifyData.Seller.ReferralMoneyType = model.CoachIncomeFinishCToCoachReferralBonus
	}

	if order.Pack == model.PackNo { // 单次课程
		if order.UserType == model.UserTypeCoach {
			if order.TransferCoachID != "" && order.TransferFee > 0 { //完成课程（转单）
				verifyData.Transfer.ServiceMoneyType = model.CoachPayFinishCToTranferService
				verifyData.Seller.ServiceMoneyType = model.CoachPayFinishCToTranferBySellService
				verifyData.Transfer.MoneyType = model.CoachIncomeFinishCToTranferByTeacher
				verifyData.Seller.MoneyType = model.CoachIncomeFinishCToTranferBySeller
				verifyData.Transfer.ReferralMoneyType = model.CoachIncomeFinishCToTranferReferralBonus

				verifyData.Transfer.UserId = order.TransferCoachID
				verifyData.Transfer.UserType = model.UserTypeCoach
				verifyData.Transfer.Money = order.TransferFee
				verifyData.Seller.Money = order.TotalFee - order.TransferFee
			} else {
				verifyData.Seller.Money = order.TotalFee
			}
		}

		if order.UserType == model.UserTypeClub { //俱乐部
			verifyData.Seller.Money = orderCourse.ClubMoney
			verifyData.Teacher.UserId = orderCourse.TeachCoachID
			verifyData.Teacher.UserType = model.UserTypeCoach
			verifyData.Teacher.Money = order.TotalFee - orderCourse.ClubMoney
		}
	}

	if order.Pack == model.PackYes { // 打包课程（打包课程没有转单）
		if order.UserType == model.UserTypeCoach {
			verifyData.Seller.Money = (orderCourse.TeachMoney + orderCourse.AreaMoney) * int64(order.Discount) / 100
		}
		if order.UserType == model.UserTypeClub { //俱乐部的打包课程，售卖金额是俱乐部金额，教学金额是教练金额
			verifyData.Seller.Money = orderCourse.ClubMoney * int64(order.Discount) / 100
			verifyData.Teacher.UserId = orderCourse.TeachCoachID
			verifyData.Teacher.UserType = model.UserTypeCoach
			verifyData.Teacher.Money = (orderCourse.TeachMoney + orderCourse.AreaMoney) * int64(order.Discount) / 100
		}
	}
	err = ImproveVerifyData(c, db, &verifyData)
	if err != nil {
		return
	}
	err = SaveVerifyData(c, db, &verifyData)
	if err != nil {
		return
	}
	return
}

// ImproveVerifyData 完善核销数据,添加服务费，返佣等信息
func ImproveVerifyData(c context.Context, db *gorm.DB, verifyData *VerifyCourseData) (err error) {
	uidReferralRecord := &model.ReferralRecords{}
	err = db.Model(&model.ReferralRecords{}).
		Where("user_id = ? and referral_user_id=? and referral_type=? and state = 0",
			verifyData.Uid, verifyData.Seller.UserId, verifyData.Seller.UserType).Last(&uidReferralRecord).Error
	var sellerServiceRatio, teacherServiceRatio, transferServiceRatio int64
	sellerServiceRatio, teacherServiceRatio, transferServiceRatio = enum.ServiceRatio, enum.ServiceRatio, enum.ServiceRatio
	if verifyData.Seller.UserType == model.UserTypeCoach {
		coach, err := QueryCoachInfo(verifyData.Seller.UserId)
		if err == nil && coach != nil {
			sellerServiceRatio = int64(coach.ServiceRate)
		}
	}
	if verifyData.Teacher.UserType == model.UserTypeCoach {
		coach, err := QueryCoachInfo(verifyData.Teacher.UserId)
		if err == nil && coach != nil {
			teacherServiceRatio = int64(coach.ServiceRate)
		}
	}
	if verifyData.Transfer.UserType == model.UserTypeCoach {
		coach, err := QueryCoachInfo(verifyData.Transfer.UserId)
		if err == nil && coach != nil {
			transferServiceRatio = int64(coach.ServiceRate)
		}
	}

	if err != nil || uidReferralRecord.ID == 0 { //购买课程的用户跟售卖课程的用户不是推荐关系，需要缴纳平台服务费
		verifyData.Seller.ServiceMoney = verifyData.Seller.Money * sellerServiceRatio / 100
	}

	if verifyData.Seller.ServiceMoney > 0 { //如果有服务费，平台就给售卖课程用户的推荐人返佣
		sellUserReferralRecord := &model.ReferralRecords{}
		err = db.Model(&model.ReferralRecords{}).
			Where("user_id = ? and user_type=? and state = 0",
				verifyData.Seller.UserId, verifyData.Seller.UserType).Last(&sellUserReferralRecord).Error
		if err == nil && sellUserReferralRecord.ID > 0 { //有服务费的情况下，售卖课程的用户有推荐人，平台需要给推荐人返佣
			verifyData.Seller.ReferralMoney = verifyData.Seller.Money * sellerServiceRatio / 100
			verifyData.Seller.ReferralUserId = sellUserReferralRecord.ReferralUserID
			verifyData.Seller.ReferralUserType = sellUserReferralRecord.ReferralType
		}
	}
	if verifyData.Teacher.Money > 0 { //如果教学金额大于0，说明有教学费用，需要缴纳平台服务费
		verifyData.Teacher.ServiceMoney = verifyData.Teacher.Money * teacherServiceRatio / 100
		if verifyData.Teacher.ServiceMoney > 0 { //如果有服务费，平台就给教学课程用户的推荐人返佣
			teachUserReferralRecord := &model.ReferralRecords{}
			err = db.Model(&model.ReferralRecords{}).
				Where("user_id = ? and user_type=? and state = 0",
					verifyData.Teacher.UserId, verifyData.Teacher.UserType).Last(&teachUserReferralRecord).Error
			if err == nil && teachUserReferralRecord.ID > 0 { //有服务费的情况下，教学课程的用户有推荐人，平台需要给推荐人返佣
				verifyData.Teacher.ReferralMoney = verifyData.Teacher.Money * teacherServiceRatio / 100
				verifyData.Teacher.ReferralUserId = teachUserReferralRecord.ReferralUserID
				verifyData.Teacher.ReferralUserType = teachUserReferralRecord.ReferralType
			}
		}
	}

	if verifyData.Transfer.Money > 0 { //如果转单金额大于0，说明有转单费用，需要缴纳平台服务费
		verifyData.Transfer.ServiceMoney = verifyData.Transfer.Money * transferServiceRatio / 100
		if verifyData.Transfer.ServiceMoney > 0 { //如果有服务费，平台就给转单课程用户的推荐人返佣
			transferUserReferralRecord := &model.ReferralRecords{}
			err = db.Model(&model.ReferralRecords{}).
				Where("user_id = ? and user_type=? and state = 0",
					verifyData.Transfer.UserId, verifyData.Transfer.UserType).Last(&transferUserReferralRecord).Error
			if err == nil && transferUserReferralRecord.ID > 0 { //有服务费的情况下，转单课程的用户有推荐人，平台需要给推荐人返佣
				verifyData.Transfer.ReferralMoney = verifyData.Transfer.Money * transferServiceRatio / 100
				verifyData.Transfer.ReferralUserId = transferUserReferralRecord.ReferralUserID
				verifyData.Transfer.ReferralUserType = transferUserReferralRecord.ReferralType
			}
		}
	}
	return nil
}

// 外面开事务
func SaveVerifyData(c context.Context, tx *gorm.DB, verifyData *VerifyCourseData) (err error) {
	err = tx.Model(&model.OrdersCourses{}).Where("order_course_id = ?", verifyData.OderCourseID).Updates(model.OrdersCourses{
		IsCheck:    model.IsCheckYes,
		TeachState: model.TeachStateFinish,
	}).Error
	if err != nil {
		return err
	}
	if verifyData.Seller.Money > 0 {
		err = SaveMoneyRecord(c, tx, verifyData.Seller, verifyData.OrderId, verifyData.OderCourseID)
		if err != nil {
			return err
		}
	}
	if verifyData.Transfer.Money > 0 { //转单金额大于0，保存转单金额
		err = SaveMoneyRecord(c, tx, verifyData.Transfer, verifyData.OrderId, verifyData.OderCourseID)
		if err != nil {
			return err
		}
	}
	if verifyData.Teacher.Money > 0 { //教学金额大于0，保存教学金额
		err = SaveMoneyRecord(c, tx, verifyData.Teacher, verifyData.OrderId, verifyData.OderCourseID)
		if err != nil {
			return err
		}
	}
	return err
}

func SaveMoneyRecord(c context.Context, db *gorm.DB, incomeData IncomeData, orderId, orderCourseId string) (err error) {
	if incomeData.Money <= 0 {
		return nil
	}
	record := model.MoneyRecords{
		UserID:        incomeData.UserId,
		UserType:      incomeData.UserType,
		Money:         incomeData.Money,
		MoneyType:     incomeData.MoneyType,
		IncomeType:    model.IncomeTypeIncome,
		RelationType:  model.RelationTypeOrder,
		RelationID:    orderId,
		OrderCourseID: orderCourseId,
	}
	err = NewMoneyRecordsDao(c, db).Create(c, &record, db)
	if err != nil {
		return
	}

	if incomeData.ServiceMoney > 0 { //平台服务费（谁的收入谁交服务费）
		record = model.MoneyRecords{
			UserID:        incomeData.UserId,
			UserType:      incomeData.UserType,
			Money:         incomeData.ServiceMoney,
			MoneyType:     incomeData.ServiceMoneyType,
			IncomeType:    model.IncomeTypePay,
			RelationType:  model.RelationTypeOrder,
			RelationID:    orderId,
			OrderCourseID: orderCourseId,
		}
		err = NewMoneyRecordsDao(c, db).Create(c, &record, db)
		if err != nil {
			return
		}

		if incomeData.ReferralMoney > 0 { //有服务费才有推荐返佣
			err = db.Model(&model.ReferralRecords{}).Where("user_id = ? and referral_user_id=? and referral_type=? and state = 0",
				incomeData.UserId, incomeData.ReferralUserId, incomeData.ReferralUserType).
				Update("profit", gorm.Expr("profit + ?", incomeData.ReferralMoney)).Error
			if err != nil {
				return
			}

			record = model.MoneyRecords{
				UserID:        incomeData.ReferralUserId,
				UserType:      incomeData.ReferralUserType,
				Money:         incomeData.ReferralMoney,
				MoneyType:     incomeData.ReferralMoneyType,
				IncomeType:    model.IncomeTypeIncome,
				RelationType:  model.RelationTypeOrder,
				RelationID:    orderId,
				OrderCourseID: orderCourseId,
			}
			err = NewMoneyRecordsDao(c, db).Create(c, &record, db)
			if err != nil {
				return
			}
		}
	}

	return nil
}

func QueryMoneyList(c *gin.Context, req forms.QueryMoneyListRequest) (list []*model.MoneyRecords, err error) {
	userId := c.GetString("user_id")
	db := global.DB.Model(&model.MoneyRecords{}).Where("user_id = ?", userId)
	if req.IncomeType != -1 {
		db = db.Where("income_type = ?", req.IncomeType)
	}

	err = db.Order("id desc").
		Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find(&list).Error
	return list, err
}

func QueryMoneyInfoByMoneyId(c *gin.Context, moneyId string) (info *model.MoneyRecords, err error) {
	userId := c.GetString("user_id")
	err = global.DB.Model(&model.MoneyRecords{}).Where("user_id = ?", userId).Where("money_id = ?", moneyId).First(&info).Error
	return info, err
}
