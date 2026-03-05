package forms

import "skis-admin-backend/model"

type QueryClubListRequest struct {
	Keyword      string `form:"keyword"`
	Verified     int    `form:"verified"`       //0-待认证和认证失败，1-认证成功
	TagID        int    `form:"tag_id"`         //标签ID
	SkiResortsId int    `form:"ski_resorts_id"` //雪地ID
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type QueryClubListResponse struct {
	Total int64          `json:"total"`
	List  []*model.Clubs `json:"list"`
}

type UpdateClubRequest struct {
	Priority      *int    `json:"priority" binding:"omitempty,min=0,max=9999"`
	Level         *int    `json:"level" binding:"omitempty,min=1,max=5"`
	FrozenDeposit *int    `json:"frozen_deposit" binding:"omitempty"`
	Verified      *int    `json:"verified" binding:"omitempty,oneof=0 1 2"`
	Remark        *string `json:"remark" binding:"binding:required_if:Verified 2,max=200"`
}

type CoachJoinClubRequest struct {
	ClubId string `json:"club_id" binding:"required"`
}
type CoachQuitClubRequest struct {
	ClubId string `json:"club_id" binding:"required"`
}

// 登录响应
type ClubsUserLoginResponse struct {
	*model.ClubsUsers
	Token string `json:"token"`
}

type ClubCheckCoachRequest struct {
	Id       int64 `json:"id" binding:"required"`
	Verified int   `json:"verified" binding:"required,omitempty,oneof=1 2"` //是否认证，0-未认证，1-认证通过,2-驳回
}

type ClubChangeTeachTimeRequest struct {
	TeachStartTime string `json:"teach_start_time"`
}
type ClubAppointmentCourseRequest struct {
	CoachId string `json:"coach_id" binding:"required"`
}
type ClubReplaceCoachCourseRequest struct {
	CoachId string `json:"coach_id" binding:"required"`
}

type ApplyClubRequest struct {
	Name             string `json:"name" binding:"required"`
	Logo             string `json:"logo" binding:"required"`
	Manager          string `json:"manager" binding:"required"`
	Phone            string `json:"phone" binding:"required"`
	SocialCreditCode string `json:"social_credit_code" binding:"required"`
	BusinessLicense  string `json:"business_license" binding:"required"`
	IdCardFront      string `json:"id_card_front" binding:"required"`
	IdCardBack       string `json:"id_card_back" binding:"required"`
}

type QueryClubOrderCoursesRequest struct {
	TeachStates []int64 `form:"teach_states"`
	Page        int     `form:"page"`
	PageSize    int     `form:"page_size"`
}

type UpdateClubsInfoRequest struct {
	Name         *string `json:"name"`
	ApprovalLogo *string `json:"approval_logo"`
	Introduction *string `json:"introduction"`
}
