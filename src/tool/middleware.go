package tool

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")
		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token != "user" && token != "admin" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		role := "user"
		if token == "admin" {
			role = "admin"
		}
		c.Set("role", role)
		c.Next()
	}
}

func AuthAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")
		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token == "user" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if token != "admin" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}
