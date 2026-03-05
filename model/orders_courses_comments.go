package model

import "gorm.io/gorm"

//CREATE TABLE `orders_courses_comments` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
//  `pid` bigint DEFAULT '0' COMMENT '评论的父ID',
//  `rank_id` int DEFAULT '0',
//  `user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '评论user_id',
//  `user_type` int DEFAULT '1' COMMENT '用户类型，1-普通用户2-教练，3-俱乐部',
//  `order_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单Id',
//  `order_course_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单课程id',
//  `course_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '课程id',
//  `replied_user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '被回复的用户ID',
//  `replied_user_type` int DEFAULT NULL COMMENT '用户类型，1-普通用户2-教练，3-俱乐部',
//  `content` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '回复内容',
//  `urls` text COLLATE utf8mb4_unicode_ci COMMENT '恢复的图片或者视频',
//  `on_shelf` tinyint DEFAULT '1' COMMENT '是否上下架，0-下架，1-上架',
//  `status` int DEFAULT '0' COMMENT '状态，此字段暂时用于黑名单的拉黑/恢复状态处理',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常，1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子评论表'

type OrdersCoursesComments struct {
	BaseModel
	Pid             int64                  `json:"-" gorm:"pid"`
	RankId          int                    `json:"-" gorm:"rank_id"`
	OrderId         string                 `json:"-" gorm:"order_id"`
	OrderCourseId   string                 `json:"-" gorm:"order_course_id"`
	GoodId          string                 `json:"-" gorm:"good_id"`
	CourseId        string                 `json:"-" gorm:"course_id"`
	UserId          string                 `json:"-" gorm:"user_id"`
	UserType        int                    `json:"-" gorm:"user_type"`
	RepliedUserId   string                 `json:"-" gorm:"replied_user_id"`
	RepliedUserType int                    `json:"-" gorm:"replied_user_type"`
	Content         string                 `json:"content" gorm:"content"`
	Urls            JsonUrlArray           `json:"urls" gorm:"urls"`
	OnShelf         int                    `json:"-" gorm:"on_shelf"`
	Reply           *OrdersCoursesComments `json:"reply,omitempty" gorm:"foreignKey:Id;references:Pid"`
	UserInfo        *Users                 `json:"user_info,omitempty" gorm:"foreignKey:Uid;references:UserId"`
	Good            *Goods                 `json:"good,omitempty" gorm:"foreignKey:GoodID;references:GoodId"`
	CoachInfo       *Coaches               `json:"coach_info,omitempty" gorm:"foreignKey:CoachId;references:UserId"`
	Comment         *OrdersCoursesComments `json:"comment,omitempty" gorm:"foreignKey:Pid;references:Id"`
}

func (OrdersCoursesComments) TableName() string {
	return "orders_courses_comments"
}

func ScopeOrderCourseCommentFields(db *gorm.DB) *gorm.DB {
	return db
	//return db.Select("id", "pid", "order_course_id", "user_id", "user_type", "content", "urls", "created_at", "updated_at")
}
