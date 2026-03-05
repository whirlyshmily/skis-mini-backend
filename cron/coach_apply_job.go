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

	// 删除过期的教练申请
	// 查询条件：已拒绝且超时的申请
	var coaches []model.Coaches
	err := global.DB.Model(&model.Coaches{}).
		Where("op_time < ? AND verified = ? AND state = ?",
			time.Now(), model.VerifiedRejected, model.StateNormal).
		Find(&coaches).Error

	if err != nil {
		global.Lg.Error("查询过期教练申请失败", zap.Error(err))
		return
	}

	global.Lg.Info("找到过期教练申请数量", zap.Int("count", len(coaches)))

	for _, coach := range coaches {
		global.Lg.Info("处理过期教练申请",
			zap.String("coach_id", coach.CoachId),
			zap.String("realname", coach.Realname),
			zap.Time("op_time", coach.OpTime))

		// 使用 Updates 只更新状态字段，避免覆盖其他字段
		err = global.DB.Model(&model.Coaches{}).
			Where("coach_id = ? AND state = ?", coach.CoachId, model.StateNormal).
			Updates(map[string]interface{}{
				"state": model.StateDeleted,
			}).Error

		if err != nil {
			global.Lg.Error("删除过期教练申请失败",
				zap.Error(err),
				zap.String("coach_id", coach.CoachId))
		} else {
			global.Lg.Info("成功删除过期教练申请",
				zap.String("coach_id", coach.CoachId))
		}
	}
}
