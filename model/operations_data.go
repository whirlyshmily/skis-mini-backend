package model

//CREATE TABLE `users_active_day` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `date` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '日期，年-月-日',
//  `cnt` int DEFAULT '0' COMMENT '活跃数量',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常， 1-删除',
//  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`),
//  KEY `date` (`date`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type UsersActiveDay struct {
	BaseModel
	Date string `gorm:"column:date" json:"date"`
	Cnt  int    `gorm:"column:cnt" json:"cnt"`
}

func (UsersActiveDay) TableName() string {
	return "users_active_day"
}

type UsersActiveMonth struct {
	BaseModel
	Date string `gorm:"column:date" json:"date"`
	Cnt  int    `gorm:"column:cnt" json:"cnt"`
}

func (UsersActiveMonth) TableName() string {
	return "users_active_month"
}

//CREATE TABLE `orders_day_statistics` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `buyer_count` int NOT NULL DEFAULT '0' COMMENT '购买用户数量',
//  `total_amount` int NOT NULL DEFAULT '0' COMMENT '总金额',
//  `date` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '日期,年-月-日',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常， 1-删除',
//  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`),
//  UNIQUE KEY `date` (`date`) USING BTREE
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type OrdersDayStatistics struct {
	BaseModel
	BuyerCount  int64  `gorm:"column:buyer_count" json:"buyer_count"`
	TotalAmount int64  `gorm:"column:total_amount" json:"total_amount"`
	Date        string `gorm:"column:date" json:"date"`
}

func (OrdersDayStatistics) TableName() string {
	return "orders_day_statistics"
}

type OrdersMonthStatistics struct {
	BaseModel
	BuyerCount  int64  `gorm:"column:buyer_count" json:"buyer_count"`
	TotalAmount int64  `gorm:"column:total_amount" json:"total_amount"`
	Date        string `gorm:"column:date" json:"date"`
}

func (OrdersMonthStatistics) TableName() string {
	return "orders_month_statistics"
}
