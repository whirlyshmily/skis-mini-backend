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

type CoachesSkiResortsDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.CoachesSkiResorts
}

func NewCoachSkiResortsDao(ctx context.Context, dbs ...*gorm.DB) *CoachesSkiResortsDao {
	dao := new(CoachesSkiResortsDao)
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

func (d *CoachesSkiResortsDao) CoachEditSkiResorts(c *gin.Context, skiResortsIds []int64) error {
	uid := c.GetString("uid")

	coachInfo, err := CoachInfoByUserId(uid)
	if err != nil {
		return enum.NewErr(enum.CoachNotExistErr, "教练不存在")
	}

	//todo: 1.查询教练场地是否被使用 2.查询场地是否存在
	if len(skiResortsIds) != 0 {
		skiResortsIds, err = QuerySkiResortByIds(c, skiResortsIds)
		if err != nil {
			global.Lg.Error("CoachEditSkiResorts QuerySkiResortByIds 场地查询失败", zap.Error(err))
			return enum.NewErr(enum.SkiResortsGetErr, "场地查询失败")
		}
	}
	for _, id := range skiResortsIds {
		value := model.CoachesSkiResorts{
			CoachID:      coachInfo.CoachId,
			SkiResortsID: id,
			State:        0,
		}
		if err = d.sourceDB.Model(d.m).Where("ski_resorts_id = ? and coach_id = ?  and state = 0", id, coachInfo.CoachId).FirstOrCreate(&value, model.CoachesSkiResorts{}).Error; err != nil {
			global.Lg.Error("CoachesSkiResortsDao FirstOrCreate  场地插入失败", zap.Error(err))
			return enum.NewErr(enum.CoachCreateSkiErr, "场地插入失败")
		}
	}
	updb := d.sourceDB.Model(d.m).Where("coach_id = ? and state = 0", coachInfo.CoachId)
	if len(skiResortsIds) != 0 {
		updb = updb.Where("ski_resorts_id not in (?)", skiResortsIds)
	}
	err = updb.Update("state", 1).Error
	if err != nil {
		global.Lg.Error("更新教练场地失败", zap.Error(err))
		return enum.NewErr(enum.CoachUpdateSkiErr, "更新教练场地失败")
	}
	clubsCoach := model.ClubsCoaches{}
	global.DB.Model(model.ClubsCoaches{}).Where("coach_id = ? and verified=1 and state=0", coachInfo.CoachId).Last(&clubsCoach)
	if clubsCoach.ClubID != "" {
		HandleClubSki(c, clubsCoach.ClubID)
	}
	return nil
}

func (d *CoachesSkiResortsDao) CoachGetSkiResorts(c *gin.Context, coachId, orderCourseId string) (coachSkiResortsInfos []SkiResortsInfo) {
	results := make([]*model.CoachesSkiResorts, 0)
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Preload("SkiResorts", "state = 0").
		Where("coach_id = ? and state = 0", coachId).Find(&results).Error
	if err != nil {
		global.Lg.Error("查询教练场地失败", zap.Error(err))
		return nil
	}

	orderCourse, err := OrderCourseInfo(c, "", orderCourseId)
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err))
		return nil
	}

	skiResortsId := make([]int64, 0)
	for _, item := range results {
		if item.SkiResorts.Id == 0 { //删除的场地
			continue
		}
		skiResortsId = append(skiResortsId, item.SkiResortsID)
	}

	skiResortsTeachTimes, err := NewSkiResortsTeachTimeDao(c, global.DB).GetBySkiResortIds(c, coachId, skiResortsId)
	srMapTime := make(map[int64][]*model.SkiResortsTeachTime)
	for _, srTeachTime := range skiResortsTeachTimes {
		srMapTime[int64(srTeachTime.SkiResortsID)] = append(srMapTime[int64(srTeachTime.SkiResortsID)], srTeachTime)
	}

	for _, item := range results {
		if item.SkiResorts.Id == 0 { //删除的场地
			continue
		}
		coachSkiResortsInfo := NewSkiResortsTeachTimeDao(c, global.DB).GetSkiResorts(c, srMapTime, orderCourse.TeachTime, item.SkiResortsID, &item.SkiResorts)
		coachSkiResortsInfos = append(coachSkiResortsInfos, coachSkiResortsInfo)
	}
	return coachSkiResortsInfos
}

func (d *CoachesSkiResortsDao) CoachGetSkiResortDate(c *gin.Context, coachId string, req *forms.CoachGetSkiResortDateRequest) (coachSkiResortDateInfos []SkiResortDateInfo) {
	if req.StartDate > req.EndDate {
		return nil
	}
	result := model.CoachesSkiResorts{}
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Preload("SkiResorts", "state = 0").
		Where("coach_id = ? and ski_resorts_id=? and state = 0", coachId, req.SkiResortsID).Last(&result).Error
	if err != nil {
		global.Lg.Error("查询教练场地失败", zap.Error(err))
		return nil
	}
	orderCourse, err := OrderCourseInfo(c, "", req.OrderCourseId)
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err))
		return nil
	}
	coachSkiResortDateInfos = NewSkiResortsTeachTimeDao(c, global.DB).GetSkiResortDate(c, coachId, req, orderCourse.TeachTime)
	return coachSkiResortDateInfos
}

func (d *CoachesSkiResortsDao) CoachGetSkiResortTime(c *gin.Context, coachId string, req *forms.CoachGetSkiResortTimeRequest) (coachSkiResortTimeInfos []SkiResortTimeInfo) {
	result := model.CoachesSkiResorts{}
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Preload("SkiResorts", "state = 0").
		Where("coach_id = ? and ski_resorts_id=? and state = 0", coachId, req.SkiResortsID).Last(&result).Error
	if err != nil {
		global.Lg.Error("查询教练场地失败", zap.Error(err))
		return nil
	}

	coachSkiResortTimeInfos = NewSkiResortsTeachTimeDao(c, global.DB).GetSkiResortTime(c, coachId, req)
	return coachSkiResortTimeInfos
}

type CoachSkiResortTimeInfo struct {
	TeachTime string `json:"teach_time"`
	IsFull    bool   `json:"is_full"`     //是否满员
	IsSetTime bool   `json:"is_set_time"` //true：这个时间有设置，false：没有设置
}

func (d *CoachesSkiResortsDao) QueryCoachIdBySkiId(skiId int) (coachIds []string, err error) {
	if err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Where("ski_resorts_id = ? and state = 0", skiId).Pluck("coach_id", &coachIds).Error; err != nil {
		global.Lg.Error("查询教练场地失败", zap.Error(err))
		return nil, err
	}
	return coachIds, nil
}

func (d *CoachesSkiResortsDao) Create(ctx context.Context, obj *model.CoachesSkiResorts) error {
	err := d.sourceDB.Model(d.m).Create(&obj).Error
	if err != nil {
		return fmt.Errorf("CoachesSkiResortsDao: %w", err)
	}
	return nil
}

func (d *CoachesSkiResortsDao) Get(ctx context.Context, fields, where string, args ...interface{}) (item *model.CoachesSkiResorts, err error) {
	err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where, args...).Last(&item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (d *CoachesSkiResortsDao) List(ctx context.Context, fields, where string, offset, limit int) ([]model.CoachesSkiResorts, error) {
	var results []model.CoachesSkiResorts
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("CoachesSkiResortsDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *CoachesSkiResortsDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("CoachesSkiResortsDao:Update where=%s: %w", where, err)
	}
	return nil
}

func (d *CoachesSkiResortsDao) Delete(ctx context.Context, where string, args ...interface{}) error {
	if len(where) == 0 {
		return gorm.ErrInvalidField
	}
	if err := d.sourceDB.Model(d.m).Where(where, args...).Delete(d.m).Error; err != nil {
		return fmt.Errorf("CoachesSkiResortsDao: Delete where=%s: %w", where, err)
	}
	return nil
}
