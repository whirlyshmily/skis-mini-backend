package dao

import (
	"context"
	"errors"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func QueryClubList(c *gin.Context, req *forms.QueryClubListRequest) (int64, []*model.Clubs, error) {
	var clubIds []string
	if req.TagID != 0 {
		cIds, err := QueryClubIdByTagId(req.TagID)
		if err != nil {
			global.Lg.Error("QueryClubList QueryClubIdByTagId 查询俱乐部ID失败", zap.Error(err))
		}
		clubIds = append(clubIds, cIds...)
	}
	if req.SkiResortsId != 0 {
		cIds, err := NewClubSkiResortsDao(c, global.DB).QueryClubIdBySkiId(req.SkiResortsId)
		if err != nil {
			global.Lg.Error("QueryClubList QueryClubIdBySkiId 查询俱乐部ID失败", zap.Error(err))
		}
		clubIds = append(clubIds, cIds...)
	}

	db := global.DB.Table("clubs").Where("state = 0")
	if req.Keyword != "" {
		db = db.Where("name like ? or club_id like ? or phone like ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	//标签和课程ID不为空时，查询标签和课程ID对应的俱乐部ID
	if req.SkiResortsId != 0 || req.TagID != 0 {
		db = db.Where("club_id in (?)", clubIds)
	}

	if req.Verified == 0 {
		db = db.Where("verified in (0,2)")
	} else {
		db = db.Where("verified = 1")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("查询俱乐部列表失败", zap.Error(err))
		return 0, nil, err
	}
	coachid := c.GetString("coach_id")
	db = db.Preload("ClubTags", "state = 0").
		Preload("Certificates", "state = 0").
		Preload("Certificates.CertificateConfig").
		Preload("ClubTags.Tag").
		Preload("ClubsCoaches", "state = 0 and coach_id=?", coachid)
	db = db.Preload("ClubsSkiResorts", "state = 0").
		Preload("ClubsSkiResorts.SkiResorts")

	var club []*model.Clubs
	if err := db.Order("id desc").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&club).Error; err != nil {
		global.Lg.Error("查询俱乐部列表失败", zap.Error(err))
		return 0, nil, err
	}

	return total, club, nil
}

func ClubCoachJoin(c *gin.Context, req *forms.CoachJoinClubRequest) error {
	clubId := req.ClubId
	coachId := c.GetString("coach_id")
	if coachId == "" {
		return enum.NewErr(enum.CoachNotExistErr, "教练ID不能为空")
	}

	var club model.Clubs
	if err := global.DB.Table("clubs").Where("club_id = ? and state = 0", clubId).First(&club).Error; err != nil {
		global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
		return enum.NewErr(enum.ClubExitErr, "俱乐部不存在")
	}

	err := global.DB.Table("clubs_coaches").Where("club_id = ? and coach_id = ? and state = 0", clubId, coachId).
		FirstOrCreate(&model.ClubsCoaches{}, map[string]interface{}{
			"club_id":  clubId,
			"coach_id": coachId,
			"state":    0,
		}).Error
	if err != nil {
		global.Lg.Error("教练加入俱乐部失败", zap.Error(err))
		return enum.NewErr(enum.ClubCoachJoinErr, "教练加入俱乐部失败")
	}
	return nil
}

func ClubCoachQuit(c *gin.Context, req *forms.CoachQuitClubRequest) error {
	clubId := req.ClubId
	coachId := c.GetString("coach_id")
	if coachId == "" {
		return enum.NewErr(enum.CoachNotExistErr, "教练ID不能为空")
	}

	var club model.Clubs
	if err := global.DB.Table("clubs").Where("club_id = ? and state = 0", clubId).First(&club).Error; err != nil {
		global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
		return enum.NewErr(enum.ClubExitErr, "俱乐部不存在")
	}

	err := global.DB.Table("clubs_coaches").Where("club_id = ? and coach_id = ? and state = 0", clubId, coachId).
		Update("state", 1).Error
	if err != nil {
		global.Lg.Error("教练退出俱乐部失败", zap.Error(err))
		return enum.NewErr(enum.ClubCoachQuitErr, "教练退出俱乐部失败")
	}
	return nil
}
func QueryClubInfoByClubId(clubId string) (*model.Clubs, error) {
	var club model.Clubs
	if err := global.DB.Table("clubs").Where("club_id = ? and state = 0", clubId).Preload("LevelInfo", "state = 0").First(&club).Error; err != nil {
		global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
		return nil, err
	}

	club.ServiceRate = enum.ServiceRatio
	if club.LevelInfo != nil {
		club.ServiceRate = club.LevelInfo.ServiceRate
	}

	return &club, nil
}

func QueryClubInfo(clubId string) (*model.Clubs, error) {
	var club model.Clubs
	if err := global.DB.Table("clubs").
		Preload("Certificates", "state = 0").
		Preload("Certificates.CertificateConfig").
		Preload("ClubTags", "state = 0").Preload("ClubTags.Tag").
		Preload("ClubsSkiResorts", "state = 0").Preload("ClubsSkiResorts.SkiResorts").
		Where("club_id = ? and state = 0", clubId).First(&club).Error; err != nil {
		global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
		return nil, err
	}
	return &club, nil
}

func QueryApplyClubInfo(c *gin.Context, uid string) (*model.Clubs, error) {
	var club *model.Clubs
	err := global.DB.Model(model.Clubs{}).
		Preload("ClubTags", "state=0").
		Preload("ClubTags.Tag", "state = 0").
		Preload("ClubsSkiResorts", "state = 0").Where("uid = ? and state = 0", uid).Last(&club).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, enum.NewErr(enum.ClubExitErr, "俱乐部不存在")
	}
	if err != nil {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		return nil, err
	}

	return club, nil
}

func UpdateClub(clubId string, data map[string]interface{}) error {
	if err := global.DB.Table("clubs").Where("club_id = ? and state = 0", clubId).Updates(data).Error; err != nil {
		global.Lg.Error("更新俱乐部失败", zap.Error(err))
		return err
	}
	return nil
}

func DeleteClub(clubId string) error {
	if err := global.DB.Table("clubs").Where("club_id = ? and state = 0", clubId).Update("state", 1).Error; err != nil {
		global.Lg.Error("删除俱乐部失败", zap.Error(err))
		return err
	}
	return nil
}

func QueryClubInfoByUid(uid string) (*model.Clubs, error) {
	var club model.Clubs
	if err := global.DB.Table("clubs").Where("uid = ? and state = 0", uid).First(&club).Error; err != nil {
		global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
		return nil, err
	}
	return &club, nil
}

func updateClubGoodsPriceAddGoods(ctx context.Context, clubId string, goodPrice int64) error {
	clubInfo, err := QueryClubInfoByClubId(clubId)
	if err != nil {
		global.Lg.Error("updateCoachGoodsPrice 查询教练信息失败", zap.Error(err))
		return err
	}

	data := make(map[string]interface{})
	if goodPrice > clubInfo.PriceMax {
		data["price_max"] = goodPrice
	}

	if clubInfo.PriceMin == 0 || goodPrice < clubInfo.PriceMin {
		data["price_min"] = goodPrice
	}

	if len(data) > 0 {
		if err = global.DB.Model(&model.Clubs{}).Where("club_id", clubId).Updates(data).Error; err != nil {
			global.Lg.Error("updateClubGoodsPriceAddGoods 添加商品价格失败", zap.Error(err))
			return err
		}
	}
	return nil
}

func updateClubGoodsPriceDelGoods(ctx context.Context, clubId string) error {
	clubInfo, err := QueryClubInfoByClubId(clubId)
	if err != nil {
		global.Lg.Error("updateCoachGoodsPrice 删除商品价格失败", zap.Error(err))
		return err
	}

	//查询教练的上下最小价格和最大价格
	minPrice, maxPrice, err := NewGoodsDao(ctx, global.DB).QueryPriceByUserId(ctx, clubId)
	if err != nil {
		global.Lg.Error("updateCoachGoodsPrice 删除商品价格失败", zap.Error(err))
		return err
	}

	clubInfo.PriceMin = minPrice
	clubInfo.PriceMax = maxPrice
	if err = global.DB.Model(&model.Clubs{}).Where("club_id", clubId).Save(clubInfo).Error; err != nil {
		global.Lg.Error("updateClubGoodsPriceDelGoods 删除商品价格失败", zap.Error(err))
		return err
	}

	return nil
}

func ApplyClub(c *gin.Context, req *forms.ApplyClubRequest) (*model.Clubs, error) {
	uid := c.GetString("uid")
	club, err := QueryClubInfoByUid(uid)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			global.Lg.Error("查询俱乐部详情失败", zap.Error(err))
			return nil, enum.NewErr(enum.ClubExitErr, "数据库查询错误")
		}
	}
	if club != nil && club.Verified == model.VerifiedVerified {
		return nil, enum.NewErr(enum.CoachHasVerifiedErr, "俱乐部已审核通过")
	}

	if club == nil {
		club = &model.Clubs{
			ClubId:       GenerateId("JLB"),
			ReferralCode: GenerateRandomString(16),
			Level:        1,
		}
	}
	club.Name = req.Name
	club.Logo = req.Logo
	club.Manager = req.Manager
	club.Phone = req.Phone
	club.SocialCreditCode = req.SocialCreditCode
	club.BusinessLicense = req.BusinessLicense
	club.IdCardFront = req.IdCardFront
	club.IdCardBack = req.IdCardBack
	club.Verified = model.VerifiedUnverified
	club.Uid = uid
	club.OpTime = time.Now()
	club.State = 0
	if err = global.DB.Model(club).Save(club).Error; err != nil {
		global.Lg.Error("创建俱乐部失败", zap.Error(err))
		return nil, err
	}
	return club, nil
}

func AddClubFinishedCourse(ctx context.Context, tx *gorm.DB, clubId string, cnt int) error {
	if err := tx.Model(&model.Clubs{}).Where("club_id = ? and state = 0", clubId).Update("finished_course", gorm.Expr("finished_course + ?", cnt)).Error; err != nil {
		global.Lg.Error("更新俱乐部完成课程失败", zap.Error(err))
		return err
	}
	return nil
}

func AddClubWroteCourseRecord(ctx context.Context, tx *gorm.DB, clubId string, cnt int) error {
	if err := tx.Model(&model.Clubs{}).Where("club_id", clubId).Update("wrote_course_record", gorm.Expr("wrote_course_record + ?", cnt)).Error; err != nil {
		global.Lg.Error("AddClubWroteCourseRecord 更新俱乐部状态失败", zap.Error(err), zap.Any("clubId", clubId))
		return err
	}
	return nil
}

func AddClubBalance(ctx context.Context, tx *gorm.DB, clubId string, amount int64) error {
	if err := tx.Model(&model.Clubs{}).Where("club_id = ? and state = 0", clubId).Updates(map[string]interface{}{
		"total_profit": gorm.Expr("total_profit + ?", amount),
		"balance":      gorm.Expr("balance + ?", amount),
	}).Error; err != nil {
		global.Lg.Error("AddClubBalance 更新俱乐部余额失败", zap.Error(err), zap.Any("clubId", clubId), zap.Any("amount", amount))
		return err
	}
	return nil
}

// UpdateClubsInfo 更新俱乐部信息
func UpdateClubsInfo(c *gin.Context, req *forms.UpdateClubsInfoRequest) error {
	// 获取当前用户的uid
	clubId := c.GetString("club_id")
	if clubId == "" {
		return enum.NewErr(enum.ParamErr, "俱乐部ID不能为空")
	}

	// 查询俱乐部信息
	var club model.Clubs
	err := global.DB.Where("club_id = ? AND state = 0", clubId).First(&club).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return enum.NewErr(enum.ClubExitErr, "俱乐部不存在")
		}
		global.Lg.Error("UpdateClubsInfo 查询俱乐部失败", zap.Error(err))
		return err
	}

	// 更新非空字段
	if req.Name != nil {
		club.Name = *req.Name
	}
	if req.ApprovalLogo != nil {
		club.ApprovalLogo = *req.ApprovalLogo
	}
	if req.Introduction != nil {
		club.Introduction = *req.Introduction
	}

	// 更新操作时间
	club.OpTime = time.Now()

	// 保存更新
	err = global.DB.Model(&model.Clubs{}).Where("club_id = ?", club.ClubId).Save(&club).Error
	if err != nil {
		global.Lg.Error("UpdateClubsInfo 更新俱乐部信息失败", zap.Error(err))
		return err
	}

	return nil
}

func QueryClubsCourses(ctx context.Context, clubId string) (courses []*model.Courses, err error) {
	//查询俱乐部的标签
	tags, err := QueryClubAllTags(clubId)
	if err != nil {
		global.Lg.Error("QueryClubAllTags 查询俱乐部标签失败", zap.Error(err))
		return nil, err
	}

	tagMapIds := make(map[int64]struct{}, 0)
	for _, v := range tags {
		tagMapIds[v.TagID] = struct{}{}
	}
	//查询俱乐部已经配置过的课程Id
	var courseIds []string
	err = global.DB.Model(&model.Goods{}).Where("user_id = ? and pack = 0 and state = 0", clubId).Pluck("course_id", &courseIds).Error
	if err != nil {
		global.Lg.Error("QueryClubsCourses 查询俱乐部已经配置过的课程Id失败", zap.Error(err))
		return nil, err
	}

	//查询满足条件的课程
	courseData := make([]*model.Courses, 0)
	db := global.DB.Model(&model.Courses{}).Preload("CoursesTags", "state = 0").Preload("CoursesTags.Tag").
		Where("state = 0 and on_shelf = 1 ")
	if len(courseIds) > 0 {
		db = db.Where("course_id not in ?", courseIds)
	}

	if err = db.Order("id desc").Find(&courseData).Error; err != nil {
		global.Lg.Error("QueryClubsCourses 查询满足条件的课程失败", zap.Error(err))
		return nil, err
	}

	for _, course := range courseData {
		dealCourseTags(course)
		isTrue := true
		for _, tag := range course.CoursesTags {
			if _, ok := tagMapIds[tag.TagID]; !ok {
				isTrue = false
				break
			}
		}
		if isTrue {
			courses = append(courses, course)
		}
	}
	return courses, nil
}

func QueryClubMatchCoachesList(c *gin.Context, req forms.QueryMatchCoachesListRequest) (coaches []*model.Coaches, err error) {
	clubId := c.GetString("club_id")
	clubCoaches, err := NewClubsCoachesDao(c, global.DB).ClubsCoachAll(c, clubId)
	if err != nil {
		return nil, err
	}
	orderCourse := model.OrdersCourses{}
	err = global.DB.Model(&model.OrdersCourses{}).Preload("CourseTags", "state = 0").
		Where("order_course_id = ?", req.OrderCourseId).First(&orderCourse).Error
	if err != nil {
		global.Lg.Error("QueryMatchCoachesList 查询订单课程失败", zap.Error(err))
		err = enum.NewErr(enum.OrdersCoursesExitErr, "订单课程不存在")
		return
	}

	coachMap := make(map[string]*model.Coaches)
	var coachIds, skiCoachIds []string // 俱乐部所有教练id，场地匹配的教练id
	for _, v := range clubCoaches {
		v.Coaches.Match = model.MatchStruct{
			SkiIsMatch:  false,
			TagIsMatch:  false,
			TimeIsMatch: false,
		}
		if len(v.Coaches.CoachesSkiResorts) > 0 {
			for _, ski := range v.Coaches.CoachesSkiResorts {
				if ski.SkiResortsID == int64(orderCourse.SkiResortsID) {
					v.Coaches.Match.SkiIsMatch = true
					skiCoachIds = append(skiCoachIds, v.CoachID)
				}
			}
		}
		coachMap[v.CoachID] = &v.Coaches
		coachIds = append(coachIds, v.CoachID)
	}

	var tagIds []int64
	for _, v := range orderCourse.CourseTags {
		tagIds = append(tagIds, v.TagID)
	}
	var tagCoachIds []string // 标签对应的教练id
	global.DB.Model(&model.CoachesTags{}).Select("coach_id").
		Where("coach_id in (?) and tag_id in (?) and verified=1 and state=0", coachIds, tagIds).
		Group("coach_id").Having("count(1)>=?", len(tagIds)).Scan(&tagCoachIds)
	for _, v := range tagCoachIds {
		if _, ok := coachMap[v]; ok {
			coachMap[v].Match.TagIsMatch = true
		}
	}

	if err != nil {
		global.Lg.Error("QueryMatchCoachesList 查询教练列表失败", zap.Error(err))
		err = enum.NewErr(enum.OrdersCoursesExitErr, "教练不存在")
		return
	}
	//查询教练时间
	teachTimeIDs := orderCourse.ClubTimeIDs
	var timeStarts []model.LocalTime
	err = global.DB.Model(&model.SkiResortsTeachTime{}).Where("id in (?)", []int64(teachTimeIDs)).Pluck("teach_start_time", &timeStarts).Error
	if err != nil {
		global.Lg.Error("QueryMatchCoachesList 查询教学时间失败", zap.Error(err))
		err = enum.NewErr(enum.OrdersCoursesExitErr, "课程时间不存在")
		return
	}

	var coachSkiResorts []struct {
		CoachId string `json:"coach_id"`
		Num     int64  `json:"num"`
	}
	err = global.DB.Model(&model.CoachesSkiResorts{}).Select("coach_id,count(1) as num").
		Where("coaches_ski_resorts.coach_id in ? and coaches_ski_resorts.ski_resorts_id=? and coaches_ski_resorts.state = 0", tagCoachIds, orderCourse.SkiResortsID).
		Joins("JOIN ski_resorts_teach_time on ski_resorts_teach_time.teach_start_time in ? "+
			"and ski_resorts_teach_time.user_id = coaches_ski_resorts.coach_id "+
			"and ski_resorts_teach_time.state = 0 and ski_resorts_teach_time.teach_state = 0 "+
			"and ski_resorts_teach_time.teach_num > 0 and ski_resorts_teach_time.ski_resorts_id = ?",
			timeStarts, orderCourse.SkiResortsID).Group("coaches_ski_resorts.coach_id").Having("num >= ?", len(timeStarts)).
		Find(&coachSkiResorts).Error

	for _, a := range coachSkiResorts {
		if _, ok := coachMap[a.CoachId]; ok {
			coachMap[a.CoachId].Match.TimeIsMatch = true
			coachMap[a.CoachId].Match.TagIsMatch = true
			coaches = append(coaches, coachMap[a.CoachId])
			delete(coachMap, a.CoachId)
		}
	}

	for _, co := range coachMap {
		coaches = append(coaches, co)
	}
	return
}
