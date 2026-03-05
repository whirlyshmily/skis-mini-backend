package dao

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math/rand"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"
)

type ClubSkiResortsDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.ClubsSkiResorts
}

func NewClubSkiResortsDao(ctx context.Context, dbs ...*gorm.DB) *ClubSkiResortsDao {
	dao := new(ClubSkiResortsDao)
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

func (d *ClubSkiResortsDao) ClubGetSkiResorts(c *gin.Context, clubId, orderCourseId string) (clubSkiResortsInfos []SkiResortsInfo) {
	results := make([]*model.ClubsSkiResorts, 0)
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Preload("SkiResorts", "state = 0").
		Where("club_id = ? and state = 0", clubId).Find(&results).Error
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

	skiResortsTeachTimes, err := NewSkiResortsTeachTimeDao(c, global.DB).GetBySkiResortIds(c, clubId, skiResortsId)
	srMapTime := make(map[int64][]*model.SkiResortsTeachTime)
	for _, srTeachTime := range skiResortsTeachTimes {
		srMapTime[int64(srTeachTime.SkiResortsID)] = append(srMapTime[int64(srTeachTime.SkiResortsID)], srTeachTime)
	}

	for _, item := range results {
		if item.SkiResorts.Id == 0 { //删除的场地
			continue
		}
		coachSkiResortsInfo := NewSkiResortsTeachTimeDao(c, global.DB).GetSkiResorts(c, srMapTime, orderCourse.TeachTime, item.SkiResortsID, &item.SkiResorts)
		clubSkiResortsInfos = append(clubSkiResortsInfos, coachSkiResortsInfo)
	}
	return
}

func (d *ClubSkiResortsDao) ClubGetSkiResortDate(c *gin.Context, clubId string, req *forms.CoachGetSkiResortDateRequest) (coachSkiResortDateInfos []SkiResortDateInfo) {
	if req.StartDate > req.EndDate {
		return nil
	}
	result := model.ClubsSkiResorts{}
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Preload("SkiResorts", "state = 0").
		Where("club_id = ? and ski_resorts_id=? and state = 0", clubId, req.SkiResortsID).Last(&result).Error
	if err != nil {
		global.Lg.Error("查询俱乐部场地失败", zap.Error(err))
		return nil
	}
	orderCourse, err := OrderCourseInfo(c, "", req.OrderCourseId)
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err))
		return nil
	}
	coachSkiResortDateInfos = NewSkiResortsTeachTimeDao(c, global.DB).GetSkiResortDate(c, clubId, req, orderCourse.TeachTime)
	return coachSkiResortDateInfos
}

func (d *ClubSkiResortsDao) ClubGetSkiResortTime(c *gin.Context, clubId string, req *forms.CoachGetSkiResortTimeRequest) (clubSkiResortTimeInfos []SkiResortTimeInfo) {
	result := model.ClubsSkiResorts{}
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Preload("SkiResorts", "state = 0").
		Where("club_id = ? and ski_resorts_id=? and state = 0", clubId, req.SkiResortsID).Last(&result).Error
	if err != nil {
		global.Lg.Error("查询教练场地失败", zap.Error(err))
		return nil
	}

	clubSkiResortTimeInfos = NewSkiResortsTeachTimeDao(c, global.DB).GetSkiResortTime(c, clubId, req)
	return clubSkiResortTimeInfos
}

type ClubSkiResortTeachDate struct {
	TeachDate      time.Time `json:"teach_date"`
	TeachNum       int       `json:"teach_num"`
	OrderCourseNum int       `json:"order_course_num"`
	TeachState     int       `json:"teach_state"`
	EventNum       int       `json:"event_num"`
}

// func (d *ClubSkiResortsDao) QuerySkiResortTeachDateList(c *gin.Context, req *forms.ClubSkiResortTeachDateListRequest) ([]ClubTeachDateData, error) {
func (d *ClubSkiResortsDao) QuerySkiResortTeachDateList(c *gin.Context, req *forms.ClubSkiResortTeachDateListRequest) (interface{}, error) {
	userId := c.GetString("user_id")
	db := global.DB.Table("ski_resorts_teach_time as srt").Where("srt.user_id = ?", userId).Where("srt.state = 0")
	if req.TeachDateStart != "" {
		db = db.Where("srt.teach_date >= ?", req.TeachDateStart)
	}
	if req.TeachDateEnd != "" {
		db = db.Where("srt.teach_date <= ?", req.TeachDateEnd)
	}
	db.Select("srt.teach_date, min(srt.teach_num) as teach_num,max(srt.teach_state) as teach_state,count(srte.id) as event_num,count(srtoc.id) as order_course_num").
		Joins("left join ski_resorts_teach_time_event as srte on srte.skirt_id=srt.id and srte.state=0").
		Joins("left join ski_resorts_teach_time_order_courses as srtoc on srtoc.skirt_id = srt.id and srtoc.state = 0").
		Group("srt.teach_date").Order("srt.teach_date asc")
	var clubSkiResortTeachDates []*ClubSkiResortTeachDate
	if err := db.Find(&clubSkiResortTeachDates).Error; err != nil {
		global.Lg.Error("QuerySkiResortTeachDateList error", zap.Error(err))
		return nil, err
	}

	teachDateStart, err := time.ParseInLocation("2006-01-02", req.TeachDateStart, time.Local)
	if err != nil {
		global.Lg.Error("Error parsing date:", zap.Error(err))
		return nil, err
	}
	teachDateEnd, err := time.ParseInLocation("2006-01-02", req.TeachDateEnd, time.Local)
	if err != nil {
		global.Lg.Error("Error parsing date:", zap.Error(err))
		return nil, err
	}
	skiTeachDates := make(map[string]*ClubSkiResortTeachDate, 0)
	for _, v := range clubSkiResortTeachDates {
		a := v.TeachDate.Format("2006-01-02")
		skiTeachDates[a] = v
	}
	TeachDateDatas := make([]ClubTeachDateData, 0)
	for teachDateStart.Before(teachDateEnd.Add(24 * time.Hour)) {
		d := teachDateStart.Format("2006-01-02")
		teachDateData := ClubTeachDateData{
			TeachDate:   d,
			ConfigState: 1,
		}
		if skiTeachDate, ok := skiTeachDates[d]; ok {
			if skiTeachDate.OrderCourseNum > 0 {
				teachDateData.ConfigState = 2
			}
			if skiTeachDate.TeachState == model.SkiTeachStateLocked {
				teachDateData.ConfigState = 3
			}
		}
		TeachDateDatas = append(TeachDateDatas, teachDateData)
		teachDateStart = teachDateStart.Add(24 * time.Hour)
	}
	return TeachDateDatas, nil
}

type ClubTeachDateData struct {
	TeachDate   string `json:"teach_date"`
	ConfigState int    `json:"config_state"` //1. 没预约，没锁定、2. 有预约、3. 有锁定
}

func (d *ClubSkiResortsDao) Create(ctx context.Context, obj *model.ClubsSkiResorts) error {
	err := d.sourceDB.Model(d.m).Create(&obj).Error
	if err != nil {
		return fmt.Errorf("ClubSkiResortsDao: %w", err)
	}
	return nil
}

func (d *ClubSkiResortsDao) QueryClubIdBySkiId(skiId int) (clubIds []string, err error) {
	if err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Where("ski_resorts_id = ? and state = 0", skiId).Pluck("club_id", &clubIds).Error; err != nil {
		global.Lg.Error("查询俱乐部标签失败", zap.Error(err))
		return nil, err
	}
	return clubIds, nil
}

func (d *ClubSkiResortsDao) Get(ctx context.Context, fields, where string) (*model.ClubsSkiResorts, error) {
	items, err := d.List(ctx, fields, where, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("ClubSkiResortsDao: Get where=%s: %w", where, err)
	}
	if len(items) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &items[0], nil
}

func (d *ClubSkiResortsDao) List(ctx context.Context, fields, where string, offset, limit int) ([]model.ClubsSkiResorts, error) {
	var results []model.ClubsSkiResorts
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("ClubSkiResortsDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *ClubSkiResortsDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("ClubSkiResortsDao:Update where=%s: %w", where, err)
	}
	return nil
}

func (d *ClubSkiResortsDao) Delete(ctx context.Context, where string, args ...interface{}) error {
	if len(where) == 0 {
		return gorm.ErrInvalidDB
	}
	if err := d.sourceDB.Where(where, args...).Delete(d.m).Error; err != nil {
		return fmt.Errorf("ClubSkiResortsDao: Delete where=%s: %w", where, err)
	}
	return nil
}
