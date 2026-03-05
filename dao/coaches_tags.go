package dao

import (
	"go.uber.org/zap"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

func QueryCoachesTagCountByTagId(tagId int64) (int64, error) {
	var count int64
	if err := global.DB.Model(&model.CoachesTags{}).Where("tag_id = ? and state = 0", tagId).Count(&count).Error; err != nil {
		global.Lg.Error("查询教练标签数量失败", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func QueryCoachIdByTagId(tagId int64) (coachIds []string, err error) {
	if err = global.DB.Model(&model.CoachesTags{}).Where("tag_id = ? and state = 0", tagId).Pluck("coach_id", &coachIds).Error; err != nil {
		global.Lg.Error("查询教练标签失败", zap.Error(err))
		return nil, err
	}
	return coachIds, nil
}
func QueryCoachIdByTagIds(tagIds []int64, coachId string) (coachTagIds []string, err error) {
	if err = global.DB.Model(&model.CoachesTags{}).
		Where("coach_id = ? and tag_id in ? and state = 0 and verified = ?", coachId, tagIds, model.VerifiedVerified).
		Pluck("tag_id", &coachTagIds).Error; err != nil {
		global.Lg.Error("查询教练标签失败", zap.Error(err))
		return nil, err
	}
	return coachTagIds, nil
}

func QueryCoachAllTags(coachId string, verified int) (coachesTags []model.CoachesTags, err error) {
	db := global.DB.Model(&model.CoachesTags{})
	if verified != -1 {
		db = db.Where("verified = ?", verified)
	}
	err = db.
		Preload("Tag").
		Where("coach_id = ? and state = 0", coachId).
		Find(&coachesTags).Error
	if err != nil {
		global.Lg.Error("查询教练全部标签失败", zap.Error(err))
		return nil, err
	}
	return coachesTags, nil
}
