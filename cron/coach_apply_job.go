package cron

import (
	"go.uber.org/zap"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"time"
)

//教练申请，如果超过op_time还没有补充材料，则删除

type CoachApplyJob struct {
}

func (c CoachApplyJob) Run() {
	global.Lg.Debug("开始执行教练申请任务")
	//删除过期的教练申请
	//先查询
	var coaches []model.Coaches
	err := global.DB.Model(&model.Coaches{}).Where("op_time < ? and verified = ? and state = 0", time.Now(), model.VerifiedReject).Find(&coaches).Error
	if err != nil {
		global.Lg.Error("删除过期的教练申请失败", zap.Error(err))
		return
	}

	for _, coach := range coaches {
		global.Lg.Info("删除过期的教练申请", zap.Any("coach", coach))
		coach.State = model.StateDeleted
		err = global.DB.Model(&model.Coaches{}).Where("coach_id = ? and state = 0", coach.CoachId).Save(&coach).Error
		if err != nil {
			global.Lg.Error("删除过期的教练申请失败", zap.Error(err))
		}
	}

}
