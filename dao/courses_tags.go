package dao

import (
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
)

//func UpdateOrCreateCourseTag(courseId int64, tagIds []int64) error {
//	var courseTags []*model.CoursesTags
//	if err := global.DB.Where("course_id = ?", courseId).Find(&courseTags).Error; err != nil {
//		global.Lg.Error("查询课程标签失败", zap.Error(err))
//		return err
//	}
//
//}

func QueryCoursesTagsCountByTagId(tagId int64) (int64, error) {
	var count int64
	if err := global.DB.Model(&model.CoursesTags{}).Where("tag_id = ? and state = 0", tagId).Count(&count).Error; err != nil {
		global.Lg.Error("查询课程标签失败", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func QueryCoursesByTagId(tagId int64) ([]string, error) {
	var courseIds []string
	if err := global.DB.Model(&model.CoursesTags{}).Where("tag_id = ? and state = 0", tagId).Pluck("course_id", &courseIds).Error; err != nil {
		global.Lg.Error("查询课程标签失败", zap.Error(err))
		return nil, err
	}
	return courseIds, nil
}
