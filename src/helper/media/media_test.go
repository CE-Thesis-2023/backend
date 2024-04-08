package media

import (
	"fmt"
	"testing"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	"github.com/CE-Thesis-2023/backend/src/models/db"
)

func TestMediaHelper_RTSPSourceUrl(t *testing.T) {
	configs := configs.MediaMtxConfigs{
		Host:     "api.mediamtx.ntranlab.com",
		MediaUrl: "103.165.142.15",
		PublishPorts: configs.MtxPorts{
			WebRTC: 8889,
		},
		ProviderPorts: configs.MtxPorts{
			Srt: 8890,
		},
		Api: 9997,
	}
	mediaHelper := NewMediaHelper(&configs)

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
	configs := configs.MediaMtxConfigs{
		Host:     "api.mediamtx.ntranlab.com",
		MediaUrl: "103.165.142.15",
		PublishPorts: configs.MtxPorts{
			WebRTC: 8889,
		},
		ProviderPorts: configs.MtxPorts{
			Srt: 8890,
		},
		Api: 9997,
	}

	mediaHelper := NewMediaHelper(&configs)
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
	configs := configs.MediaMtxConfigs{
		Host:     "api.mediamtx.ntranlab.com",
		MediaUrl: "103.165.142.15",
		PublishPorts: configs.MtxPorts{
			WebRTC: 8889,
		},
		ProviderPorts: configs.MtxPorts{
			Srt: 8890,
		},
		Api: 9997,
	}

	mediaHelper := NewMediaHelper(&configs)
	url := mediaHelper.BuildWebRTCViewStream("test")

	if url == "" {
		t.Error("url is empty")
	}

	fmt.Println(url)
}
