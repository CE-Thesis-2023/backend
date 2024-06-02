package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CE-Thesis-2023/backend/src/helper"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/Kagami/go-face"
	"github.com/Masterminds/squirrel"
	"go.uber.org/zap"
)

type ComputerVisionService struct {
	db         *custdb.LayeredDb
	builder    squirrel.StatementBuilderType
	recognizer *face.Recognizer
}

func NewComputerVisionService(db *custdb.LayeredDb, recognizer *face.Recognizer) *ComputerVisionService {
	return &ComputerVisionService{
		db: db,
		builder: squirrel.
			StatementBuilder.
			PlaceholderFormat(squirrel.Dollar),
		recognizer: recognizer,
	}
}

type DetectRequest struct {
	Base64Image string `json:"base64_image"`
}

type DetectResponse []face.Face

func (r DetectResponse) String() string {
	m, _ := json.Marshal(r)
	return string(m)
}

func (s *ComputerVisionService) Detect(ctx context.Context, req *DetectRequest) (DetectResponse, error) {
	img := s.decodeImage(req.Base64Image)
	if !s.isJpeg(img) {
		return nil, custerror.FormatInvalidArgument("image must be in jpeg format")
	}
	faces, err := s.recognizer.Recognize(img)
	if err != nil {
		return nil, custerror.FormatInternalError("failed to recognize faces")
	}
	resp := DetectResponse(faces)
	return resp, nil
}

type SearchRequest struct {
	Vector     face.Descriptor `json:"vector"`
	TopKResult int             `json:"topKResult"`
}

func (r *SearchRequest) validate() error {
	if len(r.Vector) == 0 {
		return custerror.FormatInvalidArgument("vector is missing")
	}
	if r.TopKResult <= 0 {
		return custerror.FormatInvalidArgument("topKResult must be greater than 0")
	}
	return nil

}

func (s *ComputerVisionService) Search(ctx context.Context, req *SearchRequest) ([]db.DetectablePerson, error) {
	if err := req.validate(); err != nil {
		return nil, custerror.FormatInvalidArgument("invalid search request: %v", err)
	}
	faces := make([]db.DetectablePerson, req.TopKResult)
	if err := s.doVectorSearch(ctx, req, &faces); err != nil {
		return nil, err
	}
	if len(faces) == 0 {
		return nil, custerror.FormatNotFound("no matching person found")
	}
	return faces, nil
}

func (s *ComputerVisionService) doVectorSearch(ctx context.Context, req *SearchRequest, resp interface{}) error {
	q := s.builder.Select("*").From("detectable_people")
	vt := helper.ToPgvector(req.Vector).String()
	q = q.Suffix(fmt.Sprintf("WHERE embedding <-> '%s' < 0.5 LIMIT %d",
		vt,
		req.TopKResult))
	sqlQ, args, _ := q.ToSql()
	logger.SInfo("doVectorSearch query",
		zap.String("query", sqlQ),
		zap.Any("args", args))
	if err := s.db.Select(ctx, q, resp); err != nil {
		return custerror.FormatInternalError("failed to search vector: %v", err)
	}
	logger.SDebug("doVectorSearch result",
		zap.Any("result", resp))
	return nil
}

func (s *ComputerVisionService) Record(ctx context.Context, m *db.DetectablePerson) error {
	return s.recordVector(ctx, m)
}

func (s *ComputerVisionService) Remove(ctx context.Context, id string) error {
	return s.removeVector(ctx, id)
}

func (s *ComputerVisionService) recordVector(ctx context.Context, m *db.DetectablePerson) error {
	q := s.builder.Insert("detectable_people").
		Columns(m.Fields()...).
		Values(m.Values()...)
	if err := s.db.Insert(ctx, q); err != nil {
		return custerror.FormatInternalError("failed to record person: %v", err)
	}
	return nil
}

func (s *ComputerVisionService) removeVector(ctx context.Context, id string) error {
	q := s.builder.Delete("detectable_people").
		Where(squirrel.Eq{"person_id": id})
	if err := s.db.Delete(ctx, q); err != nil {
		return custerror.FormatInternalError("failed to remove person: %v", err)
	}
	return nil
}

func (s *ComputerVisionService) decodeImage(base64Image string) []byte {
	res, _ := base64.StdEncoding.DecodeString(base64Image)
	return res
}

func (s *ComputerVisionService) isJpeg(img []byte) bool {
	return http.DetectContentType(img) == "image/jpeg"
}
