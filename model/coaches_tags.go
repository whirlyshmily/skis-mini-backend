package model

import "time"

//create table coaches_tags
//(
//id           bigint auto_increment comment 'id'
//primary key,
//coach_id     varchar(64)                            not null comment '教练id',
//tag_id       bigint                                 not null comment '标签id',
//tag_img_urls text                                   null comment '证书图片和视频链接',
//verified     tinyint      default 0                 not null comment '是否认证，0-未认证，1-认证通过,2-驳回',
//remark       varchar(255) default ''                not null comment '审核备注',
//state        tinyint      default 0                 null comment '状态，0-正常， 1-删除',
//created_at   timestamp    default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at   timestamp    default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间'
//)
//comment '教练标签审核记录表' row_format = DYNAMIC;

type CoachesTags struct {
	Id        int64     `gorm:"primaryKey" json:"id"`
	CoachId   string    `gorm:"column:coach_id" json:"coach_id"`
	TagId     int64     `gorm:"column:tag_id" json:"tag_id"`
	Verified  int8      `gorm:"column:verified" json:"verified"`
	Remark    string    `gorm:"column:remark" json:"remark"`
	State     int       `gorm:"column:state" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"-"`
	Tag       Tags      `gorm:"foreignKey:Id;references:TagId" json:"tag_info"`
}

func (CoachesTags) TableName() string {
	return "coaches_tags"
}
