package model

import "time"

//CREATE TABLE `coach_level` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `level` int DEFAULT NULL COMMENT '等级',
//  `priority` int DEFAULT '0' COMMENT '权重分',
//  `service_rate` int DEFAULT '15' COMMENT '平台服务费率，百分比',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常， 1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type CoachesLevel struct {
	Id          int64     `gorm:"primaryKey" json:"id"`
	Level       int       `gorm:"column:level" json:"level"`
	Priority    int       `gorm:"column:priority" json:"priority"`
	ServiceRate int       `gorm:"column:service_rate" json:"service_rate"`
	State       int       `gorm:"column:state" json:"-"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}

func (CoachesLevel) TableName() string {
	return "coaches_level"
}
