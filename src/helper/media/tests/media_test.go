package media_tests

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/CE-Thesis-2023/backend/src/helper/media"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
)

func initBiz() *media.MediaHelper {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	os.Setenv("CONFIG_FILE_PATH", "../../../../configs.json")
	configs.Init(ctx)
	logger.Init(ctx,
		logger.WithGlobalConfigs(&configs.Get().Logger))

	custdb.Init(ctx, configs.Get())
	mediaHelper := media.NewMediaHelper(
		&configs.Get().MediaEngine,
		&configs.Get().S3)
	return mediaHelper
}

func TestMediaHelper_RTSPSourceUrl(t *testing.T) {
	mediaHelper := initBiz()

	camera := db.Camera{
		Ip:       "10.40.30.50",
		Port:     80,
		Username: "admin",
		Password: "admin",
	}
	url := mediaHelper.BuildRTSPSourceUrl(camera)

	if url == "" {
		t.Error("url is empty")
	}

	fmt.Println(url)
}

func TestMediaHelper_SRTPublishPort(t *testing.T) {
	mediaHelper := initBiz()

	url, err := mediaHelper.BuildSRTPublishUrl("test")
	if err != nil {
		t.Error(err)
	}

	if url == "" {
		t.Error("url is empty")
	}

	fmt.Println(url)
}

func TestMediaHelper_WebRTCViewUrl(t *testing.T) {
	mediaHelper := initBiz()

	url := mediaHelper.BuildWebRTCViewStream("test")

	if url == "" {
		t.Error("url is empty")
	}

	fmt.Println(url)
}

func prepareTestImg(path string) string {
	content, err := os.ReadFile("./testdata/" + path)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(content)
}

func TestMediaHelper_UploadImage(t *testing.T) {
	mediaHelper := initBiz()
	encodedImg := prepareTestImg("bona4.jpg")

	err := mediaHelper.UploadImage(context.Background(), &media.UploadImageRequest{
		Base64Image: encodedImg,
		Path:        "bona4.jpg",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestMediaHelper_GetPresignedUrl(t *testing.T) {
	mediaHelper := initBiz()

	url, err := mediaHelper.GetPresignedUrl(
		context.Background(),
		"bona4.jpg")
	if err != nil {
		t.Error(err)
	}

	if url == "" {
		t.Error("url is empty")
	}

	fmt.Println(url)
}
