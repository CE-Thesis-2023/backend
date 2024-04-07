package transcoder

type TranscoderEventMessage struct {
	Type         string `json:"type"`
	Action       string `json:"action"`
	TranscoderId string `json:"transcoderId"`
	OpenGateId   string `json:"openGateId"`
	CameraId     string `json:"cameraId"`
	Payload      []byte `json:"payload"`
}
