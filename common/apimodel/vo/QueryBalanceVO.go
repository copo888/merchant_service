package vo

type ProxyQueryBalanceRespVO struct {
	Code    string                       `json:"code"`
	Message string                       `json:"message"`
	Data    QueryInternalBalanceResponse `json:"data"`
	traceId string                       `json:"traceId"`
}

type QueryInternalBalanceResponse struct {
	ChannelNametring   string `json:"channelNametring,optional"`
	ChannelCodingtring string `json:"channelCodingtring,optional"`
	WithdrawBalance    string `json:"withdrawBalance,optional"`
	ProxyPayBalance    string `json:"proxyPayBalance,optional"`
	UpdateTimetring    string `json:"updateTimetring,optional"`
	ErrorCodetring     string `json:"errorCodetring,optional"`
	ErrorMsgtring      string `json:"errorMsgtring,optional"`
}
