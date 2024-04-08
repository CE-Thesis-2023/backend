package media

import (
	"fmt"
	"net/url"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"go.uber.org/zap"
)

type mediaHelper struct {
	configs *configs.MediaMtxConfigs
}

func NewMediaHelper(configs *configs.MediaMtxConfigs) *mediaHelper {
	return &mediaHelper{
		configs: configs,
	}
}

func (m *mediaHelper) BuildSRTPublishUrl(streamName string) (string, error) {
	configs := m.configs

	streamUrl := &url.URL{}
	streamUrl.Scheme = "srt"
	streamUrl.Host = configs.MediaUrl
	if configs.PublishPorts.Srt != 0 {
		streamUrl.Host = fmt.Sprintf(
			"%s:%d",
			configs.MediaUrl,
			configs.PublishPorts.Srt)
	}
	queries := streamUrl.Query()
	queries.Add("streamid", fmt.Sprintf("publish:%s", streamName))
	rawQuery, err := url.QueryUnescape(queries.Encode())
	if err != nil {
		logger.SError("failed to unescape SRT input stream parameters",
			zap.Error(err))
		return "", err
	}
	streamUrl.RawQuery = rawQuery

	url := streamUrl.String()
	return url, nil
}

func (m *mediaHelper) BuildRTSPSourceUrl(camera db.Camera) string {
	u := &url.URL{}
	u.Scheme = "rtsp"
	u.Host = camera.Ip
	if camera.Port != 0 {
		u.Host = fmt.Sprintf("%s:%d", camera.Ip, camera.Port)
	}
	u = u.JoinPath("/ISAPI", "/Streaming", "channels", "101")
	u.User = url.UserPassword(camera.Username, camera.Password)
	url := u.String()
	logger.SDebug("RTSP source stream url",
		zap.String("url", url))
	return url
}

func (m *mediaHelper) BuildWebRTCViewStream(streamName string) string {
	configs := m.configs
	url := &url.URL{}
	url.Scheme = "http"
	url.Host = fmt.Sprintf("%s:%d",
		configs.MediaUrl,
		configs.PublishPorts.WebRTC)
	url = url.JoinPath(streamName)
	return url.String()
}
