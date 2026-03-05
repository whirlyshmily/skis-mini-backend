package dao

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math/rand"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

type ClubsCoachesDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.ClubsCoaches
}

func NewClubsCoachesDao(ctx context.Context, dbs ...*gorm.DB) *ClubsCoachesDao {
	dao := new(ClubsCoachesDao)
	switch len(dbs) {
	case 0:
		panic("database connection required")
	case 1:
		dao.sourceDB = dbs[0]
		dao.replicaDB = []*gorm.DB{dbs[0]}
	default:
		dao.sourceDB = dbs[0]
		dao.replicaDB = dbs[1:]
	}
	return dao
}

func (d *ClubsCoachesDao) GetAll(ctx context.Context, where string, args ...interface{}) (items []model.ClubsCoaches, err error) {
	err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Where(where, args...).Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("ClubsCoachesDao: Get where=%s: %w", where, err)
	}
	return items, nil
}
func (d *ClubsCoachesDao) Get(ctx context.Context, where string, args ...interface{}) (items *model.ClubsCoaches, err error) {
	err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Where(where, args...).First(&items).Error
	if err != nil {
		return nil, fmt.Errorf("ClubsCoachesDao: Get where=%s: %w", where, err)
	}
	return items, nil
}

func (d *ClubsCoachesDao) CoachJoinClubs(c *gin.Context, req forms.CoachJoinClubsRequest) error {
	userId := c.GetString("user_id")
	if userId == "" {
		return enum.NewErr(enum.CoachNotExistErr, "教练不存在")
	}
	clubInfo, err := QueryClubInfoByClubId(req.ClubID)
	if err != nil || clubInfo == nil {
		return enum.NewErr(enum.ClubExitErr, "俱乐部不存在")
	}

	coachClub := &model.ClubsCoaches{
		ClubID:  req.ClubID,
		CoachID: userId,
		State:   0,
	}
	//找不到审核中和审核通过的教练，直接插入
	err = d.sourceDB.Model(d.m).Where("coach_id=? and club_id=? and state=0 and  verified in ?", userId, req.ClubID, []int{model.VerifiedPass, model.VerifiedNo}).FirstOrCreate(coachClub, model.ClubsCoaches{}).Error
	if err != nil {
		global.Lg.Error("CoachJoinClubs FirstOrCreate  场地插入失败", zap.Error(err))
		return enum.NewErr(enum.ClubJoinErr, "加入俱乐部失败")
	}
	return nil
}

func (d *ClubsCoachesDao) CoachQuitClubs(c *gin.Context, req forms.CoachQuitClubsRequest) (err error) {
	userId := c.GetString("user_id")
	if userId == "" {
		return enum.NewErr(enum.CoachNotExistErr, "教练不存在")
	}
	clubCoach, err := d.Get(c, "coach_id = ? and state = 0 and club_id = ?", userId, req.ClubID)
	if err != nil {
		return enum.NewErr(enum.CoachNotExistErr, "教练不存在该俱乐部")
	}

	//TODO 检测教练是否存在该俱乐部未完成课程
	if clubCoach.Verified == model.VerifiedPass { //教练通过审核，需要检测是否存在未完成课程
		err = global.DB.Model(model.Orders{}).Where("orders.user_id = ? and orders.status = 1 and orders.state = 0", req.ClubID).
			Joins("join orders_courses on orders_courses.order_id = orders.order_id and teach_coach_id=? and orders_courses.state=0 and is_check=0", userId).
			First(&model.Orders{}).Error
		if err != gorm.ErrRecordNotFound { //存在未完成课程
			return enum.NewErr(enum.CoachQuitErr, "请先完成此俱乐部订单再退出")
		}
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(d.m).Where("coach_id = ? and club_id = ? and state = 0", userId, req.ClubID).Update("state", 1).Error
		if err != nil {
			global.Lg.Error("QueryCoachesLevelsList error", zap.Error(err))
			return enum.NewErr(enum.CoachQuitErr, "退出俱乐部失败")
		}

		if clubCoach.Verified != model.VerifiedPass { //如果教练未通过审核，清除教练出俱乐部即可
			return nil
		}

		var skiResortsTeachTime []model.SkiResortsTeachTime
		err = tx.Model(&model.SkiResortsTeachTime{}).
			Where("user_id = ? and state = 0 and user_type = ? ", userId, enum.UserTypeCoach).
			Find(&skiResortsTeachTime).Error
		if len(skiResortsTeachTime) == 0 { //该教练没有教学时间
			return nil
		}

		teachStartTimeMap := map[int][]model.LocalTime{}
		for _, v := range skiResortsTeachTime {
			if teachStartTimeMap[v.SkiResortsID] == nil {
				teachStartTimeMap[v.SkiResortsID] = []model.LocalTime{}
			}
			teachStartTimeMap[v.SkiResortsID] = append(teachStartTimeMap[v.SkiResortsID], v.TeachStartTime)
		}
		for skiResortID, teachStartTimes := range teachStartTimeMap {
			err = tx.Model(&model.SkiResortsTeachTime{}).
				Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", clubCoach.ClubID, enum.UserTypeClub, skiResortID).
				Where("teach_start_time in ?", teachStartTimes).
				Update("teach_num", gorm.Expr("teach_num - ?", 1)).Error
			if err != nil {
				global.Lg.Error("CoachQuitClubs error", zap.Error(err))
				return enum.NewErr(enum.TeachTimeErr, "更新俱乐部教学时间失败")
			}
		}
		return nil
	})
	HandleClubData(c, req.ClubID)
	return err
}

func ClubCheckCoach(c *gin.Context, req *forms.ClubCheckCoachRequest) error {
	userId := c.GetString("user_id")
	clubsCoach := &model.ClubsCoaches{}
	err := global.DB.Table("clubs_coaches").Where("club_id = ? and id = ? and state = 0", userId, req.Id).First(&clubsCoach).Error
	if err != nil {
		return enum.NewErr(enum.ClubCoachCheckErr, "审核记录不存在")
	}

	if clubsCoach.Verified == req.Verified {
		return enum.NewErr(enum.ClubCoachCheckErr, "审核状态未改变")
	}
	if clubsCoach.Verified != model.VerifiedNo {
		return enum.NewErr(enum.ClubCoachCheckErr, "请勿重复审核")
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Table("clubs_coaches").Where("club_id = ? and id = ? and state = 0", userId, req.Id).
			Update("verified", req.Verified).Error
		if err != nil {
			global.Lg.Error("ClubCheckCoach error", zap.Error(err))
			return enum.NewErr(enum.ClubCoachCheckErr, "审核教练失败")
		}
		if req.Verified == model.VerifiedReject { //审核不通过，就直接返回
			return nil
		}

		//审核通过，需要更新俱乐部的课程时间

		//获取教练的课程时间
		var coachTimes []model.SkiResortsTeachTime
		tx.Model(&model.SkiResortsTeachTime{}).Where("user_id = ? and user_type = ? and state = 0", clubsCoach.CoachID, enum.UserTypeCoach).
			Find(&coachTimes)

		//将教练教学时间根据场地分组
		dataMap := make(map[int][]model.SkiResortsTeachTime)
		timeMap := make(map[int][]model.LocalTime)
		for _, v := range coachTimes {
			if dataMap[v.SkiResortsID] == nil {
				dataMap[v.SkiResortsID] = []model.SkiResortsTeachTime{}
				timeMap[v.SkiResortsID] = []model.LocalTime{}
			}
			dataMap[v.SkiResortsID] = append(dataMap[v.SkiResortsID], v)
			timeMap[v.SkiResortsID] = append(timeMap[v.SkiResortsID], v.TeachStartTime)
		}

		clubTeachTimes := make([]model.SkiResortsTeachTime, 0)
		for skiResortsID, timeData := range dataMap { //遍历教练个个场地的教学时间
			clubTimes := make([]model.LocalTime, 0)

			tx.Model(&model.SkiResortsTeachTime{}).
				Where("user_id = ? and ski_resorts_id = ? and teach_start_time in (?) and state = 0", userId, skiResortsID, timeMap[skiResortsID]).
				Pluck("teach_start_time", &clubTimes) //获取俱乐部课程的开始时间

			clubTimeMap := make(map[model.LocalTime]struct{}) //为了避免重复添加，使用map
			for _, v := range clubTimes {
				clubTimeMap[v] = struct{}{}
			}
			if len(clubTimes) != 0 { //俱乐部存在的教学时间，teach_num+1
				tx.Model(&model.SkiResortsTeachTime{}).
					Where("user_id = ? and ski_resorts_id = ? and teach_start_time in (?) and state = 0", userId, skiResortsID, clubTimes).
					Update("teach_num", gorm.Expr("teach_num + ?", 1))
			}
			//遍历教练的课程时间，如果不存在俱乐部课程，就添加到俱乐部
			for _, v := range timeData {
				if _, ok := clubTimeMap[v.TeachStartTime]; ok {
					continue
				}
				clubTeachTimes = append(clubTeachTimes, model.SkiResortsTeachTime{
					UserID:         userId,
					UserType:       enum.UserTypeClub,
					SkiResortsID:   v.SkiResortsID,
					TeachDate:      v.TeachDate,
					TeachStartTime: v.TeachStartTime,
					TeachEndTime:   v.TeachEndTime,
					TeachNum:       1,
				})
			}
		}
		if len(clubTeachTimes) > 0 {
			err = tx.Create(&clubTeachTimes).Error
			if err != nil {
				global.Lg.Error("创建俱乐部课程失败", zap.Error(err))
				return enum.NewErr(enum.TeachTimeErr, "创建俱乐部课程失败")
			}
		}
		return err
	})
	HandleClubData(c, userId)
	return err
}

func HandleClubData(c *gin.Context, clubId string) {
	HandleClubSki(c, clubId)
	HandleClubTag(c, clubId)
	HandleClubCertificates(c, clubId)
}

// HandleClubSki 处理俱乐部的场地
func HandleClubSki(c *gin.Context, clubId string) {
	var skiResortIds, clubsSkiResortIds []int64
	global.DB.Model(&model.ClubsCoaches{}).Distinct("ski_resorts_id").
		Joins("join coaches_ski_resorts as csr on clubs_coaches.coach_id=csr.coach_id and csr.state=0").
		Where("club_id = ? and verified = ? and clubs_coaches.state = 0", clubId, model.VerifiedPass).
		Pluck("ski_resorts_id", &skiResortIds)

	global.DB.Model(&model.ClubsSkiResorts{}).Where("club_id = ? and ski_resorts_id not in ? and state = 0", clubId, skiResortIds).
		Update("state", 1)

	global.DB.Model(&model.ClubsSkiResorts{}).Select("ski_resorts_id").Where("club_id = ? and state = 0", clubId).
		Scan(&clubsSkiResortIds)

	// 计算差集：skiResortIds 中有但 clubsSkiResortIds 中没有的值
	difference := []int64{}
	if len(clubsSkiResortIds) > 0 {
		clubSet := make(map[int64]bool)
		for _, id := range clubsSkiResortIds {
			clubSet[id] = true
		}

		for _, id := range skiResortIds {
			if !clubSet[id] {
				difference = append(difference, id)
			}
		}
	} else {
		// 如果 clubsSkiResortIds 为空，则差集就是 skiResortIds 的所有元素
		difference = skiResortIds
	}
	if len(difference) == 0 {
		return
	}
	// 批量插入差值到 ClubsSkiResorts 表
	var newClubSkiResorts []model.ClubsSkiResorts
	for _, skiResortId := range difference {
		newClubSkiResorts = append(newClubSkiResorts, model.ClubsSkiResorts{
			ClubID:       clubId,
			SkiResortsID: skiResortId,
			State:        0,
		})
	}

	if len(newClubSkiResorts) > 0 {
		global.DB.Model(&model.ClubsSkiResorts{}).Create(&newClubSkiResorts)
	}
	return
}

// HandleClubTag 处理俱乐部的技能
func HandleClubTag(c *gin.Context, clubId string) {
	var tagIds, clubsTagIds []int64
	global.DB.Model(&model.ClubsCoaches{}).Distinct("tag_id").
		Joins("join coaches_tags as ct on clubs_coaches.coach_id=ct.coach_id and ct.state=0 and ct.verified=1").
		Where("club_id = ? and clubs_coaches.verified = ? and clubs_coaches.state = 0", clubId, model.VerifiedPass).
		Pluck("tag_id", &tagIds)

	global.DB.Model(&model.ClubsTags{}).Where("club_id = ? and tag_id not in ? and state = 0", clubId, tagIds).
		Update("state", 1)

	global.DB.Model(&model.ClubsTags{}).Select("tag_id").Where("club_id = ? and state = 0", clubId).
		Scan(&clubsTagIds)

	// 计算差集：skiResortIds 中有但 clubsSkiResortIds 中没有的值
	var difference []int64
	if len(clubsTagIds) > 0 {
		clubSet := make(map[int64]bool)
		for _, id := range clubsTagIds {
			clubSet[id] = true
		}

		for _, id := range tagIds {
			if !clubSet[id] {
				difference = append(difference, id)
			}
		}
	} else {
		// 如果 clubsSkiResortIds 为空，则差集就是 skiResortIds 的所有元素
		difference = tagIds
	}
	if len(difference) == 0 {
		return
	}
	// 批量插入差值到 newClubsTags 表
	var newClubsTags []model.ClubsTags
	for _, tagId := range difference {
		newClubsTags = append(newClubsTags, model.ClubsTags{
			ClubID: clubId,
			TagID:  tagId,
			State:  0,
		})
	}

	if len(newClubsTags) > 0 {
		global.DB.Model(&model.ClubsTags{}).Create(&newClubsTags)
	}
	return
}

// HandleClubCertificates 处理俱乐部的证书
func HandleClubCertificates(c *gin.Context, clubId string) {
	var coachCertificates []*model.CoachesCertificates
	global.DB.Model(&model.ClubsCoaches{}).Select("certificate_id, level").
		Joins("join coaches_certificates as cc on clubs_coaches.coach_id=cc.coach_id and cc.state=0 and cc.verified=1").
		Where("club_id = ? and clubs_coaches.verified = ? and clubs_coaches.state = 0", clubId, model.VerifiedPass).
		Group("certificate_id, level").
		Scan(&coachCertificates)

	var clubsCertificates []*model.ClubsCertificates
	global.DB.Model(&model.ClubsCertificates{}).Where("club_id = ? and state=0", clubId).Find(&clubsCertificates)

	coachCertMap := make(map[string]*model.CoachesCertificates)

	for _, cert := range coachCertificates {
		// 使用certificate_id和level作为联合键
		key := fmt.Sprintf("%d_%s", cert.CertificateID, cert.Level)
		coachCertMap[key] = cert
	}

	var delIds []int64
	for _, cert := range clubsCertificates {
		// 使用certificate_id和level作为联合键
		key := fmt.Sprintf("%d_%s", cert.CertificateID, cert.Level)
		if _, ok := coachCertMap[key]; ok {
			delete(coachCertMap, key)
		} else {
			delIds = append(delIds, cert.ID)
		}
	}
	if len(delIds) > 0 {
		global.DB.Model(&model.ClubsCertificates{}).Where("club_id = ? and tag_id not in ? and state = 0", clubId, delIds).
			Update("state", 1)
	}

	// 批量插入差值到 newClubsTags 表
	var newClubsTags []model.ClubsCertificates
	for _, cert := range coachCertMap {
		newClubsTags = append(newClubsTags, model.ClubsCertificates{
			ClubID:        clubId,
			CertificateID: cert.CertificateID,
			Level:         cert.Level,
			State:         0,
		})
	}
	if len(newClubsTags) > 0 {
		global.DB.Model(&model.ClubsCertificates{}).Create(&newClubsTags)
	}
	return
}

func (d *ClubsCoachesDao) CoachClubsList(c *gin.Context, req forms.CoachClubsListRequest) ([]*model.ClubsCoaches, error) {
	userId := c.GetString("user_id")
	var list []*model.ClubsCoaches
	db := d.sourceDB.Model(d.m).Where("coach_id = ?  and state = 0", userId)
	if len(req.ClubIDs) > 0 {
		db = db.Where("club_id in ?", req.ClubIDs)
	}

	if err := db.Order("id desc").Find(&list).Error; err != nil {
		global.Lg.Error("查询教练加入俱乐部列表失败", zap.Error(err))
		return nil, err
	}
	return list, nil
}

func (d *ClubsCoachesDao) ClubsCoachList(c *gin.Context, clubId string, req forms.ClubsCoachListRequest) ([]*model.ClubsCoaches, error) {
	var list []*model.ClubsCoaches
	db := d.sourceDB.Model(d.m).Preload("Coaches.Users").
		Preload("Coaches.CoachTags", "verified=1 and state=0").
		Preload("Coaches.CoachTags.Tag", "state = 0").
		Preload("Coaches.Certificates", "verified=1 and state=0").
		Preload("Coaches.Certificates.CertificateConfig", "state = 0").
		Preload("Coaches.CoachesSkiResorts", "state=0").
		Preload("Coaches.CoachesSkiResorts.SkiResorts").
		Preload("Coaches.LevelInfo", "state=0").
		Where("club_id = ?  and state = 0", clubId)
	if req.Verified != nil && len(req.Verified) > 0 {
		db = db.Where("verified in ?", req.Verified)
	}
	if err := db.Order("id desc").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find(&list).Error; err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		return nil, err
	}
	return list, nil
}

func (d *ClubsCoachesDao) ClubsCoachAll(c *gin.Context, clubId string) ([]*model.ClubsCoaches, error) {
	var list []*model.ClubsCoaches
	db := d.sourceDB.Model(d.m).Preload("Coaches.Users").
		Preload("Coaches.CoachTags", "verified=1 and state=0").
		Preload("Coaches.CoachTags.Tag", "state = 0").
		Preload("Coaches.Certificates", "verified=1 and state=0").
		Preload("Coaches.Certificates.CertificateConfig", "state = 0").
		Preload("Coaches.CoachesSkiResorts", "state=0").
		Preload("Coaches.CoachesSkiResorts.SkiResorts").
		Preload("Coaches.LevelInfo", "state=0").
		Where("club_id = ?  and state = 0", clubId).Where("verified = ?", 1)
	if err := db.Order("id desc").Find(&list).Error; err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		return nil, err
	}
	return list, nil
}
