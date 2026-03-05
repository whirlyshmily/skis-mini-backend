package model

import "time"

//CREATE TABLE `certificate_tags` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
//  `level` text COLLATE utf8mb4_unicode_ci NOT NULL,
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常， 1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type CertificateConfigs struct {
	Id        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Level     JSONArray `gorm:"column:level" json:"level"`
	State     int       `gorm:"column:state" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"-"`
}

func (c *CertificateConfigs) TableName() string {
	return "certificate_configs"
}
