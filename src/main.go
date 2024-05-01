package main

import (
	"context"
	custcron "github.com/CE-Thesis-2023/backend/src/internal/cron"
	"github.com/go-co-op/gocron"
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
					custhttp.WithRegistration(publicapi.ServiceRegistration()),
					custhttp.WithMiddleware(custhttp.CommonPublicMiddlewares(&configs.Public)...),
				)),
				app.WithHttpServer(custhttp.New(
					custhttp.WithMiddleware(custhttp.CommonPrivateMiddlewares(&configs.Private)...),
					custhttp.WithRegistration(privateapi.ServiceRegistration()),
					custhttp.WithGlobalConfigs(&configs.Private),
				)),
				app.WithFactoryHook(func() error {
					custdb.Init(globalCtx, configs)
					err := custdb.Migrate(custdb.Gorm(),
						&db.Transcoder{},
						&db.OpenGateIntegration{},
						&db.Camera{},
						&db.CameraGroup{},
						&db.ObjectTrackingEvent{},
						&db.OpenGateCameraSettings{},
						&db.OpenGateIntegration{},
						&db.OpenGateMqttConfiguration{},
						&db.OpenGateCameraStats{},
						&db.OpenGateDetectorStats{},
						&db.DetectablePerson{},
						&db.Snapshot{},
						&db.PersonHistory{},
						&db.TranscoderStatus{},
					)
					if err != nil {
						return err
					}
					person := &db.DetectablePerson{}
					person.Index(custdb.Gorm())

					service.Init(configs, globalCtx)
					eventsapi.Init(configs, globalCtx)
					return nil
				}),
				app.WithShutdownHook(func(ctx context.Context) {
					cancelGlobalCtx()
					custdb.Stop(ctx)
					logger.Close()
				}),
				app.WithScheduling(custcron.New(
					custcron.WithEnabled(configs.CronSchedule.Enabled),
					custcron.WithRegisterFunc(
						func(s *gocron.Scheduler) error {
							srv := service.GetWebService()
							_, err := s.Cron(configs.CronSchedule.Cron).Do(srv.DeleteOpengateCameraStats, globalCtx)
							if err != nil {
								return err
							}
							_, err = s.Cron(configs.CronSchedule.Cron).Do(srv.DeleteOpengateDetectorStats, globalCtx)
							if err != nil {
								return err
							}
							return nil
						},
					),
				)),
			}
		},
	)
}
