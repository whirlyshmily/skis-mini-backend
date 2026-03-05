package dao

import (
	"context"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func UserActive(ctx context.Context) error {
	if err := UserActiveDay(ctx); err != nil {
		global.Lg.Error("UserActive: UserActiveDay", zap.Error(err))
		return err
	}
	if err := UserActiveMonth(ctx); err != nil {
		global.Lg.Error("UserActive: UserActiveMonth", zap.Error(err))
		return err
	}
	return nil
}

func UserActiveDay(ctx context.Context) error {
	d := time.Now().Format("2006-01-02")
	var usersActiveDay model.UsersActiveDay
	if err := global.DB.Model(&model.UsersActiveDay{}).Where("date = ?", d).FirstOrCreate(&usersActiveDay, model.UsersActiveDay{
		Date: d,
		Cnt:  0,
	}).Error; err != nil {
		global.Lg.Error("UserActive: FirstOrCreate", zap.Error(err), zap.Any("data", usersActiveDay))
		return err
	}

	usersActiveDay.Cnt++
	if err := global.DB.Model(&model.UsersActiveDay{}).Where("date = ?", d).Save(&usersActiveDay).Error; err != nil {
		global.Lg.Error("UserActive: Updates", zap.Error(err), zap.Any("data", usersActiveDay))
		return err
	}
	return nil
}

func UserActiveMonth(ctx context.Context) error {
	d := time.Now().Format("2006-01")
	var usersActiveMonth model.UsersActiveMonth
	if err := global.DB.Model(&model.UsersActiveMonth{}).Where("date = ?", d).FirstOrCreate(&usersActiveMonth, model.UsersActiveMonth{
		Date: d,
		Cnt:  0,
	}).Error; err != nil {
		global.Lg.Error("UserActive: FirstOrCreate", zap.Error(err), zap.Any("data", usersActiveMonth))
		return err
	}

	usersActiveMonth.Cnt++
	if err := global.DB.Model(&model.UsersActiveMonth{}).Where("date = ?", d).Save(&usersActiveMonth).Error; err != nil {
		global.Lg.Error("UserActive: Updates", zap.Error(err), zap.Any("data", usersActiveMonth))
		return err
	}
	return nil
}

func OrderStatisticsDay(ctx context.Context, tx *gorm.DB, buyerCount, totalAmount int64, d string) error {
	var orderDayStatistics *model.OrdersDayStatistics
	if err := tx.Model(&model.OrdersDayStatistics{}).Where("date = ?", d).FirstOrCreate(&orderDayStatistics, model.OrdersDayStatistics{
		Date:        d,
		BuyerCount:  0,
		TotalAmount: 0,
	}).Error; err != nil {
		global.Lg.Error("OrdersDayStatistics: FirstOrCreate", zap.Error(err), zap.Any("data", orderDayStatistics))
		return err
	}

	if err := tx.Model(&model.OrdersDayStatistics{}).Where("date = ?", d).Updates(map[string]interface{}{
		"buyer_count":  gorm.Expr("buyer_count + ?", buyerCount),
		"total_amount": gorm.Expr("total_amount + ?", totalAmount),
	}).Error; err != nil {
		global.Lg.Error("OrdersDayStatistics: Updates", zap.Error(err), zap.Any("data", orderDayStatistics))
		return err
	}
	return nil
}

func OrderStatisticsMonth(ctx context.Context, tx *gorm.DB, buyerCount, totalAmount int64, d string) error {
	var orderMonthStatistics *model.OrdersMonthStatistics
	if err := tx.Model(&model.OrdersMonthStatistics{}).Where("date = ?", d).FirstOrCreate(&orderMonthStatistics, model.OrdersMonthStatistics{
		Date:        d,
		BuyerCount:  0,
		TotalAmount: 0,
	}).Error; err != nil {
		global.Lg.Error("OrdersDayStatistics: FirstOrCreate", zap.Error(err), zap.Any("data", orderMonthStatistics))
		return err
	}

	if err := tx.Model(&model.OrdersMonthStatistics{}).Where("date = ?", d).Updates(map[string]interface{}{
		"buyer_count":  gorm.Expr("buyer_count + ?", buyerCount),
		"total_amount": gorm.Expr("total_amount + ?", totalAmount),
	}).Error; err != nil {
		global.Lg.Error("OrdersMonthStatistics: Updates", zap.Error(err), zap.Any("data", orderMonthStatistics))
		return err
	}
	return nil
}

func OrderStatistics(ctx context.Context, tx *gorm.DB, buyerCount, totalAmount int64, d time.Time) error {
	if err := OrderStatisticsDay(ctx, tx, buyerCount, totalAmount, d.Format("2006-01-02")); err != nil {
		global.Lg.Error("OrderStatistics: OrderStatisticsDay", zap.Error(err))
		return err
	}
	if err := OrderStatisticsMonth(ctx, tx, buyerCount, totalAmount, d.Format("2006-01")); err != nil {
		global.Lg.Error("OrderStatistics: OrderStatisticsMonth", zap.Error(err))
		return err
	}
	return nil
}
