package transcoder

const (
	OPENGATE_EVENTS                 = "events"
	OPENGATE_STATS                  = "stats"
	OPENGATE_AVAILABLE              = "available"
	OPENGATE_SNAPSHOT               = "snapshot"
	OPENGATE_STATE_AUDIO            = "audio/state"
	OPENGATE_STATE_SNAPSHOTS        = "snapshots/state"
	OPENGATE_STATE_DETECT           = "detect/state"
	OPENGATE_STATE_RECORDINGS       = "recordings/state"
	OPENGATE_STATE_MOTION           = "motion/state"
	OPENGATE_STATE_PTZ_AUTOTRACKER  = "ptz_autotracker/state"
	OPENGATE_STATE_IMPROVE_CONTRAST = "improve_contrast/state"
	OPENGATE_STATE_BIRDSEYE         = "birdseye/state"
)

type TranscoderEventMessage struct {
	Type         string  `json:"type"`
	TranscoderId string  `json:"transcoderId"`
	CameraName   *string `json:"cameraName,omitempty"`

	Payload []byte `json:"payload"`
}

type OpenGateSnapshotPayload struct {
	EventId  string `json:"eventId"`
	RawImage []byte `json:"rawImage"`
}
