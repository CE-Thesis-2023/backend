package service

import (
	"context"
	"sync"

	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"github.com/Kagami/go-face"
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
		mediaHelper := media.NewMediaHelper(&c.MediaEngine, &c.S3)
		recognizer, err := face.NewRecognizer("./models")
		if err != nil {
			logger.SFatal("unable to create face recognizer",
				zap.Error(err))
		}
		computerVisionService := NewComputerVisionService(custdb.Layered(), recognizer)
		webService = NewWebService(reqreply, mediaHelper, computerVisionService)
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
