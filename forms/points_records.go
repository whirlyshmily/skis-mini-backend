package forms

import (
	"skis-admin-backend/model"
)

type QueryPointsRecordsListRequest struct {
	Keyword   string `form:"keyword"`
	PointType *int   `form:"point_type"` //0-收入，1-支出
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"page_size" binding:"required,min=1,max=100"`
	StartTime string `form:"start_time" binding:"omitempty" time_format:"2006-01-02"`
	EndTime   string `form:"end_time" binding:"omitempty" time_format:"2006-01-02"`
}

type QueryPointsRecordsListResponse struct {
	List  []*model.PointsRecords `json:"list"`
	Total int64                  `json:"total"`
}
