package model

import "time"

const (
	// 加积分
	ActionTypeBuyCourse    = 0 // 核销课程增加积分
	ActionTypeSystemAdd    = 1 // 系统增加积分
	ActionTypeOrderTimeout = 2 // 订单超时未支付，返还积分
	ActionTypeOrderRefund  = 3 // 订单退款，返还积分

	//减积分
	ActionTypeBuyCourseDeduct = 1000 // 购买课程抵扣积分
	ActionTypeSystemDeduct    = 1001 // 系统减少积分
)

//create table points_records
//(
//id             bigint auto_increment comment 'id'
//primary key,
//point_id       varchar(64)                         not null comment '积分ID',
//uid            varchar(64)                         null comment '用户uid',
//action_type    int       default 0                 not null comment '操作类型，0-完成课程增加积分，1-系统增加积分，1000-购买课程抵扣积分， 1001-系统减少积分',
//relation_id    varchar(64)                         null comment '关联id',
//points         int                                 not null comment '积分变更',
//current_points bigint    default 0                 null comment '当前积分',
//state          tinyint   default 0                 null comment '状态，0-正常， 1-删除',
//created_at     timestamp default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at     timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间',
//constraint cid
//unique (point_id)
//)
//comment '积分记录表' row_format = DYNAMIC;

type PointsRecords struct {
	ID            int64     `json:"id" gorm:"column:id"`                         // id
	PointID       string    `json:"point_id" gorm:"column:point_id"`             // 积分ID
	Uid           string    `json:"uid" gorm:"column:uid"`                       // 用户uid
	ActionType    int64     `json:"action_type" gorm:"column:action_type"`       // 操作类型，0-完成课程增加积分，1-系统增加积分，1000-购买课程抵扣积分， 1001-系统减少积分
	RelationID    string    `json:"relation_id" gorm:"column:relation_id"`       // 关联id
	Points        int64     `json:"points" gorm:"column:points"`                 // 积分变更
	CurrentPoints int64     `json:"current_points" gorm:"column:current_points"` // 当前积分
	Remark        string    `json:"remark" gorm:"column:remark"`                 //备注
	State         int8      `json:"state" gorm:"column:state"`                   // 状态，0-正常， 1-删除
	CreatedAt     LocalTime `json:"created_at" gorm:"column:created_at"`         // 创建时间
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`         // 更新时间
}

func (m *PointsRecords) TableName() string {
	return "points_records"
}
