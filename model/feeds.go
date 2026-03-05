package model

import "time"

//CREATE TABLE `feeds` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标题',
//  `club_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '俱乐部id',
//  `coach_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '教练id',
//  `coverUrl` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '封面',
//  `View` bigint DEFAULT '0' COMMENT '浏览量',
//  `priority` int DEFAULT '100' COMMENT '排序权重，默认100',
//  `curated` int DEFAULT '0' COMMENT '是否精选，0-不精选，1-精选',
//  `on_shelf` int DEFAULT '1' COMMENT '是否上架，0-下架，1-上架',
//  `state` tinyint DEFAULT '0' COMMENT '0-正常， 1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type Feeds struct {
	Id        int64        `gorm:"primaryKey" json:"id"`
	Title     string       `gorm:"column:title" json:"title"`
	UserId    string       `gorm:"column:user_id" json:"user_id"`
	UserType  int          `gorm:"column:user_type" json:"user_type"`
	Urls      JsonUrlArray `gorm:"column:urls" json:"urls"`
	Detail    string       `gorm:"column:detail" json:"detail"`
	View      int64        `gorm:"column:view" json:"view"`
	Priority  int          `gorm:"column:priority;" json:"priority"`
	Curated   int          `gorm:"column:curated" json:"curated"`
	OnShelf   int          `gorm:"column:on_shelf;default:1" json:"on_shelf"`
	State     int          `gorm:"column:state" json:"-"`
	CreatedAt time.Time    `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time    `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	Club      *Clubs       `gorm:"foreignKey:ClubId;references:UserId" json:"club_info,omitempty"`
	Coach     *Coaches     `gorm:"foreignKey:CoachId;references:UserId" json:"coach_info,omitempty"`
}

func (Feeds) TableName() string {
	return "feeds"
}
