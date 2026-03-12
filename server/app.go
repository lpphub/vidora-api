package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vidora-api/server/middleware"

	"vidora-api/infra"
	"vidora-api/module/auth"
	"vidora-api/module/tag"
	"vidora-api/module/transcode"
	"vidora-api/module/user"
	"vidora-api/module/video"
	"vidora-api/shared/mod"

	"github.com/gin-gonic/gin"
	"github.com/lpphub/goweb/ext/logx"
	"github.com/lpphub/goweb/monitor"
)

// App 应用
type App struct {
	engine *gin.Engine
	server *http.Server
}

// New 创建应用
func New() *App {
	return &App{engine: gin.New()}
}

// Run 启动应用
func (a *App) Run() {
	a.init()
	a.start()
	a.waitForShutdown()
}

func (a *App) init() {
	if err := infra.Init(); err != nil {
		panic(fmt.Sprintf("Failed to init: %v", err))
	}
	a.setupRouter()
}

func (a *App) setupRouter() {
	r := a.engine

	// 中间件
	r.Use(gin.Recovery(), logx.GinAccessLog(logx.WithSkipPaths("/metrics", "/health")))
	r.Use(middleware.Cors())

	// 系统路由
	monitor.RegisterMetrics(r)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 业务路由 - 按依赖顺序注册
	for _, m := range a.initModules() {
		m.RegisterRoutes(r.Group(""))
	}
}

func (a *App) initModules() []mod.Module {
	// 初始化所有模块
	userMod := user.Init(infra.DB)
	authMod := auth.Init(userMod.Service)
	tagMod := tag.Init(infra.DB)
	videoMod := video.Init(infra.DB, tagMod.Service)
	transcodeMod := transcode.Init(infra.DB)

	return []mod.Module{userMod, authMod, tagMod, videoMod, transcodeMod}
}

func (a *App) start() {
	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", infra.Cfg.Server.Port),
		Handler: a.engine,
	}

	go func() {
		log.Printf("Server starting on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server start failed: %v", err)
		}
	}()
}

func (a *App) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server shutdown completed")
	}
}
