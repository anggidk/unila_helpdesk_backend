package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, payload any) {
    c.JSON(http.StatusOK, gin.H{"data": payload})
}

func respondCreated(c *gin.Context, payload any) {
    c.JSON(http.StatusCreated, gin.H{"data": payload})
}

func respondError(c *gin.Context, status int, message string) {
    c.JSON(status, gin.H{"error": message})
}
