package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"skis-admin-backend/model"
)

type OrdersCoursesStateDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.OrdersCoursesState
}

func NewOrdersCoursesStateDao(ctx context.Context, dbs ...*gorm.DB) *OrdersCoursesStateDao {
	dao := new(OrdersCoursesStateDao)
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
func (d *OrdersCoursesStateDao) Last(ctx context.Context, fields, where string, args ...interface{}) (item *model.OrdersCoursesState, err error) {
	err = d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where, args...).Last(&item).Error
	return item, err
}

func (d *OrdersCoursesStateDao) Create(ctx context.Context, obj *model.OrdersCoursesState) error {
	err := d.sourceDB.Model(d.m).Create(&obj).Error
	if err != nil {
		return fmt.Errorf("OrdersCoursesStateDao: %w", err)
	}
	return nil
}

func (d *OrdersCoursesStateDao) Get(ctx context.Context, fields, where string) (*model.OrdersCoursesState, error) {
	items, err := d.List(ctx, fields, where, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("OrdersCoursesStateDao: Get where=%s: %w", where, err)
	}
	if len(items) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &items[0], nil
}

func (d *OrdersCoursesStateDao) List(ctx context.Context, fields, where string, offset, limit int) ([]model.OrdersCoursesState, error) {
	var results []model.OrdersCoursesState
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("OrdersCoursesStateDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *OrdersCoursesStateDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("OrdersCoursesStateDao:Update where=%s: %w", where, err)
	}
	return nil
}
