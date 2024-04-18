package media

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/aws/smithy-go/logging"
	"go.uber.org/zap"
)

type MediaHelper struct {
	configs       *configs.MediaMtxConfigs
	s3Configs     *configs.S3Storage
	s3Client      *s3.Client
	presignClient *s3.PresignClient
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

type s3EndpointResolver struct {
	Endpoint string
}

func (r *s3EndpointResolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	u, err := url.Parse(r.Endpoint)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}
	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}

func (m *MediaHelper) initS3Client() error {
	awsConfigs, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(m.s3Configs.Region),
		config.WithLogger(logging.NewStandardLogger(os.Stdout)),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			m.s3Configs.AccessKeyID,
			m.s3Configs.Secret,
			""),
		),
	)
	if err != nil {
		logger.SError("failed to load S3 config",
			zap.Error(err))
		return err
	}
	client := s3.NewFromConfig(awsConfigs, func(o *s3.Options) {
		o.EndpointResolverV2 = &s3EndpointResolver{
			Endpoint: m.s3Configs.Endpoint,
		}
	})
	m.s3Client = client
	m.presignClient = s3.NewPresignClient(client)
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
		u.Host = fmt.Sprintf("%s:%d", camera.Ip, camera.Port)
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
	url.Scheme = "http"
	url.Host = fmt.Sprintf("%s:%d",
		configs.MediaUrl,
		configs.PublishPorts.WebRTC)
	url = url.JoinPath(streamName)
	return url.String()
}

type UploadImageRequest struct {
	Base64Image string `json:"base64_image"`
	Path        string `json:"path"`
}

func (m *MediaHelper) UploadImage(ctx context.Context, req *UploadImageRequest) error {
	decodedImg, err := base64.StdEncoding.DecodeString(req.Base64Image)
	if err != nil {
		return custerror.FormatInvalidArgument("failed to decode image: %v", err)
	}
	reader := bytes.NewReader(decodedImg)
	_, err = m.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &m.s3Configs.Bucket,
		Key:    m.getImageBasePath(req.Path),
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

func (m *MediaHelper) GetImage(ctx context.Context, path string) (*GetImageResponse, error) {
	objRef, err := m.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &m.s3Configs.Bucket,
		Key:    m.getImageBasePath(path),
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

func (m *MediaHelper) GetPresignedUrl(ctx context.Context, path string) (string, error) {
	presignedReq, err := m.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &m.s3Configs.Bucket,
		Key:    m.getImageBasePath(path),
	},
		s3.WithPresignExpires(15*time.Minute),
	)
	if err != nil {
		return "", custerror.FormatInternalError("failed to get presigned URL: %v", err)
	}
	return presignedReq.URL, nil
}

func (m *MediaHelper) getImageBasePath(id string) *string {
	p := filepath.Join(m.s3Configs.PathPrefix, "people", id)
	return &p
}
