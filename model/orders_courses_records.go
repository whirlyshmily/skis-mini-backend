package model

//CREATE TABLE `orders_courses_records` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `uid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户id',
//  `order_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单Id',
//  `order_course_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单课程id',
//  `good_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '商品id',
//  `course_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '课程id',
//  `content` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '评论内容',
//  `urls` text COLLATE utf8mb4_unicode_ci COMMENT '图片或者视频url',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常， 1-删除',
//  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`),
//  KEY `uid` (`uid`,`order_id`,`good_id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type OrdersCoursesRecords struct {
	BaseModel
	Uid           string       `json:"uid" gorm:"uid"`
	OrderId       string       `json:"-" gorm:"order_id"`
	OrderCourseId string       `json:"-" gorm:"order_course_id"`
	GoodId        string       `json:"-" gorm:"good_id"`
	CourseId      string       `json:"-" gorm:"course_id"`
	CoachId       string       `json:"-" gorm:"coach_id"`
	Content       string       `json:"content" gorm:"content"`
	Urls          JsonUrlArray `json:"urls" gorm:"urls"`
	UserInfo      *Users       `json:"user_info,omitempty" gorm:"foreignKey:Uid;references:Uid"`
	Good          *Goods       `json:"good,omitempty" gorm:"foreignKey:GoodID;references:GoodId"`
}

func (OrdersCoursesRecords) TableName() string {
	return "orders_courses_records"
}
