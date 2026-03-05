package forms

import (
	"skis-admin-backend/model"

	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
)

type CreateOrderRequest struct {
	GoodId    string `json:"good_id" binding:"required"`
	UName     string `json:"u_name" binding:"required"`
	UPhone    string `json:"u_phone" binding:"required"`
	UsePoints int64  `json:"use_points"`
}

type CreateOrderResp struct {
	PrePayResp *jsapi.PrepayWithRequestPaymentResponse `json:"pre_pay_resp"`
	OrderId    string                                  `json:"order_id"`
	Status     int                                     `json:"status"`
	LeftPoints int64                                   `json:"left_points"`
	CostPoints int64                                   `json:"cost_points"`
	GainPoints int64                                   `json:"gain_points"`
}

type QueryOrdersListRequest struct {
	TeachStates []int64 `form:"teach_states"`
	Page        int     `form:"page"`
	PageSize    int     `form:"page_size"`
}

type QueryOrderCoursesListRequest struct {
	TeachStates []int64 `form:"teach_states"`
	Page        int     `form:"page"`
	PageSize    int     `form:"page_size"`
}

type QueryOrdersListResp struct {
	Total int64           `json:"total"`
	List  []*model.Orders `json:"list"`
}

type WhirlyOrderPayCallbackRequest struct {
	OrderId       string `json:"order_id"  binding:"required"`
	TransactionId string `json:"transaction_id" binding:"required"`
	SuccessTime   string `json:"success_time" binding:"required"`
}

type AdminOrderRefundRequest struct {
	RefundType            int64 `json:"refund_type" binding:"omitempty,oneof=0 1"` //0-全额退款，退积分，1-部分退款，不退积分
	RefundUserMoney       int64 `json:"refund_user_money"`                         //部分退款金额，必填
	RefundTeachCoachMoney int64 `json:"refund_teach_coach_money"`                  //部分退款，教练金额，必填
}

type QueryCoachOrderCoursesListResponse struct {
	Total int64                  `json:"total"`
	List  []*model.OrdersCourses `json:"list"`
}

type OrderRefundLimitResponse struct {
	RefundMaxMoney   int64 `json:"refund_max_money"`   //最大可退款金额，单位：分
	RefundTeachCoach bool  `json:"refund_teach_coach"` // 是否可退教练
}
