package db

import "time"

type Transcoder struct {
	DeviceId string `json:"deviceId" db:"DEVICE_ID,primary"`
	Name     string `json:"name" db:"NAME"`
}

type OpenGateIntegration struct {
	DeviceId     string `json:"device_id" db:"DEVICE_ID"`
	Available    bool   `json:"available" db:"AVAILABLE"`
	IsRestarting bool   `json:"isRestarting" db:"IS_RESTARTING"`
}

type ObjectTrackingEvent struct {
	EventId    string `json:"eventId" db:"EVENT_ID,primary"`
	OpenGateId string `json:"openGateId" db:"OPENGATE_ID"`
	CameraId   string `json:"cameraId" db:"CAMERA_ID"`

	EventType string `json:"eventType" db:"EVENT_TYPE"`
}

type Camera struct {
	CameraId string  `json:"cameraId" db:"CAMERA_ID,primary"`
	Name     string  `json:"name" db:"NAME"`
	Ip       string  `json:"ip" db:"IP"`
	Port     int     `json:"port" db:"PORT"`
	Username string  `json:"username" db:"USERNAME"`
	Password string  `json:"password" db:"PASSWORD"`
	Started  bool    `json:"started" db:"STARTED"`
	GroupId  *string `json:"groupId" db:"GROUP_ID,omitempty"`

	TranscoderId string `json:"transcoderId" db:"TRANSCODER_ID"`
	OpenGateId   string `json:"openGateId" db:"OPENGATE_ID"`
}

type CameraGroup struct {
	GroupId     string    `json:"groupId" db:"GROUP_ID,primary"`
	Name        string    `json:"name" db:"NAME"`
	CreatedDate time.Time `json:"createdDate" db:"CREATED_DATE"`
}

func (t *Transcoder) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"DEVICE_ID",
		"NAME",
	)
	return fs
}

func (t *Transcoder) Values() []interface{} {
	vs := []interface{}{
		t.DeviceId,
		t.Name,
	}
	return vs
}

func (t *Camera) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"CAMERA_ID",
		"NAME",
		"IP",
		"PORT",
		"USERNAME",
		"PASSWORD",
		"TRANSCODER_ID",
		"STARTED",
	)
	return fs
}

func (t *Camera) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.CameraId,
		t.Name,
		t.Ip,
		t.Port,
		t.Username,
		t.Password,
		t.TranscoderId,
		t.Started,
	)
	return vs
}

func (t *CameraGroup) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"GROUP_ID",
		"NAME",
		"CREATED_DATE",
	)
	return fs
}

func (t *CameraGroup) Values() []interface{} {
	vs := []interface{}{}
	vs = append(vs,
		t.GroupId,
		t.Name,
		t.CreatedDate,
	)
	return vs
}
