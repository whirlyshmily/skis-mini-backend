package dao

import (
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
)

func QueryCoursesList(req *forms.QueryCoursesListRequest) (int64, []*model.Courses, error) {
	var courses []*model.Courses
	var total int64
	query := global.DB.Model(&model.Courses{}).
		Preload("CoursesTags", "state = 0").
		Preload("CoursesTags.Tag", "state = 0").
		Where("state = 0")
	if req.Keyword != "" {
		query = query.Where("(title like ? or course_id like ?)", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		global.Lg.Error("QueryCoursesList error", zap.Error(err))
		return 0, nil, err
	}

	query = query.Order("id desc")
	if req.Page != 0 && req.PageSize != 0 {
		query = query.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize)
	}

	if err := query.Find(&courses).Error; err != nil {
		global.Lg.Error("QueryCoursesList error", zap.Error(err))
		return 0, nil, err
	}

	for _, course := range courses {
		dealCourseTags(course)
	}

	return total, courses, nil
}

func QueryCourseInfo(courseId string) (*model.Courses, error) {
	var course *model.Courses
	if err := global.DB.Model(&model.Courses{}).
		Preload("CoursesTags", "state = 0").
		Preload("CoursesTags.Tag", "state = 0").
		Where("course_id = ? and state = 0 and on_shelf = 1", courseId).First(&course).Error; err != nil {
		global.Lg.Error("QueryCourseInfo error", zap.Error(err))
		return nil, err
	}

	dealCourseTags(course)
	return course, nil
}

func dealCourseTags(course *model.Courses) {
	var tags []*model.Tags
	for _, tag := range course.CoursesTags {
		tags = append(tags, tag.Tag)
	}
	course.Tags = tags
}

func QueryCourseByIds(courseIds []string) ([]*model.Courses, error) {
	var courses []*model.Courses
	if err := global.DB.Where("course_id in ? and state = 0", courseIds).Find(&courses).Error; err != nil {
		global.Lg.Error("QueryCourseByIds error", zap.Error(err))
		return nil, err
	}
	return courses, nil
}
