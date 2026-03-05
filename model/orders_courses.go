package model

import (
	"fmt"
	"time"
)

//create table orders_courses
//(
//id              int auto_increment comment '自增ID' primary key,
//uid             varchar(64)                         not null comment '下单用户ID',
//order_course_id varchar(64)                           null comment '订单课程ID',
//order_id        varchar(64)                           null comment '订单ID',
//good_id         varchar(64) default ''                not null comment '商品id',
//pack              int       default 0                 null comment '是否为打包课程（0：否(good_id)，1：打包(pack_good_id)）',
//course_id       varchar(64)                           not null comment '课程ID',
//teach_time_ids  varchar(255)                        null comment '教学时间ID（ski_resorts_teach_time表的ID）',
//club_time_ids     varchar(255)                        null comment '俱乐部的教学时间ID（ski_resorts_teach_time表的ID）',
//teach_state     int       default 0                 null comment '教学状态（0：待预约，100：待上课，200：待核销，）',
//teach_start_time datetime  default CURRENT_TIMESTAMP null comment '教学开始时间',
//ski_resorts_id    int       default 0                 null comment '雪场ID',
//teach_buffer_time int       default 0                 null comment '教学缓冲时间',
//teach_coach_id  varchar(64)                         null comment '教学教练ID',
//teach_money       int       default 0                 null comment '教学费用',
//club_money        int       default 0                 null comment '俱乐部的推荐费（俱乐部的课程才有）',
//area_money        int       default 0                 null comment '场地费用',
//fault_money       int       default 0                 null comment '有责取消费',
//check_code      varchar(64)                           null comment '核销码',
//is_check        int         default 0                 not null comment '是否核销（0：未核销，1：已核销）',
//state           int         default 0                 not null comment '状态，0-正常， 1-删除',
//created_at      timestamp   default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at      timestamp   default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间',
//constraint orders_courses_pk
//unique (order_id, course_id)
//)
//comment '订单课程表' row_format = DYNAMIC;

type OrdersCourses struct {
	ID                 int                    `json:"id" gorm:"column:id"`                               // 自增ID
	Uid                string                 `gorm:"column:uid" json:"uid"`                             // 下单用户ID
	OrderCourseID      string                 `json:"order_course_id" gorm:"column:order_course_id"`     // 订单课程ID
	OrderID            string                 `json:"order_id" gorm:"column:order_id"`                   // 订单ID
	GoodID             string                 `json:"good_id" gorm:"column:good_id"`                     // 商品id
	Pack               int                    `json:"pack" gorm:"column:pack"`                           // 是否为打包课程（0：否(good_id)，1：打包(pack_good_id)）
	CourseID           string                 `json:"course_id" gorm:"column:course_id"`                 // 课程ID
	TeachCoachID       string                 `json:"teach_coach_id" gorm:"column:teach_coach_id"`       // 教学教练ID
	TeachState         TeachState             `json:"teach_state" gorm:"column:teach_state"`             // 教学状态
	TeachStartTime     LocalTime              `json:"teach_start_time" gorm:"column:teach_start_time"`   // 教学开始时间
	SkiResortsID       int                    `json:"ski_resorts_id" gorm:"column:ski_resorts_id"`       // 雪场ID
	TeachTimeIDs       JSONIntArray           `json:"teach_time_ids" gorm:"column:teach_time_ids"`       // 教练的教学时间ID（ski_resorts_teach_time表的ID）
	ClubTimeIDs        JSONIntArray           `json:"club_time_ids" gorm:"column:club_time_ids"`         // 俱乐部的教学时间ID（ski_resorts_teach_time表的ID）
	TeachBufferTime    int                    `json:"teach_buffer_time" gorm:"column:teach_buffer_time"` // 教学缓冲时间（分钟）
	TeachTime          int                    `json:"teach_time" gorm:"column:teach_time"`               // 教学时长（分钟）
	TeachMoney         int64                  `json:"teach_money" gorm:"column:teach_money"`             // 教学费用
	ClubMoney          int64                  `json:"club_money" gorm:"column:club_money"`               // 部的推荐费（部的课程才有）
	AreaMoney          int64                  `json:"area_money" gorm:"column:area_money"`               // 场地费用
	FaultMoney         int64                  `json:"fault_money" gorm:"column:fault_money"`             // 有责取消费
	PayMoney           int64                  `json:"pay_money" gorm:"column:pay_money"`                 // 实际支付金额
	CheckCode          string                 `json:"check_code" gorm:"column:check_code"`               // 核销码
	IsCheck            int                    `json:"is_check" gorm:"column:is_check"`                   // 是否核销（0：未核销，1：已核销）
	State              int                    `json:"state" gorm:"column:state"`                         // 状态，0-正常， 1-删除
	CreatedAt          time.Time              `json:"created_at" gorm:"column:created_at"`               // 创建时间
	UpdatedAt          time.Time              `json:"updated_at" gorm:"column:updated_at"`               // 更新时间
	Course             *Courses               `gorm:"foreignKey:CourseID;references:CourseID" json:"course,omitempty"`
	CourseTags         []*CoursesTags         `gorm:"foreignKey:CourseID;references:CourseID" json:"courses_tags,omitempty"`
	OrdersCoursesState OrdersCoursesState     `gorm:"-" json:"orders_courses_state"`
	Order              *Orders                `gorm:"foreignKey:OrderID;references:OrderID" json:"order,omitempty"`
	Good               *Goods                 `gorm:"foreignKey:GoodID;references:GoodID" json:"good,omitempty"`
	Tags               []*Tags                `gorm:"-" json:"tags,omitempty"`
	TeachTimes         []*SkiResortsTeachTime `gorm:"-" json:"teach_times,omitempty"`
	NewTeachTimes      []*SkiResortsTeachTime `gorm:"-" json:"new_teach_times,omitempty"`
	UserInfo           *Users                 `gorm:"foreignKey:Uid;references:Uid" json:"user_info,omitempty"`
	Comment            *OrdersCoursesComments `json:"-" gorm:"foreignKey:OrderCourseId;references:OrderCourseID"`
	CommentId          int64                  `gorm:"-" json:"comment_id"`
	Reply              *OrdersCoursesComments `json:"-" gorm:"foreignKey:OrderCourseId;references:OrderCourseID"`
	Replied            bool                   `gorm:"-" json:"replied"`
	Remark             string                 `gorm:"-" json:"remark"`
	TransferCoach      *Coaches               `gorm:"-" json:"transfer_coach"`
}

func (m *OrdersCourses) TableName() string {
	return "orders_courses"
}

const FrozenNo = 0  //无效
const FrozenYes = 1 //已冻结

const IsCheckNo = 0  //未核销
const IsCheckYes = 1 //已核销

type TeachState int

const (
	TeachStateWaitAppointment           TeachState = 0   //待预约
	TeachStateWaitUserSecondConfirmTime TeachState = 5   //教练确认预约前修改时间，待用户确认预约时间
	TeachStateWaitCoachConfirmUser      TeachState = 10  //待教练确认用户预约
	TeachStateWaitUserConfirmCoachTime  TeachState = 20  //待上课时，教练修改时间，待用户确认上课时间
	TeachStateWaitUserConfirmTransfer   TeachState = 30  //待用户确认课程转单
	TeachStateWaitCoachTransfer         TeachState = 40  //用户已同意转单，待教练转单
	TeachStateWaitConfirmTransfer       TeachState = 45  //待接单教练确认
	TeachStateWaitClass                 TeachState = 100 //待上课
	TeachStateWaitClassTransfer         TeachState = 110 //转单后待上课

	TeachStateWaitClubConfirm         TeachState = 200 //待俱乐部确认（分配教练）
	TeachStateWaitUserConfirmClubTime TeachState = 210 //待用户确认俱乐部申请修改上课时间
	TeachStateWaitCoachConfirmClub    TeachState = 220 //待教练确认俱乐部分配课程（审核后，状态回到200）

	TeachStateWaitCoachClass           TeachState = 250 //待上课（教练确认俱乐部分配课程）
	TeachStateCoachApplyTransfer       TeachState = 260 //教练申请转单
	TeachStateWaitCoachConfirmTransfer TeachState = 261 //待教练确认俱乐部转移课程（审核后，状态回到250）

	TeachStateWaitCheck    TeachState = 300  //待核销
	TeachStateAlreadyClass TeachState = 310  //已上课
	TeachStateFinish       TeachState = 400  //已完成
	TeachStateCancel       TeachState = 1000 //已退课
)

func GetUserOrderCourseOrderStr() string {
	return fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v",
		TeachStateWaitUserConfirmTransfer,
		TeachStateWaitUserConfirmClubTime,
		TeachStateWaitUserSecondConfirmTime,
		TeachStateWaitCoachClass,
		TeachStateWaitClass,
		TeachStateWaitClassTransfer,
		TeachStateWaitCoachTransfer,
		TeachStateWaitConfirmTransfer,
		TeachStateWaitUserConfirmCoachTime,
		TeachStateWaitClubConfirm,
		TeachStateWaitCoachConfirmUser,
		TeachStateWaitCoachConfirmClub,
		TeachStateCoachApplyTransfer,
		TeachStateWaitCoachConfirmTransfer,
		TeachStateWaitCheck,
		TeachStateWaitAppointment,
		TeachStateAlreadyClass,
		TeachStateFinish,
		TeachStateCancel,
	)
}

//const UserCourseOrderBy = fmt.Sprintf("%v,%s,%s,%s", TeachStateWaitUserSecondConfirmTime)

//string(rune(TeachStateWaitUserSecondConfirmTime)) + "," + string(rune(TeachStateWaitUserConfirmCoachTime)) + "," + string(rune(TeachStateWaitUserConfirmTransfer)) + "," + string(rune(TeachStateWaitUserConfirmClubTime))

var UserTeachStateStr = map[TeachState]string{
	TeachStateWaitAppointment:           "待学员预约上课时间",
	TeachStateWaitUserSecondConfirmTime: "待学员确认订单，请在%s前确认时间",
	TeachStateWaitCoachConfirmUser:      "待教练确认学员预约",
	TeachStateWaitUserConfirmCoachTime:  "时间有变更，请在%s前确认订单",
	TeachStateWaitUserConfirmTransfer:   "更换授课人，请在%s前确认订单",
	TeachStateWaitCoachTransfer:         "待课程转单",
	TeachStateWaitConfirmTransfer:       "待接单教练确认",
	TeachStateWaitClass:                 "待学员上课",
	TeachStateWaitClassTransfer:         "待学员上课。",
	TeachStateWaitClubConfirm:           "待俱乐部确认（分配教练）",
	TeachStateWaitUserConfirmClubTime:   "待学员确认俱乐部申请修改上课时间",
	TeachStateWaitCoachConfirmClub:      "待教练确定俱乐部派单",
	TeachStateWaitCoachClass:            "待课程分配",
	TeachStateCoachApplyTransfer:        "待课程转单",
	TeachStateWaitCoachConfirmTransfer:  "待教练确认俱乐部转移课程",
	TeachStateWaitCheck:                 "待学员确认订单",
	TeachStateAlreadyClass:              "待确认订单，上课7天后自动核销",
	TeachStateFinish:                    "已完成",
	TeachStateCancel:                    "已退课",
}

var CoachTeachStateStr = map[TeachState]string{
	TeachStateWaitAppointment:           "待学员预约上课时间",
	TeachStateWaitUserSecondConfirmTime: "时间变更，待学员确认",
	TeachStateWaitCoachConfirmUser:      "请在%s前确认时间",
	TeachStateWaitUserConfirmCoachTime:  "时间变更，待学员确认",
	TeachStateWaitUserConfirmTransfer:   "转单变更，待学员确认",
	TeachStateWaitCoachTransfer:         "学员同意转单，请在48小时内完成转单",
	TeachStateWaitConfirmTransfer:       "有教练转单给你，请在48小时内完成转单",
	TeachStateWaitClass:                 "待学员上课",
	TeachStateWaitClubConfirm:           "待俱乐部确认（分配教练）",
	TeachStateWaitUserConfirmClubTime:   "待学员确认俱乐部申请修改上课时间",
	TeachStateWaitCoachConfirmClub:      "待教练确定俱乐部派单",
	TeachStateWaitCoachClass:            "待课程分配",
	TeachStateCoachApplyTransfer:        "待课程转单",
	TeachStateWaitCoachConfirmTransfer:  "待教练确认俱乐部转移课程",
	TeachStateWaitCheck:                 "待学员确认订单",
	TeachStateAlreadyClass:              "待确认订单，上课7天后自动核销",
	TeachStateFinish:                    "已完成",
	TeachStateCancel:                    "已退课",
}

const (
	OrderFaultCancel  = 1 //有责取消订单
	OrderCancelNumber = 2
)

const (
	PackTeachStateWaitAppointment = iota
	PackTeachStateDoing           //进行中
	PackTeachStateWaitCheck       //待核销
	PackTeachStateFinish          //已完成
	PackTeachStateCancel          //已退课
)
