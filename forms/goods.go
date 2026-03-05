package forms

import "skis-admin-backend/model"

type QueryGoodsListRequest struct {
	UserId   string `form:"user_id"`
	OnShelf  *int32 `form:"on_shelf"`
	TagId    int64  `form:"tag_id"`
	Pack     *int   `form:"pack" binding:"omitempty,oneof=0 1 -1"` //是否为打包课程（0：否，1：打包）, -1: 所有
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type QueryGoodsListResponse struct {
	List  []*model.Goods `json:"list"`
	Total int64          `json:"total"`
}

type SaveGoodsRequest struct {
	GoodID     string   `json:"good_id"`
	Title      string   `json:"title"`
	CoverUrl   string   `json:"cover_url"`
	TeachTime  int      `json:"teach_time"`                 //教学时长
	TeachMoney int64    `json:"teach_money"`                //教学费用
	ClubMoney  int64    `json:"club_money"`                 //推荐费（俱乐部的课程才有）
	AreaMoney  int64    `json:"area_money"`                 //场地费用
	FaultMoney int64    `json:"fault_money"`                //有责取消费
	Pack       int      `json:"pack"`                       //是否为打包课程（0：否，1：打包）
	Discount   int      `json:"discount" max:"100" min:"0"` //折扣（打包课才设置）
	CourseId   string   `json:"course_id"`                  //课程ID
	GoodIds    []string `json:"good_ids"`                   //商品IDs（打包课）
}

type TakeDownGoodsRequest struct {
	GoodID string `json:"good_id" binding:"required"`
}

type PutUpGoodsRequest struct {
	GoodID string `json:"good_id" binding:"required"`
}

type DeleteGoodsRequest struct {
	GoodID string `json:"good_id" binding:"required"`
}

// BeforeEditGoodsRequest 编辑商品前检查请求
type BeforeEditGoodsRequest struct {
	GoodID string `json:"good_id" binding:"required"` // 商品ID
}

// BeforeEditGoodsResp 编辑商品前检查响应
type BeforeEditGoodsResp struct {
	HasUnfinishedOrder bool   `json:"has_unfinished_order"` // 是否存在未完成订单
	IsInPackCourse     bool   `json:"is_in_pack_course"`    // 是否在打包课中（仅单次课有效）
	Message            string `json:"message"`              // 提示信息
}
