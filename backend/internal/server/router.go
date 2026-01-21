package server

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/server/routes"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// SetupRouter 配置路由器中间件和路由
func SetupRouter(
	r *gin.Engine,
	handlers *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
	redisClient *redis.Client,
) *gin.Engine {
	// 应用中间件
	r.Use(middleware2.Logger())
	r.Use(middleware2.CORS(cfg.CORS))
	r.Use(middleware2.SecurityHeaders(cfg.Security.CSP))

	// Serve embedded frontend with settings injection if available
	if web.HasEmbeddedFrontend() {
		frontendServer, err := web.NewFrontendServer(settingService)
		if err != nil {
			log.Printf("Warning: Failed to create frontend server with settings injection: %v, using legacy mode", err)
			r.Use(web.ServeEmbeddedFrontend())
		} else {
			// Register cache invalidation callback
			settingService.SetOnUpdateCallback(frontendServer.InvalidateCache)
			r.Use(frontendServer.Middleware())
		}
	}

	uploadDir := resolveUploadDir()
	if uploadDir != "" {
		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			log.Printf("Warning: failed to create upload dir: %v", err)
		} else {
			uploads := r.Group("/uploads")
			uploads.Use(func(c *gin.Context) {
				c.Header("X-Content-Type-Options", "nosniff")
				c.Header("Content-Security-Policy", "sandbox; default-src 'none'; img-src 'self' data:")
				c.Header("Cache-Control", "no-store")
				c.Next()
			})
			uploads.Static("/", uploadDir)
		}
	}

	// 注册路由
	registerRoutes(r, handlers, jwtAuth, adminAuth, apiKeyAuth, apiKeyService, subscriptionService, opsService, cfg, redisClient)

	return r
}

func resolveUploadDir() string {
	baseDir := resolveDataDir()
	if baseDir == "" {
		return ""
	}
	return filepath.Join(baseDir, "uploads")
}

func resolveDataDir() string {
	if value := strings.TrimSpace(os.Getenv("DATA_DIR")); value != "" {
		return value
	}

	dockerDataDir := "/app/data"
	if info, err := os.Stat(dockerDataDir); err == nil && info.IsDir() {
		testFile := filepath.Join(dockerDataDir, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0o644); err == nil {
			_ = os.Remove(testFile)
			return dockerDataDir
		}
	}

	return "."
}

// registerRoutes 注册所有 HTTP 路由
func registerRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	cfg *config.Config,
	redisClient *redis.Client,
) {
	// 通用路由（健康检查、状态等）
	routes.RegisterCommonRoutes(r)

	// API v1
	v1 := r.Group("/api/v1")

	// 注册各模块路由
	routes.RegisterPublicRoutes(v1, h)
	routes.RegisterAuthRoutes(v1, h, jwtAuth, redisClient)
	routes.RegisterUserRoutes(v1, h, jwtAuth)
	routes.RegisterAdminRoutes(v1, h, adminAuth, redisClient)
	routes.RegisterGatewayRoutes(r, h, apiKeyAuth, apiKeyService, subscriptionService, opsService, cfg)
}
