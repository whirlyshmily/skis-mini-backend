package dao

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

func QueryTagsList(req *forms.QueryTagsListRequest) (int64, []*model.TagsList, error) {
	var count int64

	var tags []*model.TagsList
	db := global.DB.Table("tags").Where("tags.state = 0")
	if req.Keyword != "" {
		db = db.Where("tags.name like ?", "%"+req.Keyword+"%")
	}

	if err := db.Count(&count).Error; err != nil {
		global.Lg.Error("查询标签列表失败", zap.Error(err))
		return 0, nil, err
	}

	//这里要查询tag在courses_tag表中的数量
	result := db.Select("tags.id, tags.name, tags.state, tags.created_at, tags.updated_at, count(courses_tags.id) as course_ref_cnt, count(coaches_tags.id) as coach_ref_cnt").
		Joins("LEFT JOIN courses_tags ON tags.id = courses_tags.tag_id and courses_tags.state = 0").
		Joins("LEFT JOIN coaches_tags ON tags.id = coaches_tags.tag_id and coaches_tags.state = 0").
		Group("tags.id").Order("id desc").Offset((req.Page - 1) * req.PageSize).Find(&tags)

	global.Lg.Info("sql", zap.Any("sql", db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx
	})))

	if result.Error != nil {
		global.Lg.Error("查询标签列表失败", zap.Error(result.Error))
		return 0, nil, result.Error
	}

	return count, tags, nil
}

func CreateTag(req forms.CreateTagRequest) (*model.Tags, error) {
	//检查是否存在
	var existTag *model.Tags
	err := global.DB.Where("name = ? and state = 0", req.Name).First(&existTag).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		global.Lg.Error("创建标签失败", zap.Error(err))
		return nil, err
	}

	if err == nil {
		global.Lg.Error("创建标签失败", zap.Any("tag exist", existTag))
		return nil, enum.NewErr(enum.TagExistErr, "标签已存在")
	}

	tag := &model.Tags{
		Name: req.Name,
	}
	if err := global.DB.Create(tag).Error; err != nil {
		global.Lg.Error("创建标签失败", zap.Error(err))
		return nil, err
	}
	return tag, nil
}

func QueryTagInfo(id int64) (*model.Tags, error) {
	var tag *model.Tags
	if err := global.DB.Where("id = ? and state = 0", id).First(&tag).Error; err != nil {
		global.Lg.Error("查询标签信息失败", zap.Error(err))
		return nil, err
	}
	return tag, nil
}

func UpdateTag(id int64, req *forms.CreateTagRequest) error {
	//判断标签是否存在
	tag, err := QueryTagInfo(id)
	if err != nil {
		global.Lg.Error("查询标签失败", zap.Error(err))
		return err
	}

	//判断文件名是否存在
	var existTag *model.Tags
	err = global.DB.Where("id != ? and name = ? and state = 0", id, req.Name).First(&existTag).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		global.Lg.Error("查询标签失败", zap.Error(err))
		return err
	}

	if err == nil {
		global.Lg.Error("更新标签失败", zap.Any("tag exist", tag))
		return enum.NewErr(enum.TagExistErr, "标签已存在")
	}

	tag.Name = req.Name
	if err := global.DB.Save(tag).Error; err != nil {
		global.Lg.Error("更新标签失败", zap.Error(err))
		return err
	}

	//修改标签名称
	if err := global.DB.Model(&model.CoursesTags{}).Where("tag_id = ?", id).Update("tag_name", req.Name).Error; err != nil {
		global.Lg.Error("更新标签失败", zap.Error(err))
		return err
	}
	//修改标签名称
	if err := global.DB.Model(&model.CoachesTags{}).Where("tag_id = ?", id).Update("tag_name", req.Name).Error; err != nil {
		global.Lg.Error("更新标签失败", zap.Error(err))
		return err
	}
	return nil
}

func DeleteTag(id int64) error {
	//判断标签是否存在
	tag, err := QueryTagInfo(id)
	if err != nil {
		global.Lg.Error("查询标签失败", zap.Error(err))
		return err
	}

	//判断标签是否被引用
	courseRefCount, err := QueryCoursesTagsCountByTagId(id)
	if err != nil {
		global.Lg.Error("查询标签失败", zap.Error(err))
		return err
	}

	if courseRefCount > 0 {
		global.Lg.Error("删除标签失败", zap.Any("tag ref count", courseRefCount))
		return enum.NewErr(enum.CourseRefTagErr, "标签已被课程引用")
	}

	//判断标签是否被引用
	coachRefCount, err := QueryCoachesTagCountByTagId(id)
	if err != nil {
		global.Lg.Error("查询标签失败", zap.Error(err))
		return err
	}

	if coachRefCount > 0 {
		global.Lg.Error("删除标签失败", zap.Any("coach ref count", coachRefCount))
		return enum.NewErr(enum.CoachRefTagErr, "标签已被引用")
	}

	//删除标签
	tag.State = 1
	if err := global.DB.Save(tag).Error; err != nil {
		global.Lg.Error("删除标签失败", zap.Error(err))
		return err
	}

	return nil
}
