package main

import (
	"context"
	"fmt"
	"gin-user-center/app/common"
	"gin-user-center/app/config"
	mongo "gin-user-center/app/database/mongodb"
	"gin-user-center/app/database/mysql"
	"gin-user-center/app/database/redis"
	"gin-user-center/app/database/table"
	"gin-user-center/app/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var logger = common.Logger

func main() {
	// 默认使用mysql
	mysql.Init()
	defer mysql.DB.Close()

	table.Init()

	if config.Conf.Server.UserMongo {
		// mongo初始化
		mongo.Init()
	}
	if config.Conf.Server.UserRedis {
		// 初始化redis服务
		redis.Init()
	}

	RunServer()
}

func RunServer() {
	router := router.InitRouter()
	// router.Run(config.Conf.Server.Port)

	// 优雅关停
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Conf.Server.Port),
		Handler: router,
	}

	log.Println(fmt.Sprintf("Listening and serving HTTP on Port: %d, Pid: %d", config.Conf.Server.Port, os.Getpid()))

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	// 创建系统信号接收器
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	logger.Info("shutdown server...")

	// 创建5s的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown:", err)
	}
	logger.Info("server exiting...")
}
