package custhttp

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func CommonPublicMiddlewares(configs *configs.HttpConfigs) []interface{} {
	return []interface{}{
		cors.New(cors.Config{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Origin, Content-Type, Accept",
		}),
		limiter.New(limiter.Config{
			Max:        10,
			Expiration: 1 * time.Second,
			LimitReached: func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusTooManyRequests)
			},
		}),
		etag.New(),
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),
		recover.New(recover.ConfigDefault),
		logger.New(logger.Config{
			DisableColors: true,
			Format:        "PUBLIC ${time} [${ip}:${port}] ${latency} ${method} ${status} ${path}\n",
			TimeFormat:    time.RFC3339,
		}),
		filesystem.New(filesystem.Config{
			Root: http.Dir("./public"),
		}),
		cache.New(cache.Config{
			Expiration:   time.Minute * 1,
			CacheControl: false,
			CacheHeader:  "X-Cache",
			Methods: []string{
				fiber.MethodGet,
				fiber.MethodHead,
			},
		}),
		helmet.New(helmet.ConfigDefault),
	}
}

func CommonPrivateMiddlewares(configs *configs.HttpConfigs) []interface{} {
	username := configs.Auth.Username
	token := configs.Auth.Token

	return []interface{}{
		cors.New(cors.Config{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Origin, Content-Type, Accept",
		}),
		limiter.New(limiter.Config{
			Max:        5,
			Expiration: 1 * time.Second,
			LimitReached: func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusTooManyRequests)
			},
		}),
		etag.New(),
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),
		basicauth.New(basicauth.Config{
			Users: map[string]string{
				username: token,
			},
			ContextUsername: "username",
			ContextPassword: "token",
			Unauthorized: func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusUnauthorized)
			},
		}),
		filesystem.New(filesystem.Config{
			Root: http.Dir("./public"),
		}),
		recover.New(recover.ConfigDefault),
		logger.New(logger.Config{
			DisableColors: true,
			Format:        "ADMIN ${time} [${ip}:${port}] (${latency}) [${locals:username}:${locals:token}] ${method} ${status}  ${path}\n",
			TimeFormat:    time.RFC3339,
		}),
	}
}

func GlobalErrorHandler() func(c *fiber.Ctx, err error) error {
	return func(c *fiber.Ctx, err error) error {
		customError, yes := err.(*custerror.CustomError)
		if !yes {
			switch err {
			case sql.ErrNoRows:
				customError = custerror.ErrorNotFound
			case sql.ErrTxDone, sql.ErrConnDone:
				customError = custerror.ErrorInternal
			default:
				customError = custerror.ErrorInternal
			}
		}
		return customError.
			Parent().
			Fiber(c)
	}
}
