package dao

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

// GetOrCreateClubsUser 查询或创建用户
func GetOrCreateClubsUser(openID string, userInfo forms.UserInfo) (*model.ClubsUsers, error) {
	// 查询用户是否存在
	user, err := QueryClubsUserByOpenID(openID)
	if err != nil && err != gorm.ErrRecordNotFound {
		global.Lg.Error("GetOrCreateUser QueryUserByOpenID error", zap.Error(err))
		return nil, err
	}

	// 如果用户不存在则创建新用户
	if user == nil {
		user = &model.ClubsUsers{
			Uid:      GenerateId("U"), // 生成唯一用户ID
			OpenId:   openID,
			Nickname: userInfo.NickName,
			Avatar:   userInfo.AvatarURL,
			Gender:   userInfo.Gender,
		}
		err = global.DB.Create(user).Error
		if err != nil {
			global.Lg.Error("GetOrCreateUser Create error", zap.Error(err))
			return nil, err
		}
	} else {
		// 更新用户信息
		user.Nickname = userInfo.NickName
		user.Avatar = userInfo.AvatarURL
		user.Gender = userInfo.Gender

		err = global.DB.Save(user).Error
		if err != nil {
			global.Lg.Error("GetOrCreateUser Save error", zap.Error(err))
			return nil, err
		}
	}

	return user, nil
}

func QueryClubsUserByOpenID(openID string) (*model.ClubsUsers, error) {
	var user model.ClubsUsers
	if err := global.DB.Where("open_id = ? and state = 0", openID).First(&user).Error; err != nil {
		global.Lg.Error("QueryClubsUserByOpenID error", zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func QueryClubsUserInfo(uid string) (*model.ClubsUsers, error) {
	var user *model.ClubsUsers
	if err := global.DB.Where("uid = ? AND state = 0", uid).First(&user).Error; err != nil {
		global.Lg.Error("QueryClubsUserInfo error", zap.Error(err))
		return nil, err
	}
	return user, nil
}
