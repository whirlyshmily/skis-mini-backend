package model

import (
	"time"
)

//create table coaches_tags_review
//(
//id           bigint auto_increment comment 'id'
//primary key,
//coach_id     varchar(64)                         not null comment '教练id',
//tag_id       bigint                              not null comment '标签id',
//tag_img_urls text                                null comment '证书图片和视频链接',
//verified     tinyint   default 0                 not null comment '是否认证，0-未认证，1-认证通过,2-驳回',
//remark       varchar(255) default ''                not null comment '审核备注',
//state        tinyint   default 0                 null comment '状态，0-正常， 1-删除',
//created_at   timestamp default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at   timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间'
//)
//comment '教练标签审核记录表';

type CoachesTagsReview struct {
	ID         int64     `json:"id" gorm:"column:id"`                     // id
	CoachID    string    `json:"coach_id" gorm:"column:coach_id"`         // 教练id
	TagID      int64     `json:"tag_id" gorm:"column:tag_id"`             // 标签id
	TagImgUrls JSONArray `json:"tag_img_urls" gorm:"column:tag_img_urls"` // 证书图片和视频链接
	Verified   int8      `json:"verified" gorm:"column:verified"`         // 是否认证，0-未认证，1-认证通过,2-驳回
	Remark     string    `json:"remark" gorm:"column:remark"`             // 审核备注
	State      int8      `json:"state" gorm:"column:state"`               // 状态，0-正常， 1-删除
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`     // 创建时间
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`     // 更新时间
}

func (m *CoachesTagsReview) TableName() string {
	return "coaches_tags"
}
