package model

import "time"

//CREATE TABLE `courses_tags` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `course_id` bigint NOT NULL COMMENT '课程id',
//  `tag_id` bigint NOT NULL COMMENT '标签id',
//  `tag_name` varchar(255) NOT NULL COMMENT '标签名称',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常， 1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`),
//  UNIQUE KEY `course_id` (`course_id`,`tag_id`) USING BTREE,
//  KEY `tag_id` (`tag_id`)
//) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type CoursesTags struct {
	Id        int       `gorm:"primaryKey" json:"id"`
	CourseID  string    `gorm:"column:course_id" json:"course_id"`
	TagID     int64     `gorm:"column:tag_id" json:"tag_id"`
	State     int       `gorm:"column:state" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"-"`
	Tag       *Tags     `gorm:"foreignKey:Id;references:TagID" json:"tag"`
}

func (c *CoursesTags) TableName() string {
	return "courses_tags"
}
