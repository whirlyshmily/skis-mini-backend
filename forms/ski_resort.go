package forms

import "skis-admin-backend/model"

type CreateSkiResortRequest struct {
	Name     string `json:"name" binding:"required"`
	Province string `json:"province" binding:"required"`
	City     string `json:"city" binding:"required"`
}

type QuerySkiResortsListRequest struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type QuerySkiResortsListResponse struct {
	List  []*model.SkiResortList `json:"list"`
	Total int64                  `json:"total"`
}

type UpdateSkiResortRequest struct {
	Name     *string `json:"name" binding:"omitempty"`
	Province *string `json:"province" binding:"omitempty"`
	City     *string `json:"city" binding:"omitempty"`
	Status   *uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

type QuerySkiResortTeachTimeListRequest struct {
	SkiResortsId   int    `form:"ski_resorts_id" binding:"required"`
	UserID         string `form:"user_id"`
	TeachDateStart string `form:"teach_date_start"`
	TeachDateEnd   string `form:"teach_date_end"`
	OnlyDate       int    `form:"only_date"`
}

type QuerySkiResortTeachDateListRequest struct {
	SkiResortsId   int    `form:"ski_resorts_id" binding:"required"`
	UserID         string `form:"user_id"`
	TeachDateStart string `form:"teach_date_start"`
	TeachDateEnd   string `form:"teach_date_end"`
}

type CreateSkiResortTeachTimeRequest struct {
	TeachDates     []string `json:"teach_dates" binding:"required"`
	TeachStartTime string   `json:"teach_start_time" binding:"required"`
	TeachEndTime   string   `json:"teach_end_time" binding:"required"`
	SkiResortsId   int      `json:"ski_resorts_id" binding:"required"`
}

type UpdateSkiResortTeachStateRequest struct {
	SkiResortsId    int      `json:"ski_resorts_id" binding:"required"`
	TeachStartTimes []string `json:"teach_start_times" binding:"required"`
	TeachState      *uint8   `json:"teach_state"` //预约状态（0：可预约，1：已锁定，2：课后缓冲）
	Remark          *string  `json:"remark"`
	Title           *string  `json:"title"`
}

type DeleteSkiResortTeachTimeRequest struct {
	SkiResortsId int      `json:"ski_resorts_id" binding:"required"`
	TeachDates   []string `json:"teach_dates" binding:"required"`
}

type ScheduleEventRequest struct {
	SkiResortsId int    `form:"ski_resorts_id" binding:"required"`
	TeachDate    string `form:"teach_date" binding:"required"`
}

type ClubSkiResortTeachDateListRequest struct {
	SkiResortsId   int    `form:"ski_resorts_id" binding:"required"`
	TeachDateStart string `form:"teach_date_start"`
	TeachDateEnd   string `form:"teach_date_end"`
}
