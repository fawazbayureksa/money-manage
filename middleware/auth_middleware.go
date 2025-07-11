package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/dgrijalva/jwt-go"
    "my-api/utils"
    "strings"
    "net/http"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            utils.JSONError(c, http.StatusUnauthorized, "Authorization header is required")
            c.Abort()
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 {
            utils.JSONError(c, http.StatusUnauthorized, "Invalid token format")
            c.Abort()
            return
        }

        token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
            return []byte("secret"), nil // Use the same secret key as in GenerateToken
        })

        if err != nil || !token.Valid {
            utils.JSONError(c, http.StatusUnauthorized, "Invalid token")
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            utils.JSONError(c, http.StatusUnauthorized, "Invalid token claims")
            c.Abort()
            return
        }

        userID, ok := claims["user_id"].(float64)
        if !ok {
            utils.JSONError(c, http.StatusUnauthorized, "Invalid user ID in token")
            c.Abort()
            return
        }

        c.Set("user_id", uint(userID))
        c.Next()
    }
}