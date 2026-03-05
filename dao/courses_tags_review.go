package dao

import (
	"context"
	"encoding/json"
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

type CoachesTagsReviewDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.CoachesTagsReview
}

func NewCoachesTagsReviewDao(ctx context.Context, dbs ...*gorm.DB) *CoachesTagsReviewDao {
	dao := new(CoachesTagsReviewDao)
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

func (d *CoachesTagsReviewDao) CoachAddTagReview(c *gin.Context, req forms.CoachAddTagReviewRequest) error {
	coachInfo, err := CoachInfoByUserId(c.GetString("uid"))
	if err != nil {
		return enum.NewErr(enum.CoachNotExistErr, "教练不存在")
	}
	for _, v := range req.TagReviews {
		imgStr, _ := json.Marshal(v.TagImgUrls)

		coachesTag := model.CoachesTagsReview{}
		d.sourceDB.Model(d.m).Where("coach_id = ? and tag_id = ? and state = 0", coachInfo.CoachId, v.TagID).Last(&coachesTag)

		if coachesTag.ID != 0 {
			if coachesTag.Verified == model.VerifiedVerified {
				continue
			}
			d.sourceDB.Model(d.m).Where("id = ? and coach_id = ?", coachesTag.ID, coachInfo.CoachId).
				Updates(map[string]interface{}{
					"tag_img_urls": string(imgStr),
					"verified":     model.VerifiedUnverified,
				})
			continue
		}

		m := model.CoachesTagsReview{
			CoachID:    coachInfo.CoachId,
			TagID:      v.TagID,
			TagImgUrls: v.TagImgUrls,
		}
		if err = d.sourceDB.Model(d.m).Create(&m).Error; err != nil {
			global.Lg.Error("CoachAddTagReview FirstOrCreate  技能申请插入失败", zap.Error(err))
			return enum.NewErr(enum.CoachTagReviewCreateErr, "技能申请插入失败")
		}
	}
	return nil
}

func (d *CoachesTagsReviewDao) CoachGetTagReview(c *gin.Context, req forms.CoachGetTagReviewRequest) ([]model.CoachesTagsReview, error) {
	coachInfo, err := CoachInfoByUserId(c.GetString("uid"))
	if err != nil {
		return nil, enum.NewErr(enum.CoachNotExistErr, "教练不存在")
	}
	var ids []int64
	err = d.sourceDB.Model(d.m).Select("max(id) as id").Where("coach_id = ? and tag_id in (?) and state = 0", coachInfo.CoachId, req.TagIds).
		Group("tag_id").
		Find(&ids).Error
	if err != nil {
		global.Lg.Error("CoachGetTagReview  技能申请查询失败", zap.Error(err))
		return nil, enum.NewErr(enum.CoachTagReviewGetErr, "技能申请查询失败")
	}
	var results []model.CoachesTagsReview
	err = d.sourceDB.Model(d.m).Where("id in (?)", ids).Find(&results).Error
	if err != nil {
		global.Lg.Error("CoachGetTagReview  技能申请查询失败", zap.Error(err))
		return nil, enum.NewErr(enum.CoachTagReviewGetErr, "技能申请查询失败")
	}
	return results, nil
}

func (d *CoachesTagsReviewDao) Create(ctx context.Context, obj *model.CoachesTagsReview) error {
	err := d.sourceDB.Model(d.m).Create(&obj).Error
	if err != nil {
		return fmt.Errorf("CoachesTagsReviewDao: %w", err)
	}
	return nil
}

func (d *CoachesTagsReviewDao) Get(ctx context.Context, fields, where string) (*model.CoachesTagsReview, error) {
	items, err := d.List(ctx, fields, where, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("CoachesTagsReviewDao: Get where=%s: %w", where, err)
	}
	if len(items) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &items[0], nil
}

func (d *CoachesTagsReviewDao) List(ctx context.Context, fields, where string, offset, limit int) ([]model.CoachesTagsReview, error) {
	var results []model.CoachesTagsReview
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("CoachesTagsReviewDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *CoachesTagsReviewDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("CoachesTagsReviewDao:Update where=%s: %w", where, err)
	}
	return nil
}

func (d *CoachesTagsReviewDao) Delete(ctx context.Context, where string, args ...interface{}) error {
	if len(where) == 0 {
		return gorm.ErrInvalidField
	}
	if err := d.sourceDB.Where(where, args...).Delete(d.m).Error; err != nil {
		return fmt.Errorf("CoachesTagsReviewDao: Delete where=%s: %w", where, err)
	}
	return nil
}
