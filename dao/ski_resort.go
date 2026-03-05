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

func CreateSkiResort(req *forms.CreateSkiResortRequest) (*model.SkiResorts, error) {
	// 检查是否存在
	_, err := QuerySkiResortByName(req.Name)
	if err != nil && err != gorm.ErrRecordNotFound {
		global.Lg.Error("QuerySkiResortByName error", zap.Error(err))
		return nil, err
	}
	if err == nil {
		return nil, enum.NewErr(enum.SkiResortExistErr, "名称已存在")
	}

	skiResort := &model.SkiResorts{
		Name:     req.Name,
		Province: req.Province,
		City:     req.City,
	}
	if err = global.DB.Create(skiResort).Error; err != nil {
		global.Lg.Error("创建滑雪场失败", zap.Error(err))
		return nil, err
	}

	return skiResort, nil
}

func QuerySkiResortByIds(ctx context.Context, skiId []int64) (skiIds []int64, err error) {
	if err = global.DB.Model(model.SkiResorts{}).Where("id in ? and state = 0", skiId).Pluck("id", &skiIds).Error; err != nil {
		global.Lg.Error("查询教练场地失败", zap.Error(err))
		return nil, err
	}
	return skiIds, nil
}
func QuerySkiResortByName(name string) (*model.SkiResorts, error) {
	skiResort := &model.SkiResorts{}
	err := global.DB.Model(model.SkiResorts{}).Where("name = ? and state = 0", name).First(&skiResort).Error
	if err != nil {
		global.Lg.Error("QuerySkiResortByName error", zap.Error(err))
		return nil, err
	}
	return skiResort, nil
}

func QuerySkiResortsList(req *forms.QuerySkiResortsListRequest) (int64, []*model.SkiResortList, error) {
	var total int64
	db := global.DB.Table("ski_resorts").Where("state = 0")

	if req.Keyword != "" {
		db = db.Where("name like ?", "%"+req.Keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("QuerySkiResortsList error", zap.Error(err))
		return 0, nil, err
	}
	var skiResorts []*model.SkiResortList
	if err := db.Order("id desc").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&skiResorts).Error; err != nil {
		global.Lg.Error("QuerySkiResortsList error", zap.Error(err))
		return 0, nil, err
	}

	return total, skiResorts, nil
}

func QuerySkiResortInfo(id int64) (*model.SkiResorts, error) {
	var skiResort *model.SkiResorts
	err := global.DB.Where("id = ? and state = 0", id).First(&skiResort).Error
	if err != nil {
		global.Lg.Error("QuerySkiResortInfo error", zap.Error(err))
		return nil, err
	}
	return skiResort, nil
}

func UpdateSkiResort(id int64, req *forms.UpdateSkiResortRequest) (*model.SkiResorts, error) {
	skiResort, err := QuerySkiResortInfo(id)
	if err != nil {
		global.Lg.Error("QuerySkiResortInfo error", zap.Error(err))
		return nil, err
	}

	//查询文件名是否存在了
	var existSkiResort *model.SkiResorts
	if err = global.DB.Where("name = ? and id != ? and state = 0", req.Name, id).First(&existSkiResort).Error; err == nil {
		return nil, enum.NewErr(enum.SkiResortExistErr, "名称已存在")
	}

	if req.Name != nil {
		skiResort.Name = *req.Name
	}
	if req.Province != nil {
		skiResort.Province = *req.Province
	}
	if req.City != nil {
		skiResort.City = *req.City
	}
	if req.Status != nil {
		skiResort.Status = *req.Status
	}

	if err = global.DB.Save(skiResort).Error; err != nil {
		global.Lg.Error("UpdateSkiResort error", zap.Error(err))
		return nil, err
	}
	return skiResort, nil
}

func DeleteSkiResort(id int64) error {
	skiResort, err := QuerySkiResortInfo(id)
	if err != nil {
		global.Lg.Error("QuerySkiResortInfo error", zap.Error(err))
		return err
	}

	//todo 查询滑雪场下所有课程，是否有教练
	skiResort.State = 1
	if err = global.DB.Save(skiResort).Error; err != nil {
		global.Lg.Error("DeleteSkiResort error", zap.Error(err))
		return err
	}
	return nil
}

func QuerySkiResortTeachTimeList(c *gin.Context, req *forms.QuerySkiResortTeachTimeListRequest) ([]*model.SkiResortsTeachTime, error) {
	userId := c.GetString("user_id")
	if req.UserID != "" {
		userId = req.UserID
	}
	db := global.DB.Table("ski_resorts_teach_time").
		Where("user_id = ?", userId).Where("state = 0")
	db = db.Preload("SkiResortsTeachTimeOrderCourses", "state = 0").Preload("SkiResortsTeachTimeEvent", "state = 0")
	if req.SkiResortsId != 0 {
		db = db.Where("ski_resorts_id = ?", req.SkiResortsId)
	}
	if req.TeachDateStart != "" {
		db = db.Where("teach_date >= ?", req.TeachDateStart)
	}
	if req.TeachDateEnd != "" {
		db = db.Where("teach_date <= ?", req.TeachDateEnd)
	}
	if req.OnlyDate == 1 {
		db.Select("teach_date,min(teach_start_time) as teach_start_time, max(teach_end_time) as teach_end_time, min(teach_num) as teach_num,max(teach_state) as teach_state").
			Group("teach_date").Order("teach_date asc")
	} else {
		db.Order("teach_start_time asc")
	}

	var skiResortTeachTimes []*model.SkiResortsTeachTime
	if err := db.Find(&skiResortTeachTimes).Error; err != nil {
		global.Lg.Error("QuerySkiResortTeachTimeList error", zap.Error(err))
		return nil, err
	}
	for _, v := range skiResortTeachTimes {
		if len(v.SkiResortsTeachTimeOrderCourses) > 0 {
			v.OrderCourseID = v.SkiResortsTeachTimeOrderCourses[0].OrderCourseID
		}
		if len(v.SkiResortsTeachTimeEvent) > 0 {
			v.Title = v.SkiResortsTeachTimeEvent[0].Title
			v.Remark = v.SkiResortsTeachTimeEvent[0].Remark
		}
	}
	return skiResortTeachTimes, nil
}

func QuerySkiResortTeachDateList(c *gin.Context, req *forms.QuerySkiResortTeachDateListRequest) ([]TeachDateData, error) {
	userId := c.GetString("user_id")
	if req.UserID != "" {
		userId = req.UserID
	}
	db := global.DB.Table("ski_resorts_teach_time").Where("user_id = ?", userId).Where("state = 0")
	if req.TeachDateStart != "" {
		db = db.Where("teach_date >= ?", req.TeachDateStart)
	}
	if req.TeachDateEnd != "" {
		db = db.Where("teach_date <= ?", req.TeachDateEnd)
	}
	db.Select("min(user_id) as user_id,min(ski_resorts_id) as ski_resorts_id,teach_date, min(teach_num) as teach_num,max(teach_state) as teach_state").
		Group("teach_date").Order("teach_date asc")

	var skiResortTeachTimes []*model.SkiResortsTeachTime
	if err := db.Find(&skiResortTeachTimes).Error; err != nil {
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
	skiTeachDates := make(map[string]*model.SkiResortsTeachTime, 0)
	for _, v := range skiResortTeachTimes {
		a := time.Time(v.TeachDate).Format("2006-01-02")
		skiTeachDates[a] = v
	}
	TeachDateDatas := make([]TeachDateData, 0)
	for teachDateStart.Before(teachDateEnd.Add(24 * time.Hour)) {
		d := teachDateStart.Format("2006-01-02")
		teachDateData := TeachDateData{
			TeachDate:    d,
			ConfigState:  0,
			SkiResortsId: 0,
			Switchable:   true,
		}
		if skiTeachDate, ok := skiTeachDates[d]; ok {
			teachDateData.SkiResortsId = skiTeachDate.SkiResortsID
			teachDateData.Switchable = false
			if skiTeachDate.SkiResortsID != req.SkiResortsId { //不是一个雪场，所以要锁定
				teachDateData.ConfigState = -1
			} else {
				if skiTeachDate.TeachState != 0 { //状态不为0，所以要锁定
					teachDateData.ConfigState = -1
				} else {
					if skiTeachDate.TeachNum == 0 { // teach_num为0，说明已经被预约，所以要锁定
						teachDateData.ConfigState = -1
					}
				}
			}
		}
		TeachDateDatas = append(TeachDateDatas, teachDateData)
		teachDateStart = teachDateStart.Add(24 * time.Hour)
	}
	return TeachDateDatas, nil
}

type TeachDateData struct {
	TeachDate    string `json:"teach_date"`
	ConfigState  int    `json:"config_state"` //0：未配置日期，-1：已预约、其他雪场的时间，1：已配置日期
	SkiResortsId int    `json:"ski_resorts_id"`
	Switchable   bool   `json:"switchable"` //true：可配置，false：不可配置
}

func CreateSkiResortTeachTime(c *gin.Context, req *forms.CreateSkiResortTeachTimeRequest) (err error) {
	var startTime, endTime string

	if req.TeachStartTime[3:] != "00" && req.TeachStartTime[3:] != "30" {
		return enum.NewErr(enum.TeachTimeErr, "时间格式错误")
	}

	if c.GetInt("user_type") != enum.UserTypeCoach {
		return enum.NewErr(enum.TeachTimeErr, "目前只支持教练添加教学时间")
	}
	userId := c.GetString("user_id")

	_, err = NewCoachSkiResortsDao(c, global.DB).Get(c, "*", "coach_id = ? and ski_resorts_id = ? and state = 0", userId, req.SkiResortsId)
	if err != nil {
		global.Lg.Error("CreateSkiResortTeachTime error", zap.Error(err))
		return errors.New("请先绑定雪场")
	}
	var myTeachDate []string
	global.DB.Model(&model.SkiResortsTeachTime{}).Where("user_id = ? and ski_resorts_id = ? and state = 0", userId, req.SkiResortsId).
		Where("teach_date in ?", req.TeachDates).
		Group("teach_date").
		Pluck("teach_date", &myTeachDate)
	if len(myTeachDate) > 0 {
		return errors.New("该时间段已存在，请勿重复添加")
	}

	TeachDateMap := make(map[string]struct{})
	for _, date := range req.TeachDates { //时间段去重
		TeachDateMap[date] = struct{}{}
	}
	var skiResortsTeachTimes []model.SkiResortsTeachTime
	var startTi, endTi, dateTi time.Time
	for date := range TeachDateMap {
		startTime = date + " " + req.TeachStartTime
		endTime = date + " " + req.TeachEndTime

		startTi, err = time.ParseInLocation("2006-01-02 15:04", startTime, time.Local)
		if err != nil {
			global.Lg.Error("Error parsing date:", zap.Error(err))
			return
		}
		if startTi.Before(time.Now()) {
			return enum.NewErr(enum.TeachTimeErr, "时间不能小于当前时间")
		}
		endTi, err = time.ParseInLocation("2006-01-02 15:04", endTime, time.Local)
		if err != nil {
			global.Lg.Error("Error parsing date:", zap.Error(err))
			return
		}
		dateTi, err = time.ParseInLocation("2006-01-02", date, time.Local)
		if err != nil {
			global.Lg.Error("Error parsing date:", zap.Error(err))
			return
		}
		for i := 0; i < 48; i++ {
			da := model.SkiResortsTeachTime{
				UserID:         c.GetString("user_id"),
				UserType:       c.GetInt("user_type"),
				SkiResortsID:   req.SkiResortsId,
				TeachNum:       1,
				TeachDate:      model.LocalDate(dateTi),
				TeachStartTime: model.LocalTime(startTi),
				TeachEndTime:   model.LocalTime(startTi.Add(30 * time.Minute)),
			}
			skiResortsTeachTimes = append(skiResortsTeachTimes, da)
			startTi = startTi.Add(30 * time.Minute)
			if startTi.After(endTi) {
				break
			}
		}
	}

	//只处理审核通过的俱乐部教学时间
	clubCoachs, err := NewClubsCoachesDao(c, global.DB).GetAll(c, "coach_id = ? and  state = 0 and  verified = 1", userId)
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(&skiResortsTeachTimes).Error
		if err != nil {
			global.Lg.Error("CreateSkiResortTeachTime error", zap.Error(err))
			return enum.NewErr(enum.TeachTimeErr, "添加教学时间失败")
		}
		//俱乐部教练添加教学时间，需要更新俱乐部的可用时间
		for _, clubCoach := range clubCoachs {
			err = ClubCreateSkiResortTeachTime(c, tx, clubCoach.ClubID, req.SkiResortsId, skiResortsTeachTimes)
			if err != nil {
				global.Lg.Error("ClubCreateSkiResortTeachTime error", zap.Error(err))
				return enum.NewErr(enum.TeachTimeErr, "添加俱乐部教学时间失败")
			}
		}
		return nil
	})
	return err
}
func ClubCreateSkiResortTeachTime(c *gin.Context, tx *gorm.DB, clubId string, skiResortsID int, skiResortsTeachTimes []model.SkiResortsTeachTime) (err error) {
	if len(skiResortsTeachTimes) == 0 {
		return
	}
	teachStartTimes := make([]model.LocalTime, 0)
	for _, v := range skiResortsTeachTimes {
		teachStartTimes = append(teachStartTimes, v.TeachStartTime)
	}
	tx.Model(&model.SkiResortsTeachTime{}).
		Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", clubId, enum.UserTypeClub, skiResortsID).
		Where("teach_start_time in ?", teachStartTimes).
		Pluck("teach_start_time", &teachStartTimes)
	if len(teachStartTimes) > 0 { //查出存在的时间，则更新教学次数
		tx.Model(&model.SkiResortsTeachTime{}).
			Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", clubId, enum.UserTypeClub, skiResortsID).
			Where("teach_start_time in ?", teachStartTimes).
			Update("teach_num", gorm.Expr("teach_num + ?", 1))
	}
	if len(teachStartTimes) == len(skiResortsTeachTimes) { //时间全部存在，则不处理
		return
	}
	teachStartTimeMap := make(map[model.LocalTime]struct{}, 0)
	for _, v := range teachStartTimes {
		teachStartTimeMap[v] = struct{}{}
	}
	clubTeachTimes := make([]model.SkiResortsTeachTime, 0)
	for _, v := range skiResortsTeachTimes {
		if _, ok := teachStartTimeMap[v.TeachStartTime]; ok {
			continue
		}
		clubTeachTimes = append(clubTeachTimes, model.SkiResortsTeachTime{
			UserID:         clubId,
			UserType:       enum.UserTypeClub,
			SkiResortsID:   skiResortsID,
			TeachDate:      v.TeachDate,
			TeachStartTime: v.TeachStartTime,
			TeachEndTime:   v.TeachEndTime,
			TeachNum:       1,
		})
	}
	err = tx.Create(&clubTeachTimes).Error
	return err
}

func UpdateSkiResortTeachState(c *gin.Context, req *forms.UpdateSkiResortTeachStateRequest) (err error) {
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	if userType != enum.UserTypeCoach && userType != enum.UserTypeClub {
		return enum.NewErr(enum.TeachTimeErr, "目前只支持教练和俱乐部修改教学时间")
	}
	if req.TeachState != nil {
		if *req.TeachState != 0 && *req.TeachState != 1 && *req.TeachState != 2 {
			return enum.NewErr(enum.TeachTimeErr, "教学状态不对")
		}
		err = global.DB.Model(&model.SkiResortsTeachTime{}).Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", userId, userType, req.SkiResortsId).
			Where("teach_start_time in ?", req.TeachStartTimes).
			Update("teach_state", req.TeachState).Error
		if err != nil {
			return err
		}
	}
	upData := map[string]string{}
	if req.Title != nil {
		upData["title"] = *req.Title
	}
	if req.Remark != nil {
		upData["remark"] = *req.Remark
	}
	var skirtIds []int64
	err = global.DB.Model(&model.SkiResortsTeachTime{}).Select("id").
		Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", userId, userType, req.SkiResortsId).
		Where("teach_start_time in ?", req.TeachStartTimes).
		Find(&skirtIds).Error
	if len(upData) > 0 {
		if len(skirtIds) > 0 {
			var skiResortsTeachTimeEvents []*model.SkiResortsTeachTimeEvent
			for _, id := range skirtIds {
				skiResortsTeachTimeEvents = append(skiResortsTeachTimeEvents, &model.SkiResortsTeachTimeEvent{
					SkirtID: id,
					Title:   upData["title"],
					Remark:  upData["remark"],
				})
			}
			err = global.DB.Model(&model.SkiResortsTeachTimeEvent{}).Create(skiResortsTeachTimeEvents).Error
			if err != nil {
				global.Lg.Error("CreateSkiResortsTeachTimeEvent error", zap.Error(err), zap.Any("skiResortsTeachTimeEvents", skiResortsTeachTimeEvents))
			}
		}
	}

	if req.TeachState != nil && *req.TeachState == 0 && userType == model.UserTypeClub && len(skirtIds) > 0 {
		err = global.DB.Model(&model.SkiResortsTeachTimeEvent{}).Where("skirt_id in ?", skirtIds).Update("state", 1).Error
		if err != nil {
			global.Lg.Error("UpdateSkiResortsTeachTimeEvent error", zap.Error(err), zap.Any("skiResortsTeachTimeEvents", skirtIds))
		}
	}

	return err
}
func DeleteSkiResortTeachTime(c *gin.Context, req *forms.DeleteSkiResortTeachTimeRequest) (err error) {
	//TeachDateMap := make(map[string]struct{})
	for _, date := range req.TeachDates { //时间段去重
		if date < time.Now().Format("2006-01-02") {
			return enum.NewErr(enum.TeachTimeErr, "时间不能小于当前时间")
		}
	}
	userId := c.GetString("user_id")
	userType := c.GetInt("user_type")
	if userType != enum.UserTypeCoach {
		return enum.NewErr(enum.TeachTimeErr, "目前只支持教练删除教学时间")
	}
	err = global.DB.Model(&model.SkiResortsTeachTime{}).Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", userId, userType, req.SkiResortsId).
		Where("teach_date in ? and teach_num = 0", req.TeachDates).First(&model.SkiResortsTeachTime{}).Error
	if err == nil {
		return enum.NewErr(enum.TeachTimeErr, "该时间段已被预约，无法删除")
	}

	teachStartTimes := make([]model.LocalTime, 0)
	err = global.DB.Model(&model.SkiResortsTeachTime{}).Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", userId, userType, req.SkiResortsId).
		Where("teach_date in ?", req.TeachDates).
		Pluck("teach_start_time", &teachStartTimes).Error

	//只处理审核通过的俱乐部教学时间
	clubCoachs, err := NewClubsCoachesDao(c, global.DB).GetAll(c, "coach_id = ? and  state = 0  and  verified = 1", userId)
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(&model.SkiResortsTeachTime{}).Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", userId, userType, req.SkiResortsId).
			Where("teach_date in ?", req.TeachDates).
			Update("state", 1).Error
		if err != nil {
			global.Lg.Error("CreateSkiResortTeachTime error", zap.Error(err))
			return enum.NewErr(enum.TeachTimeErr, "添加教学时间失败")
		}
		//俱乐部教练删除教学时间，需要更新俱乐部的可用时间
		for _, clubCoach := range clubCoachs {
			err = tx.Model(&model.SkiResortsTeachTime{}).Where("user_id = ? and state = 0 and user_type = ? and ski_resorts_id = ?", clubCoach.ClubID, enum.UserTypeClub, req.SkiResortsId).
				Where("teach_start_time in ?", teachStartTimes).
				Update("teach_num", gorm.Expr("teach_num - ?", 1)).Error

			if err != nil {
				global.Lg.Error("DeleteSkiResortTeachTime error", zap.Error(err))
				return enum.NewErr(enum.TeachTimeErr, "更新俱乐部教学时间失败")
			}
		}
		return nil
	})

	return err
}

type SimpleScheduleEventData struct {
	OrderCourseID string `json:"order_course_id"`
	Title         string `json:"title"`
	Remark        string `json:"remark"`
	TeachState    int    `json:"teach_state"`
	StartTime     string `json:"start_time"` // 修改为字符串类型，只包含时分
	EndTime       string `json:"end_time"`   // 修改为字符串类型，只包含时分
}

type srtEvent struct {
	TeachStartTime time.Time `json:"teach_start_time"` // 教学开始时间
	TeachEndTime   time.Time `json:"teach_end_time"`   // 教学结束时间
	Title          string    `json:"title"`            // 标题
	Remark         string    `json:"remark"`           // 备注
	StartTime      string    `json:"start_time"`       // 修改为字符串类型，只包含时分
	EndTime        string    `json:"end_time"`         // 修改为字符串类型，只包含时分
}

type srtOrderCourses struct {
	OrderCourseID  string    `json:"order_course_id"`
	TeachStartTime time.Time `json:"teach_start_time"` // 教学开始时间
	TeachEndTime   time.Time `json:"teach_end_time"`   // 教学结束时间
	StartTime      string    `json:"start_time"`       // 修改为字符串类型，只包含时分
	EndTime        string    `json:"end_time"`         // 修改为字符串类型，只包含时分
}

func ScheduleEvent(c *gin.Context, teachDate, userId string, skiResortsId int) (interface{}, error) {
	var modelScheduleEvents []*srtEvent
	err := global.DB.Model(&model.SkiResortsTeachTime{}).
		Select("min(ski_resorts_teach_time.teach_start_time) as teach_start_time, max(ski_resorts_teach_time.teach_end_time) as teach_end_time, max(srtte.title) as title, max(srtte.remark) as remark").
		Joins("join ski_resorts_teach_time_event as srtte on srtte.skirt_id = ski_resorts_teach_time.id  and srtte.state=0").
		Where("user_id = ? and teach_date=? and ski_resorts_id=?", userId, teachDate, skiResortsId).
		Group("srtte.title").
		Find(&modelScheduleEvents).Error
	if err != nil {
		global.Lg.Error("ScheduleEvent error", zap.Error(err))
		return nil, err
	}
	for _, v := range modelScheduleEvents {
		v.StartTime = v.TeachStartTime.Format("15:04")
		v.EndTime = v.TeachEndTime.Format("15:04")
	}

	var srtOCs []*srtOrderCourses
	err = global.DB.Model(&model.SkiResortsTeachTime{}).
		Select("min(ski_resorts_teach_time.teach_start_time) as teach_start_time, max(ski_resorts_teach_time.teach_end_time) as teach_end_time, max(srtte.order_course_id) as order_course_id").
		Joins("join ski_resorts_teach_time_order_courses as srtte on srtte.skirt_id = ski_resorts_teach_time.id  and srtte.state=0").
		Where("user_id = ? and teach_date=? and ski_resorts_id=?", userId, teachDate, skiResortsId).
		Group("srtte.order_course_id").
		Find(&srtOCs).Error
	if err != nil {
		global.Lg.Error("ScheduleEvent error", zap.Error(err))
		return nil, err
	}
	for _, v := range srtOCs {
		v.StartTime = v.TeachStartTime.Format("15:04")
		v.EndTime = v.TeachEndTime.Format("15:04")
	}

	data := map[string]interface{}{}
	data["order_courses"] = srtOCs
	data["events"] = modelScheduleEvents
	return data, nil
}
