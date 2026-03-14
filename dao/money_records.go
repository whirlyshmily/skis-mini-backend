package dao

import (
	"context"
	"fmt"
	"math/rand"
	"skis-admin-backend/enum"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MoneyRecordsDao struct {
	sourceDB  *gorm.DB
	replicaDB []*gorm.DB
	m         *model.MoneyRecords
}

func NewMoneyRecordsDao(ctx context.Context, dbs ...*gorm.DB) *MoneyRecordsDao {
	dao := new(MoneyRecordsDao)
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

func (d *MoneyRecordsDao) Create(ctx context.Context, obj *model.MoneyRecords, tx *gorm.DB) error {
	if obj.Money <= 0 {
		return enum.NewErr(enum.MoneyRecordsCreateFailedErr, "创建积分记录失败，金额必须大于0")
	}
	if obj.UserID == "" {
		return enum.NewErr(enum.MoneyRecordsCreateFailedErr, "创建积分记录失败，用户ID不能为空")
	}
	if model.UserMoneyTypeStr[obj.MoneyType] != "" {
		obj.MoneyDesc = model.UserMoneyTypeStr[obj.MoneyType]
	}
	obj.MoneyID = GenerateId("ZJ")
	obj.MoneyDesc = model.UserMoneyTypeStr[obj.MoneyType]
	obj.Remark = model.UserMoneyTypeRemark[obj.MoneyType]

	err := tx.Model(d.m).Create(&obj).Error
	if err != nil {
		return enum.NewErr(enum.MoneyRecordsCreateFailedErr, "创建积分记录失败")
	}

	updateData := map[string]interface{}{}
	//资金变动记录（订单相关）
	if obj.RelationType == model.RelationTypeOrder {
		if obj.IncomeType == model.IncomeTypeIncome {
			updateData["total_profit"] = gorm.Expr("total_profit + ?", obj.Money)
			updateData["balance"] = gorm.Expr("balance + ?", obj.Money)
		} else if obj.IncomeType == model.IncomeTypePay { //目前支出的情况都是平台服务费
			updateData["paid_service_fee"] = gorm.Expr("paid_service_fee + ?", obj.Money)
			updateData["balance"] = gorm.Expr("balance - ?", obj.Money)
		}
	}

	//资金变动记录（保证金相关）
	if obj.RelationType == model.RelationTypeDeposit {
		if obj.IncomeType == model.IncomeTypeIncome {
			updateData["deposit"] = gorm.Expr("deposit + ?", obj.Money) // 收入：保证金增加
		} else if obj.IncomeType == model.IncomeTypePay {
			updateData["deposit"] = gorm.Expr("deposit - ?", obj.Money) // 支出：保证金减少
		}
	}

	if obj.UserType == enum.UserTypeCoach { //教练
		err = tx.Model(&model.Coaches{}).Where("coach_id = ? and state = 0", obj.UserID).Updates(updateData).Error
		if err != nil {
			global.Lg.Error("更新教练失败", zap.Error(err), zap.Any("data", obj))
			return err
		}
	}
	if obj.UserType == enum.UserTypeClub { //俱乐部
		if err = tx.Model(&model.Clubs{}).Where("club_id = ? and state = 0", obj.UserID).Updates(updateData).Error; err != nil {
			global.Lg.Error("更新俱乐部失败", zap.Error(err), zap.Any("data", obj))
			return enum.NewErr(enum.ClubUpdateErr, "更新俱乐部失败")
		}
	}
	return nil
}

func (d *MoneyRecordsDao) Get(ctx context.Context, fields, where string) (*model.MoneyRecords, error) {
	items, err := d.List(ctx, fields, where, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("MoneyRecordsDao: Get where=%s: %w", where, err)
	}
	if len(items) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &items[0], nil
}

func (d *MoneyRecordsDao) List(ctx context.Context, fields, where string, offset, limit int, args ...interface{}) ([]model.MoneyRecords, error) {
	var results []model.MoneyRecords
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(d.m).
		Select(fields).Where(where, args...).Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("MoneyRecordsDao: List where=%s: %w", where, err)
	}
	return results, nil
}

func (d *MoneyRecordsDao) Update(ctx context.Context, where string, update map[string]interface{}, args ...interface{}) error {
	err := d.sourceDB.Model(d.m).Where(where, args...).
		Updates(update).Error
	if err != nil {
		return fmt.Errorf("MoneyRecordsDao:Update where=%s: %w", where, err)
	}
	return nil
}

func (d *MoneyRecordsDao) Statistics(ctx context.Context, startTime, endTime string) (int64, int64, error) {
	//查询用户下所有商品的最大价格
	type result struct {
		BuyerCount  int64 `gorm:"column:buyer_count"`
		TotalAmount int64 `gorm:"column:total_amount"`
	}

	var res *result
	err := d.replicaDB[rand.Intn(len(d.replicaDB))].Model(&d.m).
		Where("money_type = 1000 and relation_type = 0 and created_at >= ? and created_at < ?", startTime, endTime).
		Select("count(1) as buyer_count, sum(money) as total_amount").
		Scan(&res).Error
	if err != nil {
		global.Lg.Error("查询商品价格范围失败", zap.Error(err))
		return 0, 0, err
	}

	return res.BuyerCount, res.TotalAmount, nil
}
