package router

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"rbackeeper/config"
	"rbackeeper/internal/controller"
	"rbackeeper/middleware"

	_ "rbackeeper/docs" // swagger 生成的文档
)

// Setup 构造并注册路由引擎
func Setup(cfg config.CFG) *gin.Engine {
	if cfg.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.Recovery(), middleware.RequestLog(), middleware.CORS())

	health := &controller.HealthController{}
	// 根路径健康探针，供负载均衡/k8s 探活
	router.GET("/healthz", health.Healthz)
	router.GET("/readyz", health.Readyz)

	base := router.Group("/api")
	{
		registerAuthRoutes(base)
		registerUserRoutes(base)
		registerRoleRoutes(base)
		registerMenuRoutes(base)
		registerDeptRoutes(base)
		registerDictRoutes(base)
		registerLogRoutes(base)
	}

	// Swagger 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

// Run 启动 HTTP 服务并支持优雅关闭；appCancel 用于在退出时取消后台 goroutine（MQ 消费者等）。
func Run(cfg config.CFG, appCancel context.CancelFunc) {
	handler := Setup(cfg)
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info().
			Str("swagger_url", fmt.Sprintf("http://127.0.0.1:%d/swagger/index.html", cfg.Server.Port)).
			Str("addr", srv.Addr).
			Msg("HTTP 服务启动")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP 服务异常退出")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("收到退出信号，开始优雅关闭")
	// 通知后台 goroutine（MQ 消费者）退出
	if appCancel != nil {
		appCancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP 服务关闭失败")
	}
	log.Info().Msg("HTTP 服务已关闭")
}
