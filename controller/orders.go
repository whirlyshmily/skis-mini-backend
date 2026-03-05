package controller

import (
	"net/http"
	"skis-admin-backend/cron"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/response"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateOrder(c *gin.Context) {
	uid := c.GetString("uid")
	openId := c.GetString("open_id")

	//创建订单
	var req forms.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("创建订单参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	order, err := dao.CreateOrder(c, uid, openId, &req)
	if err != nil {
		global.Lg.Error("创建订单失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, order)
	return

}

func TestQueryOrdersList(c *gin.Context) {
	orders, err := dao.TestQueryOrdersList()
	if err != nil {
		global.Lg.Error("查询订单列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, orders)
	return
}

func QueryOrdersList(c *gin.Context) {
	uid := c.GetString("uid")
	var req forms.QueryOrdersListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, orders, err := dao.QueryOrdersList(uid, &req)
	if err != nil {
		global.Lg.Error("查询订单列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, &forms.QueryOrdersListResp{
		Total: total,
		List:  orders,
	})
	return
}

func QueryOrderInfo(c *gin.Context) {
	uid := c.GetString("uid")
	userId := c.GetString("user_id")
	orderId := c.Param("order_id")
	order, err := dao.QueryOrderInfo("", orderId)
	if err != nil {
		global.Lg.Error("查询订单详情失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	if uid != order.Uid && order.UserID != userId {
		response.Err(c, enum.NewErr(enum.OrderNotExistErr, "只能查询自己的订单"))
		return
	}
	response.Success(c, order)
	return
}

func TestWhirly(c *gin.Context) {
	var data []model.OrdersCoursesState
	global.DB.Model(model.OrdersCoursesState{}).
		Where("created_at between now() - interval 120 day and now() - interval 1 day").
		Where(" process = ?", model.ProcessNo).
		Find(&data)

	/*global.DB.Model(model.OrdersCoursesState{}).
	Where("created_at between now() - interval 13 day and now() - interval 2 day").
	Where("operate = ? and process = ?", model.OperateUserAgreeCoachTransferCourse, model.ProcessNo).
	Find(&data)*/
	for _, ocs := range data {
		cron.OrdersCoursesStateJobProcessData(c, ocs)
	}
	response.Success(c, data)
	return
	record := model.MoneyRecords{
		Money:        100,
		IncomeType:   model.IncomeTypePay,
		RelationType: model.RelationTypeOrder,
		UserType:     model.UserTypeCoach,
		UserID:       "C20250823181716sjyrgjh",
		MoneyType:    model.CoachPayOneCRefundFaultService,
		RelationID:   "order20250909160450j3i18o",
	}
	err := dao.NewMoneyRecordsDao(c.Request.Context(), global.DB).Create(c, &record, global.DB)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "成功")
	return
}

func OrderPayCallback(c *gin.Context) {
	//{"mchid":"1726738364","appid":"wx553446f6516fa0a0","out_trade_no":"order202511072100014ef3ec","transaction_id":"4200002850202511077419076156","trade_type":"JSAPI","trade_state":"SUCCESS","trade_state_desc":"支付成功","bank_type":"CMB_DEBIT","attach":"自定义数据说明","success_time":"2025-11-07T21:00:17+08:00","payer":{"openid":"o4yE-1xyrKjYsHgl0AnTr_oEoIX8"},"amount":{"total":3,"payer_total":3,"currency":"CNY","payer_currency":"CNY"}}
	global.Lg.Info("订单支付回调", zap.Any("params", c.Request.Body))

	if err := dao.OrderPayCallback(c); err != nil {
		global.Lg.Info("支付回调失败", zap.Any("error", err.Error()))
		response.Err(c, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func WhirlyOrderPayCallback(c *gin.Context) {
	var req forms.WhirlyOrderPayCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("查询订单列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	err := dao.OrderPayCallbackSql(c, req.OrderId, req.TransactionId, req.SuccessTime)
	if err != nil {
		global.Lg.Error("查询订单列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	good1, _ := dao.NewGoodsDao(c, global.DB).QueryGoodInfo("G20250925165842bad8bd")
	response.Success(c, good1)
	return
}

func QueryOrdersRefundInfo(c *gin.Context) {
	refund, err := dao.QueryOrdersRefundInfo(c, c.GetString("uid"), c.Param("order_id"))
	if err != nil {
		global.Lg.Error("查询订单退款信息失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, refund)
	return
}

func OrderRefund(c *gin.Context) {
	uid := c.GetString("uid")
	if uid == "" {
		global.Lg.Error("用户ID不存在")
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不存在"))
		return
	}
	err := dao.OrderRefund(c, uid, c.Param("order_id"))
	if err != nil {
		global.Lg.Error("订单退款失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, nil)
}

func OrderRefundCallback(c *gin.Context) {
	global.Lg.Info("订单退款回调", zap.Any("params", c.Request.Body))

	if err := dao.OrderRefundCallback(c); err != nil {
		global.Lg.Info("退款回调失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func OrderRefundCallbackTest(c *gin.Context) {
	t := time.Now()
	req := dao.RefundCallbackResult{
		Mchid:         "160450909",
		OutRefundNo:   "refund_20251106224140e532bc",
		RefundId:      "50303505272025110673969648315",
		OutTradeNo:    "order20251106185546e48a41",
		TransactionId: "",
		RefundStatus:  "SUCCESS",
		SuccessTime:   &t,
		Amount: dao.Amount{
			Total:  100,
			Refund: 100,
		},
	}
	if err := dao.OrdersRefundCallback(c, &req); err != nil {
		global.Lg.Info("退款回调失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "成功")
	return
}

func QueryOrderCoursesList(c *gin.Context) {
	var req forms.QueryOrderCoursesListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("查询订单课程列表参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	orderCourses, err := dao.NewOrdersCoursesDao(c, global.DB).QueryOrderCoursesList(c, &req)
	if err != nil {
		global.Lg.Error("查询订单课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, map[string]interface{}{
		"list": orderCourses,
	})
	return
}

func AdminQueryOrderRefundLimit(c *gin.Context) {
	refundLimit, err := dao.AdminQueryOrderRefundLimit(c, c.Param("order_id"))
	if err != nil {
		global.Lg.Error("查询订单退款限制失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, refundLimit)
	return
}

func AdminOrderRefund(c *gin.Context) {
	var req forms.AdminOrderRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("订单退款参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.AdminOrderRefund(c, c.Param("order_id"), &req)
	if err != nil {
		global.Lg.Error("订单退款失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, nil)
	return
}
