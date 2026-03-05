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
	"strconv"
)

type CoachesCertificatesDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.CoachesCertificates
}

func NewCoachesCertificatesDao(ctx context.Context, dbs ...*gorm.DB) *CoachesCertificatesDao {
	dao := new(CoachesCertificatesDao)
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

func (d *CoachesCertificatesDao) CoachAddCertificates(c *gin.Context, req forms.CoachAddCertificatesRequest) error {
	coachInfo, err := CoachInfoByUserId(c.GetString("uid"))
	if err != nil {
		return enum.NewErr(enum.CoachNotExistErr, "教练不存在")
	}
	for _, v := range req.CertificateReviews {
		coachesCertificates := model.CoachesCertificates{}
		d.sourceDB.Model(d.m).
			Where("coach_id = ? and certificate_id = ? and  level = ? and state = 0", coachInfo.CoachId, v.CertificateId, v.Level).
			Last(&coachesCertificates)
		if coachesCertificates.ID != 0 {
			if coachesCertificates.Verified == model.VerifiedVerified {
				continue
			}
			imgStr, _ := json.Marshal(v.CertificateImgUrls)
			d.sourceDB.Model(d.m).Where("id = ? and coach_id = ?", coachesCertificates.ID, coachInfo.CoachId).
				Updates(map[string]interface{}{
					"certificate_img_urls": string(imgStr),
					"verified":             model.VerifiedUnverified,
				})
			continue
		}
		certificateInfo, err := QueryCertificateConfigInfo(v.CertificateId)
		if err != nil {
			global.Lg.Error("CoachAddTagReview QueryCertificateTagInfo  证书查询失败："+strconv.FormatInt(v.CertificateId, 10), zap.Error(err))
			return enum.NewErr(enum.CertificateExitErr, "证书查询失败")
		}
		if !IsElementInArray(certificateInfo.Level, v.Level) {
			global.Lg.Error("CoachAddTagReview IsElementInArray  证书级别错误："+strconv.FormatInt(v.CertificateId, 10), zap.Error(err))
			return enum.NewErr(enum.CertificateLevelErr, "证书级别错误")
		}

		m := model.CoachesCertificates{
			CoachID:            coachInfo.CoachId,
			CertificateID:      v.CertificateId,
			CertificateImgUrls: v.CertificateImgUrls,
			Level:              v.Level,
		}
		if err = d.sourceDB.Model(d.m).Create(&m).Error; err != nil {
			global.Lg.Error("CoachAddTagReview Create  证书申请插入失败", zap.Error(err))
			return enum.NewErr(enum.CoachCertificateReviewCreateErr, "证书申请插入失败")
		}
	}
	return nil
}

func (d *CoachesCertificatesDao) CoachGetCertificates(c *gin.Context, req forms.CoachGetCertificatesRequest) ([]model.CoachesCertificates, error) {
	coachInfo, err := CoachInfoByUserId(c.GetString("uid"))
	if err != nil {
		return nil, enum.NewErr(enum.CoachNotExistErr, "教练不存在")
	}

	var items []model.CoachesCertificates
	var ids []int64
	err = d.sourceDB.Model(d.m).Select("max(id) as id").Where("coach_id = ? and certificate_id in (?) and state = 0", coachInfo.CoachId, req.CertificateIds).
		Group("certificate_id, level").
		Find(&ids).Error

	err = d.sourceDB.Model(d.m).Where("id in (?)", ids).Order("id desc").Find(&items).Error
	if err != nil {
		global.Lg.Error("CoachGetCertificates  证书申请查询失败", zap.Error(err))
		return nil, enum.NewErr(enum.CoachCertificateReviewGetErr, "证书申请查询失败")
	}
	return items, nil
}

func (d *CoachesCertificatesDao) CoachGetAllCertificates(coachId string, verified int) (coachesCertificates []model.CoachesCertificates, err error) {
	db := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m)
	if verified != -1 {
		db = db.Where("verified = ?", verified)
	}
	err = db.Preload("CertificateConfig").
		Where("coach_id = ? and state = 0", coachId).
		Find(&coachesCertificates).Error
	if err != nil {
		global.Lg.Error("CoachGetAllCertificates 查询教练全部证书失败", zap.Error(err))
		return nil, err
	}
	return coachesCertificates, nil
}

func (d *CoachesCertificatesDao) Create(ctx context.Context, obj *model.CoachesCertificates) error {
	err := d.sourceDB.Model(d.m).Create(&obj).Error
	if err != nil {
		return fmt.Errorf("CoachesCertificatesDao: %w", err)
	}
	return nil
}

func (d *CoachesCertificatesDao) Get(ctx context.Context, fields, where string) (*model.CoachesCertificates, error) {
	items, err := d.List(ctx, fields, where, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("CoachesCertificatesDao: Get where=%s: %w", where, err)
	}
	if len(items) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &items[0], nil
}

func (d *CoachesCertificatesDao) List(ctx context.Context, fields, where string, offset, limit int) ([]model.CoachesCertificates, error) {
	var results []model.CoachesCertificates
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("CoachesCertificatesDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *CoachesCertificatesDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("CoachesCertificatesDao:Update where=%s: %w", where, err)
	}
	return nil
}
