package model

import "time"

type Tags struct {
	Id        int       `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	State     int       `gorm:"column:state" json:"-"`                                                             //0-正常，1-删除
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"-"`                // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"-"` // 更新时间
}

func (Tags) TableName() string {
	return "tags"
}

type TagsList struct {
	Tags
	CoachRefCnt  int `json:"coach_ref_cnt"`  // 教练引用数
	CourseRefCnt int `json:"course_ref_cnt"` // 课程引用数
}
