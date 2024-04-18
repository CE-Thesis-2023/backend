package helper

import (
	"encoding/json"

	"github.com/Kagami/go-face"
	"github.com/pgvector/pgvector-go"
)

func ToMap(s interface{}) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(v, &m)
	return m, nil
}

func Int(i int) *int {
	return &i
}

func ToPgvector(d face.Descriptor) pgvector.Vector {
	return pgvector.NewVector(d[:])
}
