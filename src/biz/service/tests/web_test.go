package service_tests

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func prepareTestWebBiz() *service.WebService {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	os.Setenv("CONFIG_FILE_PATH", "./configs.json")
	configs.Init(ctx)
	logger.Init(ctx,
		logger.WithGlobalConfigs(&configs.Get().Logger))

	custdb.Init(ctx, configs.Get())
	custdb.Migrate(custdb.Gorm(), &db.DetectablePerson{})

	mediaHelper := media.NewMediaHelper(
		&configs.Get().MediaEngine,
		&configs.Get().S3)

	sess, err := custmqtt.NewMQTTSession(
		context.Background(),
		&configs.Get().MqttStore)
	if err != nil {
		log.Fatal(err)
	}
	return service.NewWebService(
		sess,
		mediaHelper,
	)
}

func TestWebService_GetDetectablePeople(t *testing.T) {
	biz := prepareTestWebBiz()
	defer custdb.Stop(context.Background())

	resp, err := biz.GetDetectablePeople(context.Background(), &web.GetDetectablePeopleRequest{
		PersonIds: []string{},
	})
	if err != nil {
		t.Error(err)
	}

	logger.SInfo("GetDetectablePeople test passed",
		zap.Reflect("response", resp))
}

func TestWebService_GetDetectablePeoplePresignedUrl(t *testing.T) {
	biz := prepareTestWebBiz()
	defer custdb.Stop(context.Background())

	resp, err := biz.GetDetectablePersonImagePresignedUrl(context.Background(), &web.GetDetectablePeopleImagePresignedUrlRequest{
		PersonId: uuid.NewString(),
	})
	if err != nil {
		t.Error(err)
	}

	logger.SInfo("GetDetectablePeoplePresignedUrl test passed",
		zap.Reflect("response", resp))
}
