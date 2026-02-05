package handler

import (
    "errors"
    "net/http"

    "unila_helpdesk_backend/internal/service"

    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    auth *service.AuthService
}

type loginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type refreshRequest struct {
    RefreshToken string `json:"refresh_token"`
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
    return &AuthHandler{auth: auth}
}

func (handler *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
    router.POST("/auth/login", handler.login)
    router.POST("/auth/refresh", handler.refreshToken)
}

func (handler *AuthHandler) login(c *gin.Context) {
    var req loginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    clientType := c.GetHeader("X-Client-Type")
    result, err := handler.auth.LoginWithPasswordClient(req.Username, req.Password, clientType)
    if err != nil {
        if errors.Is(err, service.ErrAdminWebOnly) {
            respondError(c, http.StatusForbidden, err.Error())
            return
        }
        respondError(c, http.StatusUnauthorized, err.Error())
        return
    }
    respondOK(c, result)
}

func (handler *AuthHandler) refreshToken(c *gin.Context) {
    var req refreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    clientType := c.GetHeader("X-Client-Type")
    result, err := handler.auth.RefreshWithTokenClient(req.RefreshToken, clientType)
    if err != nil {
        if errors.Is(err, service.ErrAdminWebOnly) {
            respondError(c, http.StatusForbidden, err.Error())
            return
        }
        respondError(c, http.StatusUnauthorized, err.Error())
        return
    }
    respondOK(c, result)
}
