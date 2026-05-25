package middleware

import "github.com/gin-gonic/gin"

func RequestLogger() gin.HandlerFunc {
	return gin.Logger()
}
