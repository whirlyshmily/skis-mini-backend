package dao

import (
	"context"
	"errors"
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

type SkiResortsTeachTimeDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.SkiResortsTeachTime
}

func NewSkiResortsTeachTimeDao(ctx context.Context, dbs ...*gorm.DB) *SkiResortsTeachTimeDao {
	dao := new(SkiResortsTeachTimeDao)
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

func (d *SkiResortsTeachTimeDao) GetSkiResorts(ctx context.Context, srMapTime map[int64][]*model.SkiResortsTeachTime, teachTime int, skiResortsId int64, skiResort *model.SkiResorts) (result SkiResortsInfo) {
	isFull := true
	if srTachTimes, ok := srMapTime[skiResortsId]; ok {
		timeNum := 0
		lastStartTime := time.Time{}
		for _, tt := range srTachTimes {
			if timeNum == 0 || lastStartTime.Add(30*time.Minute).Equal(time.Time(tt.TeachStartTime)) {
				timeNum++
				lastStartTime = time.Time(tt.TeachStartTime)
			} else {
				timeNum = 0
			}
			if timeNum*30 >= teachTime {
				isFull = false
				break
			}
		}
	}
	result = SkiResortsInfo{
		SkiResortsID: skiResortsId,
		Name:         skiResort.Name,
		Status:       skiResort.Status,
		IsFull:       isFull,
	}
	return
}

type SkiResortsInfo struct {
	SkiResortsID int64  `json:"ski_resorts_id"` // 场地ID
	Name         string `json:"name"`
	Status       uint8  `json:"status"`  //雪场状态，0-关闭，1-开启
	IsFull       bool   `json:"is_full"` //是否满员
}

func (d *SkiResortsTeachTimeDao) GetSkiResortDate(ctx context.Context, userId string, req *forms.CoachGetSkiResortDateRequest, teachTime int) (skiResortDateInfos []SkiResortDateInfo) {
	skiResortsTeachTimes, err := d.GetBySkiResortIdDateRange(ctx, userId, req.SkiResortsID, req.StartDate, req.EndDate)
	srMapTime := make(map[model.LocalDate][]*model.SkiResortsTeachTime)
	for _, srTeachTime := range skiResortsTeachTimes {
		srMapTime[srTeachTime.TeachDate] = append(srMapTime[srTeachTime.TeachDate], srTeachTime)
	}

	startDate, err := time.ParseInLocation("2006-01-02", req.StartDate, time.Local)
	if err != nil {
		global.Lg.Error("Error parsing date:", zap.Error(err))
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", req.EndDate, time.Local)
	if err != nil {
		global.Lg.Error("Error parsing date:", zap.Error(err))
		return
	}
	endDate = endDate.Add(24 * time.Hour)
	for ; startDate.Before(endDate); startDate = startDate.Add(24 * time.Hour) {
		isFull := true
		isSetDate := false
		if teachTimes, ok := srMapTime[model.LocalDate(startDate)]; ok {
			isSetDate = true
			timeNum := 0
			lastStartTime := time.Time{}
			for _, tt := range teachTimes {
				if timeNum == 0 || (lastStartTime.Add(30*time.Minute).Equal(time.Time(tt.TeachStartTime)) && tt.TeachNum > 0 && tt.TeachState == model.SkiTeachStateWaitAppointment) {
					timeNum++
					lastStartTime = time.Time(tt.TeachStartTime)
				} else {
					timeNum = 0
				}

				if timeNum*30 >= teachTime {
					isFull = false
					break
				}
			}
		}
		skiResortDateInfos = append(skiResortDateInfos, SkiResortDateInfo{
			TeachDate: model.LocalDate(startDate),
			IsFull:    isFull,
			IsSetDate: isSetDate,
		})
	}
	return skiResortDateInfos
}

type SkiResortDateInfo struct {
	TeachDate model.LocalDate `json:"teach_date"`
	IsFull    bool            `json:"is_full"` //是否满员
	IsSetDate bool            `json:"is_set_date"`
}

func (d *SkiResortsTeachTimeDao) GetSkiResortTime(c *gin.Context, userId string, req *forms.CoachGetSkiResortTimeRequest) (skiResortTimeInfos []SkiResortTimeInfo) {
	skiResortsTeachTimes, _ := NewSkiResortsTeachTimeDao(c, global.DB).GetBySkiResortIdDateRange(c, userId, 0, req.Date, req.Date)
	srMapTime := make(map[model.LocalTime]*model.SkiResortsTeachTime)
	for _, srTeachTime := range skiResortsTeachTimes {
		srMapTime[srTeachTime.TeachStartTime] = srTeachTime
	}

	startTi, _ := time.ParseInLocation("2006-01-02 15:04:05", req.Date+" 08:00:00", time.Local)
	for i := 0; i < 28; i++ {
		isFull := true
		isSetDate := false
		if teachTime, ok := srMapTime[model.LocalTime(startTi)]; ok {
			if teachTime.TeachNum > 0 && teachTime.TeachState == model.SkiTeachStateWaitAppointment {
				isFull = false
			}
			isSetDate = true
		}

		skiResortTimeInfos = append(skiResortTimeInfos, SkiResortTimeInfo{
			TeachTime: startTi.Format("15:04"),
			IsFull:    isFull,
			IsSetTime: isSetDate,
		})
		startTi = startTi.Add(30 * time.Minute)
	}
	return skiResortTimeInfos
}

type SkiResortTimeInfo struct {
	TeachTime string `json:"teach_time"`
	IsFull    bool   `json:"is_full"`     //是否满员
	IsSetTime bool   `json:"is_set_time"` //true：这个时间有设置，false：没有设置
}

func (d *SkiResortsTeachTimeDao) dealOrderCourseTeachTime(ctx context.Context, ordersCourses *model.OrdersCourses) {
	d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Where("id in ?", []int64(ordersCourses.TeachTimeIDs)).
		Find(&ordersCourses.TeachTimes)
	if ordersCourses.TeachState == model.TeachStateWaitUserConfirmCoachTime || ordersCourses.TeachState == model.TeachStateWaitUserConfirmClubTime || ordersCourses.TeachState == model.TeachStateWaitUserSecondConfirmTime {
		newTimes, err := NewOrdersCoursesStateDao(ctx, d.replicaDB[rand.Intn(len(d.replicaDB))]).
			Last(ctx, "*", "order_course_id = ? and  state = 0 and operate in ?",
				ordersCourses.OrderCourseID,
				[]model.TeachState{model.OperateCoachChangeCourseTime, model.TeachStateWaitUserConfirmCoachTime, model.TeachStateWaitUserConfirmClubTime})
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			global.Lg.Error("获取订单课程状态时间失败", zap.Error(err))
		} else {
			d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
				Where("id in ?", []int64(newTimes.TeachTimeIDs)).
				Find(&ordersCourses.NewTeachTimes)
		}
	}

	ordersCoursesState := model.OrdersCoursesState{}
	d.sourceDB.Model(&model.OrdersCoursesState{}).
		Where("order_course_id = ? and state = 0",
			ordersCourses.OrderCourseID).
		Last(&ordersCoursesState)
	//ordersCoursesState.LastConfirmTime = ordersCoursesState.CreatedAt.Add(24 * time.Hour)
	ordersCourses.OrdersCoursesState = ordersCoursesState
	return
}
func (d *SkiResortsTeachTimeDao) GetBySkiResortIds(ctx context.Context, userId string, skiResortIds []int64) (results []*model.SkiResortsTeachTime, err error) {
	err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select("ski_resorts_id, teach_date, teach_start_time, teach_end_time, teach_num, teach_state, state").
		Where("user_id = ? and ski_resorts_id in ? and teach_start_time > ?", userId, skiResortIds, time.Now().Format("2006-01-02 15:04:05")).
		Where("state = 0 and teach_state = 0 and teach_num > 0 ").
		Order("teach_start_time asc").
		Find(&results).Error
	return results, err
}
func (d *SkiResortsTeachTimeDao) GetBySkiResortIdDateRange(ctx context.Context, userId string, skiResortId int64, startDate, endDate string) (results []*model.SkiResortsTeachTime, err error) {

	db := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select("ski_resorts_id, teach_date, teach_start_time, teach_end_time, teach_num, teach_state, state").
		Where("user_id = ?  and teach_start_time > ? and teach_start_time < ? and state = 0", userId, startDate, endDate+"  23:59:59")

	if skiResortId != 0 {
		db = db.Where("ski_resorts_id = ?", skiResortId)
	}

	err = db.Order("teach_start_time asc").
		Find(&results).Error
	return results, err
}
func (d *SkiResortsTeachTimeDao) Create(ctx context.Context, obj *model.SkiResortsTeachTime) error {
	err := d.sourceDB.Model(d.m).Create(&obj).Error
	if err != nil {
		return fmt.Errorf("SkiResortsTeachTimeDao: %w", err)
	}
	return nil
}

func (d *SkiResortsTeachTimeDao) Get(ctx context.Context, fields, where string) (*model.SkiResortsTeachTime, error) {
	items, err := d.List(ctx, fields, where, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("SkiResortsTeachTimeDao: Get where=%s: %w", where, err)
	}
	if len(items) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &items[0], nil
}

func (d *SkiResortsTeachTimeDao) List(ctx context.Context, fields, where string, offset, limit int) ([]model.SkiResortsTeachTime, error) {
	var results []model.SkiResortsTeachTime
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("SkiResortsTeachTimeDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *SkiResortsTeachTimeDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("SkiResortsTeachTimeDao:Update where=%s: %w", where, err)
	}
	return nil
}
