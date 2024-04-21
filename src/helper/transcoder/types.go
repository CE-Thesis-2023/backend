package transcoder

const (
	OPENGATE_EVENTS    = "events"
	OPENGATE_STATS     = "stats"
	OPENGATE_AVAILABLE = "available"
	OPENGATE_SNAPSHOT  = "snapshot"
)

type TranscoderEventMessage struct {
	Type         string `json:"type"`
	TranscoderId string `json:"transcoderId"`

	Payload []byte `json:"payload"`
}

type OpenGateSnapshotPayload struct {
	EventId  string `json:"eventId"`
	RawImage []byte `json:"rawImage"`
}
