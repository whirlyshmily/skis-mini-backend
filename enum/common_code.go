package enum

const ServiceRatio = 15 // 服务费比例
const ReferralRatio = 1 //  推荐人返佣比例

// OrderStatusPending 支付状态，0-待支付，1-支付成功，2-支付失败，3-申请退款，4-已退款
const (
	OrderStatusPending          = 0 //待支付
	OrderStatusPaid             = 1 //支付成功
	OrderStatusFail             = 2 //支付失败
	OrderStatusRefundProcessing = 3 //退款处理中
	OrderStatusRefundSuccess    = 4 //退款成功
	OrderStatusRefundFailed     = 5 //退款失败
)

const (
	// UserTypeUser 用户类型（1：普通用户、2：教练、3：俱乐部、4：官方、5：游客）
	UserTypeUser     = 1      //1：普通用户
	UserTypeCoach    = 2      //2：教练
	UserTypeClub     = 3      //3：俱乐部
	UserTypeOfficial = 4      //4：官方
	UserTypeTourist  = 5      //5：游客
	UserTypeCron     = 6      //6：定时任务
	UserIdCron       = "cron" //定时任务用户id
)

const (
	RefundTypeAll  = 0 //全额退款退积分
	RefundTypePart = 1 //部分退款，不退积分
)
