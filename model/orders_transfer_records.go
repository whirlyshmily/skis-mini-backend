package model

type OrdersTransferRecords struct {
	BaseModel
	OrderId          string `gorm:"order_id" json:"order_id"`                     // 订单ID
	PreviousUserId   string `gorm:"previous_user_id" json:"previous_user_id"`     // 上一个的user_id
	PreviousUserType int    `gorm:"previous_user_type" json:"previous_user_type"` // 同订单表的user_type
	CurUserId        string `gorm:"cur_user_id" json:"cur_user_id"`               // 当前的user_id
	CurUserType      int    `gorm:"cur_user_type" json:"cur_user_type"`           // 同订单表的user_type
}
