package model

import (
	"time"
)

//create table ski_resorts_teach_time
//(
//id               bigint auto_increment comment '自增ID'
//primary key,
//user_id          varchar(64)                         null comment '用户ID（包括用户ID、教练ID、俱乐部ID、官方ID）',
//user_type        int       default 2                 null comment '用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）',
//ski_resorts_id   bigint                              null comment '雪场ID',
//teach_date       date                                null comment '教学日期',
//teach_start_time datetime                            null comment '教学开始时间',
//teach_end_time   datetime                            null comment '教学结束时间',
//teach_num        int                                 null comment '教学次数（0：为不可预约，>0：可预约次数）',
//teach_state      int       default 0                 null comment '预约状态（0：可预约，1：已锁定，2：课后缓冲）',
//order_course_id  varchar(64)                         null comment '订单课程ID',
//state            int       default 0                 null comment '状态，0-正常， 1-删除',
//created_at       timestamp default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at       timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间'
//)
//comment '雪场的教学时间';

type SkiResortsTeachTime struct {
	ID                              int64                             `json:"id" gorm:"column:id"`                             // 自增ID
	UserID                          string                            `json:"user_id" gorm:"column:user_id"`                   // 用户ID（包括用户ID、教练ID、俱乐部ID、官方ID）
	UserType                        int                               `json:"user_type" gorm:"column:user_type"`               // 用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）
	SkiResortsID                    int                               `json:"ski_resorts_id" gorm:"column:ski_resorts_id"`     // 雪场ID
	TeachDate                       LocalDate                         `json:"teach_date" gorm:"column:teach_date"`             // 教学日期
	TeachStartTime                  LocalTime                         `json:"teach_start_time" gorm:"column:teach_start_time"` // 教学开始时间
	TeachEndTime                    LocalTime                         `json:"teach_end_time" gorm:"column:teach_end_time"`     // 教学结束时间
	TeachNum                        int                               `json:"teach_num" gorm:"column:teach_num"`               // 教学次数（0：为不可预约，>0：可预约次数）
	TeachState                      int                               `json:"teach_state" gorm:"column:teach_state"`           // 预约状态（0：可预约，1：已锁定，2：课后缓冲）
	State                           int                               `json:"state" gorm:"column:state"`                       // 状态，0-正常， 1-删除
	CreatedAt                       time.Time                         `json:"-" gorm:"column:created_at"`                      // 创建时间
	UpdatedAt                       time.Time                         `json:"-" gorm:"column:updated_at"`                      // 更新时间
	OrderCourseID                   string                            `json:"order_course_id" gorm:"-"`                        // 订单课程ID
	Title                           string                            `json:"title" gorm:"-"`                                  // 标题
	Remark                          string                            `json:"remark" gorm:"-"`                                 //备注
	SkiResortsTeachTimeOrderCourses []SkiResortsTeachTimeOrderCourses `json:"srt_order_courses" gorm:"foreignKey:SkirtID;references:ID"`
	SkiResortsTeachTimeEvent        []SkiResortsTeachTimeEvent        `json:"srt_order_event" gorm:"foreignKey:SkirtID;references:ID"`
}

func (m *SkiResortsTeachTime) TableName() string {
	return "ski_resorts_teach_time"
}

const (
	SkiTeachStateWaitAppointment = 0 //可预约
	SkiTeachStateLocked          = 1 //已锁定
	SkiTeachStateAfterClass      = 2 //课后缓冲
)
