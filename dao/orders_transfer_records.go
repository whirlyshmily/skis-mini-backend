package dao

import (
	"context"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateOrdersTransferRecords(ctx context.Context, tx *gorm.DB, orderId string, previousUserId string, previousUserType int, curUserId string, curUserType int) (*model.OrdersTransferRecords, error) {
	r := &model.OrdersTransferRecords{
		OrderId:          orderId,
		PreviousUserId:   previousUserId,
		PreviousUserType: previousUserType,
		CurUserId:        curUserId,
		CurUserType:      curUserType,
	}

	if err := tx.Model(&model.OrdersTransferRecords{}).Create(r).Error; err != nil {
		global.Lg.Error("CreateOrdersTransferRecords error", zap.Error(err))
		return nil, err
	}

	return r, nil
}
