package model

import (
	"time"
)

//create table goods_courses
//(
//id         int auto_increment comment '自增ID'
//primary key,
//good_id    varchar(64)                         not null comment '商品ID',
//pack_good_id varchar(64)                         null comment '打包起来的商品ID',
//course_id  varchar(64)                         not null comment '课程ID',
//state      int       default 0                 not null comment '状态（0：正常，1：删除）',
//created_at timestamp default CURRENT_TIMESTAMP not null comment '创建时间',
//updated_at timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
//constraint courses_pk
//unique (course_id),
//constraint goods_courses_pk
//unique (good_id, course_id)
//)
//comment '商品关联课程表' collate = utf8mb4_general_ci;

type GoodsCourses struct {
	ID         int       `json:"id" gorm:"column:id"`                     // 自增ID
	GoodID     string    `json:"good_id" gorm:"column:good_id"`           // 商品ID
	PackGoodID string    `json:"pack_good_id" gorm:"column:pack_good_id"` // 打包起来的商品ID
	CourseID   string    `json:"course_id" gorm:"column:course_id"`       // 课程ID
	State      int       `json:"state" gorm:"column:state"`               // 状态（0：正常，1：删除）
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`     // 创建时间
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`     // 更新时间
	PackGood   *Goods    `gorm:"foreignKey:GoodID;references:PackGoodID" json:"pack_good,omitempty"`
}

func (m *GoodsCourses) TableName() string {
	return "goods_courses"
}
