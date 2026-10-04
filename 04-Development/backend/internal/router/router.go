package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kel1/backend/internal/config"
	"kel1/backend/internal/handler"
	"kel1/backend/internal/middleware"
	"kel1/backend/internal/models"
	"kel1/backend/internal/repository"
	"kel1/backend/internal/service"
	"kel1/backend/internal/utils"
)

func Setup(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS(cfg.CORSOrigins))

	// --- Dependency wiring: repository -> service -> handler ---
	userRepo := repository.NewUserRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	userService := service.NewUserService(userRepo)

	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	r.GET("/health", healthHandler.Check)

	r.NoRoute(func(c *gin.Context) {
		utils.Error(c, http.StatusNotFound, "route tidak ditemukan", nil)
	})

	api := r.Group("/api/v1")
	{
		// Public
		auth := api.Group("/auth")
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)

		// Butuh login
		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.GET("/me", userHandler.Me)

			// Khusus admin
			users := protected.Group("/users")
			users.Use(middleware.RequireRole(models.RoleAdmin))
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		// TODO: tambahkan group route fitur baru di sini setelah topik project fix
	}

	return r
}
