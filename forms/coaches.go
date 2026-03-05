package forms

import "skis-admin-backend/model"

type QueryCoachesListRequest struct {
	Keyword      string `form:"keyword"`
	Verified     int    `form:"verified"`       //0-待认证和认证失败，1-认证成功
	TagID        int64  `form:"tag_id"`         //标签ID
	SkiResortsId int    `form:"skis_resort_id"` //雪地ID
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type QueryCoachesListResponse struct {
	List  []*model.Coaches `json:"list"`
	Total int64            `json:"total"`
}

type QueryMatchCoachesListRequest struct {
	OrderCourseId string `form:"order_course_id"`
}

type CoachRemoveTagRequest struct {
	TagID int64 `json:"tag_id" binding:"required"` //标签ID
}

type CoachAddTagReviewRequest struct {
	TagReviews []struct {
		TagID      int64    `json:"tag_id" binding:"required,min=1"`
		TagImgUrls []string `json:"tag_img_urls" binding:"required"`
	} `json:"tag_reviews"`
}
type CoachGetTagReviewRequest struct {
	TagIds []int64 `json:"tag_ids"` //标签ID
}
type CoachEditSkiResortsRequest struct {
	SkiResortsIDs []int64 `json:"ski_resorts_ids"` //雪场ID
}

type CoachJoinClubsRequest struct {
	ClubID string `json:"club_id"` //俱乐部ID
}

type CoachQuitClubsRequest struct {
	ClubID string `json:"club_id"` //俱乐部ID
}

type CoachClubsListRequest struct {
	ClubIDs []string `json:"club_ids"` //俱乐部IDs
}
type ClubsCoachListRequest struct {
	Verified []int `form:"verified"`
	Page     int   `form:"page" binding:"omitempty,min=1"`
	PageSize int   `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type CoachTransferOrderToCoachRequest struct {
	OrderCourseId string `json:"order_course_id" binding:"required"`
	CoachId       string `json:"coach_id" binding:"required"`
	TransferFee   int64  `json:"transfer_fee" binding:"min=1,max=100000000"`
}
type CoachReviewOrderFromCoachRequest struct {
	OrderCourseId string `json:"order_course_id" binding:"required"`
	BufferTime    int    `json:"buffer_time"  binding:"min=0,max=180"`
	IsAgree       bool   `json:"is_agree"`
}
type CoachConfirmOrderCoursRequest struct {
	BufferTime     int    `json:"buffer_time"  binding:"min=0,max=180"`
	TeachStartTime string `json:"teach_start_time"`
}

type CoachChangeOrderCourseTimeRequest struct {
	TeachStartTime string `json:"teach_start_time"  binding:"required"`
}

type CoachCanChangeTimeResp struct {
	CanChangeTime bool   `json:"can_change_time"`
	Reason        string `json:"reason"`
}

type CoachReviewOrderFromClubRequest struct {
	OrderCourseId string `json:"order_course_id" binding:"required"`
	BufferTime    int    `json:"buffer_time"  binding:"min=0,max=180"`
	IsAgree       bool   `json:"is_agree"`
}

type CoachReviewReplaceFromClubRequest struct {
	OrderCourseId string `json:"order_course_id" binding:"required"`
	BufferTime    int    `json:"buffer_time"  binding:"min=0,max=180"`
	IsAgree       bool   `json:"is_agree"`
}

type CreateCoachRequest struct {
	ReferralCode           string               `json:"referral_code" binding:"max=16"`           // 邀请码
	Realname               string               `json:"realname" binding:"required,min=2,max=10"` // 真实姓名
	IdCard                 string               `json:"id_card" binding:"required,len=18"`        // 身份证号
	Phone                  string               `json:"phone" binding:"required,len=11"`          // 手机号
	IdCardPhoto            string               `json:"id_card_photo" binding:"required"`
	Certificates           []CertificateReviews `json:"certificates" binding:"required"`
	Introduction           string               `json:"introduction"`
	SupplementaryMaterials []string             `json:"supplementary_materials"`
}

type AdminTransferOrderToCoachRequest struct {
	CoachId string `json:"coach_id" binding:"required"`
}
type EditCoachInfoRequest struct {
	EditFields   []string `json:"edit_fields"`
	Introduction string   `json:"introduction"`
}

type QueryCoachGoodsRequest struct {
	OnShelf *int `form:"on_shelf"`
	Pack    *int `form:"pack" binding:"oneof=0 1"` //是否为打包课程（0：否，1：打包）, -1: 所有
}
type CoachGetSkiResortsRequest struct {
	OrderCourseId string `form:"order_course_id" binding:"required"`
}
type CoachGetSkiResortDateRequest struct {
	SkiResortsID  int64  `form:"ski_resorts_id" binding:"required"` //雪场ID
	OrderCourseId string `form:"order_course_id" binding:"required"`
	StartDate     string `form:"start_date"  binding:"required"`
	EndDate       string `form:"end_date"  binding:"required"`
}
type CoachGetSkiResortTimeRequest struct {
	SkiResortsID int64  `form:"ski_resorts_id" binding:"required"` //雪场ID
	Date         string `form:"date"  binding:"required"`
}
