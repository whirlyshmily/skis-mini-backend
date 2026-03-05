package dao

import (
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateFeed(c *gin.Context, req *forms.CreateFeedRequest) (feed *model.Feeds, err error) {
	feed = &model.Feeds{
		Title:    req.Title,
		Urls:     req.Urls,
		Detail:   req.Detail,
		UserId:   c.GetString("user_id"),
		UserType: c.GetInt("user_type"),
	}
	err = global.DB.Model(&model.Feeds{}).Create(feed).Error
	if err != nil {
		global.Lg.Error("创建动态失败", zap.Error(err), zap.Any("req", req))
		return
	}
	return
}
func QueryFeedList(userId string, userType int, req *forms.QueryFeedListRequest) (int64, []*model.Feeds, error) {
	var feeds []*model.Feeds
	var total int64
	db := global.DB.Model(&model.Feeds{}).Where("state = 0")
	//根据教练ID查询
	if userId != "" {
		db = db.Where("user_id = ? and user_type = ?", userId, userType)
	}
	if req.Curated != nil { //是否精选，0-不精选，1-精选
		db = db.Where("curated = ?", req.Curated)
	}
	if req.OnShelf != nil { //是否自己发布的，0-不是，1-是
		db = db.Where("on_shelf = ?", req.OnShelf)
	}
	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("查询动态列表失败", zap.Error(err))
		return 0, nil, err
	}

	db = db.Preload("Club.ClubTags.Tag").Preload("Coach.CoachTags.Tag").Preload("Coach.Users").Order("priority desc, id desc")
	if req.Page > 0 && req.PageSize > 0 {
		db = db.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize)
	}

	if err := db.Find(&feeds).Error; err != nil {
		global.Lg.Error("查询动态列表失败", zap.Error(err))
		return 0, nil, err
	}

	return total, feeds, nil

}

func QueryFeedInfo(id int64) (*model.Feeds, error) {
	feed := &model.Feeds{}
	if err := global.DB.
		Preload("Club.ClubTags.Tag").
		Preload("Coach.CoachTags.Tag").
		Preload("Coach.Users").
		Preload("Coach.Certificates", "state=0").
		Preload("Coach.Certificates.CertificateConfig").
		Preload("Coach.CoachesSkiResorts", "state=0").
		Preload("Coach.CoachesSkiResorts.SkiResorts").
		Where("id = ? AND state = 0", id).First(feed).Error; err != nil {
		global.Lg.Error("查询动态失败", zap.Error(err))
		return nil, err
	}
	return feed, nil
}

func UpdateFeed(id int64, userId string, req *forms.UpdateFeedRequest) (*model.Feeds, error) {
	feed, err := QueryFeedInfo(id)
	if err != nil {
		global.Lg.Error("查询动态失败", zap.Error(err))
		return nil, err
	}

	if feed.UserId != userId {
		return nil, enum.NewErr(enum.ParamErr, "无权限操作动态")
	}

	if req.Title != nil {
		feed.Title = *req.Title
	}

	if req.Urls != nil {
		feed.Urls = *req.Urls
	}

	if req.Detail != nil {
		feed.Detail = *req.Detail
	}

	if req.Curated != nil {
		feed.Curated = *req.Curated
	}

	if req.OnShelf != nil {
		feed.OnShelf = *req.OnShelf
	}

	if err = global.DB.Model(model.Feeds{}).Where("id = ?", id).Save(feed).Error; err != nil {
		global.Lg.Error("更新动态失败", zap.Error(err))
		return nil, err
	}

	return feed, nil
}

func DeleteFeed(id int64, userId string) error {
	feed, err := QueryFeedInfo(id)
	if err != nil {
		global.Lg.Error("查询动态失败", zap.Error(err))
		return err
	}

	if feed.UserId != userId {
		return enum.NewErr(enum.ParamErr, "无权限操作动态")
	}

	feed.State = 1
	if err = global.DB.Model(model.Feeds{}).Where("id = ?", id).Save(feed).Error; err != nil {
		global.Lg.Error("删除动态失败", zap.Error(err))
		return err
	}

	return nil
}

func FeedAddView(id int64) {
	if err := global.DB.Model(model.Feeds{}).Where("id = ?", id).Update("view", gorm.Expr("view + ?", 1)).Error; err != nil {
		global.Lg.Error("动态添加浏览量失败", zap.Error(err))
	}
}
