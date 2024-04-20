package custhttp

import (
	"net/http"
	"time"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/zap"
)

func CommonPublicMiddlewares(configs *configs.HttpConfigs) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		CORSMiddleware(),
		ginzap.Ginzap(logger.Logger(), time.RFC3339, true),
		gzip.Gzip(gzip.BestCompression),
		gin.Recovery(),
	}
}

func CommonPrivateMiddlewares(configs *configs.HttpConfigs) []gin.HandlerFunc {
	username := configs.Auth.Username
	token := configs.Auth.Token

	return []gin.HandlerFunc{
		CORSMiddleware(),
		ginzap.Ginzap(logger.Logger(), time.RFC3339, true),
		gzip.Gzip(gzip.BestCompression),
		gin.Recovery(),
		gin.BasicAuth(gin.Accounts{
			username: token,
		}),
	}
}

func ToHTTPErr(err error, c *gin.Context) {
	customError, yes := err.(*custerror.CustomError)
	if yes {
		customError.Gin(c)
		return
	}
	madeCustom := custerror.NewError(
		err.Error(),
		http.StatusInternalServerError)
	madeCustom.Gin(c)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, HEAD, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
