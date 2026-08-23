package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRouter(handler *Handler, auth gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	metrics := &Metrics{}
	router.Use(RequestID(), Security(), CORS("http://localhost:5173"), RateLimit(120, time.Minute), metrics.Middleware(), Recovery(), Timeout(15*time.Second))
	router.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	router.GET("/metrics", metrics.Handler)
	api := router.Group("/api/v1")
	api.POST("/auth/login", handler.Login)
	api.POST("/auth/refresh", handler.Refresh)
	api.Use(auth)
	api.POST("/teams", handler.CreateTeam)
	api.POST("/libraries", handler.CreateLibrary)
	api.POST("/artifacts", handler.Upload)
	api.GET("/artifacts", handler.ListArtifacts)
	api.GET("/artifacts/:id/content", handler.DownloadArtifact)
	api.POST("/analyses", handler.Analyze)
	api.POST("/precheck-plans", handler.GeneratePlan)
	api.POST("/precheck-plans/:id/submit", handler.SubmitPlan)
	api.POST("/precheck-plans/:id/review", handler.ReviewPlan)
	api.POST("/precheck-plans/:id/sign", handler.SignPlan)
	return router
}
