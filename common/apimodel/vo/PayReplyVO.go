package vo

type PayReplyVO struct {
	PayPageInfo    string `json:"payPageInfo, optional"`
	PayPageType    string `json:"payPageType, optional"`
	ChannelOrderNo string `json:"channelOrderNo, optional"`
	OrderAmount    string `json:"orderAmount, optional"`
	RealAmount     string `json:"realAmount, optional"`
	Status         string `json:"status, optional"`
	IsCheckOutMer  bool   `json:"isCheckOutMer, optional"`
}

type PayReplBodyVO struct {
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Data    PayReplyVO `json:"data, omitempty"`
	Trace   string     `json:"trace"`
}
