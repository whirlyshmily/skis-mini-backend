package model

import (
	"time"
)

// create table coaches_certificate_review
// (
// id                   bigint auto_increment comment 'id'
// primary key,
// coach_id             varchar(64)                           not null comment '教练id',
// certificate_id       bigint                                not null comment '证书ID',
// certificate_img_urls text                                  null comment '证书图片和视频链接',
// level                varchar(32) default ”                not null comment '等级',
// verified             tinyint     default 0                 null comment '是否认证，0-未认证，1-认证通过,2-驳回',
// remark               varchar(255)                          null comment '审核备注',
// state                tinyint     default 0                 null comment '状态，0-正常， 1-删除',
// created_at           timestamp   default CURRENT_TIMESTAMP null comment '创建时间',
// updated_at           timestamp   default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间'
// )
// comment '教练证书审核记录表';

type CoachesCertificates struct {
	ID                 int64              `json:"id" gorm:"column:id"`                                     // id
	CoachID            string             `json:"coach_id" gorm:"column:coach_id"`                         // 教练id
	CertificateID      int64              `json:"certificate_id" gorm:"column:certificate_id"`             // 证书ID
	CertificateImgUrls JSONArray          `json:"certificate_img_urls" gorm:"column:certificate_img_urls"` // 证书图片和视频链接
	Level              string             `json:"level" gorm:"column:level"`                               // 等级
	Verified           int8               `json:"verified" gorm:"column:verified"`                         // 是否认证，0-未认证，1-认证通过,2-驳回
	Remark             string             `json:"remark" gorm:"column:remark"`                             // 审核备注
	State              int8               `json:"-" gorm:"column:state"`                                   // 状态，0-正常， 1-删除
	CreatedAt          time.Time          `json:"-" gorm:"column:created_at"`                              // 创建时间
	UpdatedAt          time.Time          `json:"-" gorm:"column:updated_at"`                              // 更新时间
	CertificateConfig  CertificateConfigs `gorm:"foreignKey:Id;references:CertificateID" json:"certificate_info"`
}

//Certificates           []CoachesCertificates `gorm:"foreignKey:CoachID;references:CoachId" json:"certificates"`

func (m *CoachesCertificates) TableName() string {
	return "coaches_certificates"
}
