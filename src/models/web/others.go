package web

type VideoPlayerParameters struct {
	Info     *GetStreamInfoResponse `json:"info"`
	Pagename string                 `json:"pagename"`
}

type TranscoderStreamConfiguration struct {
	CameraId   string `json:"cameraId"`
	SourceUrl  string `json:"sourceUrl"`
	PublishUrl string `json:"publishUrl"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	Fps        int    `json:"fps"`
}
