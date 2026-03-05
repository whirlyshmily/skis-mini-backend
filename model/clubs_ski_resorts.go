package model

import (
	"time"
)

//create table clubs_ski_resorts
//(
//id             int auto_increment comment '自增ID'
//primary key,
//club_id        varchar(64)                         null comment '俱乐部ID',
//ski_resorts_id bigint                              not null comment '场地ID',
//state          int       default 0                 not null comment '状态，0-正常， 1-删除',
//created_at     timestamp default CURRENT_TIMESTAMP not null comment '创建时间',
//updated_at     timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
//constraint club_id_pk
//unique (club_id),
//constraint ski_resorts_id_pk
//unique (ski_resorts_id)
//)
//comment '俱乐部场地关联表';

type ClubsSkiResorts struct {
	ID           int        `json:"id" gorm:"column:id"`                         // 自增ID
	ClubID       string     `json:"club_id" gorm:"column:club_id"`               // 俱乐部ID
	SkiResortsID int64      `json:"ski_resorts_id" gorm:"column:ski_resorts_id"` // 场地ID
	State        int        `json:"state" gorm:"column:state"`                   // 状态，0-正常， 1-删除
	CreatedAt    time.Time  `json:"-" gorm:"column:created_at"`                  // 创建时间
	UpdatedAt    time.Time  `json:"-" gorm:"column:updated_at"`                  // 更新时间
	SkiResorts   SkiResorts `gorm:"foreignKey:SkiResortsID;references:Id" json:"ski_resorts"`
}

func (m *ClubsSkiResorts) TableName() string {
	return "clubs_ski_resorts"
}
