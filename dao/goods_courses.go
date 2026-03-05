package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

type GoodsCoursesDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.GoodsCourses
}

func NewGoodsCoursesDao(ctx context.Context, dbs ...*gorm.DB) *GoodsCoursesDao {
	dao := new(GoodsCoursesDao)
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
func (d *GoodsCoursesDao) SaveAll(ctx context.Context, obj []*model.GoodsCourses) error {
	err := d.sourceDB.Model(d.m).Save(&obj).Error
	if err != nil {
		return fmt.Errorf("GoodsCoursesDao: %w", err)
	}
	return nil
}

func (d *GoodsCoursesDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("GoodsCoursesDao:Update where=%s: %w", where, err)
	}
	return nil
}

// GetCourseOnShelf 查询商品关联课程是否存在下架的课程
func (d *GoodsCoursesDao) GetCourseOnShelf(ctx context.Context, good *model.Goods, onShelf int) int64 {
	var count int64
	if good.Pack == model.PackNo {
		global.DB.Model(model.Goods{}).Where("good_id = ? and goods.state = 0", good.GoodID).
			Joins("join courses on courses.course_id = goods.course_id and courses.state = 0 and courses.on_shelf = ?", onShelf).Count(&count)
	} else {
		d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Where("good_id = ? and goods_courses.state = 0", good.GoodID).
			Joins("join courses on courses.course_id = goods_courses.course_id and courses.state = 0 and on_shelf = ?", onShelf).Count(&count)
		if count == 0 {
			d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Where("goods_courses.good_id = ? and goods_courses.state = 0", good.GoodID).
				Joins("join goods on goods.good_id = goods_courses.pack_good_id and goods.state = 0 and on_shelf = ?", onShelf).Count(&count)
		}
	}

	return count
}

// GetCourseTags 获取商品关联课程标签
func (d *GoodsCoursesDao) GetCourseTags(ctx context.Context, good *model.Goods) ([]int64, error) {
	var tags []int64
	goodId := good.GoodID
	if good.Pack == model.PackNo {
		err := global.DB.Model(model.Goods{}).Where("good_id = ? and  goods.state = 0", goodId).
			Joins("join courses_tags on courses_tags.course_id = goods.course_id and  courses_tags.state = 0").
			Distinct("tag_id").Pluck("tag_id", &tags).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("GoodsCoursesDao: GetCourseTags where=%s: %w", goodId, err)
		}
	} else {
		err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).Where("goods_courses.good_id = ? and  goods_courses.state = 0", goodId).
			Joins("join courses_tags on courses_tags.course_id = goods_courses.course_id and  courses_tags.state = 0").
			Distinct("tag_id").Pluck("tag_id", &tags).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("GoodsCoursesDao: GetCourseTags where=%s: %w", goodId, err)
		}
	}

	return tags, nil
}

func (d *GoodsCoursesDao) Create(ctx context.Context, obj *model.GoodsCourses) error {
	err := d.sourceDB.Model(d.m).Create(&obj).Error
	if err != nil {
		return fmt.Errorf("GoodsCoursesDao: %w", err)
	}
	return nil
}

func (d *GoodsCoursesDao) Get(ctx context.Context, fields, where string) (*model.GoodsCourses, error) {
	items, err := d.List(ctx, fields, where, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("GoodsCoursesDao: Get where=%s: %w", where, err)
	}
	if len(items) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &items[0], nil
}

func (d *GoodsCoursesDao) List(ctx context.Context, fields, where string, offset, limit int) ([]model.GoodsCourses, error) {
	var results []model.GoodsCourses
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("GoodsCoursesDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *GoodsCoursesDao) Delete(ctx context.Context, where string, args ...interface{}) error {
	if len(where) == 0 {
		return gorm.ErrInvalidField
	}
	if err := d.sourceDB.Where(where, args...).Delete(d.m).Error; err != nil {
		return fmt.Errorf("GoodsCoursesDao: Delete where=%s: %w", where, err)
	}
	return nil
}
