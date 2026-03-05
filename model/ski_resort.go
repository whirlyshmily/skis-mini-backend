package model

import "time"

//CREATE TABLE `ski_resorts` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '雪场名字',
//  `province` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '省份',
//  `city` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '城市',
//  `status` tinyint(1) unsigned zerofill DEFAULT '0' COMMENT '状态，0-开启，1-关闭',
//  `state` tinyint(1) unsigned zerofill DEFAULT '0' COMMENT '状态，0-正常，1-删除',
//  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='雪场'

type SkiResorts struct {
	Id           int64     `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"column:name" json:"name"`
	Province     string    `gorm:"column:province" json:"province"`
	City         string    `gorm:"column:city" json:"city"`
	District     string    `gorm:"column:district" json:"district"`
	Detail       string    `gorm:"column:detail" json:"detail"`
	LocationCode string    `gorm:"column:location_code" json:"location_code"`
	Longitude    float64   `gorm:"column:longitude" json:"longitude"`
	Latitude     float64   `gorm:"column:latitude" json:"latitude"`
	Description  string    `gorm:"column:description" json:"description"`
	Status       uint8     `gorm:"column:status" json:"status"`
	State        uint8     `gorm:"column:state" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"-"`
}

func (SkiResorts) TableName() string {
	return "ski_resorts"
}

type SkiResortList struct {
	SkiResorts
	CoachRefCount      int64 `json:"coach_ref_count"`
	ReserveCourseCount int64 `json:"reserve_course_count"`
}
