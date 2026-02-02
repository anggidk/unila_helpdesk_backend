package middleware

import (
    "net/http"
    "strings"

    "unila_helpdesk_backend/internal/domain"
    "unila_helpdesk_backend/internal/repository"
    "unila_helpdesk_backend/internal/service"

    "github.com/gin-gonic/gin"
)

const ContextUserKey = "authUser"

func AuthMiddleware(auth *service.AuthService, users *repository.UserRepository, required bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
        if header == "" {
            if required {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token tidak ditemukan"})
                return
            }
            c.Next()
            return
        }

        parts := strings.SplitN(header, " ", 2)
        if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
            if required {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "format token tidak valid"})
                return
            }
            c.Next()
            return
        }

        claims, err := auth.ParseToken(parts[1])
        if err != nil {
            if required {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid"})
                return
            }
            c.Next()
            return
        }

        user, err := users.FindByID(claims.UserID)
        if err != nil {
            if required {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user tidak ditemukan"})
                return
            }
            c.Next()
            return
        }

        c.Set(ContextUserKey, *user)
        c.Next()
    }
}

func RequireRole(role domain.UserRole) gin.HandlerFunc {
    return func(c *gin.Context) {
        raw, exists := c.Get(ContextUserKey)
        if !exists {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token dibutuhkan"})
            return
        }
        user := raw.(domain.User)
        if user.Role != role {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "akses ditolak"})
            return
        }
        c.Next()
    }
}

func GetUser(c *gin.Context) (domain.User, bool) {
    raw, exists := c.Get(ContextUserKey)
    if !exists {
        return domain.User{}, false
    }
    user, ok := raw.(domain.User)
    return user, ok
}
