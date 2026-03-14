package dao

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"skis-admin-backend/enum"
	"skis-admin-backend/model"
	"time"

	"gorm.io/gorm"
)

func GenerateId(prefix string) string {
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), GenerateSecureRandomString(6))
}

// 生成a-z,0-9的随机字符串
func GenerateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func IsElementInArray(arr []string, element string) bool {
	for _, item := range arr {
		if item == element {
			return true
		}
	}
	return false
}

// 生成指定长度的随机字符串（使用hex编码的随机字节）
func GenerateSecureRandomString(n int) string {
	bytes := make([]byte, n/2) // 因为hex编码，所以实际字节长度是n/2
	_, err := rand.Read(bytes)
	if err != nil {
		return GenerateRandomString(n)
	}
	return hex.EncodeToString(bytes)
}

func RandomStringBySeed(seed int64, length int) string {
	rand.Seed(seed)
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// 取2个字符串切片的交集
func StrIntersection(s1, s2 []string) []string {
	m := make(map[string]bool)
	for _, v := range s1 {
		m[v] = true
	}
	for _, v := range s2 {
		m[v] = true
	}

	var intersection []string
	for k := range m {
		intersection = append(intersection, k)
	}
	return intersection

}

func GetStartTime(startTime string) string {
	return startTime + " 00:00:00"
}

func GetEndTime(endTime string) string {
	return endTime + " 23:59:59"
}

// SafeUpdateTeachState 安全更新课程教学状态（带乐观锁）
// expectedState: 期望的当前状态，只有当数据库中的状态等于此值时才允许更新
// 返回值: 如果 RowsAffected == 0 表示状态已被其他操作修改，发生了并发冲突
func SafeUpdateTeachState(tx *gorm.DB, orderCourseId string, expectedState model.TeachState, updates map[string]interface{}) error {
	result := tx.Model(model.OrdersCourses{}).
		Where("order_course_id = ? AND teach_state = ?", orderCourseId, expectedState).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return enum.NewErr(enum.OptimisticLockErr, "课程状态已变更，请刷新后重试")
	}
	return nil
}

// SafeUpdateTeachStateWithState 安全更新课程教学状态并设置新状态（带乐观锁）
func SafeUpdateTeachStateWithState(tx *gorm.DB, orderCourseId string, expectedState, newState model.TeachState) error {
	return SafeUpdateTeachState(tx, orderCourseId, expectedState, map[string]interface{}{
		"teach_state": newState,
	})
}
