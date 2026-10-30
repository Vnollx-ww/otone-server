package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"otome-server/internal/module/auth"
	"otome-server/internal/module/player"
	"otome-server/internal/module/story"
	"otome-server/internal/shared/apperr"
	"otome-server/internal/shared/config"
	"otome-server/internal/shared/httpx"
	"otome-server/internal/transport/http/middleware"
)

type dependencies struct {
	config *config.Config
	log    *slog.Logger
	db     *sqlx.DB
	redis  *redis.Client
}

func newRouter(deps dependencies) *gin.Engine {
	if deps.config.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(
		middleware.RequestID(),
		middleware.AccessLog(deps.log),
		middleware.Recovery(deps.log),
	)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/readyz", readiness(deps.db, deps.redis))

	tokens := auth.NewTokenManager(deps.config.JWTSecret, deps.config.AccessTokenTTL, deps.config.RefreshTokenTTL)
	authHandler := auth.NewHandler(auth.NewService(auth.NewRepository(deps.db), tokens))
	playerHandler := player.NewHandler(player.NewService(player.NewRepository(deps.db)))
	storyHandler := story.NewHandler(story.NewService(story.NewRepository(deps.db)))

	v1 := router.Group("/api/v1")
	authHandler.Register(v1)

	protected := v1.Group("")
	protected.Use(middleware.Auth(tokens))
	playerHandler.Register(protected)
	storyHandler.Register(protected)

	router.NoRoute(func(c *gin.Context) {
		httpx.Fail(c, apperr.NotFound("route not found"))
	})
	return router
}

func readiness(db *sqlx.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "dependency": "mysql"})
			return
		}
		if err := rdb.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "dependency": "redis"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "mysql": "ok", "redis": "ok"})
	}
}
