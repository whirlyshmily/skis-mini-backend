package test

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"skis-admin-backend/cron"
	"skis-admin-backend/global"
	"skis-admin-backend/initialize"
	"skis-admin-backend/model"
	"testing"
	"time"
)

// TestMain 是测试包的主入口，用于初始化和清理
func TestMain(m *testing.M) {
	// 获取项目根目录（假设 test 目录在项目根目录下）
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取工作目录失败：%v\n", err)
		os.Exit(1)
	}

	// 向上一级目录找到项目根路径
	projectRoot := filepath.Dir(wd)

	// 切换到项目根目录，这样 viper 才能找到 env.toml
	if err := os.Chdir(projectRoot); err != nil {
		fmt.Printf("切换目录失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("测试运行目录：%s\n", projectRoot)

	// 1. 初始化配置
	initialize.InitConfig()

	// 2. 初始化日志
	if err := initialize.InitLogger(); err != nil {
		fmt.Printf("初始化日志失败：%v\n", err)
		os.Exit(1)
	}

	// 3. 初始化数据库
	if err := initialize.InitMysqlDB(); err != nil {
		global.Lg.Error("初始化数据库失败", zap.Error(err))
		os.Exit(1)
	}

	// 4. 运行测试
	m.Run()

	// 5. 测试完成后的清理工作（如果有）
	// 例如关闭数据库连接等
}

func TestPingRoute(t *testing.T) {
	job := cron.OrdersCoursesStateJob{}
	job.Run()

	return
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

func TestPingRoute1(t *testing.T) {
	today := time.Now().Format("2006-01-02") + " 23:59:59"
	var data []model.OrdersCourses

	// 执行查询
	global.DB.Table("orders_courses").
		Where("teach_time < ?", today).
		Where("is_check = ? and state = 0", model.IsCheckNo).
		Find(&data)

	// 记录日志
	t.Logf("查询条件：teach_time < %s", today)
	t.Logf("查询结果：%d 条记录", len(data))

	// 添加断言
	if data == nil {
		t.Fatal("查询结果为 nil")
	}

	// 如果有数据，打印第一条记录
	if len(data) > 0 {
		t.Logf("第一条记录：%+v", data[0])
	}

	fmt.Println(len(data))
}
