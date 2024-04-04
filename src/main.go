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
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"github.com/CE-Thesis-2023/backend/src/models/db"

	"go.uber.org/zap"
)

func main() {
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
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					custdb.Init(ctx, configs)

					custdb.Migrate(custdb.Gorm(),
						&db.Transcoder{},
						&db.Camera{},
						&db.CameraGroup{},
						&db.ObjectTrackingEvent{},
						&db.OpenGateIntegration{},
					)
					service.Init()

					custmqtt.InitClient(
						context.Background(),
						custmqtt.WithClientGlobalConfigs(&configs.MqttStore),
						custmqtt.WithOnReconnection(eventsapi.Register),
						custmqtt.WithOnConnectError(func(err error) {
							logger.Error("MQTT Connection failed", zap.Error(err))
						}),
						custmqtt.WithClientError(eventsapi.ClientErrorHandler),
						custmqtt.WithOnServerDisconnect(eventsapi.DisconnectHandler),
						custmqtt.WithHandlerRegister(eventsapi.RouterHandler()),
					)
					return nil
				}),
				app.WithShutdownHook(func(ctx context.Context) {
					custdb.Stop(ctx)
					custmqtt.StopClient(ctx)
					logger.Close()
				}),
			}
		},
	)
}
