package custhttp

import (
	"net/http"
	"time"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/zap"
)

func CommonPublicMiddlewares(configs *configs.HttpConfigs) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.Default(),
		ginzap.Ginzap(logger.Logger(), time.RFC3339, true),
		gzip.Gzip(gzip.BestCompression),
		gin.Recovery(),
	}
}

func CommonPrivateMiddlewares(configs *configs.HttpConfigs) []gin.HandlerFunc {
	username := configs.Auth.Username
	token := configs.Auth.Token

	return []gin.HandlerFunc{
		cors.Default(),
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
	return
}
