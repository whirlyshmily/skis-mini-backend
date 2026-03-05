package model

import (
	"time"
)

//create table coach_ski_resorts
//(
//id             int auto_increment comment '自增ID'
//primary key,
//coach_id       varchar(64)                         not null comment '教练ID',
//ski_resorts_id bigint                              not null comment '场地ID',
//state          int       default 0                 not null comment '状态，0-正常， 1-删除',
//created_at     timestamp default CURRENT_TIMESTAMP not null comment '创建时间',
//updated_at     timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
//)
//comment '教练场地关联表';

type CoachesSkiResorts struct {
	ID           int64      `json:"id" gorm:"column:id;primary_key;auto_increment"` // 自增ID
	CoachID      string     `json:"coach_id" gorm:"column:coach_id"`                // 教练ID
	SkiResortsID int64      `json:"ski_resorts_id" gorm:"column:ski_resorts_id"`    // 场地ID
	State        int        `json:"state" gorm:"column:state"`                      // 状态，0-正常， 1-删除
	CreatedAt    time.Time  `json:"-" gorm:"column:created_at"`                     // 创建时间
	UpdatedAt    time.Time  `json:"-" gorm:"column:updated_at"`                     // 更新时间
	SkiResorts   SkiResorts `gorm:"foreignKey:SkiResortsID;references:Id" json:"ski_resorts"`
}

func (m *CoachesSkiResorts) TableName() string {
	return "coaches_ski_resorts"
}
