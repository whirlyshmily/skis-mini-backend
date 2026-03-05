package model

import "time"

// create table clubs_tags
// (
// id         bigint auto_increment comment 'id'
// primary key,
// club_id    varchar(64)                         not null comment '俱乐部',
// tag_id     bigint                              not null comment '标签id',
// state      tinyint   default 0                 null comment '状态，0-正常， 1-删除',
// created_at timestamp default CURRENT_TIMESTAMP null comment '创建时间',
// updated_at timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间',
// constraint club_id
// unique (club_id, tag_id),
// constraint coach_id
// unique (club_id, tag_id)
// )

type ClubsTags struct {
	ID        int64     `json:"id" gorm:"column:id"`           // id
	ClubID    string    `json:"club_id" gorm:"column:club_id"` // 俱乐部
	TagID     int64     `json:"tag_id" gorm:"column:tag_id"`   // 标签id
	State     int8      `json:"state" gorm:"column:state"`     // 状态，0-正常， 1-删除
	CreatedAt time.Time `json:"-" gorm:"column:created_at"`    // 创建时间
	UpdatedAt time.Time `json:"-" gorm:"column:updated_at"`    // 更新时间
	Tag       Tags      `gorm:"foreignKey:Id;references:TagID" json:"tag_info"`
}

func (m *ClubsTags) TableName() string {
	return "clubs_tags"
}
