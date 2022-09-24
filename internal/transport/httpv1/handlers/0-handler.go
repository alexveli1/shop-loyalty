package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alexveli/diploma/internal/config"
	"github.com/alexveli/diploma/internal/service"
	"github.com/alexveli/diploma/pkg/auth"
	"github.com/alexveli/diploma/pkg/hash"
)

type Handler struct {
	services     *service.Services
	tokenManager *auth.Manager
	hasher       hash.Hasher
}

func NewHandler(services *service.Services, tokenManager *auth.Manager, hasher hash.Hasher) *Handler {
	return &Handler{
		services:     services,
		tokenManager: tokenManager,
		hasher:       hasher,
	}
}

func (h *Handler) Init(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.Use(
		gin.Recovery(),
		gin.Logger(),
		corsMiddleware,
	)

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
	api := router.Group("/api/user")
	{
		api.POST("/register", h.UserRegister)
		api.POST("/login", h.UserLogin)
	}
	orders := api.Group("/orders")
	{
		orders.POST("/", h.ProcessOrder)
		orders.GET("/", h.GetOrders)
		orders.Use(h.JwtAuthMiddleware())
	}
	balance := api.Group("/balance")
	{
		balance.GET("/", h.GetBalance)
		balance.POST("/withdraw", h.Withdraw)
		balance.Use(h.JwtAuthMiddleware())
	}
	withdrawals := api.Group("/withdrawals")
	{
		withdrawals.GET("/", h.GetWithdrawls)
		withdrawals.Use(h.JwtAuthMiddleware())
	}
}

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
