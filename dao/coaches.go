package dao

import (
	"context"
	"errors"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func QueryCoachesList(c *gin.Context, req forms.QueryCoachesListRequest) (coaches []*model.Coaches, err error) {
	var coachIds []string
	if req.TagID != 0 {
		cIds, err := QueryCoachIdByTagId(req.TagID)
		if err != nil {
			global.Lg.Error("QueryClubList QueryClubIdByTagId 查询俱乐部ID失败", zap.Error(err))
			return []*model.Coaches{}, err
		}
		if len(cIds) == 0 {
			return []*model.Coaches{}, nil
		}
		coachIds = append(coachIds, cIds...)
	}

	if req.SkiResortsId != 0 {
		cIds, err := NewCoachSkiResortsDao(c, global.DB).QueryCoachIdBySkiId(req.SkiResortsId)
		if err != nil {
			global.Lg.Error("QueryClubList QueryClubIdBySkiId 查询俱乐部ID失败", zap.Error(err))
			return []*model.Coaches{}, err
		}
		if len(cIds) == 0 {
			return []*model.Coaches{}, nil
		}

		//取交集
		coachIds = StrIntersection(coachIds, cIds)
	}

	db := global.DB.Model(&model.Coaches{}).Scopes(model.ScopeCoachFields).Where("state = 0")
	if req.Keyword != "" { //todo: 添加搜索字段
		db = db.Where("(realname LIKE ? or coach_id LIKE ?)", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	//标签和场地ID不为空时，查询标签和场地ID对应的教练ID
	if len(coachIds) > 0 {
		db = db.Where("coach_id in (?)", coachIds)
	}
	if req.Verified == 0 {
		db = db.Where("verified in (0,2)")
	} else {
		db = db.Where("verified = 1")
	}
	err = db.Preload("Users", func(db *gorm.DB) *gorm.DB {
		return db.Scopes(model.ScopeUserSensitiveFields)
	}).Preload("CoachTags", "verified=1 and state=0").
		Preload("CoachTags.Tag", "state = 0").
		Preload("Certificates", "verified=1 and state=0").
		Preload("Certificates.CertificateConfig").
		Preload("CoachesSkiResorts", "state=0").
		Preload("CoachesSkiResorts.SkiResorts").
		Order("id desc").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find(&coaches).Error
	if err != nil {
		global.Lg.Error("查询教练列表失败", zap.Error(err))
		return
	}
	return
}

func QueryMatchCoachesList(c *gin.Context, req forms.QueryMatchCoachesListRequest) (coaches []*model.Coaches, err error) {
	orderCourse := model.OrdersCourses{}
	userId := c.GetString("user_id")
	err = global.DB.Model(&model.OrdersCourses{}).Preload("CourseTags", "state = 0").
		Where("order_course_id = ?", req.OrderCourseId).First(&orderCourse).Error
	if err != nil {
		global.Lg.Error("QueryMatchCoachesList 查询订单课程失败", zap.Error(err))
		err = enum.NewErr(enum.OrdersCoursesExitErr, "订单课程不存在")
		return
	}
	var tagIds []int64
	for _, v := range orderCourse.CourseTags {
		tagIds = append(tagIds, v.TagID)
	}
	var coachIds []string
	global.DB.Model(&model.CoachesTags{}).Select("coach_id").
		Where(" tag_id in (?) and verified=1 and state=0", tagIds).
		Group("coach_id").Having("count(1)>=?", len(tagIds)).Scan(&coachIds)
	for i := 0; i < len(coachIds); i++ {
		if coachIds[i] == userId {
			coachIds = append(coachIds[:i], coachIds[i+1:]...)
			i--
		}
	}

	err = global.DB.Model(&model.Coaches{}).Scopes(model.ScopeCoachFields).
		Preload("Users").
		Preload("CoachTags", "state = 0").
		Preload("CoachTags.Tag", "state = 0").
		Preload("CoachesSkiResorts", "state=0").Preload("CoachesSkiResorts.SkiResorts").
		Where("state = 0 and verified = 1 and coach_id in (?)", coachIds).
		Find(&coaches).Error

	if err != nil {
		global.Lg.Error("QueryMatchCoachesList 查询教练列表失败", zap.Error(err))
		err = enum.NewErr(enum.OrdersCoursesExitErr, "教练不存在")
		return
	}
	//查询教练时间
	teachTimeIDs := orderCourse.ClubTimeIDs
	if orderCourse.TeachTimeIDs != nil {
		teachTimeIDs = orderCourse.TeachTimeIDs
	}
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
		Where("coaches_ski_resorts.coach_id in ? and coaches_ski_resorts.state = 0", coachIds).
		Joins("JOIN ski_resorts_teach_time on ski_resorts_teach_time.teach_start_time in ? "+
			"and ski_resorts_teach_time.user_id = coaches_ski_resorts.coach_id "+
			"and ski_resorts_teach_time.state = 0 and ski_resorts_teach_time.teach_state = 0 "+
			"and ski_resorts_teach_time.teach_num > 0 and ski_resorts_teach_time.ski_resorts_id = ?",
			timeStarts, orderCourse.SkiResortsID).Group("coaches_ski_resorts.coach_id").Having("num >= ?", len(timeStarts)).
		Find(&coachSkiResorts).Error

	for _, coach := range coaches {
		for _, a := range coachSkiResorts {
			if a.CoachId == coach.CoachId {
				coach.TimeIsMatch = true
				coach.Match = model.MatchStruct{
					SkiIsMatch:  false,
					TagIsMatch:  false,
					TimeIsMatch: true,
				}
				break
			}
		}
	}
	return
}
func CheckCoachTag(coachId string, orderTagIds []int64) (coach *model.Coaches, err error) {
	err = global.DB.Model(model.Coaches{}).Preload("CoachTags", "tag_id in ? and state=0 and verified = ?", orderTagIds, model.VerifiedVerified).
		Where("coach_id = ? and state = 0", coachId).First(&coach).Error
	if err != nil {
		return nil, enum.NewErr(enum.OrdersCoursesExitErr, "教练不存在")
	}
	if coach.Verified != model.VerifiedVerified {
		return nil, enum.NewErr(enum.CoachVerifyErr, "教练未通过审核")
	}
	if len(coach.CoachTags) < len(orderTagIds) {
		return nil, enum.NewErr(enum.OrdersCoursesExitErr, "教练没有该课程标签")
	}
	return
}

func QueryCoachInfo(coachId string) (*model.Coaches, error) {
	var coach *model.Coaches
	err := global.DB.Model(model.Coaches{}).Scopes(model.ScopeCoachFields).
		Preload("Users", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(model.ScopeUserSensitiveFields)
		}).
		Preload("CoachTags", "verified=1 and state=0").
		Preload("CoachTags.Tag", "state = 0").
		Preload("Certificates", "verified=1 and state=0").
		Preload("Certificates.CertificateConfig", "state = 0").
		Preload("CoachesSkiResorts", "state=0").Preload("CoachesSkiResorts.SkiResorts").Preload("LevelInfo", "state=0").
		First(&coach, "coach_id = ? and state = 0", coachId).Error
	if err != nil {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		return nil, err
	}

	coach.ServiceRate = 15 //默认值
	if coach.LevelInfo != nil {
		coach.ServiceRate = coach.LevelInfo.ServiceRate
	}

	return coach, nil
}

func QueryLoginCoachInfo(coachId string) (*model.Coaches, error) {
	var coach *model.Coaches
	err := global.DB.Preload("Users").
		Preload("CoachTags", "verified=1 and state=0").
		Preload("CoachTags.Tag", "state = 0").
		Preload("Certificates", "verified=1 and state=0").
		Preload("Certificates.CertificateConfig", "state = 0").
		Preload("CoachesSkiResorts", "state=0").Preload("CoachesSkiResorts.SkiResorts").Preload("LevelInfo", "state=0").
		First(&coach, "coach_id = ? and state = 0", coachId).Error
	if err != nil {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		return nil, err
	}

	coach.ServiceRate = 15 //默认值
	if coach.LevelInfo != nil {
		coach.ServiceRate = coach.LevelInfo.ServiceRate
	}

	return coach, nil
}
func CoachInfoByCoachId(coachId string) (*model.Coaches, error) {
	var coach *model.Coaches
	err := global.DB.First(&coach, "coach_id = ? and state = 0", coachId).Error
	if err != nil {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		return nil, err
	}
	return coach, nil
}

func CoachInfoByUserId(uid string) (*model.Coaches, error) {
	var coach *model.Coaches
	err := global.DB.Model(model.Coaches{}).Preload("Certificates").First(&coach, "uid = ? and state = 0", uid).Error
	if err != nil {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		return nil, err
	}
	return coach, nil
}

func CoachRemoveTag(c *gin.Context, req forms.CoachRemoveTagRequest) error {
	uid := c.GetString("uid")

	//校验教练技能是否存在
	coachInfo, err := CoachInfoByUserId(uid)
	if err != nil {
		global.Lg.Error("CoachRemoveTag CoachInfoByUserId 查询教练信息失败", zap.Error(err))
		return enum.NewErr(enum.CoachNotExistErr, "教练不存在")
	}

	//校验教练技能是否被使用
	err = global.DB.Model(&model.CoachesTags{}).Where("coach_id = ? and tag_id = ? and state = 0", coachInfo.CoachId, req.TagID).
		First(&model.CoachesTags{}).Error
	if err != nil {
		global.Lg.Error("CoachRemoveTag CoachInfoByUserId 查询教练信息失败", zap.Error(err))
		return enum.NewErr(enum.CoachTagNoExistErr, "教练技能不存在")
	}
	var num int64
	err = global.DB.Model(&model.Goods{}).Where("goods.user_id = ? and goods.state = 0", uid).
		Joins("LEFT JOIN goods_courses ON goods.good_id = goods_courses.good_id and goods_courses.state = 0").
		Joins("LEFT JOIN course_tags ON goods_courses.course_id = course_tags.course_id and course_tags.state = 0 and course_tags.tag_id = ?", req.TagID).
		Count(&num).Error

	if err != nil {
		global.Lg.Error("校验教练技能失败", zap.Error(err))
		return err
	}
	if num > 0 {
		return enum.NewErr(enum.CoachCourseTagExistErr, "请先删除教练技能下的课程")
	}
	err = global.DB.Model(&model.CoachesTags{}).Where("coach_id = ? and tag_id = ? and state = 0", coachInfo.CoachId, req.TagID).
		Update("state", 1).Error
	if err != nil {
		global.Lg.Error("删除教练技能失败", zap.Error(err))
		return enum.NewErr(enum.CoachDelTagErr, "删除教练技能失败")
	}
	return nil
}

func QueryApplyInfo(c *gin.Context, uid string) (*model.Coaches, error) {
	var coach *model.Coaches
	err := global.DB.Model(model.Coaches{}).
		Preload("CoachTags", "state=0").
		Preload("CoachTags.Tag", "state = 0").
		Preload("Certificates", "state = 0").
		Preload("Certificates.CertificateConfig", "state = 0").Where("uid = ? and state = 0", uid).First(&coach).Error
	if err != nil {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		return nil, err
	}

	return coach, nil
}

func ApplyCoach(ctx context.Context, uid string, req *forms.CreateCoachRequest) (*model.Coaches, error) {
	coach, err := CoachInfoByUserId(uid)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		return nil, err
	}

	if coach != nil && coach.State == model.VerifiedVerified {
		global.Lg.Error("该用户通过教练认证")
		return nil, enum.NewErr(enum.CoachHasVerifiedErr, "已通过认证")
	}

	if coach == nil {
		coach = &model.Coaches{
			CoachId:      GenerateId("JL"),
			ReferralCode: GenerateRandomString(16),
		}

		var referralUserId string
		var referralUserType int
		//第一次申请，要检测推荐码是否存在
		if req.ReferralCode != "" {
			_, referralUserId, referralUserType, err = QueryUserIdReferralCode(ctx, req.ReferralCode)
			if err != nil {
				global.Lg.Error("GetOrCreateUser QueryUserIdReferralCode error", zap.Error(err), zap.String("referralCode", req.ReferralCode))
				return nil, err
			}
		}
		coach.ApplyReferralCode = req.ReferralCode
		coach.ReferralUserId = referralUserId
		coach.ReferralUserType = referralUserType
	}

	coach.Uid = uid
	coach.Realname = req.Realname
	coach.IdCard = req.IdCard
	coach.IdCardPhoto = req.IdCardPhoto
	coach.Phone = req.Phone
	coach.Introduction = req.Introduction
	coach.SupplementaryMaterials = req.SupplementaryMaterials
	coach.Verified = model.VerifiedUnverified //未认证
	coach.OpTime = time.Now()
	coach.Remark = "" //remark 清空
	coach.Level = 1   //默认是1级

	if err = global.DB.Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(coach).Save(coach).Error; err != nil {
			global.Lg.Error("创建教练失败", zap.Error(err))
			return err
		}
		var certificates []model.CoachesCertificates
		for _, tmp := range req.Certificates {
			certificate := model.CoachesCertificates{
				ID:                 tmp.Id,
				CoachID:            coach.CoachId,
				CertificateID:      tmp.CertificateId,
				CertificateImgUrls: tmp.CertificateImgUrls,
				Level:              tmp.Level,
			}
			certificates = append(certificates, certificate)
		}
		if err = tx.Model(model.CoachesCertificates{}).Save(&certificates).Error; err != nil {
			global.Lg.Error("创建教练证书失败", zap.Error(err))
			return err
		}

		return nil

	}); err != nil {
		global.Lg.Error("创建教练失败", zap.Error(err))
		return nil, err
	}

	return coach, nil

}

func AdminMatchCoachesList(c *gin.Context, orderId string) (coaches []*model.Coaches, err error) {
	order, err := QueryOrderInfo("", orderId)
	if err != nil {
		global.Lg.Error("QueryMatchCoachesList 查询订单失败", zap.Error(err))
		return nil, err
	}

	orderCourse := order.OrdersCourses[0]
	var tagIds []int64
	for _, v := range orderCourse.CourseTags {
		tagIds = append(tagIds, v.TagID)
	}
	err = global.DB.Model(&model.Coaches{}).Preload("CoachesTags", "state = 0").Preload("CoachesTags.Tag").
		Where("state = 0 and verified = 1 and coach_id in"+
			" (select coach_id from coaches_tags where tag_id in (?) and verified=1 group by coach_id having count(1)>=?)", tagIds, len(tagIds)).
		Find(&coaches).Error
	if err != nil {
		global.Lg.Error("QueryMatchCoachesList 查询教练列表失败", zap.Error(err))
		err = enum.NewErr(enum.OrdersCoursesExitErr, "教练不存在")
		return
	}
	var coachIds []string
	for _, v := range coaches {
		coachIds = append(coachIds, v.CoachId)
	}
	//查询教练时间
	teachTimeIDs := orderCourse.ClubTimeIDs
	if orderCourse.TeachTimeIDs != nil {
		teachTimeIDs = orderCourse.TeachTimeIDs
	}
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
		Where("coaches_ski_resorts.coach_id in ? and coaches_ski_resorts.state = 0 and coaches_ski_resorts.ski_resorts_id = ?",
			coachIds, orderCourse.SkiResortsID).
		Joins("JOIN ski_resorts_teach_time on ski_resorts_teach_time.teach_start_time in ? and ski_resorts_teach_time.user_id = coaches_ski_resorts.coach_id and ski_resorts_teach_time.state = 0 and ski_resorts_teach_time.teach_state = 0 and ski_resorts_teach_time.teach_num > 0",
			timeStarts).Group("coaches_ski_resorts.coach_id").Having("num >= ?", len(timeStarts)).
		Find(&coachSkiResorts).Error

	for _, coach := range coaches {
		for _, a := range coachSkiResorts {
			if a.CoachId == coach.CoachId {
				coach.TimeIsMatch = true
				coach.Match = model.MatchStruct{
					SkiIsMatch:  false,
					TagIsMatch:  false,
					TimeIsMatch: true,
				}
				break
			}
		}
	}
	return
}

func EditCoachInfo(c *gin.Context, req forms.EditCoachInfoRequest) (err error) {
	coachId := c.GetString("coach_id")
	if coachId == "" {
		return errors.New("请先申请成为教练")
	}
	upData := make(map[string]interface{}, 0)
	for _, v := range req.EditFields {
		switch v {
		case "introduction":
			upData[v] = req.Introduction
		}
	}
	if len(upData) == 0 {
		return errors.New("请传修改的字段")
	}
	err = global.DB.Model(&model.Coaches{}).Where("coach_id", coachId).Where("state", model.StateNormal).
		Updates(upData).Error
	if err != nil {
		global.Lg.Error("EditCoachInfo 更新教练信息失败", zap.Error(err), zap.Any("coachId", coachId), zap.Any("upData", upData), zap.Any("req", req))
	}
	return err
}

func QueryCoachCourses(ctx context.Context, coachId string) (courses []*model.Courses, err error) {
	//查询教练的标签
	tags, err := QueryCoachAllTags(coachId, 1)
	if err != nil {
		global.Lg.Error("QueryCoachCourses 查询教练标签失败", zap.Error(err))
		return nil, err
	}

	tagMapIds := make(map[int64]struct{}, 0)
	for _, v := range tags {
		tagMapIds[v.TagId] = struct{}{}
	}
	//查询教练已经配置过的课程Id
	var courseIds []string
	err = global.DB.Model(&model.Goods{}).Where("user_id = ? and pack = 0 and state = 0", coachId).Pluck("course_id", &courseIds).Error
	if err != nil {
		global.Lg.Error("QueryCoachCourses 查询教练已经配置过的课程Id失败", zap.Error(err))
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
		global.Lg.Error("QueryCoachCourses 查询满足条件的课程失败", zap.Error(err))
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

func updateCoachGoodsPriceAddGoods(ctx context.Context, coachId string, goodPrice int64) error {
	coachInfo, err := CoachInfoByCoachId(coachId)
	if err != nil {
		global.Lg.Error("updateCoachGoodsPrice 查询教练信息失败", zap.Error(err))
		return err
	}

	data := make(map[string]interface{})
	if goodPrice > coachInfo.PriceMax {
		data["price_max"] = goodPrice
	}

	if coachInfo.PriceMin == 0 || goodPrice < coachInfo.PriceMin {
		data["price_min"] = goodPrice
	}

	if len(data) > 0 {
		if err = global.DB.Model(&model.Coaches{}).Where("coach_id", coachId).Updates(data).Error; err != nil {
			global.Lg.Error("updateCoachGoodsPrice 添加商品价格失败", zap.Error(err))
			return err
		}
	}
	return nil
}

func updateCoachGoodsPriceDelGoods(ctx context.Context, coachId string) error {
	coachInfo, err := CoachInfoByCoachId(coachId)
	if err != nil {
		global.Lg.Error("updateCoachGoodsPrice 删除商品价格失败", zap.Error(err))
		return err
	}

	//查询教练的上下最小价格和最大价格
	minPrice, maxPrice, err := NewGoodsDao(ctx, global.DB).QueryPriceByUserId(ctx, coachId)
	if err != nil {
		global.Lg.Error("updateCoachGoodsPrice 删除商品价格失败", zap.Error(err))
		return err
	}

	coachInfo.PriceMin = minPrice
	coachInfo.PriceMax = maxPrice
	if err = global.DB.Model(&model.Coaches{}).Where("coach_id", coachId).Save(coachInfo).Error; err != nil {
		global.Lg.Error("updateCoachGoodsPrice 删除商品价格失败", zap.Error(err))
		return err
	}

	return nil
}

func AdminQueryMatchCoachesList(c *gin.Context, orderId string) (coaches []*model.Coaches, err error) {
	order, err := QueryOrderInfo("", orderId)
	if err != nil {
		global.Lg.Error("AdminQueryMatchCoachesList 查询订单信息失败", zap.Error(err))
		return nil, err
	}

	orderCourseId := order.OrdersCourses[0].OrderCourseID

	req := forms.QueryMatchCoachesListRequest{
		OrderCourseId: orderCourseId,
	}
	return QueryMatchCoachesList(c, req)
}

func AddCoachFinishedCourse(ctx context.Context, tx *gorm.DB, coachId string, cnt int) error {
	if err := tx.Model(&model.Coaches{}).Where("coach_id", coachId).Update("finished_course", gorm.Expr("finished_course + ?", cnt)).Error; err != nil {
		global.Lg.Error("AddFinishedCourse 更新教练状态失败", zap.Error(err), zap.Any("coachId", coachId))
		return err
	}
	return nil
}

func AddCoachWroteCourseRecord(ctx context.Context, tx *gorm.DB, coachId string, cnt int) error {
	if err := tx.Model(&model.Coaches{}).Where("coach_id", coachId).Update("wrote_course_record", gorm.Expr("wrote_course_record + ?", cnt)).Error; err != nil {
		global.Lg.Error("AddCoachWroteCourseRecord 更新教练状态失败", zap.Error(err), zap.Any("coachId", coachId))
		return err
	}
	return nil
}

func AddCoachTotalProfit(ctx context.Context, tx *gorm.DB, coachId string, profit int64) error {
	if err := tx.Model(&model.Coaches{}).Where("coach_id", coachId).Update("total_profit", gorm.Expr("total_profit + ?", profit)).Error; err != nil {
		global.Lg.Error("AddCoachTotalProfit 添加教练利润失败", zap.Error(err), zap.Any("coachId", coachId))
		return err
	}

	return nil
}

func AddCoachPaidServiceFee(ctx context.Context, tx *gorm.DB, coachId string, fee int64) error {
	if err := tx.Model(&model.Coaches{}).Where("coach_id", coachId).Update("paid_service_fee", gorm.Expr("paid_service_fee + ?", fee)).Error; err != nil {
		global.Lg.Error("AddCoachPaidServiceFee 添加教练服务费用失败", zap.Error(err), zap.Any("coachId", coachId))
		return err
	}

	return nil
}

func CoachInfoByCoachIdWithLevel(coachId string) (*model.Coaches, error) {
	var coach *model.Coaches
	err := global.DB.Model(&model.Coaches{}).Preload("LevelInfo", "state=0").First(&coach, "coach_id = ? and state = 0", coachId).Error
	if err != nil {
		global.Lg.Error("查询教练信息失败", zap.Error(err))
		return nil, err
	}

	coach.ServiceRate = enum.ServiceRatio
	if coach.LevelInfo != nil {
		coach.ServiceRate = coach.LevelInfo.ServiceRate
	}

	return coach, nil
}

func AddCoachBalance(ctx context.Context, tx *gorm.DB, coachId string, amount int64) error {
	if err := tx.Model(&model.Coaches{}).Where("coach_id = ? and state = 0", coachId).Updates(map[string]interface{}{
		"balance":      gorm.Expr("balance + ?", amount),
		"total_profit": gorm.Expr("total_profit + ?", amount)}).Error; err != nil {
		global.Lg.Error("AddCoachBalance 更新教练余额失败", zap.Error(err), zap.Any("coachId", coachId), zap.Any("amount", amount))
		return err
	}
	return nil
}
