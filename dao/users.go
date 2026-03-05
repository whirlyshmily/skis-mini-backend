package dao

import (
	"context"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetOrCreateUser 查询或创建用户
func GetOrCreateUser(ctx context.Context, openID string, userInfo forms.UserInfo, referralCode string) (*model.Users, error) {
	// 查询用户是否存在
	user, err := QueryUserByOpenID(openID)
	if err != nil && err != gorm.ErrRecordNotFound {
		global.Lg.Error("GetOrCreateUser QueryUserByOpenID error", zap.Error(err))
		return nil, err
	}

	if err == gorm.ErrRecordNotFound {
		var referralUserId string
		var referralUserType int
		if referralCode != "" {
			_, referralUserId, referralUserType, err = QueryUserIdReferralCode(ctx, referralCode)
			if err != nil {
				global.Lg.Error("GetOrCreateUser QueryUserInfoByReferralCode error", zap.Error(err), zap.String("referralCode", referralCode))
				return nil, err
			}
		}

		if err = global.DB.Transaction(func(tx *gorm.DB) error {
			user = &model.Users{
				Uid:      GenerateId("U"), // 生成唯一用户ID
				OpenId:   openID,
				Nickname: "E滑萌新", //这里没取用户真是的昵称
				Avatar:   userInfo.AvatarURL,
				Gender:   userInfo.Gender,
				Country:  userInfo.Country,
				Province: userInfo.Province,
				City:     userInfo.City,
			}

			err = tx.Create(user).Error
			if err != nil {
				global.Lg.Error("GetOrCreateUser Create error", zap.Error(err))
				return err
			}

			if referralUserId != "" {
				referralRecord := &model.ReferralRecords{
					UserID:         user.Uid,
					UserType:       model.UserTypeUser,
					ReferralUserID: referralUserId,
					ReferralType:   referralUserType,
					ReferralCode:   referralCode,
				}
				if err = CreateReferralRecord(ctx, tx, referralRecord); err != nil {
					global.Lg.Error("CreateReferralRecord error", zap.Error(err), zap.Any("referralRecord", referralRecord))
					return err
				}
			}

			return nil
		}); err != nil {
			global.Lg.Error("GetOrCreateUser Create error", zap.Error(err))
			return nil, err
		}

	} else {
		// 更新用户信息
		if user.Nickname == "" && userInfo.NickName != "" {
			user.Nickname = userInfo.NickName
		}
		if user.Avatar == "" && userInfo.AvatarURL != "" {
			user.Avatar = userInfo.AvatarURL
		}
		if user.Gender == 0 && userInfo.Gender != 0 {
			user.Gender = userInfo.Gender
		}
		if user.Country == "" && userInfo.Country != "" {
			user.Country = userInfo.Country
		}
		if user.Province == "" && userInfo.Province != "" {
			user.Province = userInfo.Province
		}
		if user.City == "" && userInfo.City != "" {
			user.City = userInfo.City
		}

		err = global.DB.Save(user).Error
		if err != nil {
			global.Lg.Error("GetOrCreateUser Save error", zap.Error(err))
			return nil, err
		}
	}

	return user, nil
}

// QueryUserByOpenID 根据OpenID查询用户
func QueryUserByOpenID(openID string) (*model.Users, error) {
	var user model.Users
	if err := global.DB.Where("open_id = ? and state = 0", openID).First(&user).Error; err != nil {
		global.Lg.Error("QueryUserByOpenID error", zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func QueryUserInfo(uid string) (*model.Users, error) {
	var user *model.Users
	if err := global.DB.Table("users").Where("uid = ? AND state = 0", uid).First(&user).Error; err != nil {
		global.Lg.Error("QueryUserInfo error", zap.Error(err))
		return nil, err
	}
	return user, nil
}

// UpdateUserPhone 更新用户手机号
func UpdateUserPhone(uid, phoneNumber string) (*model.Users, error) {
	var user model.Users
	err := global.DB.Where("uid = ? AND state = 0", uid).First(&user).Error
	if err != nil {
		global.Lg.Error("UpdateUserPhone error", zap.Error(err))
		return nil, err
	}

	// 更新手机号
	user.Phone = phoneNumber
	err = global.DB.Model(&model.Users{}).Where("uid = ?", uid).Save(&user).Error
	if err != nil {
		global.Lg.Error("UpdateUserPhone error", zap.Error(err))
		return nil, err
	}

	return &user, nil
}

func UpdateUserInfo(ctx context.Context, uid string, req *forms.UpdateUserInfoRequest) (*model.Users, error) {
	var user model.Users
	err := global.DB.Where("uid = ? AND state = 0", uid).First(&user).Error
	if err != nil {
		global.Lg.Error("UpdateUserPhone error", zap.Error(err))
		return nil, err
	}

	if req.NickName != nil {
		user.Nickname = *req.NickName
	}

	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}

	if req.Gender != nil {
		user.Gender = *req.Gender
	}

	if req.Birthday != nil {
		user.Birthday = *req.Birthday
	}
	err = global.DB.Model(&model.Users{}).Where("uid = ?", uid).Save(&user).Error
	if err != nil {
		global.Lg.Error("UpdateUserPhone error", zap.Error(err))
		return nil, err
	}

	return &user, nil
}

func SubUserPoints(tx *gorm.DB, uid string, points int64) error {
	if err := tx.Model(&model.Users{}).Where("uid = ?", uid).UpdateColumn("left_points", gorm.Expr("left_points - ?", points)).Error; err != nil {
		global.Lg.Error("SubUserPoints error", zap.Error(err))
		return err
	}

	return nil
}

func AddUserPoints(tx *gorm.DB, uid string, points int64) error {
	if err := tx.Model(&model.Users{}).Where("uid = ?", uid).UpdateColumn("left_points", gorm.Expr("left_points + ?", points)).Error; err != nil {
		global.Lg.Error("AddUserPoints error", zap.Error(err))
		return err
	}

	return nil
}

func QueryUserIdReferralCode(ctx context.Context, referralCode string) (string, string, int, error) {
	//先查询教练表，教练表不存在就查询俱乐部表
	var coach model.Coaches
	if err := global.DB.Model(&model.Coaches{}).Where("referral_code = ? and state = 0", referralCode).First(&coach).Error; err == nil {
		return coach.Uid, coach.CoachId, model.UserTypeCoach, nil
	}

	//再查询俱乐部表
	var club model.Clubs
	if err := global.DB.Model(&model.Clubs{}).Where("referral_code = ? and state = 0", referralCode).First(&club).Error; err == nil {
		return club.Uid, club.ClubId, model.UserTypeClub, nil
	}

	return "", "", 0, enum.NewErr(enum.ReferralCodeNotExistErr, "推荐码不存在")
}

func GetPointsLevel(points int64) int {
	//0～499【Lv1】
	//500～999【Lv2】
	//1000～2999【Lv3】
	//3000～5999【Lv4】
	//6000～9999【Lv5】
	//9999～15999【Lv6】
	//16000以上【Lv7】
	if points <= 499 {
		return 1
	} else if points <= 999 {
		return 2
	} else if points <= 2999 {
		return 3
	} else if points <= 5999 {
		return 4
	} else if points <= 9999 {
		return 5
	} else if points <= 15999 {
		return 6
	} else {
		return 7
	}
}
