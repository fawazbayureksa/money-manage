package middleware

import (
    "net/http"
    "strings"

    "my-api/utils"

    "github.com/dgrijalva/jwt-go"
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            utils.LogWarningf("Auth failed: Missing authorization header from %s", c.ClientIP())
            utils.JSONError(c, http.StatusUnauthorized, "Authorization header is required")
            c.Abort()
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 {
            utils.LogWarningf("Auth failed: Invalid token format from %s", c.ClientIP())
            utils.JSONError(c, http.StatusUnauthorized, "Invalid token format")
            c.Abort()
            return
        }

        token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
            return []byte("secret"), nil // Use the same secret key as in GenerateToken
        })

        if err != nil || !token.Valid {
            utils.LogWarningf("Auth failed: Invalid token from %s - %v", c.ClientIP(), err)
            utils.JSONError(c, http.StatusUnauthorized, "Invalid token")
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            utils.LogWarningf("Auth failed: Invalid token claims from %s", c.ClientIP())
            utils.JSONError(c, http.StatusUnauthorized, "Invalid token claims")
            c.Abort()
            return
        }

        userID, ok := claims["user_id"].(float64)
        if !ok {
            utils.LogWarningf("Auth failed: Invalid user ID in token from %s", c.ClientIP())
            utils.JSONError(c, http.StatusUnauthorized, "Invalid user ID in token")
            c.Abort()
            return
        }

        utils.LogInfof("Auth success: User %d authenticated from %s", uint(userID), c.ClientIP())
        c.Set("user_id", uint(userID))
        c.Next()
    }
}