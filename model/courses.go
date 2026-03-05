package model

import "time"

//CREATE TABLE `courses` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '课程名字',
//  `course_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '课程ID',
//  `coverUrl` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主图',
//  `detail` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '课程详情',
//  `price_min` int DEFAULT '0' COMMENT '最低积分',
//  `price_max` int DEFAULT '0' COMMENT '最高积分',
//  `ref_coach_cnt` int DEFAULT '0' COMMENT '推荐教练数量',
//  `ref_club_cnt` int DEFAULT '0' COMMENT '推荐俱乐部数量',
//  `finished_cnt` int DEFAULT '0' COMMENT '已完成数量',
//  `unfinished_cnt` int DEFAULT '0' COMMENT '未完成数量',
//  `canceled_cnt` int DEFAULT '0' COMMENT '已取消数量',
//  `points_deduct` tinyint DEFAULT '0' COMMENT '是否开启低分抵扣，0-不开启，1-开启',
//  `on_shelf` tinyint DEFAULT '1' COMMENT '是否上架，0-下架，1-上架',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常， 1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type Courses struct {
	Id            int64          `gorm:"primaryKey" json:"id"`
	Title         string         `gorm:"column:title" json:"title"`
	CourseID      string         `gorm:"column:course_id" json:"course_id"`
	CoverUrl      string         `gorm:"column:cover_url" json:"cover_url"`
	Detail        string         `gorm:"column:detail" json:"detail"`
	PointsDeduct  int            `gorm:"column:points_deduct" json:"points_deduct"`
	OnShelf       int            `gorm:"column:on_shelf" json:"on_shelf"`
	PriceMin      int64          `gorm:"column:price_min" json:"price_min"`
	PriceMax      int64          `gorm:"column:price_max" json:"price_max"`
	RefCoachCnt   int            `gorm:"column:ref_coach_cnt" json:"ref_coach_cnt"`
	RefClubCnt    int            `gorm:"column:ref_club_cnt" json:"ref_club_cnt"`
	FinishedCnt   int            `gorm:"column:finished_cnt" json:"finished_cnt"`
	UnFinishedCnt int            `gorm:"column:unfinished_cnt" json:"unfinished_cnt"`
	CanceledCnt   int            `gorm:"column:canceled_cnt" json:"canceled_cnt"`
	State         int            `gorm:"column:state" json:"-"`
	CreatedAt     time.Time      `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	CoursesTags   []*CoursesTags `gorm:"foreignKey:CourseID;references:CourseID" json:"-"`
	Tags          []*Tags        `gorm:"-" json:"tags,omitempty"`
}

func (Courses) TableName() string {
	return "courses"
}
