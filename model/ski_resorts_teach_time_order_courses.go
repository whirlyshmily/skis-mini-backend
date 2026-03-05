package model

import "time"

//create table ski_resorts_teach_time_order_courses
//(
//id              bigint auto_increment comment '自增ID'
//primary key,
//skirt_id        bigint                                null comment '雪场教学时间的ID',
//order_course_id varchar(64) default ''                not null comment '订单课程ID',
//state           int         default 0                 null comment '状态，0-正常， 1-删除',
//created_at      timestamp   default CURRENT_TIMESTAMP null comment '创建时间',
//updated_at      timestamp   default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时候'
//)
//comment '预约雪场的教学时间对应订单课程';

type SkiResortsTeachTimeOrderCourses struct {
	ID            int64     `json:"id" gorm:"column:id"`                           // 自增ID
	SkirtID       int64     `json:"skirt_id" gorm:"column:skirt_id"`               // 雪场教学时间的ID
	OrderCourseID string    `json:"order_course_id" gorm:"column:order_course_id"` // 订单课程ID
	State         int       `json:"-" gorm:"column:state"`                         // 状态，0-正常， 1-删除
	CreatedAt     time.Time `json:"-" gorm:"column:created_at"`                    // 创建时间
	UpdatedAt     time.Time `json:"-" gorm:"column:updated_at"`                    // 更新时候
}

func (m *SkiResortsTeachTimeOrderCourses) TableName() string {
	return "ski_resorts_teach_time_order_courses"
}
