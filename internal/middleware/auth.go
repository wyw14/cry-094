package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Auth struct{ secret []byte }

func NewAuth(secret []byte) *Auth { return &Auth{secret: append([]byte(nil), secret...)} }
func (a *Auth) Require() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "), func(token *jwt.Token) (any, error) { return a.secret, nil })
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}
		subject, err := claims.GetSubject()
		if err != nil || subject == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token subject"})
			return
		}
		c.Set("actor_id", subject)
		c.Next()
	}
}
