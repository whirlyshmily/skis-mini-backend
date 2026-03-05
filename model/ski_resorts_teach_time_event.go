package model

import "time"

//create table ski_resorts_teach_time_event
//(
//id         bigint auto_increment comment '自增ID'
//primary key,
//skirt_id  bigint                               null comment '雪场教学时间的ID',
//title      varchar(64)                         null comment '标题',
//remark     varchar(255)                        not null comment '备注',
//state      int       default 0                 null comment '状态，0-正常， 1-删除',
//created_at timestamp default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时候'
//)
//comment '预约雪场的教学时间日程事件';

type SkiResortsTeachTimeEvent struct {
	ID        int64     `json:"id" gorm:"column:id"`             // 自增ID
	SkirtID   int64     `json:"skirt_id" gorm:"column:skirt_id"` // 雪场教学时间的ID
	Title     string    `json:"title" gorm:"column:title"`       // 标题
	Remark    string    `json:"remark" gorm:"column:remark"`     // 备注
	State     int       `json:"-" gorm:"column:state"`           // 状态，0-正常， 1-删除
	CreatedAt time.Time `json:"-" gorm:"column:created_at"`      // 创建时间
	UpdatedAt time.Time `json:"-" gorm:"column:updated_at"`      // 更新时候
}

func (m *SkiResortsTeachTimeEvent) TableName() string {
	return "ski_resorts_teach_time_event"
}
