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
	registerAuthRoutes(rg, deps)
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
