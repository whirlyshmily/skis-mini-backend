package dao

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

func SRTOrderCourses(tx *gorm.DB, skirtIds []int64, orderCourseId string, state int) (err error) {
	if len(skirtIds) == 0 {
		return nil
	}
	if state == 1 { //取消预约
		err = tx.Model(model.SkiResortsTeachTime{}).Where("id in ?", skirtIds).
			Updates(map[string]interface{}{
				"teach_num": gorm.Expr("teach_num + ?", 1),
			}).Error
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SkiResortsTeachTime: %w", zap.Error(err), zap.Any("skirtIds", skirtIds), zap.Any("orderCourseId", orderCourseId))
			return err
		}

		//课后缓冲时间也要恢复
		err = tx.Model(model.SkiResortsTeachTime{}).Where("id in ? and teach_state= ? ", skirtIds, model.SkiTeachStateAfterClass).
			Updates(map[string]interface{}{
				"teach_state": model.SkiTeachStateWaitAppointment,
			}).Error
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SkiTeachStateAfterClass: %w", zap.Error(err), zap.Any("skirtIds", skirtIds), zap.Any("orderCourseId", orderCourseId))
			return err
		}

		err = tx.Table("ski_resorts_teach_time_order_courses").
			Where("skirt_id in ? and order_course_id = ? and state = 0", skirtIds, orderCourseId).Update("state", 1).Error
	} else { //预约课程，缓冲时间在外面处理
		err = tx.Model(model.SkiResortsTeachTime{}).Where("id in ?", skirtIds).
			Updates(map[string]interface{}{"teach_num": gorm.Expr("teach_num - ?", 1)}).Error
		if err != nil {
			global.Lg.Error("OrdersCoursesDao: SkiResortsTeachTime: %w", zap.Error(err), zap.Any("skirtIds", skirtIds), zap.Any("orderCourseId", orderCourseId))
			return err
		}

		var sRTData []*model.SkiResortsTeachTimeOrderCourses
		for _, id := range skirtIds {
			sRTData = append(sRTData, &model.SkiResortsTeachTimeOrderCourses{
				SkirtID:       id,
				OrderCourseID: orderCourseId,
			})
		}
		err = global.DB.Model(&model.SkiResortsTeachTimeOrderCourses{}).Create(sRTData).Error
	}
	return err
}
