package model

//CREATE TABLE `feeds_comments` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
//  `pid` bigint DEFAULT '0' COMMENT '评论的父ID',
//  `rank_id` int DEFAULT '0',
//  `user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '评论user_id',
//  `user_type` int DEFAULT '1' COMMENT '用户类型，1-普通用户2-教练，3-俱乐部',
//  `feed_id` int NOT NULL COMMENT '动态圈ID',
//  `replied_user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '被回复的用户ID',
//  `replied_user_type` int DEFAULT NULL COMMENT '用户类型，1-普通用户2-教练，3-俱乐部',
//  `content` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '回复内容',
//  `urls` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '恢复的图片或者视频',
//  `on_shelf` tinyint DEFAULT '1' COMMENT '是否上下架，0-下架，1-上架',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常，1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`)
//) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子评论表'

type FeedsComments struct {
	BaseModel
	Pid             int64        `gorm:"column:pid;type:bigint(20);default:0" json:"pid"`
	RankId          int          `gorm:"column:rank_id;type:int(11);default:0" json:"rank_id"`
	UserId          string       `gorm:"column:user_id;type:varchar(64);not null" json:"user_id"`
	UserType        int          `gorm:"column:user_type;type:int(11);default:1" json:"user_type"`
	FeedId          int64        `gorm:"column:feed_id;type:int(11);not null" json:"feed_id"`
	RepliedUserId   string       `gorm:"column:replied_user_id;type:varchar(64);default:''" json:"replied_user_id"`
	RepliedUserType int          `gorm:"column:replied_user_type;type:int(11);default:0" json:"replied_user_type"`
	Content         string       `json:"content" gorm:"content"`
	Urls            JsonUrlArray `json:"urls" gorm:"urls"`
	OnShelf         int          `gorm:"column:on_shelf;type:tinyint(1);default:1" json:"on_shelf"`
	Right           bool         `gorm:"-" json:"right"`
}
