package transcoder

type TranscoderEventMessage struct {
	Type         string `json:"type"`
	Action       string `json:"action"`
	TranscoderId string `json:"transcoderId"`
	OpenGateId   string `json:"openGateId"`
	CameraId     string `json:"cameraId"`
	GroupId      string `json:"groupId"`
	Payload      []byte `json:"payload"`
}
