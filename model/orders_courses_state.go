package model

import (
	"time"
)

//create table orders_courses_state
//(
//id               bigint auto_increment comment '自增ID'
//primary key,
//order_course_id  varchar(64)                         null comment '订单课程ID',
//user_id          varchar(64)                         null comment '操作状态的用户（包括用户ID、教练ID、俱乐部ID、官方ID）',
//user_type        int                                 null comment '用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）',
//operate          int                                 null comment '操作变更（参考model）',
//coach_id         varchar(64)                         null comment '教练ID（俱乐部分配教练的ID，教练转单的教练ID）',
//teach_start_time datetime                            null comment '教学开始时间',
//teach_time_ids   varchar(1024)                       null comment '教学时间ID（ski_resorts_teach_time表的ID）',
//transfer_fee     bigint    default 0                 null comment '转单金额',
//process          int       default 1                 null comment '处理（0：待处理，1：已处理）',
//remark           varchar(255)                        null comment '备注',
//state            int       default 0                 null comment '状态，0-正常， 1-删除',
//created_at       timestamp default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at       timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间'
//)
//comment '订单课程状态';

type OrdersCoursesState struct {
	ID              int64        `json:"id" gorm:"column:id"`                                      // 自增ID
	OrderCourseID   string       `json:"order_course_id" gorm:"column:order_course_id"`            // 订单课程ID
	UserID          string       `json:"user_id" gorm:"column:user_id"`                            // 操作状态的用户（包括用户ID、教练ID、俱乐部ID、官方ID）
	UserType        int          `json:"user_type" gorm:"column:user_type"`                        // 用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）
	Operate         int          `json:"operate" gorm:"column:operate"`                            // 操作变更（参考model）
	CoachID         string       `json:"coach_id" gorm:"column:coach_id"`                          // 教练ID（俱乐部分配教练的ID，教练转单的教练ID）
	TeachStartTime  LocalTime    `json:"teach_start_time" gorm:"column:teach_start_time"`          // 教学开始时间
	TeachTimeIDs    JSONIntArray `json:"teach_time_ids" gorm:"column:teach_time_ids;default null"` // 教学时间ID（ski_resorts_teach_time表的ID）
	TransferFee     int64        `json:"transfer_fee" gorm:"column:transfer_fee;default 0"`        // 转单金额
	Process         int          `json:"process" gorm:"column:process;default 1"`                  // 处理（0：待处理，1：已处理）
	Remark          string       `json:"remark" gorm:"column:remark"`                              // 备注
	State           int          `json:"state" gorm:"column:state"`                                // 状态，0-正常， 1-删除
	CreatedAt       time.Time    `json:"created_at" gorm:"column:created_at"`                      // 创建时间
	UpdatedAt       time.Time    `json:"updated_at" gorm:"column:updated_at"`                      // 更新时间
	LastConfirmTime time.Time    `json:"last_confirm_time" gorm:"-"`                               //最后确认时间
}

func (m *OrdersCoursesState) TableName() string {
	return "orders_courses_state"
}

const (
	OperateUserAppointment                 = 1   //用户预约课程
	OperateUserCancelBeforeClubConfirm     = 98  //用户在俱乐部未确认前取消
	OperateUserCancelBeforeCoachConfirm    = 99  //用户在教练未确认前取消
	OperateUserCancelNoResponsibility      = 100 //用户无责取消
	OperateUserCancelResponsibility        = 101 //用户有责取消
	OperateUserAgreeChangeTimeBeforeC      = 102 //用户同意教练确认预约之前修改的上课时间
	OperateUserDisagreeChangeTimeBeforeC   = 103 //用户不同意教练确认预约之前修改上的课时间
	OperateUserAgreeCoachChangeCourse      = 110 //用户同意教练修改上课时间
	OperateUserDisagreeCoachChangeCourse   = 111 //用户不同意教练修改上课时间
	OperateUserAgreeCoachTransferCourse    = 120 //用户同意教练转单课程
	OperateUserDisagreeCoachTransferCourse = 121 //用户不同意教练转单课程
	OperateUserCancelCourse                = 130 //用户退课
	OperateUserVerifyCourse                = 150 //用户核销课程

	OperateCoachCancelBeforeCoachConfirm = 190 //教练在教练未确认前取消
	OperateCoachCancelNoResponsibility   = 191 //教练无责取消
	OperateCoachChangeCourseTime         = 199 //待用户确认课程时，教练修改上课时间	定时任务
	OperateCoachConfirmCourse            = 200 //教练确认课程
	OperateCoachChangeCourse             = 210 //待上课时，教练修改上课时间	定时任务
	OperateCoachTransferCourse           = 220 //教练向用户申请转单课程
	OperateCoachCancelCourseTransfer     = 221 //教练取消转单课程
	OperateCoachTransferToCoach          = 230 //教练转单课程给其他教练
	OperateCoachAgreeTransferCourse      = 231 //教练同意转单课程
	OperateCoachDisagreeTransferCourse   = 232 //教练不同意转单课程
	OperateCoachCancelCourse             = 250 //教练取消课程

	OperateClubChangeUserCourseTime         = 300 //俱乐部申请修改用户上课时间
	OperateUserAgreeClubChangeCourseTime    = 310 //用户同意俱乐部修改上课时间
	OperateUserDisagreeClubChangeCourseTime = 320 //用户不同意俱乐部修改上课时间
	OperateClubAppointCoach                 = 330 //俱乐部安排教练上课
	OperateCoachAgreeClubCourse             = 331 //教练同意俱乐部课程
	OperateCoachDisagreeClubCourse          = 332 //教练不同意俱乐部课程
	OperateCoachApplyTransferCourse         = 340 //教练申请转单课程
	OperateClubTransferToCoach              = 341 //俱乐部转移课程给教练
	OperateCoachAgreeClubTransferToCoach    = 342 //教练同意俱乐部转单课程
	OperateCoachDisagreeClubTransferToCoach = 343 //教练不同意俱乐部转单课程

	OperateCronCancelCourse                = 400 //定时任务取消课程
	OperateCronCancelClubCourse            = 401 //定时任务取消俱乐部安排给教练的课程
	OperateCronCancelCoachChangeCourse     = 402 //定时任务取消教练修改上课时间
	OperateCronCancelClubChangeCourseTime  = 403 //定时任务取消俱乐部修改上课时间
	OperateCronCancelCoachTransferCourse   = 404 //定时任务取消教练转单课程
	OperateCronCancelTransferCourseToCoach = 405 //定时任务取消教练转单给其他教练的课程

	OperateCoachVerifyCourse = 500 //教练核销课程

	OperateCronVerifyCourse = 600 //定时任务核销课程

	OperateAdminOrderTransfer = 700 //管理台转单

)

var OCSOperateStr = map[int]string{
	OperateUserAppointment:                  "用户预约课程",
	OperateUserCancelBeforeClubConfirm:      "用户在俱乐部未确认前取消",
	OperateUserCancelBeforeCoachConfirm:     "用户在教练未确认前取消",
	OperateUserCancelNoResponsibility:       "用户无责取消",
	OperateUserCancelResponsibility:         "用户有责取消",
	OperateCoachConfirmCourse:               "教练确认课程",
	OperateCoachCancelCourse:                "教练取消课程",
	OperateCoachVerifyCourse:                "教练核销课程",
	OperateUserVerifyCourse:                 "用户核销课程",
	OperateCoachChangeCourseTime:            "待用户确认课程时，教练修改上课时间",
	OperateCoachChangeCourse:                "待上课，教练修改上课时间",
	OperateUserAgreeChangeTimeBeforeC:       "用户同意教练确认预约之前修改的上课时间",
	OperateUserDisagreeChangeTimeBeforeC:    "用户不同意教练确认预约之前修改上的课时间",
	OperateUserAgreeCoachChangeCourse:       "用户同意教练修改上课时间",
	OperateUserDisagreeCoachChangeCourse:    "用户不同意教练修改上课时间",
	OperateCoachTransferCourse:              "教练向用户申请转单课程",
	OperateUserAgreeCoachTransferCourse:     "用户同意教练转单课程",
	OperateUserDisagreeCoachTransferCourse:  "用户不同意教练转单课程",
	OperateCoachTransferToCoach:             "教练转单课程给其他教练",
	OperateCoachAgreeTransferCourse:         "教练同意转单课程",
	OperateCoachDisagreeTransferCourse:      "教练不同意转单课程",
	OperateClubChangeUserCourseTime:         "俱乐部修改上课时间",
	OperateUserAgreeClubChangeCourseTime:    "用户同意俱乐部修改上课时间",
	OperateUserDisagreeClubChangeCourseTime: "用户不同意俱乐部修改上课时间",
	OperateClubAppointCoach:                 "俱乐部安排教练上课",
	OperateCoachAgreeClubCourse:             "教练同意俱乐部课程",
	OperateCoachDisagreeClubCourse:          "教练不同意俱乐部课程",
	OperateCoachCancelCourseTransfer:        "教练取消转单课程",
	OperateCoachCancelBeforeCoachConfirm:    "教练在教练未确认前取消",
	OperateCoachCancelNoResponsibility:      "教练无责取消",
	OperateCoachApplyTransferCourse:         "教练申请转单课程",
	OperateClubTransferToCoach:              "俱乐部转移课程给教练",
	OperateCoachAgreeClubTransferToCoach:    "教练同意俱乐部转单课程",
	OperateCoachDisagreeClubTransferToCoach: "教练不同意俱乐部转单课程",
	OperateCronCancelCourse:                 "定时任务取消课程",
	OperateCronCancelClubCourse:             "定时任务取消俱乐部安排给教练的课程",
	OperateCronCancelCoachChangeCourse:      "定时任务取消教练修改上课时间",
	OperateCronCancelClubChangeCourseTime:   "定时任务取消俱乐部修改上课时间",
	OperateCronCancelCoachTransferCourse:    "定时任务取消教练转单课程",
	OperateCronCancelTransferCourseToCoach:  "定时任务取消教练转单给其他教练的课程",
	OperateCronVerifyCourse:                 "定时任务核销课程",
	OperateUserCancelCourse:                 "用户退课",
	OperateAdminOrderTransfer:               "管理台转单",
}

const (
	OrderCourseUserCancelNumber  = 2 // 用户无责取消课程次数
	OrderCourseCoachCancelNumber = 5 // 教练无责取消课程次数
)
const (
	ProcessNo  = 0 //待处理
	ProcessYes = 1 //处理完成
)
