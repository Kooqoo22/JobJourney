package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/Kooqoo22/JobJourney/backend/config"
	authhandler "github.com/Kooqoo22/JobJourney/backend/internal/auth/handler"
	authrepo "github.com/Kooqoo22/JobJourney/backend/internal/auth/repository"
	authusecase "github.com/Kooqoo22/JobJourney/backend/internal/auth/usecase"
	"github.com/Kooqoo22/JobJourney/backend/internal/database"
	"github.com/Kooqoo22/JobJourney/backend/internal/middleware"
	apphandler "github.com/Kooqoo22/JobJourney/backend/internal/application/handler"
	apprepo "github.com/Kooqoo22/JobJourney/backend/internal/application/repository"
	appusecase "github.com/Kooqoo22/JobJourney/backend/internal/application/usecase"
	statshandler "github.com/Kooqoo22/JobJourney/backend/internal/stats/handler"
	statsrepo "github.com/Kooqoo22/JobJourney/backend/internal/stats/repository"
	statsusecase "github.com/Kooqoo22/JobJourney/backend/internal/stats/usecase"
	profilehandler "github.com/Kooqoo22/JobJourney/backend/internal/profile/handler"
	profilerepo "github.com/Kooqoo22/JobJourney/backend/internal/profile/repository"
	profileusecase "github.com/Kooqoo22/JobJourney/backend/internal/profile/usecase"
	"github.com/Kooqoo22/JobJourney/backend/pkg/mailer"
	"github.com/Kooqoo22/JobJourney/backend/pkg/token"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type Dependencies struct {
	Config *config.Config
	DB     *sqlx.DB
	Token  *token.Manager
}

func New(deps Dependencies) *gin.Engine {
	if deps.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.Logger(),
		middleware.CORS(deps.Config.CORS),
		middleware.ErrorHandler(),
	)

	r.GET("/health", healthHandler(deps.DB))

	api := r.Group("/api/v1")
	registerRoutes(api, deps)

	return r
}

func registerRoutes(rg *gin.RouterGroup, deps Dependencies) {
	authMW := middleware.Auth(deps.Token)
	registerAuthRoutes(rg, deps)
	registerProfileRoutes(rg, deps, authMW)
	registerApplicationRoutes(rg, deps, authMW)
	registerStatsRoutes(rg, deps, authMW)
}

func registerStatsRoutes(rg *gin.RouterGroup, deps Dependencies, authMW gin.HandlerFunc) {
	repo := statsrepo.New(deps.DB)
	uc := statsusecase.New(repo)
	h := statshandler.New(uc)

	stats := rg.Group("/stats", authMW)
	stats.GET("/summary", h.GetSummary)
	stats.GET("/applications", h.GetAnalytics)
}

func registerProfileRoutes(rg *gin.RouterGroup, deps Dependencies, authMW gin.HandlerFunc) {
	repo := profilerepo.New(deps.DB)
	tx := database.NewTxManager(deps.DB)
	uc := profileusecase.New(repo, tx)
	h := profilehandler.New(uc)

	profile := rg.Group("/profile", authMW)
	profile.GET("", h.GetProfile)
	profile.PUT("", h.UpdateProfile)
	profile.PATCH("/password", h.ChangePassword)
	profile.PATCH("/preferences", h.UpdatePreferences)
	profile.DELETE("", h.DeleteAccount)
}

func registerApplicationRoutes(rg *gin.RouterGroup, deps Dependencies, authMW gin.HandlerFunc) {
	repo := apprepo.New(deps.DB)
	eventRepo := apprepo.NewEventRepository(deps.DB)
	tx := database.NewTxManager(deps.DB)
	uc := appusecase.New(repo, eventRepo, tx)
	h := apphandler.New(uc)

	apps := rg.Group("/applications", authMW)
	apps.GET("", h.ListApplications)
	apps.POST("", h.CreateApplication)
	apps.GET("/:id", h.GetApplication)
	apps.PUT("/:id", h.UpdateApplication)
	apps.DELETE("/:id", h.DeleteApplication)
	apps.PATCH("/:id/restore", h.RestoreApplication)
	apps.PATCH("/:id/status", h.ChangeStatus)
	apps.PATCH("/:id/archive", h.ToggleArchive)
	apps.GET("/:id/events", h.ListEvents)
	apps.POST("/:id/events", h.CreateEvent)
	apps.PUT("/:id/events/:event_id", h.UpdateEvent)
	apps.DELETE("/:id/events/:event_id", h.DeleteEvent)
}

func registerAuthRoutes(rg *gin.RouterGroup, deps Dependencies) {
	txManager := database.NewTxManager(deps.DB)
	userRepo := authrepo.NewUserRepository(deps.DB)
	emailTokenRepo := authrepo.NewEmailTokenRepository(deps.DB)
	refreshTokenRepo := authrepo.NewRefreshTokenRepository(deps.DB)

	authUsecase := authusecase.New(authusecase.Deps{
		UserRepo:        userRepo,
		EmailTokens:     emailTokenRepo,
		Refresh:         refreshTokenRepo,
		Tx:              txManager,
		Tokens:          deps.Token,
		Mailer:          mailer.NewSMTP(deps.Config.SMTP),
		FrontendURL:     deps.Config.App.FrontendURL,
		DefaultTimezone: deps.Config.App.DefaultTimezone,
		VerifyTokenTTL:  deps.Config.Auth.VerifyTokenTTL,
		ResetTokenTTL:   deps.Config.Auth.ResetTokenTTL,
	})
	h := authhandler.NewAuthHandler(authUsecase)

	auth := rg.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/verify-email", h.VerifyEmail)
	auth.POST("/resend-verification", h.ResendVerification)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)
	auth.POST("/forgot-password", h.ForgotPassword)
	auth.POST("/reset-password", h.ResetPassword)
	auth.POST("/logout-all", middleware.Auth(deps.Token), h.LogoutAll)
}

func healthHandler(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, utils.NewMessage("database unavailable"))
			return
		}
		c.JSON(http.StatusOK, utils.NewSuccess("service healthy", gin.H{"status": "ok"}))
	}
}
