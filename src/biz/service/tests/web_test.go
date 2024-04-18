package service_tests

import (
	"context"
	"log"
	"testing"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/backend/src/internal/mqtt"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"go.uber.org/zap"
)

func prepareTestWebBiz() *service.WebService {
	cvs := prepareTestBiz()

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
		cvs,
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
		PersonId: "8de74c24-8aee-4fa3-9f4a-76e4c89981c2",
	})
	if err != nil {
		t.Error(err)
	}

	logger.SInfo("GetDetectablePeoplePresignedUrl test passed",
		zap.Reflect("response", resp))
}

func TestWebService_AddDetectablePerson(t *testing.T) {
	biz := prepareTestWebBiz()
	defer custdb.Stop(context.Background())

	encodedImg := prepareTestImg("joy.jpg")

	resp, err := biz.AddDetectablePerson(context.Background(), &web.AddDetectablePersonRequest{
		Name:        "Nguyen Tran",
		Age:         "22",
		Base64Image: encodedImg,
	})
	if err != nil {
		t.Error(err)
	}

	logger.SInfo("AddDetectablePerson test passed",
		zap.Reflect("response", resp))
}

func TestWebService_DeleteDetectablePerson(t *testing.T) {
	biz := prepareTestWebBiz()
	defer custdb.Stop(context.Background())

	err := biz.DeleteDetectablePerson(context.Background(), &web.DeleteDetectablePersonRequest{
		PersonId: "8de74c24-8aee-4fa3-9f4a-76e4c89981c2",
	})
	if err != nil {
		t.Error(err)
	}
}
