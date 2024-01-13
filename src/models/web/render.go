package web

type VideoPlayerParameters struct {
	Info     *GetStreamInfoResponse `json:"info"`
	Pagename string                `json:"pagename"`
}
