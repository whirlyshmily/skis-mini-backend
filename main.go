package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"skis-admin-backend/global"
	"skis-admin-backend/initialize"
	"sync"
	"syscall"
	"time"
)

func main() {
	// 1. 初始化配置和组件
	if err := initializeApplication(); err != nil {
		os.Exit(1)
	}

	// 2. 创建带有优雅关闭的HTTP服务器
	runServerWithGracefulShutdown()
}

func initializeApplication() error {
	// 1.初始化yaml配置
	initialize.InitConfig()

	// 2.初始化日志信息
	if err := initialize.InitLogger(); err != nil {
		fmt.Printf("初始化日志失败: %v", err)
		return err
	}

	// 4.初始化mysql
	if err := initialize.InitMysqlDB(); err != nil {
		global.Lg.Error("[initializeApplication] 初始化mysql失败:", zap.Error(err))
		return err
	}

	// 7.初始化路由
	initialize.InitRouters()

	if err := initialize.InitWxPayClient(); err != nil {
		global.Lg.Error("[initializeApplication] 初始化微信支付失败:", zap.Error(err))
		return err
	}

	initialize.InitOss()

	if err := initialize.InitCron(); err != nil {
		global.Lg.Error("[initializeApplication] 初始化定时器失败:", zap.Error(err))
		return err
	}

	return nil
}

func runServerWithGracefulShutdown() {
	// 等待组用于等待所有后台服务关闭
	var wg sync.WaitGroup

	// 3. 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", global.Config.Port),
		Handler: initialize.InitRouters(),
	}

	// 4. 设置信号捕获
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 5. 启动HTTP服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		global.Lg.Info("服务器启动", zap.Int("port", global.Config.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.Lg.Error("HTTP服务器错误", zap.Error(err))
			sigChan <- os.Interrupt
			return
		}
	}()

	// 等待终止信号
	<-sigChan
	global.Lg.Info("收到终止信号，开始优雅关闭...")

	// 6. 开始优雅关闭流程
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// 关闭HTTP服务器
	if err := server.Shutdown(shutdownCtx); err != nil {
		global.Lg.Error("HTTP服务器关闭错误", zap.Error(err))
	}

	// 等待所有服务完成关闭
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// 等待完成或超时
	select {
	case <-done:
		global.Lg.Info("所有服务已完全关闭")
	case <-shutdownCtx.Done():
		global.Lg.Warn("优雅关闭超时，强制退出")
	}
}
