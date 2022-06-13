package vo

type WithdrawCallBackVO struct {
	MerchantId  string `json:"merchantId"`
	OrderNo     string `json:"orderNo"`
	OrderAmount string `json:"orderAmount"`
	OrderTime   string `json:"orderTime"`
	ReviewTime  string `json:"reviewTime"`
	Fee         string `json:"fee"`
	OrderStatus string `json:"orderStatus"`
	DiorOrderNo string `json:"diorOrderNo"`
	Sign        string `json:"sign"`
}
