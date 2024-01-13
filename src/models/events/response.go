package events

import "github.com/CE-Thesis-2023/ltd/src/models/events"

type CommandResponse struct {
	Type events.CommandType     `json:"commandType"`
	Info map[string]interface{} `json:"info"`
}
