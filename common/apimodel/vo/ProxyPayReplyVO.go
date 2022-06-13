package vo

type ProxyPayRespVO struct {
	Code    string                  `json:"code"`
	Message string                  `json:"message"`
	Data    ChannelAppProxyResponse `json:"data"`
	traceId string                  `json:"traceId"`
}

type ProxyQueryRespVO struct {
	Code    string                       `json:"code"`
	Message string                       `json:"message"`
	Data    ChannelAppProxyQeuryResponse `json:"data"`
	traceId string                       `json:"traceId"`
}

//代付下单结果回传商户 DTO物件()
type ProxyPayOrderRespVO struct {
	MerchantId  string `json:"merchantId"`
	OrderNo     string `json:"orderNo"`    //商户订单号
	PayOrderNo  string `json:"payOrderNo"` //这里指我们订单号
	OrderStatus string `json:"orderStatus"`
	Sign        string `json:"sign"`
}

//代付回调返回给渠道Channel App
type ProxyPayCallBackRespVO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
	traceId string `json:"traceId"`
}

//渠道app 代付返回物件
type ChannelAppProxyResponse struct {
	ChannelOrderNo string `json:"channelOrderNo"`
	OrderStatus    string `json:"orderStatus"`
}

type ChannelAppProxyQeuryResponse struct {
	Status           int     `json:"status"`
	ChannelOrderNo   string  `json:"channelOrderNo"`
	OrderStatus      string  `json:"orderStatus"`
	CallBackStatus   string  `json:"callBackStatus"`
	ChannelReplyDate string  `json:"channelReplyDate"`
	ChannelCharge    float64 `json:"channelCharge"`
}

type ProxyPayCallBackMerRespVO struct {
	MerchantId   string `json:"merchantId"`
	OrderNo      string `json:"orderNo"`      /** 商户唯一订单号  */
	PayOrderNo   string `json:"payOrderNo"`   /** 平台订单号  */
	OrderStatus  string `json:"orderStatus"`  /** 订单状态  */
	OrderAmount  string `json:"orderAmount"`  /** 代付金额(不含手续费)  */
	Fee          string `json:"fee"`          /** 手续费  */
	PayOrderTime string `json:"payOrderTime"` /** 交易完成时间  */
	Sign         string `json:"sign"`
}

//from channel app response
