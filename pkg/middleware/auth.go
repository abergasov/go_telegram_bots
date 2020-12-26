package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthControllerMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token != c.GetHeader("Token") {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "Invalid request"})
			c.Abort()
			return
		}
		c.Next()
	}
}
