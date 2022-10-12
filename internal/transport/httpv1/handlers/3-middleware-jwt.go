package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := h.tokenManager.TokenValid(c)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)

			return
		}
		c.Next()
	}
}
