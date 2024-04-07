package events

type CommandResponse struct {
	Type string                 `json:"commandType"`
	Info map[string]interface{} `json:"info"`
}
