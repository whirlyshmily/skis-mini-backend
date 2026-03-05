package dao

import (
	"context"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func QueryReferralRecordInfoByReferralCode(ctx context.Context, referralCode string) (*model.ReferralRecords, error) {
	var referralRecord *model.ReferralRecords
	if err := global.DB.Model(&model.ReferralRecords{}).Where("referral_code = ? and state = 0", referralCode).First(&referralRecord).Error; err != nil {
		global.Lg.Error("QueryReferralRecordInfoByUid failed ", zap.Error(err), zap.String("referralCode", referralCode))
		return nil, err
	}
	return referralRecord, nil
}

func QueryReferralRecordInfoByUserId(ctx context.Context, userId string, userType int) (*model.ReferralRecords, error) {
	var referralRecord *model.ReferralRecords
	if err := global.DB.Model(&model.ReferralRecords{}).Where("user_id = ? and user_type = ?", userId, userType).Last(&referralRecord).Error; err != nil {
		global.Lg.Error("QueryReferralRecordInfoByUserId failed ", zap.Error(err), zap.String("userId", userId), zap.Int("userType", userType))
		return nil, err
	}
	return referralRecord, nil
}

func CreateReferralRecord(ctx context.Context, tx *gorm.DB, referralRecord *model.ReferralRecords) error {
	if err := tx.Model(&model.ReferralRecords{}).Create(referralRecord).Error; err != nil {
		global.Lg.Error("CreateReferralRecord failed ", zap.Error(err))
		return err
	}
	return nil
}

func QueryReferralRecordsByReferralUserId(ctx context.Context, referralUserId string, referralType int) ([]*model.ReferralRecords, error) {
	var referralRecords []*model.ReferralRecords
	if err := global.DB.Model(&model.ReferralRecords{}).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(model.ScopeUserSensitiveFields)
		}).
		Preload("Coach", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(model.ScopeCoachFields).Preload("CoachTags", "verified=1 and state=0").
				Preload("CoachTags.Tag", "state = 0").
				Preload("Certificates", "verified=1 and state=0").
				Preload("Certificates.CertificateConfig", "state = 0").
				Preload("CoachesSkiResorts", "state=0").Preload("CoachesSkiResorts.SkiResorts").Preload("LevelInfo", "state=0")
		}).Where("referral_user_id = ? and referral_type = ? and state = 0", referralUserId, referralType).Find(&referralRecords).Error; err != nil {
		global.Lg.Error("QueryReferralRecordsByReferralUserId failed ", zap.Error(err), zap.String("referralUserId", referralUserId), zap.Int("referralType", referralType))
		return nil, err
	}
	return referralRecords, nil
}
