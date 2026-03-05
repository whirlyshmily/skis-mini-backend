package forms

type QueryMoneyListRequest struct {
	IncomeType int `form:"income_type"` //收入类型（0：收入，1：支出， -1：全部）
	Page       int `form:"page"`
	PageSize   int `form:"page_size"`
}

type QueryMoneyInfoRequest struct {
	MoneyId string `form:"money_id"`
}

type MoneyOperateWithdrawRequest struct {
	Money int64 `json:"money" binding:"required"`          //提现金额
	Type  int   `json:"type" binding:"required,oneof=1 2"` //资金类型（1：保证金，2：余额）
}

// MoneyOperateRechargeRequest 定义了充值请求的数据结构
// money 表示充值金额，单位为分，必须大于0
type MoneyOperateRechargeRequest struct {
	Money int64 `json:"money" binding:"required,min=1"` //充值金额，单位为分，必须大于0
}
