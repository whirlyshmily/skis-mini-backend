package dao

import (
	"go.uber.org/zap"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

func QueryCoachesLevelsList() ([]*model.CoachesLevel, error) {
	var levels []*model.CoachesLevel
	if err := global.DB.Table("coaches_level").Where("state = 0").Find(&levels).Error; err != nil {
		global.Lg.Error("QueryCoachesLevelsList error", zap.Error(err))
		return nil, err
	}
	return levels, nil
}
