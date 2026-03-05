package model

import "time"

// create table referral_records
// (
// id               bigint auto_increment comment 'id'
// primary key,
// user_id          varchar(64)                         not null comment '用户id',
// user_type        int                                 null comment '用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）',
// referral_user_id varchar(64)                         not null comment '推荐人的user_id（包括教练和俱乐部）',
// referral_type    tinyint   default 0                 null comment '推荐类型，0-普通用户，1-教练，2-俱乐部',
// referral_code    varchar(64)                         not null comment '推荐码',
// state            tinyint   default 0                 null comment '状态，0-正常， 1-删除',
// created_at       timestamp default CURRENT_TIMESTAMP not null comment '创建时间',
// updated_at       timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
// );
const (
	ReferralTypeUser  = 0 //普通用户
	ReferralTypeCoach = 1 //教练
	ReferralTypeClub  = 2 //俱乐部
)

type ReferralRecords struct {
	ID             int64     `json:"id" gorm:"column:id"`                                         // id
	UserID         string    `json:"user_id" gorm:"column:user_id"`                               // 用户id
	UserType       int       `json:"user_type" gorm:"column:user_type"`                           // 用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）
	ReferralUserID string    `json:"referral_user_id" gorm:"column:referral_user_id"`             // 推荐人的user_id（包括教练和俱乐部）
	ReferralType   int       `json:"referral_type" gorm:"column:referral_type"`                   // 推荐类型，0-普通用户，1-教练，2-俱乐部、3：俱乐部、4：官方
	ReferralCode   string    `json:"referral_code" gorm:"column:referral_code"`                   // 推荐码
	Profit         int64     `json:"profit" gorm:"column:profit"`                                 // 利润
	State          int8      `json:"-" gorm:"column:state"`                                       // 状态，0-正常， 1-删除
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`                         // 创建时间
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`                         // 更新时间
	User           *Users    `json:"user,omitempty" gorm:"foreignKey:UserID;references:Uid"`      // 用户信息
	Coach          *Coaches  `json:"coach,omitempty" gorm:"foreignKey:UserID;references:CoachId"` // 教练信息
}

func (m *ReferralRecords) TableName() string {
	return "referral_records"
}
