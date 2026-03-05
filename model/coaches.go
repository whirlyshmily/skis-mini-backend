package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

//CREATE TABLE `coaches` (
//  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
//  `coach_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '教练id',
//  `uid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'uid',
//  `realname` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '真实姓名',
//  phone                   varchar(20)                         null comment '手机号',
//  `id_card` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '身份证',
//  `introduction` text COLLATE utf8mb4_unicode_ci COMMENT '介绍',
//  `recommender_uid` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '推荐人uid',
//  `id_card_photo` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '手持身份证照片',
//  `certificate` json DEFAULT NULL COMMENT '证书，json格式',
//  `supplementary_materials` text COLLATE utf8mb4_unicode_ci COMMENT '补充材料',
//  `verified` tinyint DEFAULT '0' COMMENT '是否认证，0-未认证，10已认证',
//  `priority` int DEFAULT '100' COMMENT '权重',
//  `total_fee` int DEFAULT '0' COMMENT '总收益',
//  `balance` int DEFAULT '0' COMMENT '余额',
//  `deposit` int DEFAULT '0' COMMENT '保证金',
//  `frozen_deposit` tinyint DEFAULT '0' COMMENT '是否冻结保证金，0-不冻结，1-冻结',
//  `paid_service_fee` int DEFAULT '0' COMMENT '缴纳平台服务费',
//  `finished_course` int DEFAULT '0' COMMENT '完成课程',
//  `wrote_course_record` int DEFAULT NULL COMMENT '填写课程记录',
//  `level` int DEFAULT '1' COMMENT '教练等级',
//  `state` tinyint DEFAULT '0' COMMENT '状态，0-正常， 1-删除',
//  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//  PRIMARY KEY (`id`),
//  UNIQUE KEY `coach_id` (`coach_id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

type CertificateItem struct {
	Name string `gorm:"column:name" json:"name"`
	Url  string `gorm:"column:url" json:"url"`
}

type Certificate []CertificateItem

// 实现Scan接口
func (c *Certificate) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for JSONArray: %T", value)
	}
	return json.Unmarshal(bytes, c)
}

// 实现Value接口
func (c *Certificate) Value() (driver.Value, error) {
	return json.Marshal(c)
}

const VerifiedUnverified = 0
const VerifiedVerified = 1
const VerifiedRejected = 2

type Coaches struct {
	Id                     int64                 `gorm:"primaryKey" json:"id"`
	CoachId                string                `gorm:"column:coach_id" json:"coach_id"`
	Uid                    string                `gorm:"column:uid" json:"uid"`
	Realname               string                `gorm:"column:realname" json:"realname,omitempty"`
	Phone                  string                `gorm:"column:phone" json:"phone,omitempty"`
	IdCard                 string                `gorm:"column:id_card" json:"idCard,omitempty"`
	Introduction           string                `gorm:"column:introduction" json:"introduction,omitempty"`
	IdCardPhoto            string                `gorm:"column:id_card_photo" json:"id_card_photo"`
	SupplementaryMaterials JSONArray             `gorm:"column:supplementary_materials" json:"supplementary_materials,omitempty"`
	ApplyReferralCode      string                `gorm:"column:apply_referral_code" json:"apply_referral_code"`
	ReferralUserId         string                `gorm:"column:referral_user_id" json:"-"`   //对外不返回
	ReferralUserType       int                   `gorm:"column:referral_user_type" json:"-"` //对外不返回
	Priority               int64                 `gorm:"column:priority" json:"priority,omitempty"`
	TotalProfit            int64                 `gorm:"column:total_profit" json:"total_profit"`
	Balance                int64                 `gorm:"column:balance" json:"balance,omitempty"`
	Deposit                int64                 `gorm:"column:deposit" json:"deposit"`
	FrozenDeposit          int64                 `gorm:"column:frozen_deposit" json:"frozen_deposit"`
	PaidServiceFee         int64                 `gorm:"column:paid_service_fee" json:"paid_service_fee"`
	FinishedCourse         int64                 `gorm:"column:finished_course" json:"finished_course"`
	WroteCourseRecord      int64                 `gorm:"column:wrote_course_record" json:"wrote_course_record"`
	Level                  int64                 `json:"level" gorm:"column:level" json:"level"`
	Verified               int64                 `gorm:"column:verified" json:"verified"`
	OpTime                 time.Time             `gorm:"column:op_time" json:"op_time"`
	Remark                 string                `gorm:"column:remark" json:"remark,omitempty"`
	ReferralCode           string                `gorm:"column:referral_code" json:"referral_code,omitempty"`
	PriceMin               int64                 `gorm:"column:price_min" json:"price_min"`
	PriceMax               int64                 `gorm:"column:price_max" json:"price_max"`
	State                  int                   `gorm:"column:state" json:"-"`
	CreatedAt              time.Time             `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt              time.Time             `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	Users                  *Users                `gorm:"foreignKey:Uid;references:Uid" json:"user_info,omitempty"`
	CoachTags              []CoachesTags         `gorm:"foreignKey:CoachId;references:CoachId" json:"coach_tags"`
	TimeIsMatch            bool                  `gorm:"-" json:"time_is_match"`
	Certificates           []CoachesCertificates `gorm:"foreignKey:CoachID;references:CoachId" json:"certificates"`
	CoachesSkiResorts      []CoachesSkiResorts   `gorm:"foreignKey:CoachID;references:CoachId" json:"coaches_ski_resorts"`
	LevelInfo              *CoachesLevel         `gorm:"foreignKey:Level;references:Level" json:"-"`
	ServiceRate            int                   `gorm:"-" json:"service_rate,omitempty"`
	Match                  MatchStruct           `gorm:"-" json:"match"`
}
type MatchStruct struct {
	TimeIsMatch bool `json:"time_is_match"` //时间匹配
	TagIsMatch  bool `json:"tag_is_match"`  //标签匹配
	SkiIsMatch  bool `json:"ski_is_match"`  //滑雪场匹配
}

func (Coaches) TableName() string {
	return "coaches"
}

func ScopeCoachFields(db *gorm.DB) *gorm.DB {
	return db.Select("id", "realname", "phone", "coach_id", "uid", "price_min", "price_max", "level", "verified")
}
