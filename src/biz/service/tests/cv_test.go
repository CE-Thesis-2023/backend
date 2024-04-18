package service_tests

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/CE-Thesis-2023/backend/src/biz/service"
	"github.com/CE-Thesis-2023/backend/src/helper"
	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/backend/src/internal/db"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/Kagami/go-face"
	"github.com/google/uuid"
)

func prepareTestBiz() *service.ComputerVisionService {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	os.Setenv("CONFIG_FILE_PATH", "./configs.json")
	configs.Init(ctx)
	logger.Init(ctx,
		logger.WithGlobalConfigs(&configs.Get().Logger))

	custdb.Init(ctx, configs.Get())
	custdb.Migrate(custdb.Gorm(), &db.DetectablePerson{})

	recognizer, err := face.NewRecognizer("../../../../models")
	if err != nil {
		log.Fatal(err)
	}
	return service.NewComputerVisionService(
		custdb.NewLayeredDb(ctx, configs.Get()),
		recognizer)
}

func prepareTestImg(path string) string {
	content, err := os.ReadFile("./testdata/" + path)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(content)
}

func TestComputerVisionService_DetectFaces(t *testing.T) {
	biz := prepareTestBiz()
	defer custdb.Stop(context.Background())
	img := prepareTestImg("nayoung.jpg")

	faces, err := biz.Detect(context.Background(), &service.DetectRequest{
		Base64Image: img,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(faces) > 1 {
		t.Fatal("expected 1 face, got", len(faces))
	}

	if len(faces) == 0 {
		t.Fatal("expected 1 face, got 0")
	}
}

func TestComputerVisionService_RecordFaces(t *testing.T) {
	biz := prepareTestBiz()
	defer custdb.Stop(context.Background())
	img := prepareTestImg("nayoung.jpg")

	faces, err := biz.Detect(context.Background(), &service.DetectRequest{
		Base64Image: img,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(faces) > 1 {
		t.Fatal("expected 1 face, got", len(faces))
	}

	if len(faces) == 0 {
		t.Fatal("expected 1 face, got 0")
	}

	people, err := biz.Search(context.Background(), &service.SearchRequest{
		Vector:     faces[0].Descriptor,
		TopKResult: 1,
	})
	switch err {
	case nil:
		t.Log("person already exists")
		return
	default:
		if !errors.Is(err, custerror.ErrorNotFound) {
			t.Fatal(err)
		}
	}

	if len(people) == 0 {
		if err := biz.Record(context.Background(), &db.DetectablePerson{
			PersonId: uuid.NewString(),
			Name:     "Joy",
			Age:      "25",
			ImageUrl: "https://example.com/bona.jpg",
			Embedding: helper.ToPgvector(
				faces[0].
					Descriptor),
		}); err != nil {
			t.Fatal(err)
		}
	}

	people, err = biz.Search(context.Background(), &service.SearchRequest{
		Vector:     faces[0].Descriptor,
		TopKResult: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(people) == 0 {
		t.Fatal("expected 1 person, got 0")
	}
}

func TestComputerVisionService_SearchExistingFace(t *testing.T) {
	biz := prepareTestBiz()
	defer custdb.Stop(context.Background())
	tests := []string{
		"nayoung.jpg",
		"joy.jpg",
		"bona4.jpg",
		"chaeyeon2.jpg",
		"eunseo2.jpg",
		"jimin.jpg",
	}
	vectors := make([]face.Descriptor, len(tests))
	for i, test := range tests {
		faces, err := biz.Detect(context.Background(), &service.DetectRequest{
			Base64Image: prepareTestImg(test),
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(faces) == 0 {
			t.Fatalf("expected 1 face, got 0 for %s", test)
		}
		vectors[i] = faces[0].Descriptor
	}

	for i, test := range tests {
		testname := fmt.Sprintf("TestComputerVisionService_SearchExistingFace:%s", test)
		t.Run(testname, func(t *testing.T) {
			v := vectors[i]
			res, err := biz.Search(context.Background(), &service.SearchRequest{
				Vector:     v,
				TopKResult: 1,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(res) == 0 {
				t.Fatal("expected 1 person, got 0")
			}
		})
	}
}

/**
 * Is roughly 0.0095476s -> 9.5476ms
 */
func BenchmarkComputerVisionService_SearchExistingFaces(b *testing.B) {
	img := prepareTestImg("nayoung.jpg")
	biz := prepareTestBiz()
	defer custdb.Stop(context.Background())
	faces, err := biz.Detect(context.Background(), &service.DetectRequest{
		Base64Image: img,
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i += 1 {
		resp, err := biz.Search(context.Background(), &service.SearchRequest{
			Vector:     faces[0].Descriptor,
			TopKResult: 1,
		})
		if err != nil {
			b.Fatal(err)
		}
		if len(resp) == 0 {
			b.Fatal("expected 1 person, got 0")
		}
	}
}

/**
 * Is roughly 103.112ms
 */
func BenchmarkComputerVisionService_DetectFaces(b *testing.B) {
	img := prepareTestImg("nayoung.jpg")
	biz := prepareTestBiz()
	defer custdb.Stop(context.Background())
	b.ResetTimer()

	for i := 0; i < b.N; i += 1 {
		_, err := biz.Detect(context.Background(), &service.DetectRequest{
			Base64Image: img,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
