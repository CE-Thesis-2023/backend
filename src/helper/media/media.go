package media

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"go.uber.org/zap"
)

type MediaHelper struct {
	configs   *configs.MediaMtxConfigs
	s3Configs *configs.S3Storage
	s3Client  *s3.S3
}

func NewMediaHelper(configs *configs.MediaMtxConfigs, s3 *configs.S3Storage) *MediaHelper {
	h := &MediaHelper{
		configs:   configs,
		s3Configs: s3,
	}
	if err := h.initS3Client(); err != nil {
		logger.SFatal("failed to init S3 client",
			zap.Error(err))
	}
	return h
}

func (m *MediaHelper) initS3Client() error {
	awsConfigs := aws.NewConfig().
		WithEndpoint(m.s3Configs.Endpoint).
		WithRegion(m.s3Configs.Region).
		WithS3ForcePathStyle(true).
		WithCredentials(credentials.NewStaticCredentials(
			m.s3Configs.AccessKeyID,
			m.s3Configs.Secret,
			""))
	sess, err := session.NewSession(awsConfigs)
	if err != nil {
		return err
	}
	m.s3Client = s3.New(sess)
	return nil
}

func (m *MediaHelper) BuildSRTPublishUrl(streamName string) (string, error) {
	configs := m.configs

	streamUrl := &url.URL{}
	streamUrl.Scheme = "srt"
	streamUrl.Host = configs.MediaUrl
	if configs.ProviderPorts.Srt != 0 {
		streamUrl.Host = fmt.Sprintf(
			"%s:%d",
			configs.MediaUrl,
			configs.ProviderPorts.Srt)
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

func (m *MediaHelper) BuildRTSPSourceUrl(camera db.Camera) string {
	u := &url.URL{}
	u.Scheme = "rtsp"
	u.Host = camera.Ip
	if camera.Port != 0 {
		u.Host = fmt.Sprintf("%s:%d", camera.Ip, 554)
	}
	u = u.JoinPath("/ISAPI", "/Streaming", "channels", "101")
	u.User = url.UserPassword(camera.Username, camera.Password)
	url := u.String()
	logger.SDebug("RTSP source stream url",
		zap.String("url", url))
	return url
}

func (m *MediaHelper) BuildWebRTCViewStream(streamName string) string {
	configs := m.configs
	url := &url.URL{}
	url.Scheme = "https"
	url.Host = fmt.Sprintf("%s:%d",
		configs.MediaUrl,
		configs.PublishPorts.WebRTC)
	url = url.JoinPath(streamName)
	return url.String()
}

type AssetsType string

var (
	AssetsTypePeople   AssetsType = "people"
	AssetsTypeEvents   AssetsType = "events"
	AssetsTypeSnapshot AssetsType = "snapshot"
)

type UploadImageRequest struct {
	Base64Image string     `json:"base64_image"`
	Path        string     `json:"path"`
	Type        AssetsType `json:"type"`
}

func (m *MediaHelper) UploadImage(ctx context.Context, req *UploadImageRequest) error {
	decodedImg, err := base64.StdEncoding.DecodeString(req.Base64Image)
	if err != nil {
		return custerror.FormatInvalidArgument("failed to decode image: %v", err)
	}
	p := m.getImageBasePath(req.Type, req.Path)
	reader := bytes.NewReader(decodedImg)
	_, err = m.s3Client.PutObjectWithContext(
		ctx, &s3.PutObjectInput{
			Bucket: &m.s3Configs.Bucket,
			Key:    p,
			Body:   reader,
		})
	if err != nil {
		return custerror.FormatInternalError("failed to upload image: %v", err)
	}
	return nil
}

type GetImageResponse struct {
	Base64Image string `json:"base64_image"`
}

type GetImageRequest struct {
	Path string     `json:"path"`
	Type AssetsType `json:"type"`
}

func (m *MediaHelper) GetImage(ctx context.Context, req *GetImageRequest) (*GetImageResponse, error) {
	objRef, err := m.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: &m.s3Configs.Bucket,
		Key:    m.getImageBasePath(req.Type, req.Path),
	})
	if err != nil {
		return nil, custerror.FormatInternalError("failed to get image: %v", err)
	}
	rawImg, err := io.ReadAll(objRef.Body)
	if err != nil {
		return nil, custerror.FormatInternalError("failed to read image: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(rawImg)
	return &GetImageResponse{
		Base64Image: encoded,
	}, nil
}

type GetPresignedUrlRequest struct {
	Path string     `json:"path"`
	Type AssetsType `json:"type"`
}

func (m *MediaHelper) GetPresignedUrl(ctx context.Context, req *GetPresignedUrlRequest) (string, error) {
	objRef, _ := m.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &m.s3Configs.Bucket,
		Key:    m.getImageBasePath(req.Type, req.Path),
	})
	presignedUrl, err := objRef.Presign(15 * time.Minute)
	if err != nil {
		return "", custerror.FormatInternalError("failed to presign URL: %v", err)
	}
	return presignedUrl, nil
}

func (m *MediaHelper) GetImageBasePath(t AssetsType, id string) string {
	return *m.getImageBasePath(t, id)
}

func (m *MediaHelper) getImageBasePath(t AssetsType, id string) *string {
	cat := "people"
	switch t {
	case AssetsTypeEvents:
		cat = "events"
	case AssetsTypePeople:
		cat = "people"
	case AssetsTypeSnapshot:
		cat = "snapshot"
	}
	p := filepath.Join(
		m.s3Configs.PathPrefix,
		cat,
		id) + ".jpg"
	return &p
}

type DeleteImageRequest struct {
	Path string     `json:"path"`
	Type AssetsType `json:"type"`
}

func (m *MediaHelper) DeleteImage(ctx context.Context, req *DeleteImageRequest) error {
	_, err := m.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: &m.s3Configs.Bucket,
		Key:    m.getImageBasePath(req.Type, req.Path),
	})
	if err != nil {
		return custerror.FormatInternalError("failed to delete image: %v", err)
	}
	return nil
}
