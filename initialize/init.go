package initialize

import (
	"fmt"
	"skis-admin-backend/config"
	cron2 "skis-admin-backend/cron"
	"skis-admin-backend/global"
	"skis-admin-backend/middlewares"
	"skis-admin-backend/router"
	"skis-admin-backend/services"
	"skis-admin-backend/utils"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/*
* 初始化配置项
 */
func InitConfig() {
	// 实例化viper
	v := viper.New()
	v.SetConfigName("env")
	v.SetConfigType("toml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	// 声明一个ServerConfig类型的实例
	cfg := &config.Config{}
	// 给serverConfig初始值
	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	// 传递给全局变量
	global.Config = cfg
	color.Blue("initConfig", global.Config)
}

/*
* 初始化路由
 */
func InitRouters() *gin.Engine {
	Router := gin.Default()

	// 加载自定义中间件
	// Router.Use(middlewares.GinLogger(), middlewares.GinRecovery(true), middlewares.CORSMiddleware())
	Router.Use(middlewares.GinLogger(), middlewares.GinRecovery(true))
	// 添加全局健康检查路由
	Router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, skis-mini-backend!",
		})
	})
	// 路由分组
	ApiGroup := Router.Group("/api/")
	router.FeedsRouter(ApiGroup)

	router.UsersRouter(ApiGroup)
	router.ClubsRouter(ApiGroup)
	router.TagsRouter(ApiGroup)
	router.SkiResortRouter(ApiGroup)
	router.CoursesRouter(ApiGroup)
	router.CoachLevelRouter(ApiGroup)
	router.CoachesRouter(ApiGroup)
	router.PointsRecordsRouter(ApiGroup)
	router.CertificateConfigsRouter(ApiGroup)
	router.OrdersRouter(ApiGroup)
	router.GoodsRouter(ApiGroup)
	router.MoneyRouter(ApiGroup)
	router.OssRouter(ApiGroup)
	router.OrdersCoursesRouter(ApiGroup)
	router.OrdersCoursesRecordsRouter(ApiGroup)
	router.AdminRouter(ApiGroup)

	return Router
}

// InitLogger 初始化Logger
func InitLogger() error {
	// 实例化zap配置

	cfg := zap.NewDevelopmentConfig()
	// 配置日志的输出地址
	cfg.OutputPaths = []string{
		fmt.Sprintf("%slog_%s.log", global.Config.Log.Path, utils.GetNowFormatTodayTime()),
		"stdout", // "stdout" 表示同时将日志输出到标准输出流（控制台）。这样就可以将日志同时输出到文件和控制台
	}

	cfg.Level.SetLevel(zapcore.Level(global.Config.Log.Level))

	// 创建logger实例
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger) // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
	global.Lg = logger         // 注册到全局变量中
	return nil
}

// 初始化Mysql
func InitMysqlDB() error {
	mysqlInfo := global.Config.Mysql
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlInfo.Username, mysqlInfo.Password, mysqlInfo.Host,
		mysqlInfo.Port, mysqlInfo.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: global.NewZapLogger(global.Lg),
	})
	if err != nil {
		global.Lg.Error("[InitMysqlDB] 链接mysql异常:", zap.Error(err))
		return err
	}
	global.DB = db

	return nil
}

func InitWxPayClient() error {
	return services.InitWxPayClient()
}

func InitCron() error {
	c := cron.New()

	// 添加一个自定义任务
	job, err := c.AddJob("@every 1m", cron2.MyJob{})
	if err != nil {
		return err
	}
	fmt.Println("=================", job)

	orderCourseJob, err := c.AddJob("@every 5m", cron2.OrdersCoursesStateJob{})
	if err != nil {
		return err
	}
	fmt.Println("=================", orderCourseJob)

	verifyCourseJob, err := c.AddJob("@every 5m", cron2.VerifyCourseJob{})
	if err != nil {
		return err
	}
	fmt.Println("=================", verifyCourseJob)

	coachApplyJob, err := c.AddJob("@every 5m", cron2.CoachApplyJob{})
	if err != nil {
		return err
	}
	fmt.Println("=================", coachApplyJob)

	clubJob, err := c.AddJob("@every 10m", cron2.ClubJob{})
	if err != nil {
		return err
	}
	fmt.Println("=================", clubJob)

	// 启动调度器
	c.Start()
	return nil
}

func InitOss() {
	services.InitOss()
}
