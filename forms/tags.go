package forms

import "skis-admin-backend/model"

type QueryTagsListRequest struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type QueryTagsListResponse struct {
	List  []*model.TagsList `json:"list"`
	Total int64             `json:"total"`
}

type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}
