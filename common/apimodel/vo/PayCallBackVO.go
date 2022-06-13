package vo

type PayCallBackVO struct {
	AccessType   string `json:"accessType"`
	Language     string `json:"language"`
	MerchantId   string `json:"merchantId"`
	OrderNo      string `json:"orderNo"`
	OrderAmount  string `json:"orderAmount"`
	OrderTime    string `json:"orderTime"`
	PayOrderTime string `json:"payOrderTime"`
	Fee          string `json:"fee"`
	OrderStatus  string `json:"orderStatus"`
	PayOrderId   string `json:"payOrderId"`
	Sign         string `json:"sign"`
}
