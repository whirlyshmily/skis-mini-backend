package model

import "time"

// create table orders
// (
// id           bigint auto_increment comment 'id'
// primary key,
// uid          varchar(64)                         not null comment '下单用户ID',
//u_name            varchar(32)                         null comment '用户称呼',
//u_phone           varchar(20)                         null comment '用户手机',
// good_id      varchar(64)                         not null comment '商品id',
// order_id     varchar(64)                         not null comment '内部订单号',
// out_order_id varchar(64)                         null comment '外部订单号',
// user_id      varchar(64)                         null comment '售卖课程的用户ID（包括教练和俱乐部）',
// user_type    int       default 2                 null comment '用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）',
// coach_id     varchar(64)                         null comment '上课的教练ID',
// total_fee    bigint                              not null comment '总金额，单位：分',
// club_money        int        default 0                 null comment '俱乐部的推荐费（俱乐部的课程才有）',
// paid_fee     bigint    default 0                 null comment '实际支付金额，单位：分',
// transfer_fee      bigint    default 0                 null comment '转单价格（分）',
// transfer_coach_id varchar(64)                         null comment '接受转单的教练ID',
// used_credits bigint    default 0                 null comment '使用积分',
// credits_fee  bigint    default 0                 null comment '积分抵扣金额',
// credit_id    varchar(64)                         null comment '积分流水的id',
// teach_time   int                                 null comment '教学时长',
// pack        int          default 0                 not null comment '是否为打包课程（0：否，1：打包）',
// discount          int                                  null comment '折扣（打包课才设置）',
// status       int       default 0                 null comment '支付状态，0-待支付，1-支付成功，2-支付失败，3-申请退款，4-已退款',
// pay_time     timestamp                           null comment '支付时间',
// state        tinyint   default 0                 null comment '状态，0-正常，1-删除',
// created_at   timestamp default CURRENT_TIMESTAMP null comment '创建时间',
// updated_at   timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间',
// constraint order_id
// unique (order_id)
// )
// comment '订单表' row_format = DYNAMIC;

type Orders struct {
	Id              int64           `gorm:"column:id" json:"id"`                               // id
	Uid             string          `gorm:"column:uid" json:"uid"`                             // 下单用户ID
	UName           string          `json:"u_name" gorm:"column:u_name"`                       // 用户称呼
	UPhone          string          `json:"u_phone" gorm:"column:u_phone"`                     // 用户手机
	GoodID          string          `gorm:"column:good_id" json:"good_id"`                     // 商品id
	OrderID         string          `gorm:"column:order_id" json:"order_id"`                   // 内部订单号
	TransactionId   string          `gorm:"column:transaction_id" json:"transaction_id"`       // 微信支付订单号
	UserID          string          `gorm:"column:user_id" json:"user_id"`                     // 售卖课程的用户ID（包括教练和俱乐部）
	UserType        int             `gorm:"column:user_type" json:"user_type"`                 // 用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）
	TotalFee        int64           `gorm:"column:total_fee" json:"total_fee"`                 // 总金额，单位：分
	PaidFee         int64           `gorm:"column:paid_fee" json:"paid_fee"`                   // 实际支付金额，单位：分
	TransferFee     int64           `json:"transfer_fee" gorm:"column:transfer_fee"`           // 转单价格（分）
	TransferCoachID string          `json:"transfer_coach_id" gorm:"column:transfer_coach_id"` // 接受转单的教练ID
	UsedPoints      int64           `gorm:"column:used_points" json:"used_points"`             // 使用门店积分
	PointsFee       int64           `gorm:"column:points_fee" json:"points_fee"`               // 积分抵扣金额
	TeachTime       int             `gorm:"column:teach_time" json:"teach_time"`               // 教学时长
	PayTime         *time.Time      `gorm:"column:pay_time" json:"pay_time"`                   // 支付时间
	Pack            int             `gorm:"column:pack" json:"pack"`                           // 是否为打包课程（0：否，1：打包）
	Discount        int             `gorm:"column:discount" json:"discount"`                   // 折扣（打包课才设置）
	TeachState      int             `gorm:"column:teach_state" json:"teach_state"`             //教学状态，0-待预约，1-进行中，2-已结束，3-退课
	FrozenMoney     int64           `gorm:"column:frozen_money" json:"frozen_money"`           // 冻结资金
	Progress        string          `gorm:"column:progress" json:"progress"`                   // 订单进度，已完成/总数
	Status          int             `gorm:"column:status" json:"status"`                       // 支付状态，0-待支付，1-支付成功，2-支付失败，3-退款处理中，4-退款成功, 5-退款失败
	State           int             `gorm:"column:state" json:"-"`                             //状态，0-正常，1-删除
	CreatedAt       time.Time       `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	Goods           *Goods          `gorm:"foreignKey:GoodID;references:GoodID" json:"good,omitempty"`
	GoodsCourses    []GoodsCourses  `gorm:"foreignKey:GoodID;references:GoodID" json:"goods_courses"`
	OrdersCourses   []OrdersCourses `gorm:"foreignKey:OrderID;references:OrderID" json:"orders_courses"` // 订单课程表
	Club            *Clubs          `gorm:"foreignKey:ClubId;references:UserID" json:"club,omitempty"`
	Coach           *Coaches        `gorm:"foreignKey:CoachId;references:UserID" json:"coach,omitempty"`
	PointRecord     *PointsRecords  `gorm:"foreignKey:RelationID;references:OrderID" json:"-"`
	PointId         string          `json:"point_id" gorm:"-"`
	Refund          *OrdersRefund   `gorm:"foreignKey:OrderId;references:OrderID" json:"refund,omitempty"`
}

func (m *Orders) TableName() string {
	return "orders"
}

const OrderStatusRefundIng = 3
const OrderStatusRefund = 4
