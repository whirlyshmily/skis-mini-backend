package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrdersCoursesDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.OrdersCourses
}

func NewOrdersCoursesDao(ctx context.Context, dbs ...*gorm.DB) *OrdersCoursesDao {
	dao := new(OrdersCoursesDao)
	switch len(dbs) {
	case 0:
		panic("database connection required")
	case 1:
		dao.sourceDB = dbs[0]
		dao.replicaDB = []*gorm.DB{dbs[0]}
	default:
		dao.sourceDB = dbs[0]
		dao.replicaDB = dbs[1:]
	}
	return dao
}

func (d *OrdersCoursesDao) QueryOrderCoursesList(c *gin.Context, req *forms.QueryOrderCoursesListRequest) ([]*model.OrdersCourses, error) {
	uid := c.GetString("uid")
	userId := c.GetString("user_id")

	db := d.sourceDB.Model(d.m).
		Preload("Order.Coach.Users").
		Preload("Order.Coach.CoachTags.Tag").
		Preload("Order.Club.Users").
		Preload("Order.Club.ClubTags.Tag").
		Preload("Course").
		Preload("Comment", "user_type = ?", model.UserTypeUser).
		Preload("Reply", "user_type = ?", model.UserTypeCoach).
		Preload("Good").
		Preload("Good.CourseTags", "state = 0").
		Preload("Good.CourseTags.Tag", "state = 0").
		Where("state = 0").Where("uid = ?", uid)

	if len(req.TeachStates) > 0 {
		db = db.Where("teach_state in ?", req.TeachStates)
	}

	var ordersCourses []*model.OrdersCourses
	if err := db.Order(fmt.Sprintf("field(teach_state,%s) asc", model.GetUserOrderCourseOrderStr())).Order("id desc").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find(&ordersCourses).Error; err != nil {
		global.Lg.Error("QueryOrdersList error", zap.Error(err))
		return nil, err
	}
	d.ProcessUserCourseData(userId, ordersCourses)
	return ordersCourses, nil
}

func (d *OrdersCoursesDao) ProcessUserCourseData(userId string, ordersCourses []*model.OrdersCourses) {
	for _, v := range ordersCourses {
		dealGoodTags(v.Good)
		dealOrderCourseCommented(v)
		if str, ok := model.UserTeachStateStr[v.TeachState]; ok {
			v.Remark = str
		}
		if v.TeachState == model.TeachStateWaitConfirmTransfer && v.TeachCoachID != userId {
			v.Remark = "待教练确认转单"
		}

		stateToOperate := map[model.TeachState]int{
			model.TeachStateWaitUserSecondConfirmTime: model.OperateCoachChangeCourseTime,
			model.TeachStateWaitUserConfirmCoachTime:  model.OperateCoachChangeCourse,
			model.TeachStateWaitUserConfirmTransfer:   model.OperateCoachTransferCourse,
		}

		if op, ok := stateToOperate[v.TeachState]; ok {
			ordersCoursesState := model.OrdersCoursesState{}
			d.sourceDB.Model(&model.OrdersCoursesState{}).
				Where("order_course_id = ? and operate = ? and state = 0",
					v.OrderCourseID, op).
				Last(&ordersCoursesState)
			ordersCoursesState.LastConfirmTime = ordersCoursesState.CreatedAt.Add(24 * time.Hour)
			v.OrdersCoursesState = ordersCoursesState
			v.Remark = fmt.Sprintf(v.Remark, ordersCoursesState.LastConfirmTime.Format("2006-01-02 15:04"))
		}
	}

}

// ExitOrdersCoursesByGoodId  查询是否存在未核销的课程
func (d *OrdersCoursesDao) ExitOrdersCoursesByGoodId(ctx context.Context, goodId string) (*model.OrdersCourses, error) {
	var ordersCourses model.OrdersCourses
	err := d.sourceDB.Model(d.m).Where("good_id = ? and is_check=0 and state=0", goodId).First(&ordersCourses).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, enum.NewErr(enum.OrdersCoursesExitErr, "存在未核销的课程")
	}
	return &ordersCourses, nil
}

func (d *OrdersCoursesDao) AppointmentCourse(c *gin.Context, req *forms.AppointmentCourseRequest) (err error) {
	if req.TeachStartTime[14:16] != "00" && req.TeachStartTime[14:16] != "30" {
		return enum.NewErr(enum.TeachTimeErr, "时间格式错误,"+req.TeachStartTime[14:16]+"不是00或30")
	}
	startTi, err := time.ParseInLocation("2006-01-02 15:04", req.TeachStartTime[:16], time.Local)
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "时间格式错误")
	}
	if startTi.Before(time.Now()) {
		return enum.NewErr(enum.OrdersCoursesExitErr, "预约时间不能小于当前时间")
	}
	uid := c.GetString("uid")
	order, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return err
	}
	if order.Uid != uid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "只能预约自己的课程")
	}
	if order.Status != enum.OrderStatusPaid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单未支付")
	}
	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState != model.TeachStateWaitAppointment {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已预约")
	}

	var courseNum int64
	d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Where("order_id=? and state=0 and teach_state < ? and id < ?",
			order.OrderID, model.TeachStateWaitCheck, orderCourse.ID).
		Count(&courseNum)
	if courseNum > 0 { //打包课程，前面有课程未完成
		return enum.NewErr(enum.OrdersCoursesExitErr, "前面有课程未完成")
	}
	//上课时间段30 分钟一个时间段，不到30分钟按30分钟计算
	teachCount := int(math.Ceil(float64(orderCourse.TeachTime) / 30))
	var teachStartTimes []model.LocalTime

	upStartTi := startTi
	for i := 0; i < teachCount; i++ {
		teachStartTimes = append(teachStartTimes, model.LocalTime(upStartTi))
		upStartTi = upStartTi.Add(30 * time.Minute)
	}

	var skiNum int64
	global.DB.Model(&model.SkiResortsTeachTime{}).
		Where("user_id = ? and user_type = ? and ski_resorts_id != ? and teach_start_time in ? and teach_num = 0 and state = 0",
			order.UserID, order.UserType, req.SkiResortsId, teachStartTimes).
		Count(&skiNum)
	if skiNum > 0 {
		return enum.NewErr(enum.OrdersCoursesExitErr, "当前时间段与该教练其他雪场课程有冲突")
	}

	var ids []int64
	global.DB.Model(&model.SkiResortsTeachTime{}).
		Where("user_id = ? and user_type = ? and ski_resorts_id = ? and teach_start_time in ? and teach_state = ? and teach_num > 0 and state = 0",
			order.UserID, order.UserType, req.SkiResortsId, teachStartTimes, model.SkiTeachStateWaitAppointment).
		Pluck("id", &ids)
	if len(ids) != teachCount {
		return enum.NewErr(enum.OrdersCoursesExitErr, "时间冲突，请重新预约")
	}

	upData := map[string]interface{}{
		"teach_start_time": req.TeachStartTime[:16],
		"teach_state":      model.TeachStateWaitCoachConfirmUser,
		"ski_resorts_id":   req.SkiResortsId,
	}
	idsStr, _ := json.Marshal(ids)
	if order.UserType == enum.UserTypeClub {
		upData["teach_state"] = model.TeachStateWaitClubConfirm
		upData["club_time_ids"] = string(idsStr)
	} else {
		upData["teach_time_ids"] = string(idsStr)
	}
	//下面预约课程
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		// 使用乐观锁安全更新状态
		err = SafeUpdateTeachState(tx, req.OrderCourseId, model.TeachStateWaitAppointment, upData)
		if err != nil {
			return err
		}
		err = SRTOrderCourses(tx, ids, orderCourse.OrderCourseID, 0)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("req", req))
			return err
		}
		insrtocsData := model.OrdersCoursesState{
			OrderCourseID:  req.OrderCourseId,
			UserID:         uid,
			UserType:       enum.UserTypeUser,
			Operate:        model.OperateUserAppointment,
			Remark:         model.OCSOperateStr[model.OperateUserAppointment],
			TeachStartTime: model.LocalTime(startTi),
			TeachTimeIDs:   ids,
			Process:        model.ProcessNo,
		}
		err = tx.Model(model.OrdersCoursesState{}).Create(&insrtocsData).Error
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: 取消预约: %w", zap.Error(err), zap.Any("req", req))
			return err
		}
		teachState := model.TeachStateWaitCoachConfirmUser
		if order.UserType == enum.UserTypeClub {
			teachState = model.TeachStateWaitClubConfirm
		}
		err = UpdateOrderTeachState(c, tx, &order, teachState)
		return err
	})
	return nil
}

func (d *OrdersCoursesDao) BeforeCancelAppointmentCourse(c *gin.Context, req *forms.CancelAppointmentCourseRequest) (resp forms.BeforeCancelAppointmentCourseResp, err error) {
	uid := c.GetString("uid")

	order, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return
	}
	if order.Uid != uid {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "只能取消自己的课程")
	}
	if orderCourse.IsCheck == model.IsCheckYes {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	//TODO 判断状态等下处理
	if orderCourse.TeachState >= model.TeachStateFinish {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "课程已完成")
	}
	if order.Status != enum.OrderStatusPaid {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "订单未支付")
	}
	if orderCourse.TeachState != model.TeachStateWaitClass {
		return resp, enum.NewErr(model.OrderFaultCancel, "待上课状态才能取消预约")
	}

	resp = forms.BeforeCancelAppointmentCourseResp{
		Liability:    0,
		RefundMoney:  0,
		RefundPoints: 0,
	}
	//教练和俱乐部还没确认时，学员可以取消预约
	if orderCourse.TeachState == model.TeachStateWaitCoachConfirmUser || orderCourse.TeachState == model.TeachStateWaitClubConfirm {
		resp.Liability = 1
		return resp, nil
	}

	//离上课时间，2个自然日内，只能有责取消预约
	if time.Now().Add(2 * 24 * time.Hour).After(time.Time(orderCourse.TeachStartTime)) {
		resp.Liability = 2
		orderRefund, err := QueryOrderInfo("", order.OrderID)
		if err != nil {
			return resp, err
		}
		refundMoney, err := GetRefundMoney(*orderRefund, orderRefund.OrdersCourses)
		if err != nil {
			return resp, err
		}
		resp.RefundMoney = refundMoney.Money
		resp.RefundPoints = refundMoney.UsedPoints
		//退款金额
		return resp, nil
	}

	resp.Liability = 1
	year, month, _ := time.Now().Date()
	thisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	monthOneDay := thisMonth.Format("2006-01-02")
	var cancelNum int64 //用户无责取消预约的次数
	err = global.DB.Model(model.OrdersCoursesState{}).
		Where("user_id  = ? and  operate = ? and state = 0", order.Uid, model.OperateUserCancelNoResponsibility).
		Where("created_at > ? ", monthOneDay).
		Count(&cancelNum).Error
	if err != nil {
		return resp, enum.NewErr(enum.OrdersCoursesExitErr, "未找到预约记录")
	}
	if cancelNum >= model.OrderCourseUserCancelNumber {
		resp.Liability = 3
	}
	return resp, nil
}

type CancelAppointmentCourseRequest struct {
}

func (d *OrdersCoursesDao) CancelAppointmentCourse(c *gin.Context, req *forms.CancelAppointmentCourseRequest) (err error) {
	uid := c.GetString("uid")

	order, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return err
	}
	if order.Uid != uid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "只能取消自己的课程")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}

	//TODO 判断状态等下处理
	if orderCourse.TeachState >= model.TeachStateFinish {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已完成")
	}

	if order.Status != enum.OrderStatusPaid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单未支付")
	}
	if orderCourse.TeachState == model.TeachStateWaitAppointment {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程未预约")
	}

	if order.UserType == enum.UserTypeCoach {
		err = CancelCoachCourse(c, order, orderCourse)
	}

	if order.UserType == enum.UserTypeClub {
		err = CancelClubCourse(c, order, orderCourse)
	}
	if err != nil {
		global.Lg.Error("OrdersCoursesDao: 取消预约: %w", zap.Error(err), zap.Any("req", req))
		return err
	}
	return nil
}
func CancelCoachCourse(c *gin.Context, order model.Orders, orderCourse model.OrdersCourses) (err error) {
	operate := model.OperateUserCancelBeforeCoachConfirm
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
			Where("user_id  = ? and  operate = ? and state = 0", order.Uid, model.OperateUserCancelNoResponsibility).
			Where("created_at > ? ", monthOneDay).
			Order("id desc").Count(&cancelNum).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "未找到预约记录")
		}
		if cancelNum >= model.OrderCourseUserCancelNumber {
			return enum.NewErr(model.OrderFaultCancel, "您已用完本月无责取消机会，请联系客服")
		}
		operate = model.OperateUserCancelNoResponsibility
	}
	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        order.Uid,
		UserType:      enum.UserTypeUser,
		TeachTimeIDs:  model.JSONIntArray{},
		Operate:       operate,
		Remark:        model.OCSOperateStr[operate],
		Process:       model.ProcessYes,
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = CancelCoachCourseSql(c, tx, order, orderCourse, inspectorate)
		return err
	})
	return err
}

func CancelCoachCourseSql(c context.Context, tx *gorm.DB, order model.Orders, orderCourse model.OrdersCourses, inspectorate model.OrdersCoursesState) (err error) {
	//取消预约
	err = tx.Model(model.Orders{}).Where("order_id = ?", order.OrderID).
		Updates(map[string]interface{}{
			"transfer_fee":      0,
			"transfer_coach_id": "",
		}).Error

	if err != nil {
		global.Lg.Error("OrdersCoursesDao: 取消预约1: %w", zap.Error(err), zap.Any("order", order))
		return err
	}

	teachState := model.TeachStateWaitAppointment
	if inspectorate.Operate == model.OperateUserCancelCourse {
		teachState = model.TeachStateCancel
	}

	// 使用乐观锁安全更新状态
	err = SafeUpdateTeachState(tx, orderCourse.OrderCourseID, orderCourse.TeachState, map[string]interface{}{
		"teach_start_time":  gorm.Expr("NULL"),
		"teach_state":       teachState,
		"teach_buffer_time": 0,
		"teach_time_ids":    model.JSONIntArray{},
		"club_time_ids":     model.JSONIntArray{},
		"ski_resorts_id":    0,
		"teach_coach_id":    order.UserID, //可能转移过订单，这里要更新，把教学教练改回卖课教练ID
	})
	if err != nil {
		return err
	}

	err = SRTOrderCourses(tx, orderCourse.TeachTimeIDs, orderCourse.OrderCourseID, 1)
	if err != nil {
		global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
		return err
	}

	err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
	if err != nil {
		global.Lg.Error("OrdersCoursesDao:  取消预约: %w", zap.Error(err), zap.Any("inspectorate", inspectorate))
		return err
	}
	err = UpdateOrderTeachState(c, tx, &order, model.TeachStateWaitAppointment)
	if err != nil {
		global.Lg.Error("OrdersCoursesDao:  取消预约: %w", zap.Error(err), zap.Any("order", order))
		return err
	}

	err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and process=?",
		orderCourse.OrderCourseID, model.ProcessNo).
		Updates(map[string]interface{}{
			"process": model.ProcessYes,
		}).Error
	return err
}
func CancelClubCourse(c *gin.Context, order model.Orders, orderCourse model.OrdersCourses) (err error) {
	operate := model.OperateUserCancelBeforeClubConfirm
	//如果俱乐部还没确认，取消预约没有限制
	if orderCourse.TeachState != model.TeachStateWaitClubConfirm {
		if time.Now().Add(2 * 24 * time.Hour).After(time.Time(orderCourse.TeachStartTime)) {
			return enum.NewErr(model.OrderFaultCancel, "离上课时间，2个自然日内，只能有责取消预约")
		}
		if orderCourse.TeachState != model.TeachStateWaitCoachClass {
			return enum.NewErr(model.OrderFaultCancel, "待上课状态才能取消预约")
		}
		year, month, _ := time.Now().Date()
		thisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		monthOneDay := thisMonth.Format("2006-01-02")
		var cancelNum int64 //用户无责取消预约的次数
		err = global.DB.Model(model.OrdersCoursesState{}).
			Where("user_id  = ? and  operate = ? and state = 0", order.Uid, model.OperateUserCancelNoResponsibility).
			Where("created_at > ? ", monthOneDay).
			Order("id desc").Count(&cancelNum).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "未找到预约记录")
		}
		if cancelNum >= model.OrderCourseUserCancelNumber {
			return enum.NewErr(model.OrderFaultCancel, "您已用完本月无责取消机会，请联系客服")
		}
		operate = model.OperateUserCancelNoResponsibility
	}

	inspectorate := model.OrdersCoursesState{
		OrderCourseID: orderCourse.OrderCourseID,
		UserID:        order.Uid,
		UserType:      enum.UserTypeUser,
		TeachTimeIDs:  model.JSONIntArray{},
		Operate:       operate,
		Remark:        model.OCSOperateStr[operate],
		Process:       model.ProcessYes,
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = CancelClubCourseSql(c, tx, order, orderCourse, inspectorate)
		return err
	})
	return err
}
func CancelClubCourseSql(c context.Context, tx *gorm.DB, order model.Orders, orderCourse model.OrdersCourses, inspectorate model.OrdersCoursesState) (err error) {
	teachState := model.TeachStateWaitAppointment
	if inspectorate.Operate == model.OperateUserCancelCourse {
		teachState = model.TeachStateCancel
	}
	// 使用乐观锁安全更新状态
	err = SafeUpdateTeachState(tx, orderCourse.OrderCourseID, orderCourse.TeachState, map[string]interface{}{
		"teach_start_time":  gorm.Expr("NULL"),
		"teach_state":       teachState,
		"teach_buffer_time": 0,
		"teach_time_ids":    model.JSONIntArray{},
		"club_time_ids":     model.JSONIntArray{},
		"ski_resorts_id":    0,
		"teach_coach_id":    "",
	})
	if err != nil {
		return err
	}

	timeIds := orderCourse.ClubTimeIDs
	if len(orderCourse.TeachTimeIDs) > 0 {
		timeIds = append(timeIds, orderCourse.TeachTimeIDs...)
	}
	err = SRTOrderCourses(tx, timeIds, orderCourse.OrderCourseID, 1)
	if err != nil {
		global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
		return err
	}
	err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
	if err != nil {
		global.Lg.Error("OrdersCoursesDao:  取消预约: %w", zap.Error(err), zap.Any("inspectorate", inspectorate))
		return err
	}
	err = UpdateOrderTeachState(c, tx, &order, model.TeachStateWaitAppointment)
	if err != nil {
		global.Lg.Error("OrdersCoursesDao:  取消预约: %w", zap.Error(err), zap.Any("order", order))
		return err
	}

	err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and process=?",
		orderCourse.OrderCourseID, model.ProcessNo).
		Updates(map[string]interface{}{
			"process": model.ProcessYes,
		}).Error
	return err
}
func UserVerifyCourses(c *gin.Context, orderCourseId string) (err error) {
	uid := c.GetString("uid")
	orderCourse := model.OrdersCourses{}
	err = global.DB.Model(model.OrdersCourses{}).
		Where("order_course_id = ? and state = 0", orderCourseId).First(&orderCourse).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单课程不存在")
	}
	order := model.Orders{}
	err = global.DB.Model(model.Orders{}).Where("order_id = ? and uid = ? and state = 0", orderCourse.OrderID, uid).First(&order).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单不是您的，请不要乱操作")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState != model.TeachStateWaitClass && orderCourse.TeachState != model.TeachStateWaitCheck && orderCourse.TeachState != model.TeachStateWaitCoachClass && orderCourse.TeachState != model.TeachStateAlreadyClass {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程状态错误")
	}
	year, month, day := time.Now().Date()
	thisDay := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	if thisDay.Before(time.Time(orderCourse.TeachStartTime)) {
		return enum.NewErr(enum.OrdersCoursesExitErr, "刚上完课，明天才能核销")
	}

	insrtocsData := model.OrdersCoursesState{
		OrderCourseID: orderCourseId,
		UserID:        c.GetString("user_id"),
		UserType:      c.GetInt("user_type"),
		Operate:       model.OperateUserVerifyCourse,
		Remark:        model.OCSOperateStr[model.OperateUserVerifyCourse],
	}
	err = CompleteCourseSplitMoney(c, orderCourse, order, insrtocsData)
	return err
}

func ReviewTeachTime(c *gin.Context, req *forms.ReviewTeachTimeRequest) (err error) {
	uid := c.GetString("uid")
	order, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单课程不存在")
	}
	if order.Uid != uid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单不是您的，请不要乱操作")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}

	if orderCourse.TeachState == model.TeachStateWaitUserConfirmCoachTime || orderCourse.TeachState == model.TeachStateWaitUserSecondConfirmTime {
		return ReviewCoachTeachTime(c, &orderCourse, &order, req.IsAgree)
	}
	if orderCourse.TeachState == model.TeachStateWaitUserConfirmClubTime {
		return ReviewClubTeachTime(c, &orderCourse, &order, req.IsAgree)
	}
	return enum.NewErr(enum.OrdersCoursesExitErr, "课程暂不需要您确认")
}
func ReviewCoachTeachTime(c *gin.Context, orderCourse *model.OrdersCourses, order *model.Orders, isAgree bool) (err error) {
	uid := order.Uid
	orderCourseId := orderCourse.OrderCourseID
	oldOperate := model.OperateCoachChangeCourse
	newOperate := model.OperateUserAgreeCoachChangeCourse
	teachState := model.TeachStateWaitClass
	if orderCourse.TeachState == model.TeachStateWaitUserSecondConfirmTime {
		oldOperate = model.OperateCoachChangeCourseTime
		newOperate = model.OperateUserAgreeChangeTimeBeforeC
	}
	userId := order.UserID
	if orderCourse.TeachCoachID != "" {
		userId = orderCourse.TeachCoachID
	}
	//查出教学教练修改时间的申请
	ocsData := model.OrdersCoursesState{}
	err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and user_id = ? and operate=? and state=0",
		orderCourseId, userId, oldOperate).
		Last(&ocsData).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "没有申请修改时间")
	}

	if ocsData.Process == model.ProcessYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "请勿重复操作")
	}
	if isAgree == true { //同意教练修改上课时间
		idsStr, err := json.Marshal(ocsData.TeachTimeIDs)
		if err != nil {
			global.Lg.Error("解析数据失败", zap.Error(err), zap.Any("ocsData", ocsData))
			return enum.NewErr(enum.OrdersCoursesExitErr, "解析数据失败")
		}
		err = global.DB.Transaction(func(tx *gorm.DB) error {
			inspectorate := model.OrdersCoursesState{
				OrderCourseID: orderCourseId,
				UserID:        uid,
				UserType:      model.UserTypeUser,
				Operate:       newOperate,
				Remark:        model.OCSOperateStr[newOperate],
				Process:       model.ProcessYes,
			}
			err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "插入记录失败")
			}
			// 使用乐观锁安全更新状态
			err = SafeUpdateTeachState(tx, orderCourseId, orderCourse.TeachState, map[string]interface{}{
				"teach_start_time": ocsData.TeachStartTime,
				"teach_state":      teachState,
				"teach_time_ids":   string(idsStr),
			})
			if err != nil {
				return err
			}
			//同意教练修改的课程时间，那就要把之前锁定的时间释放掉
			err = SRTOrderCourses(tx, orderCourse.TeachTimeIDs, orderCourseId, 1)
			if err != nil {
				global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
				return err
			}
			err = tx.Model(model.OrdersCoursesState{}).Where("id = ?", ocsData.ID).
				Updates(map[string]interface{}{
					"process": model.ProcessYes,
				}).Error
			return err
		})
	} else { //不同意教练修改的上课时间
		newOperate = model.OperateUserDisagreeCoachChangeCourse
		if orderCourse.TeachState == model.TeachStateWaitUserSecondConfirmTime {
			newOperate = model.OperateUserDisagreeChangeTimeBeforeC
		}
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourseId,
			UserID:        uid,
			UserType:      model.UserTypeUser,
			Operate:       newOperate,
			Remark:        model.OCSOperateStr[newOperate],
			Process:       model.ProcessYes,
		}
		err = UserDisAgreeCoachTeachTimeSql(c, *orderCourse, inspectorate, ocsData)
		if err != nil {
			return err
		}
	}
	return err
}

func UserDisAgreeCoachTeachTimeSql(c context.Context, orderCourse model.OrdersCourses, inspectorate model.OrdersCoursesState, ocs model.OrdersCoursesState) (err error) {
	teachState := model.TeachStateWaitClass
	if orderCourse.TeachState == model.TeachStateWaitUserSecondConfirmTime {
		teachState = model.TeachStateWaitCoachConfirmUser
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "插入记录失败")
		}

		// 使用乐观锁安全更新状态
		err = SafeUpdateTeachState(tx, orderCourse.OrderCourseID, orderCourse.TeachState, map[string]interface{}{
			"teach_state": teachState,
		})
		if err != nil {
			return err
		}
		//拒绝教练修改的课程时间，那就要把之前锁定的时间释放掉
		err = SRTOrderCourses(tx, ocs.TeachTimeIDs, orderCourse.OrderCourseID, 1)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return err
		}
		err = tx.Model(model.OrdersCoursesState{}).Where("id = ?", ocs.ID).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		return err
	})
	return err
}

func ReviewClubTeachTime(c *gin.Context, orderCourse *model.OrdersCourses, order *model.Orders, isAgree bool) (err error) {
	uid := order.Uid
	orderCourseId := orderCourse.OrderCourseID
	//查出俱乐部修改时间的申请
	ocsClubData := model.OrdersCoursesState{}
	err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and user_id = ? and operate=? and process = ? and state=0",
		orderCourseId, order.UserID, model.OperateClubChangeUserCourseTime, model.ProcessNo).
		Last(&ocsClubData).Error
	if err != nil {
		return enum.NewErr(enum.OrdersCoursesExitErr, "没有申请修改时间记录")
	}
	ocsCoachData := model.OrdersCoursesState{}
	if orderCourse.TeachCoachID != "" {
		err = global.DB.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and user_id = ? and operate=? and process = ? and state=0",
			orderCourseId, orderCourse.TeachCoachID, model.OperateClubChangeUserCourseTime, model.ProcessNo).
			Last(&ocsCoachData).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "没有教练的申请修改时间记录")
		}
	}
	ids := []int64{
		ocsClubData.ID,
	}
	if ocsCoachData.ID != 0 {
		ids = append(ids, ocsCoachData.ID)
	}

	if isAgree == true { //同意俱乐部修改上课时间
		err = global.DB.Transaction(func(tx *gorm.DB) error {
			clubIdsStr, err := json.Marshal(ocsClubData.TeachTimeIDs)
			if err != nil {
				global.Lg.Error("解析数据失败", zap.Error(err), zap.Any("ocsCoachData", ocsCoachData))
				return enum.NewErr(enum.OrdersCoursesExitErr, "解析数据失败")
			}
			upOrdersCourses := map[string]interface{}{
				"teach_start_time": ocsClubData.TeachStartTime,
				"teach_state":      model.TeachStateWaitClubConfirm,
				"club_time_ids":    string(clubIdsStr),
			}
			insrtocsData := model.OrdersCoursesState{
				OrderCourseID: orderCourseId,
				UserID:        uid,
				UserType:      model.UserTypeUser,
				Operate:       model.OperateUserAgreeClubChangeCourseTime,
				Remark:        model.OCSOperateStr[model.OperateUserAgreeClubChangeCourseTime],
				Process:       model.ProcessYes,
			}
			err = tx.Model(model.OrdersCoursesState{}).Create(&insrtocsData).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "插入同意俱乐部记录失败")
			}

			if orderCourse.TeachCoachID != "" { //已经分配教练
				upOrdersCourses["teach_state"] = model.TeachStateWaitCoachClass
				idsStr, err := json.Marshal(ocsCoachData.TeachTimeIDs)
				if err != nil {
					global.Lg.Error("解析数据失败", zap.Error(err), zap.Any("ocsCoachData", ocsCoachData))
					return enum.NewErr(enum.OrdersCoursesExitErr, "解析数据失败")
				}
				upOrdersCourses["teach_time_ids"] = string(idsStr)
			}

			// 使用乐观锁安全更新状态
			err = SafeUpdateTeachState(tx, orderCourseId, orderCourse.TeachState, upOrdersCourses)
			if err != nil {
				return err
			}

			//同意俱乐部修改的课程时间，那就要把之前锁定俱乐部和教练的时间释放掉
			err = SRTOrderCourses(tx, append(orderCourse.ClubTimeIDs, []int64(orderCourse.TeachTimeIDs)...), orderCourse.OrderCourseID, 1)
			if err != nil {
				global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
				return err
			}
			err = tx.Model(model.OrdersCoursesState{}).Where("id in (?)", ids).
				Updates(map[string]interface{}{
					"process": model.ProcessYes,
				}).Error
			return err
		})
	} else { //不同意教练修改的上课时间
		inspectorate := model.OrdersCoursesState{
			OrderCourseID: orderCourseId,
			UserID:        uid,
			UserType:      model.UserTypeUser,
			Operate:       model.OperateUserDisagreeClubChangeCourseTime,
			Remark:        model.OCSOperateStr[model.OperateUserDisagreeClubChangeCourseTime],
			Process:       model.ProcessYes,
		}
		err = UserDisAgreeClubTeachTimeSql(c, *orderCourse, inspectorate, ocsClubData.TeachTimeIDs, ocsCoachData.TeachTimeIDs)
	}
	return err
}

func UserDisAgreeClubTeachTimeSql(c context.Context, orderCourse model.OrdersCourses, inspectorate model.OrdersCoursesState, clubTimeIds model.JSONIntArray, coachTimeIds model.JSONIntArray) (err error) {
	err = global.DB.Transaction(func(tx *gorm.DB) error {

		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "插入记录失败")
		}

		// 使用乐观锁安全更新状态
		err = SafeUpdateTeachState(tx, orderCourse.OrderCourseID, orderCourse.TeachState, map[string]interface{}{
			"teach_state": model.TeachStateWaitClubConfirm,
		})
		if err != nil {
			return err
		}
		//拒绝俱乐部和教练修改的课程时间，那就要把之前锁定的时间释放掉
		err = SRTOrderCourses(tx, append(clubTimeIds, []int64(coachTimeIds)...), orderCourse.OrderCourseID, 1)
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SRTOrderCourses: %w", zap.Error(err), zap.Any("orderCourse", orderCourse))
			return err
		}
		err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process = ? and state=0",
			orderCourse.OrderCourseID, model.OperateClubChangeUserCourseTime, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error

		return err
	})
	return err
}
func ReviewCoachTransferOrder(c *gin.Context, req *forms.ReviewCoachTransferOrderRequest) (err error) {
	uid := c.GetString("uid")
	order, orderCourse, err := GetOrderCourses(req.OrderCourseId)
	if err != nil {
		return err
	}
	if order.Uid != uid {
		return enum.NewErr(enum.OrdersCoursesExitErr, "订单不是您的，请不要乱操作")
	}

	if orderCourse.IsCheck == model.IsCheckYes {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程已核销")
	}
	if orderCourse.TeachState != model.TeachStateWaitUserConfirmTransfer {
		return enum.NewErr(enum.OrdersCoursesExitErr, "课程暂不需要您确认转单")
	}
	if req.IsAgree { //同意教练转单
		err = global.DB.Transaction(func(tx *gorm.DB) error {
			inspectorate := model.OrdersCoursesState{
				OrderCourseID: req.OrderCourseId,
				UserID:        uid,
				UserType:      model.UserTypeUser,
				Operate:       model.OperateUserAgreeCoachTransferCourse,
				Remark:        model.OCSOperateStr[model.OperateUserAgreeCoachTransferCourse],
				Process:       model.ProcessNo,
			}
			err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
			if err != nil {
				return enum.NewErr(enum.OrdersCoursesExitErr, "插入记录失败")
			}
			// 使用乐观锁安全更新状态
			err = SafeUpdateTeachState(tx, req.OrderCourseId, model.TeachStateWaitUserConfirmTransfer, map[string]interface{}{
				"teach_state": model.TeachStateWaitCoachTransfer,
			})
			if err != nil {
				return err
			}
			err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
				orderCourse.OrderCourseID, model.OperateCoachTransferCourse, model.ProcessNo).
				Updates(map[string]interface{}{
					"process": model.ProcessYes,
				}).Error
			return err
		})
	} else { //拒绝教练转单
		insrtocsData := model.OrdersCoursesState{
			OrderCourseID: req.OrderCourseId,
			UserID:        c.GetString("uid"),
			UserType:      model.UserTypeUser,
			Operate:       model.OperateUserDisagreeCoachTransferCourse,
			Remark:        model.OCSOperateStr[model.OperateUserDisagreeCoachTransferCourse],
			Process:       model.ProcessYes,
		}
		err = CancelCoachTransferOrderSql(c, orderCourse, insrtocsData)
	}

	return err
}

func CancelCoachTransferOrderSql(c context.Context, orderCourse model.OrdersCourses, inspectorate model.OrdersCoursesState) (err error) {
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(model.OrdersCoursesState{}).Create(&inspectorate).Error
		if err != nil {
			return enum.NewErr(enum.OrdersCoursesExitErr, "插入记录失败")
		}
		// 使用乐观锁安全更新状态
		err = SafeUpdateTeachState(tx, orderCourse.OrderCourseID, model.TeachStateWaitUserConfirmTransfer, map[string]interface{}{
			"teach_state": model.TeachStateWaitClass,
		})
		if err != nil {
			return err
		}
		err = tx.Model(model.OrdersCoursesState{}).Where("order_course_id = ? and operate=? and process=?",
			orderCourse.OrderCourseID, model.OperateCoachTransferCourse, model.ProcessNo).
			Updates(map[string]interface{}{
				"process": model.ProcessYes,
			}).Error
		return err
	})
	return err
}

func (d *OrdersCoursesDao) QueryGoodsCourses(ctx context.Context, goodId string) ([]*model.OrdersCourses, error) {
	var ordersCourses []*model.OrdersCourses
	err := d.sourceDB.Model(d.m).Where("good_id = ? and state=0", goodId).Find(&ordersCourses).Error
	if err != nil {
		global.Lg.Error("查询商品课程失败", zap.Error(err), zap.String("good_id", goodId))
		return nil, err
	}
	return ordersCourses, nil
}

func QueryOrderCourseInfo(ctx context.Context, uid, orderCourseId string) (*model.OrdersCourses, error) {
	var orderCourse *model.OrdersCourses
	db := global.DB.Model(model.OrdersCourses{}).
		Preload("Order.Coach.Users").
		Preload("Comment", "user_type = ?", model.UserTypeUser).
		Preload("Reply", "user_type = ?", model.UserTypeCoach).
		Preload("UserInfo").
		Preload("Good").
		Preload("Good.CourseTags", "state = 0").
		Preload("Good.CourseTags.Tag", "state = 0").
		Where("order_course_id = ? and state = 0", orderCourseId)
	if uid != "" {
		db = db.Where("uid = ?", uid)
	}
	err := db.First(&orderCourse).Error
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err), zap.String("uid", uid), zap.String("order_course_id", orderCourseId))
		return nil, err
	}
	global.Lg.Info("ordercourse", zap.Any("ordercourse", orderCourse))
	dealGoodTags(orderCourse.Good)
	dealOrderCourseCommented(orderCourse)

	NewSkiResortsTeachTimeDao(ctx, global.DB).dealOrderCourseTeachTime(ctx, orderCourse)
	return orderCourse, nil
}

func OrderCourseInfo(ctx context.Context, uid, orderCourseId string) (*model.OrdersCourses, error) {
	var orderCourse *model.OrdersCourses
	db := global.DB.Model(model.OrdersCourses{}).Where("order_course_id = ? and state = 0", orderCourseId)
	if uid != "" {
		db = db.Where("uid = ?", uid)
	}
	err := db.First(&orderCourse).Error
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err), zap.String("uid", uid), zap.String("order_course_id", orderCourseId))
		return nil, err
	}

	return orderCourse, nil
}

func dealOrderCourseCommented(orderCourse *model.OrdersCourses) {
	if orderCourse.Comment != nil {
		orderCourse.CommentId = orderCourse.Comment.Id
	}

	if orderCourse.Reply != nil {
		orderCourse.Replied = true
	}
}

func (d *OrdersCoursesDao) QueryCoachOrderCourses(c *gin.Context, req *forms.QueryOrderCoursesListRequest) (int64, []*model.OrdersCourses, error) {
	userId := c.GetString("user_id")
	//先手动查，先冗余
	var orderIds []string
	if err := global.DB.Model(&model.Orders{}).Where("user_id = ? and state = 0", userId).Pluck("order_id", &orderIds).Error; err != nil {
		global.Lg.Error("QueryCoachOrderCourses error", zap.Error(err))
		return 0, nil, err
	}
	var transferCoachIds []string
	if err := global.DB.Model(&model.Orders{}).Where("transfer_coach_id = ? and state = 0", userId).Pluck("order_id", &transferCoachIds).Error; err == nil {
		orderIds = append(orderIds, transferCoachIds...)
	}

	var teachCoachIds []string
	if err := global.DB.Model(&model.OrdersCourses{}).Where("teach_coach_id = ? and state = 0", userId).Pluck("order_id", &teachCoachIds).Error; err == nil {
		orderIds = append(orderIds, teachCoachIds...)
	}
	orderIds = UniqueSlice(orderIds)

	if len(orderIds) == 0 {
		return 0, []*model.OrdersCourses{}, nil
	}

	db := d.sourceDB.Model(d.m).
		Preload("Order.Coach.Users").
		Preload("Comment", "user_type = ?", model.UserTypeUser).
		Preload("Reply", "user_type = ?", model.UserTypeCoach).
		Preload("Good").
		Preload("Good.CourseTags", "state = 0").
		Preload("Good.CourseTags.Tag", "state = 0").
		Where("state = 0").Where("order_id in ?", orderIds)

	if len(req.TeachStates) > 0 {
		db = db.Where("teach_state in ?", req.TeachStates)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("QueryOrdersList error", zap.Error(err))
		return 0, nil, err
	}

	var ordersCourses []*model.OrdersCourses
	if err := db.Order("id desc").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find(&ordersCourses).Error; err != nil {
		global.Lg.Error("QueryOrdersList error", zap.Error(err))
		return 0, nil, err
	}

	//处理数据
	d.ProcessCoachCourseData(ordersCourses)
	return total, ordersCourses, nil
}

func UniqueSlice[T comparable](slice []T) []T {
	keys := make(map[T]struct{}) // 使用struct{}作为值类型，因为只需要键的唯一性
	list := []T{}
	for _, entry := range slice {
		if _, exists := keys[entry]; !exists {
			keys[entry] = struct{}{} // 空结构体用作占位符
			list = append(list, entry)
		}
	}
	return list
}

func (d *OrdersCoursesDao) QueryClubOrderCourses(c *gin.Context, req *forms.QueryClubOrderCoursesRequest) (int64, []*model.OrdersCourses, error) {
	userId := c.GetString("user_id")
	//先手动查，先冗余
	var orderIds []string
	if err := global.DB.Model(&model.Orders{}).Where("user_id = ? and state = 0", userId).Pluck("order_id", &orderIds).Error; err != nil {
		global.Lg.Error("QueryCoachOrderCourses error", zap.Error(err))
		return 0, nil, err
	}

	if len(orderIds) == 0 {
		return 0, []*model.OrdersCourses{}, nil
	}

	db := d.sourceDB.Model(d.m).
		Preload("Order.Club").
		Preload("Comment", "user_type = ?", model.UserTypeUser).
		Preload("Reply", "user_type = ?", model.UserTypeClub).
		Preload("Good").
		Preload("Good.CourseTags", "state = 0").
		Preload("Good.CourseTags.Tag", "state = 0").
		Where("state = 0").Where("order_id in ?", orderIds)

	if len(req.TeachStates) > 0 {
		db = db.Where("teach_state in ?", req.TeachStates)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("QueryOrdersList error", zap.Error(err))
		return 0, nil, err
	}

	var ordersCourses []*model.OrdersCourses
	if err := db.Order("id desc").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find(&ordersCourses).Error; err != nil {
		global.Lg.Error("QueryOrdersList error", zap.Error(err))
		return 0, nil, err
	}

	//处理数据
	d.ProcessCoachCourseData(ordersCourses)
	return total, ordersCourses, nil
}
func (d *OrdersCoursesDao) ProcessCoachCourseData(ordersCourses []*model.OrdersCourses) {
	for _, v := range ordersCourses {
		dealGoodTags(v.Good)
		dealOrderCourseCommented(v)
		if str, ok := model.CoachTeachStateStr[v.TeachState]; ok {
			v.Remark = str
		}

		stateToOperate := map[model.TeachState]int{
			model.TeachStateWaitCoachConfirmUser: model.OperateUserAppointment,
		}
		if op, ok := stateToOperate[v.TeachState]; ok {
			ordersCoursesState := model.OrdersCoursesState{}
			d.sourceDB.Model(&model.OrdersCoursesState{}).
				Where("order_course_id = ? and operate = ? and state = 0",
					v.OrderCourseID, op).
				Last(&ordersCoursesState)
			ordersCoursesState.LastConfirmTime = ordersCoursesState.CreatedAt.Add(24 * time.Hour)
			v.OrdersCoursesState = ordersCoursesState
			v.Remark = fmt.Sprintf(v.Remark, ordersCoursesState.LastConfirmTime.Format("2006-01-02 15:04"))
		}
	}

}
