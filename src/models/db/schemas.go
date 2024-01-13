package db

type Transcoder struct {
	DeviceId string `json:"deviceId" db:"device_id,primary"`
	Name     string `json:"name" db:"name"`
}

type Camera struct {
	CameraId string `json:"cameraId" db:"camera_id,primary"`
	Name     string `json:"name" db:"name"`
	Ip       string `json:"ip" db:"ip"`
	Port     int    `json:"port" db:"port"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Started  bool   `json:"started" db:"started"`

	TranscoderId string `json:"transcoderId" db:"transcoder_id"`
}

func (t *Transcoder) Fields() []string {
	fs := []string{}
	fs = append(fs,
		"device_id",
		"name",
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
		"camera_id",
		"name",
		"ip",
		"port",
		"username",
		"password",
		"transcoder_id",
		"started",
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
