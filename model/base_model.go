package model

import "time"

type BaseModel struct {
	Id        int64     `gorm:"primaryKey" json:"id"`
	State     uint8     `gorm:"column:state" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}
