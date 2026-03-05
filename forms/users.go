package forms

import "skis-admin-backend/model"

// 登录请求参数
type LoginRequest struct {
	Code string `json:"code" binding:"required"` // 微信登录凭证
	UserInfo
	ReferralCode string `json:"referral_code"` // 邀请码
}

type UserInfo struct {
	NickName  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
	Gender    int    `json:"gender"`
	Country   string `json:"country"`
	Province  string `json:"province"`
	City      string `json:"city"`
}

// 微信接口返回结果
type WechatLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// 登录响应
type LoginResponse struct {
	*model.Users
	Token string `json:"token"`
}

// 获取用户手机号请求参数
type UpdateUserPhoneRequest struct {
	Code          string `json:"code" binding:"required"`
	EncryptedData string `json:"encryptedData" binding:"required"` // 加密数据
	Iv            string `json:"iv" binding:"required"`            // 加密算法初始向量
}

// 微信手机号解密后数据结构
type WechatPhoneInfo struct {
	Errcode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`     // 用户绑定的手机号
		PurePhoneNumber string `json:"purePhoneNumber"` // 没有区号的手机号
		CountryCode     string `json:"countryCode"`     // 区号
		Watermark       struct {
			AppID     string `json:"appid"`
			Timestamp int64  `json:"timestamp"`
		} `json:"watermark"`
	} `json:"phone_info"`
}

type QueryUsersListReq struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
}

type QueryUsersListResp struct {
	List  []*model.Users `json:"list"`
	Total int64          `json:"total"`
}

type AppointmentCourseRequest struct {
	OrderCourseId  string `json:"order_course_id" binding:"required"`
	TeachStartTime string `json:"teach_start_time" binding:"required"`
	SkiResortsId   int    `json:"ski_resorts_id" binding:"required"`
}

type BeforeCancelAppointmentCourseResp struct {
	Liability    int   `json:"is_liability"`  // 是否是责任（1：无责、2：有责、3：无责次数用完）
	RefundMoney  int64 `json:"refund_money"`  // 退款金额
	RefundPoints int64 `json:"refund_points"` // 退款积分
}

type CancelAppointmentCourseRequest struct {
	OrderCourseId string `json:"order_course_id" binding:"required"`
}

type ReviewCoachTeachTimeRequest struct {
	OrderCourseId string `json:"order_course_id" binding:"required"`
	IsAgree       bool   `json:"is_agree"`
}
type ReviewTeachTimeRequest struct {
	OrderCourseId string `json:"order_course_id" binding:"required"`
	IsAgree       bool   `json:"is_agree"`
}

type ReviewCoachTransferOrderRequest struct {
	OrderCourseId string `json:"order_course_id" binding:"required"`
	IsAgree       bool   `json:"is_agree"`
}

type UpdateUserInfoRequest struct {
	NickName *string `json:"nickname"`
	Avatar   *string `json:"avatar"`
	Gender   *int    `json:"gender"`
	Birthday *string `json:"birthday"`
}
