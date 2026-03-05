package dao

import (
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func QueryPointsRecordsList(uid string, req *forms.QueryPointsRecordsListRequest) (int64, []*model.PointsRecords, error) {
	var total int64
	db := global.DB.Model(&model.PointsRecords{}).Where("uid = ? and state = 0", uid)

	if req.PointType != nil {
		if *req.PointType == 1 { //收入
			db.Where("action_type < ?", 1000)
		} else { //支出
			db.Where("action_type >= ?", 1000)
		}
	}

	if req.StartTime != "" {
		db.Where("created_at >= ?", GetStartTime(req.StartTime))
	}

	if req.EndTime != "" {
		db.Where("created_at <= ?", GetEndTime(req.EndTime))
	}

	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("QueryPointsRecordsList Count error", zap.Error(err))
		return 0, nil, err
	}

	var list []*model.PointsRecords
	if err := db.Order("id desc").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list).Error; err != nil {
		global.Lg.Error("QueryPointsRecordsList Find error", zap.Error(err))
		return 0, nil, err
	}
	return total, list, nil
}
func QueryPointsRecordInfo(uid string, pointId string) (*model.PointsRecords, error) {
	var record *model.PointsRecords
	if err := global.DB.Model(&model.PointsRecords{}).Where("uid = ? and point_id = ? and state = 0", uid, pointId).First(&record).Error; err != nil {
		global.Lg.Error("QueryPointsRecordInfo Find error", zap.Error(err))
		return nil, err
	}

	return record, nil
}

// pay_uid        varchar(64)                         null comment '支付uid',
// action_type    int       default 0                 not null comment '操作类型，0-完成课程增加积分，1-购买课程抵扣积分，2-系统增加积分，3-系统减少积分',
// relation_id    int       default 0                 null comment '关联id',
// receive_uid    varchar(255)                        null comment '接收uid',
// points         int                                 not null comment '积分变更',
// current_points bigint    default 0                 null comment '当前积分',

func CreatePointsRecord(tx *gorm.DB, uid, relationId string, actionType, points, currentPoints int64, remark string) (*model.PointsRecords, error) {
	record := &model.PointsRecords{
		PointID:       GenerateId("JF"),
		RelationID:    relationId,
		Uid:           uid,
		ActionType:    actionType,
		Points:        points,
		CurrentPoints: currentPoints,
		Remark:        remark,
	}
	if err := tx.Table("points_records").Create(record).Error; err != nil {
		global.Lg.Error("CreatePointsRecord error", zap.Error(err))
		return nil, err
	}
	return record, nil
}
