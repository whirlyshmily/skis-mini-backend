package cron

import (
	"skis-admin-backend/dao"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

type ClubJob struct {
}

func (m ClubJob) Run() {
	var clubs []*model.Clubs
	err := global.DB.Table("club").Select("club_id").Where("verified=1 and state = 0").Find(&clubs).Error
	if err != nil {
		return
	}
	for _, club := range clubs {
		dao.HandleClubData(nil, club.ClubId)
	}

}
