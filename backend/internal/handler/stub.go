package handler

import (
	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

func NotImplemented(c *gin.Context) {
	response.NotImplemented(c)
}

func SystemInfo(c *gin.Context) {
	response.Success(c, gin.H{
		"name":    "ai-review",
		"version": "0.0.0-dev",
	})
}
