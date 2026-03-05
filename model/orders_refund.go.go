package model

import "time"

//create table orders_refund
//(
//id           bigint auto_increment comment 'id'
//primary key,
//user_id               varchar(64)                         not null comment 'user_id,用户ID或者教练ID或者俱乐部ID',
//user_type             tinyint   default 1                 null comment '1-用户，2-教练，3-俱乐部',
//order_id     varchar(64)                         not null comment '订单ID',
//refund_id    varchar(64)                         not null comment '退款ID',
//refund_time  timestamp                           null comment '退款时间',
//refund_money int                                 null comment '退款金额',
//channel      varchar(32)                         null comment '渠道',
//remark       varchar(64)                         null comment '备注',
//status       varchar(255)                        null comment '对应微信退款状态',
//state        tinyint   default 0                 null comment '状态，0-正常， 1-删除',
//created_at   timestamp default CURRENT_TIMESTAMP not null comment '创建时间',
//updated_at   timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
//constraint uid
//unique (uid, order_id)
//)
//comment '订单退款记录表';

const (
	StateNormal  = 0 //正常
	StateDeleted = 1 //删除
)

type OrdersRefund struct {
	Id                  uint64     `gorm:"primarykey" json:"id"`
	UserId              string     `gorm:"column:user_id" json:"user_id"`
	UserType            int        `gorm:"column:user_type" json:"user_type"`
	OrderId             string     `gorm:"column:order_id" json:"-"`
	RefundId            string     `gorm:"column:refund_id" json:"refund_id"`
	RefundTime          *time.Time `gorm:"column:refund_time" json:"refund_time"`
	RefundMoney         int64      `gorm:"column:refund_money" json:"refund_money"`
	RefundPoints        int64      `gorm:"column:refund_points" json:"refund_points"`
	RefundType          int        `gorm:"column:refund_type" json:"refund_type"` //0-用户退款，1-管理台退款
	Channel             string     `gorm:"column:channel" json:"channel"`
	RefundTransactionId string     `gorm:"column:refund_transaction_id" json:"refund_transaction_id"`
	Remark              string     `gorm:"column:remark" json:"remark"`
	Status              string     `gorm:"column:status" json:"status"`
	State               int        `gorm:"column:state" json:"-"`
	CreatedAt           LocalTime  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (m *OrdersRefund) TableName() string {
	return "orders_refund"
}

const (
	RefundTypeUser  = 0
	RefundTypeAdmin = 1
)
