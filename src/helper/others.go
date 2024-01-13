package helper

import "github.com/bytedance/sonic"

func ToMap(s interface{}) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	v, err := sonic.Marshal(s)
	if err != nil {
		return nil, err
	}
	sonic.Unmarshal(v, &m)
	return m, nil
}

func Int(i int) *int {
	return &i
}
