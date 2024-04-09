package main

import (
	"context"
	"time"

	eventsapi "github.com/CE-Thesis-2023/backend/src/api/events"
	privateapi "github.com/CE-Thesis-2023/backend/src/api/private"
	publicapi "github.com/CE-Thesis-2023/backend/src/api/public"
	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/internal/app"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	custhttp "github.com/CE-Thesis-2023/backend/src/internal/http"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"

	"go.uber.org/zap"
)

func main() {
	// terrible code but a workaround
	globalCtx, cancelGlobalCtx := context.WithCancel(context.Background())

	app.Run(
		time.Second*10,
		func(configs *configs.Configs, zl *zap.Logger) []app.Optioner {
			return []app.Optioner{
				app.WithHttpServer(custhttp.New(
					custhttp.WithGlobalConfigs(&configs.Public),
					custhttp.WithErrorHandler(custhttp.GlobalErrorHandler()),
					custhttp.WithRegistration(publicapi.ServiceRegistration()),
					custhttp.WithMiddleware(custhttp.CommonPublicMiddlewares(&configs.Public)...),
				)),
				app.WithHttpServer(custhttp.New(
					custhttp.WithErrorHandler(custhttp.GlobalErrorHandler()),
					custhttp.WithMiddleware(custhttp.CommonPrivateMiddlewares(&configs.Private)...),
					custhttp.WithRegistration(privateapi.ServiceRegistration()),
					custhttp.WithGlobalConfigs(&configs.Private),
				)),
				app.WithFactoryHook(func() error {
					custdb.Init(globalCtx, configs)
					custdb.Migrate(custdb.Gorm(),
						&db.Transcoder{},
						&db.OpenGateIntegration{},
						&db.Camera{},
						&db.CameraGroup{},
						&db.ObjectTrackingEvent{},
						&db.OpenGateCameraSettings{},
						&db.OpenGateIntegration{},
						&db.OpenGateMqttConfiguration{},
					)

					service.Init(configs, globalCtx)
					eventsapi.Init(configs, globalCtx)
					return nil
				}),
				app.WithShutdownHook(func(ctx context.Context) {
					cancelGlobalCtx()
					custdb.Stop(ctx)
					logger.Close()
				}),
			}
		},
	)
}
