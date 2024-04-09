package service

import (
	"context"
	"sync"

	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"go.uber.org/zap"
)

var once sync.Once

var (
	privateService *PrivateService
	webService     *WebService
)

func Init(c *configs.Configs, ctx context.Context) {
	once.Do(func() {
		reqreply, err := custmqtt.NewMQTTSession(ctx, &c.MqttStore)
		if err != nil {
			logger.SFatal("unable to create mqtt session",
				zap.Error(err))
			return
		}
		mediaHelper := media.NewMediaHelper(&c.MediaEngine)
		webService = NewWebService(reqreply, mediaHelper)
		privateService = NewPrivateService(
			reqreply,
			webService,
			mediaHelper,
			&c.MqttStore)
	})
}

func GetWebService() *WebService {
	return webService
}

func GetPrivateService() *PrivateService {
	return privateService
}
