package dao

import (
	"go.uber.org/zap"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

func QueryClubCertificatesByClubId(clubId string) (club *model.ClubsTags, err error) {
	if err = global.DB.Model(&model.ClubsTags{}).Preload("Tag").Where("club_id = ? and state = 0", clubId).Find(&club).Error; err != nil {
		global.Lg.Error("查询俱乐部标签失败", zap.Error(err))
		return nil, err
	}
	return club, nil
}

/*
func QueryClubIdByTagId(tagId int) (clubIds []string, err error) {
	if err = global.DB.Model(&model.ClubsTags{}).Where("tag_id = ? and state = 0", tagId).Pluck("club_id", &clubIds).Error; err != nil {
		global.Lg.Error("查询俱乐部标签失败", zap.Error(err))
		return nil, err
	}
	return clubIds, nil
}

func QueryClubIdByTagIds(tagIds []int64, clubId string) (ClubCertificatesIds []string, err error) {
	if err = global.DB.Model(&model.ClubsTags{}).
		Where("club_id = ? and tag_id in ? and state = 0 and  verified = ?", clubId, tagIds, model.VerifiedVerified).
		Pluck("tag_id", &ClubCertificatesIds).Error; err != nil {
		global.Lg.Error("查询俱乐部标签失败", zap.Error(err))
		return nil, err
	}
	return ClubCertificatesIds, nil
}

func QueryClubAllTags(clubId string) (clubsTags []model.ClubsTags, err error) {
	db := global.DB.Model(&model.ClubsTags{})
	err = db.
		Preload("Tag").
		Where("club_id = ? and state = 0", clubId).
		Find(&clubsTags).Error
	if err != nil {
		global.Lg.Error("查询俱乐部全部标签失败", zap.Error(err))
		return nil, err
	}
	return clubsTags, nil
}
*/
