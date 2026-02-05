package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
    allowAll := false
    for _, origin := range allowedOrigins {
        if origin == "*" {
            allowAll = true
            break
        }
    }

    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        if allowAll {
            c.Header("Access-Control-Allow-Origin", "*")
        } else if origin != "" {
            for _, allowed := range allowedOrigins {
                if strings.EqualFold(origin, allowed) {
                    c.Header("Access-Control-Allow-Origin", origin)
                    break
                }
            }
        }

        c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Client-Type")
        c.Header("Access-Control-Allow-Credentials", "true")

        if c.Request.Method == http.MethodOptions {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
