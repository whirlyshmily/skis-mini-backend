package model

import (
	"time"
)

//create table money_operate
//(
//id           bigint auto_increment comment '自增ID'
//primary key,
//operate_id   varchar(64)                         null comment '操作ID',
//user_id      varchar(64)                         null comment '用户ID（包括用户ID、教练ID、俱乐部ID、官方ID）',
//user_type    int                                 null comment '用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）',
//money        int                                 null comment '资金变化（单位：分）',
//operate_type int       default 0                 null comment '资金操作类型（0：提取，1：充值）',
//type         int                                 null comment '资金类型（1：保证金，2：余额）',
//transaction_id varchar(64)                         null comment '微信支付订单号  ',
//package_info   varchar(1024)                       null comment '跳转领取页面的package信息',
//pay_time       timestamp                           null comment '支付时间',
//pay_log        varchar(2048)                       null comment '日志',
//status       int                                 null comment '状态（0：提现中，1：提现成功，2：提现失败、10：支付中，11：支付成功:12：支付失败）',
//state        int       default 0                 null comment '状态（0：正常，1：删除）',
//created_at   timestamp default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at   timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间',
//constraint money_operate_pk_2
//unique (operate_id)
//)
//comment '资金操作（保证金充值提现、余额提现）';

type MoneyOperate struct {
	ID            int64     `json:"id" gorm:"column:id"`                          // 自增ID
	OperateID     string    `json:"operate_id" gorm:"column:operate_id"`          // 操作ID
	UserID        string    `json:"user_id" gorm:"column:user_id"`                // 用户ID（包括用户ID、教练ID、俱乐部ID、官方ID）
	UserType      int       `json:"user_type" gorm:"column:user_type"`            // 用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）
	Money         int64     `json:"money" gorm:"column:money"`                    // 资金变化（单位：分）
	OperateType   int       `json:"operate_type" gorm:"column:operate_type"`      // 资金操作类型（0：提取，1：充值）
	Type          int       `json:"type" gorm:"column:type"`                      // 资金类型（1：保证金，2：余额）
	TransactionID string    `json:"transaction_id" gorm:"column:transaction_id"`  // 微信支付订单号
	PackageInfo   string    `json:"package_info" gorm:"column:package_info"`      // 跳转领取页面的package信息
	PayTime       time.Time `json:"pay_time" gorm:"column:pay_time;default:null"` // 支付时间
	PayLog        string    `json:"pay_log" gorm:"column:pay_log"`                // 日志
	Status        int       `json:"status" gorm:"column:status"`                  // 状态（0：提现中，1：提现成功，2：提现失败、10：支付中，11：支付成功:12：支付失败）
	State         int       `json:"state" gorm:"column:state"`                    // 状态（0：正常，1：删除）
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`          // 创建时间
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`          // 更新时间
}

func (m *MoneyOperate) TableName() string {
	return "money_operate"
}

const (
	OperateTypeWithdraw = 0 // 提现
	OperateTypeRecharge = 1 // 充值
)
const (
	TypeDeposit = 1 // 保证金
	TypeBalance = 2 // 余额
)
const (
	StatusWithdrawing = 0  // 提现中
	StatusWithdrawed  = 1  // 提现成功
	StatusWithdrawErr = 2  // 提现失败
	StatusRecharging  = 10 // 充值中
	StatusRecharged   = 11 // 充值成功
	StatusRechargeErr = 12 // 充值失败
)
