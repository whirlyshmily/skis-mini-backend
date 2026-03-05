package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/response"
	"skis-admin-backend/services"
	"time"
)

func MoneyTest(c *gin.Context) {
	orderCourseId := c.Param("order_course_id")
	if orderCourseId == "" {
		response.Err(c, enum.NewErr(enum.ParamErr, "订单课程ID不能为空"))
		return
	}
	orderCourse := model.OrdersCourses{}
	global.DB.Model(&model.OrdersCourses{}).Where("order_course_id = ?", orderCourseId).First(&orderCourse)
	insrtocsData := model.OrdersCoursesState{
		OrderCourseID: orderCourseId,
		UserID:        c.GetString("user_id"),
		UserType:      c.GetInt("user_type"),
		Operate:       model.OperateUserVerifyCourse,
		Remark:        model.OCSOperateStr[model.OperateUserVerifyCourse],
	}
	order := model.Orders{}
	err := global.DB.Model(model.Orders{}).Where("order_id = ? and state = 0", orderCourse.OrderID).First(&order).Error
	if err != nil {
		global.Lg.Error("没有找到订单", zap.Error(err), zap.Any("orderCourse", orderCourse))
		return
	}
	err = dao.CompleteCourseSplitMoney(c, orderCourse, order, insrtocsData)
	if err != nil {
		global.Lg.Error("核销课程失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "核销课程成功")
}
func QueryMoneyList(c *gin.Context) {
	var req forms.QueryMoneyListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	list, err := dao.QueryMoneyList(c, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, list)
	return
}

func QueryMoneyInfo(c *gin.Context) {
	moneyId := c.Param("money_id")
	info, err := dao.QueryMoneyInfoByMoneyId(c, moneyId)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, info)
	return
}

func MoneyOperateWithdraw(c *gin.Context) {
	var req forms.MoneyOperateWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	data, err := dao.MoneyOperateWithdraw(c, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, map[string]interface{}{"data": data, "mch_id": global.Config.Mch.MchId})
	return
}
func MoneyOperateRecharge(c *gin.Context) {
	var req forms.MoneyOperateRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	prePayResp, err := dao.MoneyOperateRecharge(c, &req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, map[string]interface{}{"pre_pay_resp": prePayResp})
	return
}

func DepositPayCallback(c *gin.Context) {
	if err := dao.MoneyOperateRechargeCallback(c, c.Request); err != nil {
		global.Lg.Info("保证金支付回调失败", zap.Any("error", err.Error()))
		response.Err(c, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func DepositTestPay(c *gin.Context) {

	resp, err := services.CreateTransferBill(global.Config.UserMiniProgram.AppId, "U20250910115251cycBh121", "1005", "o4yE-1_aA_F-6M2AzthxsBOdhIRM", 30, "理财返佣", "尹永明", global.Config.Mch.OrderRefundNotifyUrl, "是大法官")
	response.Success(c, map[string]interface{}{"resp": resp, "err": err})
	return
	aa, _ := dao.NewGoodsDao(c, global.DB).GetMaxPriceByUserId("123")
	global.Lg.Info("aa", zap.Any("aa", aa))
	response.Success(c, aa)
	return

	payTime, err := time.Parse(time.RFC3339, "2025-09-11T17:27:06+08:00")
	if err != nil {
		payTime = time.Now()
		global.Lg.Error("时间转换失败", zap.Error(err))
	}
	global.DB.Model(&model.MoneyOperate{}).Where("operate_id = ?", "M20250911174830NrNkPQ").Updates(map[string]interface{}{
		"status":         model.StatusRecharged,
		"transaction_id": "4200002858202509117400013140",
		"pay_time":       payTime,
	})
	response.Success(c, payTime)
	return
}

func TransferBills(c *gin.Context) {
	if err := dao.TransferBillsCallback(c, c.Request); err != nil {
		global.Lg.Info("转账回调", zap.Any("error", err.Error()))
		response.Err(c, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}
