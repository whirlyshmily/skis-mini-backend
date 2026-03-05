package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// LocalTime 自定义时间格式
type LocalTime time.Time

func (t *LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return []byte(fmt.Sprintf("\"%v\"", tTime.Format("2006-01-02 15:04:05"))), nil
}
func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	//判断给定时间是否和默认零时间的时间戳相同
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}
func (t *LocalTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// LocalDate 自定义时间格式
type LocalDate time.Time

func (t *LocalDate) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return []byte(fmt.Sprintf("\"%v\"", tTime.Format("2006-01-02"))), nil
}
func (t LocalDate) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	//判断给定时间是否和默认零时间的时间戳相同
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}
func (t *LocalDate) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalDate(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// JSONArray 定义JSON数组类型
type JSONArray []string

// Scan 实现Scan接口
func (j *JSONArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for JSONArray: %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现Value接口
func (j JSONArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// JSONIntArray 定义JSON数组类型
type JSONIntArray []int64

// Scan 实现Scan接口
func (j *JSONIntArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for JSONArray: %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现Value接口
func (j JSONIntArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type UrlJsonItem struct {
	Category string `json:"category"`
	Url      string `json:"url"`
}

type JsonUrlArray []UrlJsonItem

// Scan 实现Scan接口
func (j *JsonUrlArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for JSONArray: %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现Value接口
func (j JsonUrlArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}
