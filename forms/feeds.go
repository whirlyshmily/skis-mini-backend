package forms

import "skis-admin-backend/model"

type QueryFeedListRequest struct {
	Curated  *int `form:"curated"`  //是否精选，0-不精选，1-精选， -1-全部
	OnShelf  *int `form:"on_shelf"` //是否自己发布的，0-不是，1-是，-1-全部
	Page     int  `form:"page"`
	PageSize int  `form:"page_size"`
}

type QueryFeedListResponse struct {
	List  []*model.Feeds `json:"list"`
	Total int64          `json:"total"`
}

type UpdateFeedRequest struct {
	Title   *string             `json:"title" `
	Urls    *model.JsonUrlArray `json:"urls"`
	Detail  *string             `json:"detail"`
	Curated *int                `json:"curated" binding:"omitempty,oneof=0 1"`
	OnShelf *int                `json:"on_shelf" binding:"omitempty,oneof=0 1"`
}

type CreateFeedRequest struct {
	Title  string             `json:"title" binding:"required"`
	Urls   model.JsonUrlArray `json:"urls" binding:"required"`
	Detail string             `json:"detail" binding:"required"`
}
