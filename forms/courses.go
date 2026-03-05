package forms

import "skis-admin-backend/model"

type CreateCourseTagItem struct {
	TagID   int64  `json:"id" binding:"required"`
	TagName string `json:"name" binding:"required"`
}

type QueryCoursesListRequest struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type QueryCoursesListResponse struct {
	List  []*model.Courses `json:"list"`
	Total int64            `json:"total"`
}
