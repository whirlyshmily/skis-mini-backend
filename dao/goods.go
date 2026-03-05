package dao

import (
	"context"
	"fmt"
	"math/rand"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GoodsDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.Goods
}

func NewGoodsDao(ctx context.Context, dbs ...*gorm.DB) *GoodsDao {
	dao := new(GoodsDao)
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

func (d *GoodsDao) CreateGoods(c *gin.Context, req forms.SaveGoodsRequest) (*model.Goods, error) {
	//参数校验
	var course *model.Courses
	var err error
	if req.Pack == model.PackNo {
		course, err = QueryCourseInfo(req.CourseId)
		if err != nil {
			global.Lg.Error("查询课程信息失败", zap.Error(err))
			return nil, err
		}
	}

	good, packGoods, err := d.CheckGoods(c, req)
	if err != nil {
		return nil, err
	}
	good.GoodID = GenerateId("DCK")
	if req.Pack == model.PackNo {
		good.GoodID = GenerateId("DBK")
		good.CourseID = req.CourseId
	}

	err = d.sourceDB.Model(d.m).Create(good).Error
	if err != nil {
		return nil, enum.NewErr(enum.GoodsAddErr, "添加商品失败")
	}

	if req.Pack == model.PackYes {
		var goodCourses []*model.GoodsCourses
		for _, pGood := range packGoods {
			goodCourses = append(goodCourses, &model.GoodsCourses{
				GoodID:     good.GoodID,
				PackGoodID: pGood.GoodID,
				CourseID:   pGood.CourseID,
			})
		}
		err = NewGoodsCoursesDao(c, d.sourceDB).SaveAll(c, goodCourses)
		if err != nil {
			return nil, enum.NewErr(enum.GoodsCoursesAddErr, "添加打包课程失败")
		}
	}

	//修改课程的统计信息
	if req.Pack == model.PackNo && course != nil { //如果是单次课，要更新课程的最小价格和最大价格
		if err = d.addCourseRef(c, course, good, true); err != nil {
			return nil, err
		}
	}

	//还要更新教练或者俱乐部的商品最小价格和最大价格
	userType := c.GetInt("user_type")
	userId := c.GetString("user_id")
	if userType == model.UserTypeCoach {
		if err = updateCoachGoodsPriceAddGoods(c, userId, good.DiscountMoney); err != nil {
			global.Lg.Error("updateCoachGoodsPriceAddGoods failed", zap.Error(err))
			return nil, err
		}
	} else {
		if err = updateClubGoodsPriceAddGoods(c, userId, good.DiscountMoney); err != nil {
			global.Lg.Error("updateClubGoodsPriceAddGoods failed", zap.Error(err))
			return nil, err
		}
	}
	return good, nil
}

func (d *GoodsDao) EditGoods(c *gin.Context, req forms.SaveGoodsRequest) (*model.Goods, error) {
	if req.GoodID == "" {
		return nil, enum.NewErr(enum.ParamErr, "商品ID不能为空")
	}

	good, _, err := d.CheckGoods(c, req)
	if err != nil {
		return nil, err
	}

	if err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Where("good_id = ? and user_id = ? ", req.GoodID, good.UserID).First(&model.Goods{}).Error; err != nil {
		return nil, enum.NewErr(enum.GoodsNoExistErr, "商品不存在")
	}
	if good.FaultMoney == 0 {
		d.sourceDB.Model(d.m).Where("good_id = ?  and user_id = ? and user_type = ? ", req.GoodID, good.UserID, good.UserType).
			Updates(map[string]interface{}{
				"fault_money": 0,
			})
	}

	err = d.sourceDB.Model(d.m).Where("good_id = ?  and user_id = ? and user_type = ? ", req.GoodID, good.UserID, good.UserType).Updates(good).Error
	if err != nil {
		return nil, enum.NewErr(enum.GoodsEditErr, "修改商品失败")
	}

	// 如果是单次课，需要更新包含该单次课的所有打包课信息
	if good.Pack == model.PackNo {
		err = d.updatePackCoursesContainingGood(c, req.GoodID, good.UserID, good.UserType)
		if err != nil {
			global.Lg.Error("更新包含该单次课的打包课失败", zap.Error(err))
		}
	}

	userType := c.GetInt("user_type")
	userId := c.GetString("user_id")
	if userType == model.UserTypeCoach {
		if err = updateCoachGoodsPriceDelGoods(c, userId); err != nil {
			global.Lg.Error("updateCoachGoodsPriceDelGoods failed", zap.Error(err))
			return nil, err
		}
	} else {
		if err = updateClubGoodsPriceDelGoods(c, userId); err != nil {
			global.Lg.Error("updateClubGoodsPriceDelGoods failed", zap.Error(err))
			return nil, err
		}
	}
	return good, nil
}

// updatePackCoursesContainingGood 更新包含指定单次课的所有打包课信息
func (d *GoodsDao) updatePackCoursesContainingGood(c *gin.Context, goodID, userID string, userType int) error {
	// 查询包含该单次课的所有打包课
	var packGoods []struct {
		GoodID   string `json:"good_id"`
		Discount int    `json:"discount"`
	}
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&model.GoodsCourses{}).
		Joins("JOIN goods ON goods.good_id = goods_courses.good_id").
		Where("goods_courses.pack_good_id = ? AND goods_courses.state = 0 AND goods.pack = ? AND goods.state = 0 AND goods.user_id = ? AND goods.user_type = ?",
			goodID, model.PackYes, userID, userType).
		Select("DISTINCT goods.good_id, discount").
		Find(&packGoods).Error

	if err != nil {
		return err
	}

	if len(packGoods) == 0 {
		return nil // 没有找到包含该单次课的打包课
	}

	// 获取所有相关的打包课ID
	var packGoodIDs []string
	for _, pg := range packGoods {
		packGoodIDs = append(packGoodIDs, pg.GoodID)
	}

	// 重新计算每个打包课的信息
	for _, packGood := range packGoods {
		// 查询打包课包含的所有单次课
		var packCourses []*model.GoodsCourses
		err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&model.GoodsCourses{}).
			Where("good_id = ? AND state = 0", packGood.GoodID).
			Find(&packCourses).Error

		if err != nil {
			global.Lg.Error("查询打包课包含的单次课失败", zap.String("打包课ID", packGood.GoodID), zap.Error(err))
			continue
		}

		// 计算新的总价格和总时长
		var totalTeachTime int
		var totalTeachMoney, totalAreaMoney, totalClubMoney int64

		var packGoodIDs []string
		for _, pc := range packCourses {
			packGoodIDs = append(packGoodIDs, pc.PackGoodID)
		}

		if len(packGoodIDs) > 0 {
			// 查询所有单次课的详细信息
			var singleGoods []*model.Goods
			err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
				Where("good_id IN ? AND state = 0", packGoodIDs).
				Find(&singleGoods).Error

			if err != nil {
				global.Lg.Error("查询单次课详情失败", zap.Strings("单次课ID", packGoodIDs), zap.Error(err))
				continue
			}

			// 计算总和
			for _, sg := range singleGoods {
				totalTeachTime += sg.TeachTime
				totalTeachMoney += sg.TeachMoney
				totalAreaMoney += sg.AreaMoney
				totalClubMoney += sg.ClubMoney
			}
		}
		totalMoney := totalTeachMoney + totalAreaMoney + totalClubMoney

		// 更新打包课信息
		updateData := map[string]interface{}{
			"teach_time":     totalTeachTime,
			"teach_money":    totalTeachMoney,
			"area_money":     totalAreaMoney,
			"club_money":     totalClubMoney,
			"total_money":    totalMoney,
			"discount_money": totalMoney * int64(packGood.Discount) / 100, // Assuming default 100% discount
		}

		err = d.sourceDB.Model(d.m).
			Where("good_id = ? AND user_id = ? AND user_type = ?", packGood.GoodID, userID, userType).
			Updates(updateData).Error

		if err != nil {
			global.Lg.Error("更新打包课信息失败", zap.String("打包课ID", packGood.GoodID), zap.Error(err))
			continue
		}
	}

	global.Lg.Info("更新包含单次课的打包课", zap.String("单次课ID", goodID), zap.Strings("打包课ID", packGoodIDs))

	return nil
}

func (d *GoodsDao) CheckGoods(c *gin.Context, req forms.SaveGoodsRequest) (*model.Goods, []*model.Goods, error) {
	userType := c.GetInt("user_type")
	userId := c.GetString("user_id")
	if userId == "" {
		return nil, nil, enum.NewErr(enum.RoleNotExistErr, "用户不存在")
	}
	if userType == enum.UserTypeCoach {
		coach, err := CoachInfoByCoachId(userId)
		if err != nil || coach == nil {
			return nil, nil, enum.NewErr(enum.CoachNotExistErr, "教练不存在")
		}
		if coach.Deposit < req.AreaMoney+req.TeachMoney {
			return nil, nil, enum.NewErr(enum.CoachDepositLackErr, "课程总价不可超保证金请修改教学、场地费用")
		}
	} else if userType == enum.UserTypeClub {
		clubInfo, err := QueryClubInfoByClubId(userId)
		if err != nil {
			return nil, nil, enum.NewErr(enum.ClubExitErr, "俱乐部不存在")
		}
		if clubInfo.Deposit < req.AreaMoney+req.TeachMoney {
			return nil, nil, enum.NewErr(enum.ClubDepositLackErr, "课程总价不可超保证金请修改教学、场地费用")
		}
	} else {
		return nil, nil, enum.NewErr(enum.RoleNotExistErr, "用户没有权限")
	}
	var (
		detail = ""
	)
	var pointsDeduct int
	var packGoods []*model.Goods
	totalMoney := req.TeachMoney + req.AreaMoney + req.ClubMoney
	if req.Pack == model.PackNo { //单次课程
		var num int64
		d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
			Where("course_id = ?  and good_id != ? and user_id = ? and user_type = ? and pack = ? and state = 0", req.CourseId, req.GoodID, userId, userType, model.PackNo).
			Count(&num)
		if num > 0 {
			return nil, nil, enum.NewErr(enum.GoodsPackErr, "单次课程已存在")
		}
		course, err := QueryCourseInfo(req.CourseId)
		if err != nil {
			return nil, nil, enum.NewErr(enum.CourseExitErr, "课程不存在")
		}

		if req.TeachTime < 60 || req.TeachTime > 480 {
			return nil, nil, enum.NewErr(enum.GoodsTeachTimeErr, "单次课程时长必须在60-480分钟")
		}
		if req.FaultMoney > req.AreaMoney {
			return nil, nil, enum.NewErr(enum.GoodsFaultMoneyErr, "单次课程场地费用必须在1元-场地费用")
		}

		req.Discount = 100
		req.Title = course.Title
		req.CoverUrl = course.CoverUrl
		detail = course.Detail
		pointsDeduct = course.PointsDeduct
	} else {
		arrGoodIds := "'" + strings.Join(req.GoodIds, "','") + "'"
		d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
			Where("good_id in ? and user_id = ? and user_type = ? and pack = ? and on_shelf = 1 and state = 0", req.GoodIds, userId, userType, model.PackNo).
			Order(fmt.Sprintf("field(good_id,%s) asc", arrGoodIds)).
			Find(&packGoods)
		if len(packGoods) != len(req.GoodIds) {
			return nil, nil, enum.NewErr(enum.GoodsPackErr, "打包课程包含非自己发布或已下架的课程")
		}

		if len(packGoods) <= 1 {
			return nil, nil, enum.NewErr(enum.ParamErr, "打包课程至少包含两个课程")
		}
		if req.Title == "" || req.CoverUrl == "" {
			return nil, nil, enum.NewErr(enum.ParamErr, "打包课程必须包含标题和封面")
		}

		var countGoodsCourses []struct {
			GoodID string `json:"good_id"`
			Num    int    `json:"num"`
		}
		d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&model.GoodsCourses{}).
			Select("good_id, count(*) as num").
			Where("pack_good_id in ? and good_id != ?", req.GoodIds, req.GoodID).
			Group("good_id").Having("num = ?", len(req.GoodIds)).
			Find(&countGoodsCourses)

		if len(countGoodsCourses) != 0 {
			var goodsIds []string
			for _, v := range countGoodsCourses {
				goodsIds = append(goodsIds, v.GoodID)
			}
			d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&model.GoodsCourses{}).
				Select("good_id, count(*) as num").
				Where("good_id in ?", goodsIds).
				Group("good_id").Having("num = ?", len(req.GoodIds)).
				Find(&countGoodsCourses)
			if len(countGoodsCourses) != 0 {
				return nil, nil, enum.NewErr(enum.GoodsPackErr, "您已配置类似课程，请重新选择")
			}
		}

		detail = packGoods[0].Detail
		teachTime := 0
		var teachMoney, areaMoney, referralMoney int64
		for _, pGood := range packGoods {
			teachTime += pGood.TeachTime
			teachMoney += pGood.TeachMoney
			areaMoney += pGood.AreaMoney
			referralMoney += pGood.ClubMoney
		}
		totalMoney = teachMoney + areaMoney + referralMoney
		req.TeachTime = teachTime
		req.TeachMoney = teachMoney
		req.AreaMoney = areaMoney
		req.ClubMoney = referralMoney
		req.FaultMoney = 0 //打包课程没有有责取
	}
	good := &model.Goods{
		UserID:        userId,
		UserType:      userType,
		Title:         req.Title,
		CoverUrl:      req.CoverUrl,
		PointsDeduct:  pointsDeduct,
		TeachTime:     req.TeachTime,
		TeachMoney:    req.TeachMoney,
		ClubMoney:     req.ClubMoney,
		AreaMoney:     req.AreaMoney,
		FaultMoney:    req.FaultMoney,
		TotalMoney:    totalMoney,
		DiscountMoney: totalMoney * int64(req.Discount) / 100,
		Pack:          req.Pack,
		Discount:      req.Discount,
		Detail:        detail,
		OnShelf:       model.OnShelfYes,
	}
	return good, packGoods, nil
}

// BeforeEditGoods 编辑商品前检查
func (d *GoodsDao) BeforeEditGoods(c *gin.Context, goodID string) (*forms.BeforeEditGoodsResp, error) {
	if goodID == "" {
		return nil, enum.NewErr(enum.ParamErr, "商品ID不能为空")
	}

	// 查询商品信息
	var good model.Goods
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Where("good_id = ? and state = 0", goodID).First(&good).Error
	if err != nil {
		return nil, enum.NewErr(enum.GoodsNotExistErr, "商品不存在")
	}

	response := &forms.BeforeEditGoodsResp{}
	orderCourse, err := NewOrdersCoursesDao(c, d.sourceDB).ExitOrdersCoursesByGoodId(c, goodID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if orderCourse != nil && orderCourse.ID > 0 {
		response.HasUnfinishedOrder = true
		response.Message = "商品存在未完成订单，无法进行此操作"
	} else {
		response.HasUnfinishedOrder = false
	}

	// 如果是单次课，检查是否在打包课中
	if good.Pack == model.PackNo {
		isInPackCourse, err := d.checkInPackCourse(c, goodID)
		if err != nil {
			return nil, err
		}
		response.IsInPackCourse = isInPackCourse
	} else {
		response.IsInPackCourse = false
	}

	// 设置提示信息
	if response.HasUnfinishedOrder {
		response.Message = "商品存在未完成订单，无法进行此操作"
	} else if response.IsInPackCourse {
		response.Message = "单次课已被包含在打包课中，无法进行此操作"
	} else {
		response.Message = "商品可以正常操作"
	}
	return response, nil
}

// checkInPackCourse 检查单次课是否在打包课中
func (d *GoodsDao) checkInPackCourse(c *gin.Context, goodID string) (bool, error) {
	var count int64
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&model.GoodsCourses{}).
		Where("pack_good_id = ? and state = 0", goodID).
		Count(&count).Error
	if err != nil {
		global.Lg.Error("检查打包课关联失败", zap.Error(err))
		return false, enum.NewErr(enum.DBErr, "检查打包课关联失败")
	}

	return count > 0, nil
}

// findConflictPackCourses 查找商品冲突的打包课
func (d *GoodsDao) findConflictPackCourses(c *gin.Context, goodID string) ([]string, error) {
	// 查询商品信息，确认是单次课
	var good model.Goods
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Where("good_id = ? and state = 0", goodID).First(&good).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, enum.NewErr(enum.GoodsNotExistErr, fmt.Sprintf("商品 %s 不存在", goodID))
		}
		global.Lg.Error("查询商品信息失败", zap.Error(err))
		return nil, enum.NewErr(enum.DBErr, "查询商品信息失败")
	}

	// 如果是打包课本身，不需要检查冲突
	if good.Pack == model.PackYes {
		return nil, nil
	}

	// 查询该单次课是否已存在于其他打包课中
	var conflictPackIDs []string
	err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&model.GoodsCourses{}).
		Joins("JOIN goods ON goods.good_id = goods_courses.good_id").
		Where("goods_courses.pack_good_id = ? AND goods_courses.state = 0 AND goods.pack = ? AND goods.state = 0 AND goods.good_id != ?",
			goodID, model.PackYes, goodID).
		Pluck("DISTINCT goods_courses.good_id", &conflictPackIDs).Error

	if err != nil {
		global.Lg.Error("查询冲突打包课失败", zap.Error(err))
		return nil, enum.NewErr(enum.DBErr, "查询冲突打包课失败")
	}

	return conflictPackIDs, nil
}

func (d *GoodsDao) TakeDownGoods(c *gin.Context, req forms.TakeDownGoodsRequest) error {
	userType := c.GetInt("user_type")
	userId := c.GetString("user_id")

	good, err := d.GetByCoachCourseId(c, userId, req.GoodID)
	if err != nil {
		return err
	}
	if err = d.sourceDB.Model(d.m).Where("good_id = ? and user_id = ? and user_type = ? ", req.GoodID, userId, userType).Update("on_shelf", 0).Error; err != nil {
		return enum.NewErr(enum.GoodsTakeDownErr, "下架失败")
	}

	// 如果是单次课，下架包含该单次课的所有打包课
	if good.Pack == model.PackNo {
		err = d.takeDownPackCoursesContainingGood(c, req.GoodID, userId, userType)
		if err != nil {
			global.Lg.Error("下架包含该单次课的打包课失败", zap.Error(err))
			// 这里不返回错误，因为主商品下架已经成功
		}
	}

	if userType == model.UserTypeCoach {
		if err := updateCoachGoodsPriceDelGoods(c, userId); err != nil {
			global.Lg.Error("updateCoachGoodsPriceDelGoods failed", zap.Error(err))
			return err
		}
	} else {
		if err := updateClubGoodsPriceDelGoods(c, userId); err != nil {
			global.Lg.Error("updateClubGoodsPriceDelGoods failed", zap.Error(err))
			return err
		}
	}

	return nil
}

// takeDownPackCoursesContainingGood 下架包含指定单次课的所有打包课
func (d *GoodsDao) takeDownPackCoursesContainingGood(c *gin.Context, goodID, userID string, userType int) error {
	// 查询包含该单次课的所有打包课
	var packGoodIDs []string
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&model.GoodsCourses{}).
		Joins("JOIN goods ON goods.good_id = goods_courses.good_id").
		Where("goods_courses.pack_good_id = ? AND goods_courses.state = 0 AND goods.pack = ? AND goods.state = 0 AND goods.user_id = ? AND goods.user_type = ?",
			goodID, model.PackYes, userID, userType).
		Pluck("DISTINCT goods_courses.good_id", &packGoodIDs).Error

	if err != nil {
		return err
	}

	if len(packGoodIDs) == 0 {
		return nil // 没有找到包含该单次课的打包课
	}

	// 下架所有相关的打包课
	err = d.sourceDB.Model(d.m).
		Where("good_id IN ? and user_id = ? and user_type = ? and state = 0", packGoodIDs, userID, userType).
		Update("on_shelf", 0).Error

	if err != nil {
		return err
	}

	global.Lg.Info("下架包含单次课的打包课", zap.String("单次课ID", goodID), zap.Strings("打包课ID", packGoodIDs))

	return nil
}

func (d *GoodsDao) PutUpGoods(c *gin.Context, req forms.PutUpGoodsRequest) error {
	userType := enum.UserTypeCoach
	userId := c.GetString("coach_id")
	if userId == "" {
		userId = c.GetString("club_id")
		if userId == "" {
			return enum.NewErr(enum.RoleNotExistErr, "用户没有权限")
		}
		userType = enum.UserTypeClub
	}

	good := &model.Goods{}
	if err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Where("good_id = ?  and state = 0", req.GoodID).First(&good).Error; err != nil {
		global.Lg.Error("查询商品失败", zap.Error(err))
		return err
	}

	if good.OnShelf == model.OnShelfAdminNo {
		global.Lg.Info("商品已由管理台下架，无法上架", zap.String("商品ID", req.GoodID))
		return enum.NewErr(enum.GoodsPutUpErr, "商品已由管理台下架，无法上架")
	}

	onShelf := NewGoodsCoursesDao(c, d.sourceDB).GetCourseOnShelf(c, good, 0)
	if onShelf > 0 {
		return enum.NewErr(enum.GoodsPutUpErr, "商品包含下架课程，无法上架")
	}

	courseTagIds, err := NewGoodsCoursesDao(c, d.sourceDB).GetCourseTags(c, good)
	if err != nil {
		return enum.NewErr(enum.GoodsPutUpErr, "获取课程标签失败")
	}
	if len(courseTagIds) > 0 {
		if userType == enum.UserTypeCoach { //教练
			coachTagIds, err := QueryCoachIdByTagIds(courseTagIds, userId)
			fmt.Println("coachTagIds", coachTagIds)
			if err != nil || len(coachTagIds) != len(courseTagIds) {
				return enum.NewErr(enum.GoodsPutUpErr, "教练标签与课程标签不匹配")
			}
		} else if userType == enum.UserTypeClub { //俱乐部
			clubTagIds, err := QueryClubIdByTagIds(courseTagIds, userId)
			if err != nil || len(clubTagIds) != len(courseTagIds) {
				return enum.NewErr(enum.GoodsPutUpErr, "俱乐部标签与课程标签不匹配")
			}
		}
	}

	err = d.sourceDB.Model(d.m).Where("good_id = ? and user_id = ? and user_type = ? ", req.GoodID, userId, userType).Update("on_shelf", 1).Error
	if err != nil {
		return enum.NewErr(enum.GoodsPutUpErr, "上架失败")
	}

	if userType == model.UserTypeCoach {
		if err = updateCoachGoodsPriceDelGoods(c, userId); err != nil {
			global.Lg.Error("updateCoachGoodsPriceDelGoods failed", zap.Error(err))
			return err
		}
	} else {
		if err = updateClubGoodsPriceDelGoods(c, userId); err != nil {
			global.Lg.Error("updateClubGoodsPriceDelGoods failed", zap.Error(err))
			return err
		}
	}

	return nil
}

func (d *GoodsDao) DeleteGoods(c *gin.Context, req forms.DeleteGoodsRequest) error {
	if req.GoodID == "" {
		return enum.NewErr(enum.ParamErr, "商品ID不能为空")
	}
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")

	good, err := d.QueryGoodInfo(req.GoodID)
	if err != nil {
		return err
	}

	if good.UserID != userId {
		return enum.NewErr(enum.GoodsDeleteErr, "只能删除自己发布的商品")
	}
	orderCourse, err := NewOrdersCoursesDao(c, d.sourceDB).ExitOrdersCoursesByGoodId(c, req.GoodID)
	if err != nil {
		return err
	}
	if orderCourse != nil && orderCourse.ID > 0 {
		return enum.NewErr(enum.GoodsDeleteErr, "课程有未核销订单，建议先下架，完成后再删除")
	}

	if good.Pack == model.PackNo {
		gc := &model.GoodsCourses{}
		global.DB.Model(model.GoodsCourses{}).
			Where("pack_good_id = ? and state = 0", req.GoodID).
			First(gc)
		if gc.ID > 0 {
			return enum.NewErr(enum.GoodsDeleteErr, "此课程关联打包课程，请先删除、修改关联打包课程在删除")
		}
	}

	err = d.sourceDB.Model(d.m).Where("good_id = ? and user_id = ? and user_type = ? ", req.GoodID, userId, userType).Update("state", 1).Error
	if err != nil {
		return enum.NewErr(enum.GoodsDeleteErr, "删除失败")
	}

	if good.Pack == model.PackNo {
		course, err := QueryCourseInfo(good.CourseID)
		if err != nil {
			return err
		}
		return d.subCourseRef(c, course, good)
	}

	if userType == model.UserTypeCoach {
		if err = updateCoachGoodsPriceDelGoods(c, userId); err != nil {
			global.Lg.Error("updateCoachGoodsPriceDelGoods failed", zap.Error(err))
			return err
		}
	} else {
		if err = updateClubGoodsPriceDelGoods(c, userId); err != nil {
			global.Lg.Error("updateClubGoodsPriceDelGoods failed", zap.Error(err))
			return err
		}
	}

	return nil
}

func (d *GoodsDao) Create(ctx context.Context, obj *model.Goods) error {
	err := d.sourceDB.Model(d.m).Create(&obj).Error
	if err != nil {
		return fmt.Errorf("GoodsDao: %w", err)
	}
	return nil
}
func (d *GoodsDao) QueryGoodInfo(goodId string) (*model.Goods, error) {
	good := &model.Goods{}
	if err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Preload("CourseTags", "state = 0").
		Preload("CourseTags.Tag", "state = 0").
		Preload("GoodsCourses", "state = 0", func(db *gorm.DB) *gorm.DB {
			return db.Order("goods_courses.id asc")
		}).
		Preload("GoodsCourses.PackGood", "state = 0").
		Preload("GoodsCourses.PackGood.CourseTags", "state = 0").
		Preload("GoodsCourses.PackGood.CourseTags.Tag", "state = 0").
		Where("good_id = ?  and state = 0", goodId).First(&good).Error; err != nil {
		global.Lg.Error("查询商品失败", zap.Error(err))
		return nil, err
	}

	dealGoodTags(good)

	return good, nil
}

func (d *GoodsDao) QueryGoodBeforeBuy(uid string, good *model.Goods) (resp model.QueryGoodBeforeBuyResp, err error) {
	resp = model.QueryGoodBeforeBuyResp{
		PayFee: good.DiscountMoney,
	}
	if good.PointsDeduct == model.PointsDeductNo || good.Pack == model.PackYes { //积分抵扣和套餐不返回积分抵扣
		return resp, nil
	}
	user, err := QueryUserInfo(uid)
	if err != nil {
		return resp, err
	}
	if user.LeftPoints >= good.DiscountMoney {
		resp.PointsFee = good.DiscountMoney
		resp.PayFee = 0
	} else {
		resp.PointsFee = user.LeftPoints
		resp.PayFee = good.DiscountMoney - user.LeftPoints
	}
	return resp, nil
}

// GetMaxPriceByUserId GetMaxPrice 获取用户上架中最贵的商品
func (d *GoodsDao) GetMaxPriceByUserId(userId string) (*model.Goods, error) {
	good := &model.Goods{}
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Where("user_id = ? and state = 0 and on_shelf = 1", userId).Order("(teach_money+area_money) desc").First(&good).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		global.Lg.Error("查询商品失败", zap.Error(err))
		return nil, err
	}
	return good, nil
}

func (d *GoodsDao) ListByUserId(ctx context.Context, req forms.QueryGoodsListRequest) ([]*model.Goods, error) {
	var results []*model.Goods
	db := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Preload("CourseTags", "state = 0").
		Preload("CourseTags.Tag", "state = 0").
		Preload("GoodsCourses", "state = 0", func(db *gorm.DB) *gorm.DB {
			return db.Order("goods_courses.id asc")
		}).
		Preload("GoodsCourses.PackGood", "state = 0").
		Preload("GoodsCourses.PackGood.CourseTags", "state = 0").
		Preload("GoodsCourses.PackGood.CourseTags.Tag", "state = 0").
		Select("*").Where("user_id = ? ", req.UserId).Where("state = ?", 0)

	if req.Pack != nil && *req.Pack != -1 {
		db = db.Where("pack = ?", req.Pack)
	}

	//如果选择了标签
	var courseIds []string
	var err error
	if req.TagId != 0 {
		//查询课程标
		courseIds, err = QueryCoursesByTagId(req.TagId)
		if err != nil {
			global.Lg.Error("查询课程失败", zap.Error(err))
			return nil, err
		}

		//选择了标签，但是没有数据，就不用往下走了
		if len(courseIds) == 0 {
			return []*model.Goods{}, nil
		}

		if req.Pack != nil && *req.Pack == 0 { //查询单次课，就不用查询包课
			db = db.Where("course_id in ?", courseIds)
		} else {
			var packGoodIds []string
			err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Where("user_id = ? and state = 0 and course_id in ?", req.UserId, courseIds).Pluck("good_id", &packGoodIds).Error
			if err != nil {
				global.Lg.Error("查询商品失败", zap.Error(err))
				return nil, err
			}

			//查询打包课的id
			var goodIds []string
			err = global.DB.Model(&model.GoodsCourses{}).Where("pack_good_id in ? and state = 0", packGoodIds).Pluck("good_id", &goodIds).Error
			if err != nil {
				global.Lg.Error("查询商品失败", zap.Error(err))
				return nil, err
			}

			if len(goodIds) > 0 {
				db = db.Where("good_id in ? or course_id in ?", goodIds, courseIds)
			} else {
				db = db.Where("course_id in ?", courseIds)
			}
		}
	}

	if req.OnShelf != nil {
		db = db.Where("on_shelf = ?", req.OnShelf)
	}

	db = db.Order("on_shelf desc").Order("id desc")

	if req.Page != 0 && req.PageSize != 0 {
		db = db.Offset(req.PageSize * (req.Page - 1)).Limit(req.PageSize)
	}

	err = db.Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("GoodsDao: ListByUserId where=%s: %w", req.UserId, err)
	}

	for _, v := range results {
		dealGoodTags(v)
	}

	return results, nil
}

func dealGoodTags(good *model.Goods) {
	tagMap := make(map[int]*model.Tags)
	if good.Pack == model.PackNo {
		for _, v := range good.CourseTags {
			tagMap[v.Tag.Id] = v.Tag
		}
	} else {
		for _, packGood := range good.GoodsCourses {
			var courseTags []*model.Tags
			if packGood.PackGood != nil {
				for _, v := range packGood.PackGood.CourseTags {
					courseTags = append(courseTags, v.Tag)
					tagMap[v.Tag.Id] = v.Tag
				}
				packGood.PackGood.Tags = courseTags
			}
		}
	}

	tags := make([]*model.Tags, 0)
	for _, v := range tagMap {
		tags = append(tags, v)
	}
	good.Tags = tags
}

func (d *GoodsDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("GoodsDao:Update where=%s: %w", where, err)
	}
	return nil
}

func (d *GoodsDao) addCourseRef(ctx context.Context, course *model.Courses, good *model.Goods, isCreate bool) error {
	//修改课程的统计信息
	data := make(map[string]interface{})
	if good.DiscountMoney > course.PriceMax {
		data["price_max"] = good.DiscountMoney
	}

	if course.PriceMin == 0 || course.PriceMin > good.DiscountMoney {
		data["price_min"] = good.DiscountMoney
	}

	if isCreate {
		switch good.UserType {
		case enum.UserTypeCoach:
			data["ref_coach_cnt"] = gorm.Expr("ref_coach_cnt + ?", 1)
		case enum.UserTypeClub:
			data["ref_club_cnt"] = gorm.Expr("ref_club_cnt + ?", 1)
		}
	}

	if err := d.sourceDB.Model(&model.Courses{}).Where("course_id = ?", course.CourseID).Updates(data).Error; err != nil {
		global.Lg.Error("更新课程统计信息失败", zap.Error(err))
		return err
	}

	return nil
}

func (d *GoodsDao) subCourseRef(ctx context.Context, course *model.Courses, good *model.Goods) error {
	data := make(map[string]interface{})
	switch good.UserType {
	case enum.UserTypeCoach:
		if course.RefCoachCnt > 0 {
			data["ref_coach_cnt"] = gorm.Expr("ref_coach_cnt - ?", 1)
		}
	case enum.UserTypeClub:
		if course.RefClubCnt > 0 {
			data["ref_club_cnt"] = gorm.Expr("ref_club_cnt - ?", 1)
		}
	}

	money := good.TeachMoney + good.AreaMoney
	if course.PriceMax < money {
		//统计goods表里面加起来的最大值
		var priceMax int64
		if err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
			Where("course_id = ? and state = 0", course.CourseID).
			Select("max(teach_money+area_money)").First(&priceMax).Error; err != nil {
			global.Lg.Error("查询商品失败", zap.Error(err))
			return err
		}
		data["price_max"] = priceMax
	}

	if course.PriceMin > money {
		var priceMin int64
		if err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
			Where("course_id = ? and state = 0", course.CourseID).
			Select("mix(teach_money+area_money)").First(&priceMin).Error; err != nil {
			global.Lg.Error("查询商品失败", zap.Error(err))
			return err
		}
		data["price_min"] = priceMin
	}

	if err := d.sourceDB.Model(&model.Courses{}).Where("course_id = ?", course.CourseID).Updates(data).Error; err != nil {
		global.Lg.Error("更新课程统计信息失败", zap.Error(err))
		return err
	}

	return nil
}
func (d *GoodsDao) GetByCoachCourseId(ctx context.Context, userId string, goodId string) (goods *model.Goods, err error) {
	err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Where("user_id = ? and good_id = ? and  state = 0", userId, goodId).
		First(&goods).Error
	if err != nil {
		global.Lg.Error("GoodsDao: GetByCoachCourseId", zap.Error(err))
		return nil, err
	}
	return goods, nil
}

func (d *GoodsDao) Get(ctx context.Context, fields, where string) (*model.Goods, error) {
	items, err := d.List(ctx, fields, where, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("GoodsDao: Get where=%s: %w", where, err)
	}
	if len(items) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &items[0], nil
}
func (d *GoodsDao) List(ctx context.Context, fields, where string, offset, limit int) ([]model.Goods, error) {
	var results []model.Goods
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("GoodsDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *GoodsDao) QueryPriceByUserId(ctx context.Context, userId string) (int64, int64, error) {
	//查询用户下所有商品的最大价格
	type result struct {
		MinPrice int64 `gorm:"column:min_price"`
		MaxPrice int64 `gorm:"column:max_price"`
	}

	var res result
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&d.m).
		Where("user_id = ? and state = 0 and on_shelf = 1", userId).
		Select("MIN(discount_money) as min_price, MAX(discount_money) as max_price").
		Scan(&res).Error
	if err != nil {
		global.Lg.Error("查询商品价格范围失败", zap.Error(err))
		return 0, 0, err
	}

	return res.MinPrice, res.MaxPrice, nil
}

func (d *GoodsDao) AddGoodsFinishedCnt(goodId string, finishedCourse int64) error {
	if err := d.sourceDB.Model(&d.m).Where("good_id = ?", goodId).Updates(map[string]interface{}{
		"finished_cnt":   gorm.Expr("finished_cnt + ?", finishedCourse),   //已完成+1
		"unfinished_cnt": gorm.Expr("unfinished_cnt - ?", finishedCourse), //未完成-1
	}).Error; err != nil {
		global.Lg.Error("更新商品完成课程数失败", zap.Error(err))
		return err
	}
	return nil
}

func (d *GoodsDao) AddGoodsUnFinishedCnt(goodId string, cnt int64) error {
	if err := d.sourceDB.Model(&d.m).Where("good_id = ?", goodId).Updates(map[string]interface{}{
		"unfinished_cnt": gorm.Expr("unfinished_cnt + ?", cnt), //未完成+1
	}).Error; err != nil {
		global.Lg.Error("增加商品完成课程数失败", zap.Error(err))
		return err
	}
	return nil
}

func (d *GoodsDao) AddGoodsCanceledCnt(goodId string, cnt int64) error {
	if err := d.sourceDB.Model(&d.m).Where("good_id = ?", goodId).Updates(map[string]interface{}{
		"canceled_cnt": gorm.Expr("canceled_cnt + ?", cnt), //已取消+1
	}).Error; err != nil {
		global.Lg.Error("更新商品未完成课程数失败", zap.Error(err))
		return err
	}
	return nil
}
