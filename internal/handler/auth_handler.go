package handler

import (
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

type guestRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type refreshRequest struct {
    RefreshToken string `json:"refresh_token"`
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
    return &AuthHandler{auth: auth}
}

func (handler *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
    router.POST("/auth/login", handler.login)
    router.POST("/auth/guest", handler.guestLogin)
    router.POST("/auth/refresh", handler.refreshToken)
}

func (handler *AuthHandler) login(c *gin.Context) {
    var req loginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    result, err := handler.auth.LoginWithPassword(req.Username, req.Password)
    if err != nil {
        respondError(c, http.StatusUnauthorized, err.Error())
        return
    }
    respondOK(c, result)
}

func (handler *AuthHandler) guestLogin(c *gin.Context) {
    var req guestRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    result, err := handler.auth.GuestLogin(req.Name, req.Email)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
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
    result, err := handler.auth.RefreshWithToken(req.RefreshToken)
    if err != nil {
        respondError(c, http.StatusUnauthorized, err.Error())
        return
    }
    respondOK(c, result)
}
