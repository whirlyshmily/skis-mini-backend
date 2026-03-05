package test

import (
	"fmt"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"testing"
	"time"
)

func TestPingRoute(t *testing.T) {
	today := time.Now().Format("2006-01-02") + " 23:59:59"
	var data []model.OrdersCourses
	global.DB.Table("orders_courses").
		Where("teach_time < ?", today).
		Where(" is_check = ? and state = 0", model.IsCheckNo).
		Find(&data)
	fmt.Println(len(data))

	//aa, _ := dao.NewGoodsDao(context.Background(), global.DB).GetMaxPriceByUserId("123")
	//fmt.Println(aa)
}
