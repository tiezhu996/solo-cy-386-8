package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/config"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
	"github.com/marketpal/marketpal/internal/service"
)

// Register 统一注册全部路由（/healthz + /api/v1）。
func Register(
	r *gin.Engine,
	cfg *config.Config,
	logger *slog.Logger,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	addressHandler *handler.AddressHandler,
	cartHandler *handler.CartHandler,
	orderHandler *handler.OrderHandler,
	messageHandler *handler.MessageHandler,
	reviewHandler *handler.ReviewHandler,
	auditHandler *handler.AuditHandler,
	wsHandler *handler.WSHandler,
	uploadHandler *handler.UploadHandler,
	auditService *service.AuditService,
) {
	// 全局中间件：请求追踪、结构化日志、错误处理、限流、操作审计。
	r.Use(
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.ErrorHandler(logger),
		middleware.RateLimit(300),
		middleware.Audit(auditService),
	)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "marketpal-backend"})
	})
	// 上传目录静态托管。
	r.Static("/uploads", cfg.UploadDir)

	api := r.Group("/api/v1")
	{
		RegisterUserRoutes(api, userHandler, cfg.JWTSecret)
		RegisterProductRoutes(api, productHandler, cfg.JWTSecret)
		RegisterAddressRoutes(api, addressHandler, cfg.JWTSecret)
		RegisterCartRoutes(api, cartHandler, cfg.JWTSecret)
		RegisterOrderRoutes(api, orderHandler, cfg.JWTSecret)
		RegisterMessageRoutes(api, messageHandler, wsHandler, cfg.JWTSecret)
		RegisterReviewRoutes(api, reviewHandler, cfg.JWTSecret)
		RegisterAuditRoutes(api, auditHandler, cfg.JWTSecret)
		RegisterUploadRoutes(api, uploadHandler, cfg.JWTSecret)
	}}
