package forms

import "skis-admin-backend/model"

type QueryFeedsCommentsListRequest struct {
	LastId   int `form:"last_id"`
	PageSize int `form:"page_size"`
}

type CreateFeedsCommentsRequest struct {
	Content string             `json:"content"`
	Urls    model.JsonUrlArray `json:"urls"`
}

type UpdateFeedsCommentsRequest struct {
	Content string             `json:"content"`
	Urls    model.JsonUrlArray `json:"urls"`
}
