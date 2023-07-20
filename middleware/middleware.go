package middleware

import (
	"net/http"
	"time"

	"github.com/galihfebrizki/dbo-api/config"
	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/responses"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1000, nil))
			c.Abort()
			return
		}

		// Extract the token from the header
		tokenString := authHeader[len("Bearer "):]

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Get().Secret), nil
		})

		if err != nil {
			logrus.WithField(helper.GetRequestIDContext(helper.GetGinContext(c))).Error(err)
			c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
			c.Abort()
			return
		}

		// Validate the token
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			if claims.UserId != "" && claims.Username == "" {
				c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
				c.Abort()
				return
			}

			c.Set("Username", claims.Username)
			c.Set("UserId", claims.UserId)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1001, nil))
			c.Abort()
		}
	}
}

func GenerateJWTToken(username string, userId string) (string, string, error) {
	secret := config.Get().Secret
	// Define the secret key used to sign the JWT
	secretKey := []byte(secret)

	// Calculate the expiration time to 23:59 of the current day
	now := time.Now()
	expirationTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 59, time.UTC)

	// Create the claims for the token
	claims := Claims{
		Username: username,
		UserId:   userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	return tokenString, expirationTime.Format("02-01-2006 15:04:05"), nil
}

func GetToken(c *gin.Context) string {
	// Get the token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, *responses.NewGenericResponse(1000, nil))
		c.Abort()
		return ""
	}

	return authHeader[len("Bearer "):]
}
