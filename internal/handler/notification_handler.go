package handler

import (
	"net/http"

	"unila_helpdesk_backend/internal/middleware"
	"unila_helpdesk_backend/internal/service"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notifications *service.NotificationService
}

func NewNotificationHandler(notifications *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifications: notifications}
}

func (handler *NotificationHandler) RegisterRoutes(auth *gin.RouterGroup) {
	auth.GET("/notifications", handler.listNotifications)
	auth.POST("/notifications/fcm", handler.registerFcm)
	auth.POST("/notifications/fcm/unregister", handler.unregisterFcm)
}

func (handler *NotificationHandler) listNotifications(c *gin.Context) {
	user, ok := middleware.GetUser(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "token dibutuhkan")
		return
	}
	result, err := handler.notifications.List(user)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(c, result)
}

func (handler *NotificationHandler) registerFcm(c *gin.Context) {
	user, ok := middleware.GetUser(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "token dibutuhkan")
		return
	}
	var req service.FCMRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "payload tidak valid")
		return
	}
	if err := handler.notifications.RegisterToken(user, req); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(c, gin.H{"registered": true})
}

func (handler *NotificationHandler) unregisterFcm(c *gin.Context) {
	user, ok := middleware.GetUser(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "token dibutuhkan")
		return
	}
	var req service.FCMUnregisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "payload tidak valid")
		return
	}
	if err := handler.notifications.UnregisterToken(user, req); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(c, gin.H{"unregistered": true})
}
