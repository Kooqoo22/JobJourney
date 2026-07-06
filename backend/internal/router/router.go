package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/Kooqoo22/JobJourney/backend/config"
	"github.com/Kooqoo22/JobJourney/backend/internal/middleware"
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
	_ = rg
	_ = deps
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
