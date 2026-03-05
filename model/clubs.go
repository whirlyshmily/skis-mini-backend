package model

import "time"

//CREATE TABLE `clubs` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '教练名',
//  `uid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'uid',
//  `phone` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号',
//  `level` int DEFAULT '1' COMMENT '俱乐部等级',
//  `balance` int DEFAULT '0' COMMENT '钱包余额',
//  `deposit` int DEFAULT '0' COMMENT '保证金',
//  `frozen_deposit` tinyint DEFAULT '0' COMMENT '是否冻结保证金，0-不冻结，1-冻结',
//  `paid_service_fee` int DEFAULT '0' COMMENT '缴纳平台服务费',
//  `finished_course` int DEFAULT '0' COMMENT '完成课程',
//  `wrote_course_record` int DEFAULT '0' COMMENT '填写课程记录',
//  `priority` int DEFAULT '100' COMMENT '优先级',
//  `verified` tinyint DEFAULT '0' COMMENT '是否认证，0-未认证，10已认证',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常，1-删除',
//  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='俱乐部表'

type Clubs struct {
	Id                int64               `gorm:"primaryKey" json:"id"`
	Name              string              `gorm:"column:name" json:"name"`
	Logo              string              `gorm:"column:logo" json:"logo"`
	ApprovalLogo      string              `gorm:"column:approval_logo" json:"approval_logo"`
	ClubId            string              `gorm:"column:club_id" json:"club_id"`
	Uid               string              `gorm:"column:uid" json:"uid"`
	Manager           string              `gorm:"column:manager" json:"manager"`
	Phone             string              `gorm:"column:phone" json:"phone"`
	Introduction      string              `gorm:"column:introduction" json:"introduction"`
	SocialCreditCode  string              `gorm:"column:social_credit_code" json:"social_credit_code"`
	BusinessLicense   string              `gorm:"column:business_license" json:"business_license"`
	IdCardFront       string              `gorm:"column:id_card_front" json:"id_card_front"`
	IdCardBack        string              `gorm:"column:id_card_back" json:"id_card_back"`
	Level             int                 `gorm:"column:level" json:"level"`
	TotalProfit       int64               `gorm:"column:total_profit" json:"total_profit"`
	ReferralProfit    int64               `gorm:"column:referral_profit" json:"referral_profit"`
	Balance           int64               `gorm:"column:balance" json:"balance"`
	Deposit           int64               `gorm:"column:deposit" json:"deposit"`
	FrozenDeposit     int                 `gorm:"column:frozen_deposit" json:"frozen_deposit"`
	PaidServiceFee    int                 `gorm:"column:paid_service_fee" json:"paid_service_fee"`
	FinishedCourse    int                 `gorm:"column:finished_course" json:"finished_course"`
	WroteCourseRecord int                 `gorm:"column:wrote_course_record" json:"wrote_course_record"`
	Priority          int                 `gorm:"column:priority" json:"priority"`
	Verified          int                 `gorm:"column:verified" json:"verified"`
	OpTime            time.Time           `gorm:"column:op_time" json:"op_time"`
	Remark            string              `gorm:"column:remark" json:"remark"`
	ReferralCode      string              `gorm:"column:referral_code" json:"referral_code"`
	PriceMin          int64               `gorm:"column:price_min" json:"price_min"`
	PriceMax          int64               `gorm:"column:price_max" json:"price_max"`
	State             int                 `gorm:"column:state" json:"-"`                                                             //0-正常，1-删除
	CreatedAt         time.Time           `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"-"`                // 创建时间
	UpdatedAt         time.Time           `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"-"` // 更新时间
	ClubTags          []ClubsTags         `gorm:"foreignKey:ClubID;references:ClubId" json:"clubs_tags"`
	ClubsCoaches      []ClubsCoaches      `gorm:"foreignKey:ClubID;references:ClubId" json:"clubs_coaches"`
	ClubsSkiResorts   []ClubsSkiResorts   `gorm:"foreignKey:ClubID;references:ClubId" json:"clubs_ski_resorts"`
	Certificates      []ClubsCertificates `gorm:"foreignKey:ClubID;references:ClubId" json:"certificates"`
	ServiceRate       int                 `gorm:"-" json:"service_rate"`
	LevelInfo         *CoachesLevel       `gorm:"foreignKey:Level;references:Level" json:"-"`
	Users             *Users              `gorm:"foreignKey:Uid;references:Uid" json:"user_info,omitempty"`
}

func (Clubs) TableName() string {
	return "clubs"
}
