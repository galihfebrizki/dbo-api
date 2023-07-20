package helper

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/random"
)

func GetGinContext(c *gin.Context) context.Context {
	ctx := context.WithValue(c.Request.Context(), RequestIDContextKey, c.GetHeader(XRequestIDHeaderKey))
	return ctx
}

func GetContext(c context.Context) context.Context {
	ctx := context.WithValue(c, RequestIDContextKey, random.String(32))
	return ctx
}

func SetRequestIDToContext(c context.Context, reqId string) context.Context {
	ctx := context.WithValue(c, RequestIDContextKey, reqId)
	return ctx
}

func GetRequestIDContext(ctx context.Context) (string, interface{}) {
	val := ctx.Value(RequestIDContextKey)
	if val == nil {
		val = ""
	}

	return RequestIDContextKey, ctx.Value(RequestIDContextKey)
}

func GenerateRandomString(length int) string {
	return random.String(uint8(length))
}

func MD5(message string) string {
	hash := md5.New()

	hash.Write([]byte(message))
	hashBytes := hash.Sum(nil)

	return hex.EncodeToString(hashBytes)
}
