package dao

import (
	"context"
	"errors"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func QueryOrdersCoursesRecords(req *forms.QueryOrdersCoursesRecordsRequest) (int64, []*model.OrdersCoursesRecords, error) {
	db := global.DB.Model(&model.OrdersCoursesRecords{}).
		Preload("UserInfo", func(db *gorm.DB) *gorm.DB {
			return model.ScopeUserSensitiveFields(db)
		}).
		Preload("Good").
		Preload("Good.CourseTags", "state = 0").
		Preload("Good.CourseTags.Tag", "state = 0").
		Where("state = ?", model.StateNormal)
	if req.Uid != "" {
		db = db.Where("uid = ?", req.Uid)
	}

	if req.OrderId != "" {
		db = db.Where("order_id = ?", req.OrderId)
	}

	if req.OrderCourseId != "" {
		db = db.Where("order_course_id = ?", req.OrderCourseId)
	}

	if req.GoodId != "" {
		db = db.Where("good_id = ?", req.GoodId)
	}

	if req.CourseId != "" {
		db = db.Where("course_id = ?", req.CourseId)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("QueryOrdersCoursesRecords Count error: %v", zap.Error(err))
		return 0, nil, err
	}

	var ordersCoursesRecords []*model.OrdersCoursesRecords
	if err := db.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&ordersCoursesRecords).Error; err != nil {
		global.Lg.Error("QueryOrdersCoursesRecords Find error: %v", zap.Error(err))
		return 0, nil, err
	}

	for _, v := range ordersCoursesRecords {
		dealGoodTags(v.Good)
	}

	return total, ordersCoursesRecords, nil
}

func CreateOrdersCoursesRecord(ctx context.Context, coachId, orderCoursesId string, req *forms.CreateOrdersCoursesRecordRequest) error {
	orderCourse, err := OrderCourseInfo(ctx, "", orderCoursesId)
	if err != nil {
		global.Lg.Error("CreateOrdersCoursesRecord QueryOrderCourseInfo error: %v", zap.Error(err))
		return err
	}

	if coachId != orderCourse.TeachCoachID {
		global.Lg.Error("CreateOrdersCoursesRecord coachId not match TeachCoachID", zap.String("coachId", coachId), zap.String("TeachCoachID", orderCourse.TeachCoachID))
		return enum.NewErr(enum.AuthErr, "没有权限填写记录")
	}

	order, err := QueryOrderInfo("", orderCourse.OrderID)
	if err != nil {
		global.Lg.Error("CreateOrdersCoursesRecord QueryOrderInfo error: %v", zap.Error(err))
		return err
	}

	//查询之前是否已经填写了课程记录
	var ocr *model.OrdersCoursesRecords
	err = global.DB.Model(&model.OrdersCoursesRecords{}).Where("order_course_id = ? and coach_id = ?", orderCourse.OrderCourseID, coachId).First(&ocr).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Lg.Error("CreateOrdersCoursesRecord already wrote", zap.String("order_course_id", orderCourse.OrderCourseID))
		return err
	}

	var records []*model.OrdersCoursesRecords
	for _, v := range req.Records {
		records = append(records, &model.OrdersCoursesRecords{
			Uid:           orderCourse.Uid,
			OrderCourseId: orderCourse.OrderCourseID,
			OrderId:       orderCourse.OrderID,
			GoodId:        orderCourse.GoodID,
			CourseId:      orderCourse.CourseID,
			CoachId:       coachId,
			Content:       v.Content,
			Urls:          v.Urls,
		})
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(&model.OrdersCoursesRecords{}).Create(&records).Error; err != nil {
			global.Lg.Error("CreateOrdersCoursesRecord Create error: %v", zap.Error(err))
			return err
		}

		if ocr != nil { //之前已经填写过，不重复统计
			return nil
		}

		// 单次课程写多条，统计只算一次
		if err = AddCoachWroteCourseRecord(ctx, tx, coachId, 1); err != nil {
			global.Lg.Error("CreateOrdersCoursesRecord AddCoachFinishedCourse error: %v", zap.Error(err))
			return err
		}

		//如果课程是俱乐部的，那俱乐部的填写记录也要+1
		if order.UserType == model.UserTypeClub {
			if err = AddClubWroteCourseRecord(ctx, tx, order.UserID, 1); err != nil {
				global.Lg.Error("CreateOrdersCoursesRecord AddClubWroteCourseRecord error: %v", zap.Error(err))
				return err
			}
		}

		return nil
	})

	if err != nil {
		global.Lg.Error("CreateOrdersCoursesRecord Transaction error: %v", zap.Error(err))
		return err
	}

	return nil
}

func DeleteOrdersCoursesRecord(ctx context.Context, orderCourseRecordId int64, uid string) error {
	if err := global.DB.Model(&model.OrdersCoursesRecords{}).Where("id = ? AND uid = ?", orderCourseRecordId, uid).Update("state", model.StateDeleted).Error; err != nil {
		global.Lg.Error("DeleteOrdersCoursesRecord Delete error: %v", zap.Error(err))
		return err
	}
	return nil
}
