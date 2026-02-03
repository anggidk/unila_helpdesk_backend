package handler

import (
    "net/http"

    "unila_helpdesk_backend/internal/domain"
    "unila_helpdesk_backend/internal/middleware"
    "unila_helpdesk_backend/internal/service"

    "github.com/gin-gonic/gin"
)

type TicketHandler struct {
    tickets *service.TicketService
}

type commentRequest struct {
    Message string `json:"message"`
}

func NewTicketHandler(tickets *service.TicketService) *TicketHandler {
    return &TicketHandler{tickets: tickets}
}

func (handler *TicketHandler) RegisterRoutes(public *gin.RouterGroup, auth *gin.RouterGroup) {
    auth.GET("/tickets", handler.listTickets)
    public.GET("/tickets/search", handler.searchTickets)
    public.GET("/tickets/:id", handler.getTicket)
    public.POST("/tickets/guest", handler.createGuestTicket)
    auth.POST("/tickets", handler.createTicket)
    auth.POST("/tickets/:id", handler.updateTicket)
    auth.POST("/tickets/:id/delete", handler.deleteTicket)
    auth.POST("/tickets/:id/comments", handler.addComment)
}

func (handler *TicketHandler) listTickets(c *gin.Context) {
    user, ok := middleware.GetUser(c)
    if !ok {
        respondError(c, http.StatusUnauthorized, "token dibutuhkan")
        return
    }
    result, err := handler.tickets.ListTickets(user)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, result)
}

func (handler *TicketHandler) searchTickets(c *gin.Context) {
    query := c.Query("q")
    user, hasUser := middleware.GetUser(c)
    guestOnly := false
    if hasUser && user.Role == domain.RoleGuest {
        guestOnly = true
    }

    result, err := handler.tickets.SearchTickets(query, guestOnly)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, result)
}

func (handler *TicketHandler) getTicket(c *gin.Context) {
    ticketID := c.Param("id")
    user, hasUser := middleware.GetUser(c)
    var userPtr *domain.User
    if hasUser {
        userPtr = &user
    }
    result, err := handler.tickets.GetTicket(userPtr, ticketID)
    if err != nil {
        respondError(c, http.StatusForbidden, err.Error())
        return
    }
    respondOK(c, result)
}

func (handler *TicketHandler) createTicket(c *gin.Context) {
    user, ok := middleware.GetUser(c)
    if !ok {
        respondError(c, http.StatusUnauthorized, "token dibutuhkan")
        return
    }
    var req service.TicketCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    result, err := handler.tickets.CreateTicket(c, user, req)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondCreated(c, result)
}

func (handler *TicketHandler) createGuestTicket(c *gin.Context) {
    var req service.GuestTicketCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    result, err := handler.tickets.CreateGuestTicket(c, req)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondCreated(c, result)
}

func (handler *TicketHandler) updateTicket(c *gin.Context) {
    user, ok := middleware.GetUser(c)
    if !ok {
        respondError(c, http.StatusUnauthorized, "token dibutuhkan")
        return
    }
    var req service.TicketUpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    result, err := handler.tickets.UpdateTicket(c, user, c.Param("id"), req)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondOK(c, result)
}

func (handler *TicketHandler) deleteTicket(c *gin.Context) {
    user, ok := middleware.GetUser(c)
    if !ok {
        respondError(c, http.StatusUnauthorized, "token dibutuhkan")
        return
    }
    if err := handler.tickets.DeleteTicket(user, c.Param("id")); err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondOK(c, gin.H{"deleted": true})
}

func (handler *TicketHandler) addComment(c *gin.Context) {
    user, ok := middleware.GetUser(c)
    if !ok {
        respondError(c, http.StatusUnauthorized, "token dibutuhkan")
        return
    }
    var req commentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    result, err := handler.tickets.AddComment(user, c.Param("id"), req.Message)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondOK(c, result)
}
