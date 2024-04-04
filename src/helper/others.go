package helper

import "encoding/json"

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
