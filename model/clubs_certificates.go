package model

import "time"

//create table clubs_certificates
//(
//club_id              varchar(64)                           not null comment '俱乐部id',
//id                   bigint auto_increment comment 'id'
//primary key,
//certificate_id       bigint                                not null comment '证书ID',
//certificate_img_urls text                                  null comment '证书图片和视频链接',
//level                varchar(32) default ''                not null comment '等级',
//state                tinyint     default 0                 null comment '状态，0-正常， 1-删除',
//created_at           timestamp   default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at           timestamp   default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间'
//)
//comment '俱乐部证书审核记录表' row_format = DYNAMIC;

type ClubsCertificates struct {
	ID                int64              `json:"id" gorm:"column:id"`                         // id
	ClubID            string             `json:"club_id" gorm:"column:club_id"`               // 俱乐部id
	CertificateID     int64              `json:"certificate_id" gorm:"column:certificate_id"` // 证书ID
	Level             string             `json:"level" gorm:"column:level"`                   // 等级
	State             int8               `json:"state" gorm:"column:state"`                   // 状态，0-正常， 1-删除
	CreatedAt         time.Time          `json:"created_at" gorm:"column:created_at"`         // 创建时间
	UpdatedAt         time.Time          `json:"updated_at" gorm:"column:updated_at"`         // 更新时间
	CertificateConfig CertificateConfigs `gorm:"foreignKey:Id;references:CertificateID" json:"certificate_info"`
}

func (m *ClubsCertificates) TableName() string {
	return "clubs_certificates"
}
