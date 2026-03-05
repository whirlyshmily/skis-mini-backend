package model

import (
	"time"
)

//create table clubs_coaches
//(
//id         bigint auto_increment comment 'id'
//primary key,
//club_id    varchar(64)                         not null comment '俱乐部ID',
//coach_id   varchar(64)                         not null comment '教练ID',
//verified   tinyint   default 0                 null comment '是否认证，0-未认证，1-认证通过,2-驳回',
//state      int       default 0                 null comment '状态，0-正常， 1-删除',
//created_at timestamp default CURRENT_TIMESTAMP null,
//updated_at timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP
//);

type ClubsCoaches struct {
	ID        int64     `json:"id" gorm:"column:id"`             // id
	ClubID    string    `json:"club_id" gorm:"column:club_id"`   // 俱乐部ID
	CoachID   string    `json:"coach_id" gorm:"column:coach_id"` // 教练ID
	Verified  int       `json:"verified" gorm:"column:verified"` // 是否认证，0-未认证，1-认证通过,2-驳回
	State     int       `json:"state" gorm:"column:state"`       // 状态，0-正常， 1-删除
	CreatedAt time.Time `json:"-" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"-" gorm:"column:updated_at"`
	Coaches   Coaches   `gorm:"foreignKey:CoachId;references:CoachID" json:"coach_info"`
	Clubs     Clubs     `gorm:"foreignKey:ClubId;references:ClubID" json:"club_info"`
}

func (m *ClubsCoaches) TableName() string {
	return "clubs_coaches"
}

const VerifiedNo = 0
const VerifiedPass = 1
const VerifiedReject = 2
