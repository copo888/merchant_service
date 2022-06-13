package vo

type ChannelRespBodyVO struct {
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Data    PayReplyVO `json:"data,omitempty"`
	Trace   string     `json:"trace"`
}
