package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	// UserTypeUser 用户类型(1:普通用户、2:教练、3:俱乐部、4:官方)
	UserTypeUser     = 1 //1:普通用户 2:教练 3:俱乐部 4:官方
	UserTypeCoach    = 2 //2:教练
	UserTypeClub     = 3 //3:俱乐部
	UserTypeOfficial = 4 //4:官方
)

//CREATE TABLE `users` (
//  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
//  `uid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户ID',
//  `nickname` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '昵称',
//  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
//  `avatar` varchar(256) DEFAULT '' COMMENT '头像',
//  `gender` varchar(255) DEFAULT NULL COMMENT '性别',
//  `accumulated_points` int DEFAULT '0' COMMENT '积分',
//  `left_points` int DEFAULT '0' COMMENT '积分',
//  `level` int DEFAULT '1' COMMENT '级别',
//  `birthday` date DEFAULT NULL COMMENT '生日',
//  `state` tinyint NOT NULL DEFAULT '0' COMMENT '状态，0-正常，1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`),
//  UNIQUE KEY `users_uid_index` (`uid`) USING BTREE
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='app用户表'

type Users struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Uid               string    `gorm:"column:uid" json:"uid"`
	UnionId           string    `gorm:"column:union_id" json:"union_id,omitempty"`
	OpenId            string    `gorm:"column:open_id" json:"open_id,omitempty"`
	Nickname          string    `gorm:"column:nickname" json:"nickname"`
	Phone             string    `gorm:"column:phone" json:"phone,omitempty"`
	Avatar            string    `gorm:"column:avatar" json:"avatar"`
	Gender            int       `gorm:"column:gender" json:"gender,omitempty"`
	Birthday          string    `gorm:"column:birthday" json:"birthday,omitempty"`
	Country           string    `gorm:"column:country" json:"country,omitempty"`
	Province          string    `gorm:"column:province" json:"province,omitempty"`
	City              string    `gorm:"column:city" json:"city,omitempty"`
	AccumulatedPoints int64     `gorm:"column:accumulated_points" json:"accumulated_points,omitempty"` // 积分
	LeftPoints        int64     `gorm:"column:left_points" json:"left_points,omitempty"`               // 积分
	Level             int       `gorm:"column:level" json:"level,omitempty"`
	State             int       `gorm:"column:state" json:"-"`                                                                      //0-正常，1-删除
	CreatedAt         time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`                // 创建时间
	UpdatedAt         time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"` // 更新时间
}

func (Users) TableName() string {
	return "users"
}

func ScopeUserSensitiveFields(db *gorm.DB) *gorm.DB {
	return db.Select("id", "uid", "nickname", "avatar", "gender", "level")
}
