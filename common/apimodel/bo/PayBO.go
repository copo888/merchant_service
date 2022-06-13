package bo

type PayBO struct {
	OrderNo           string `json:"orderNo"`
	PayType           string `json:"payType"`
	ChannelPayType    string `json:"channelPayType"`
	TransactionAmount string `json:"transactionAmount"`
	BankCode          string `json:"bankCode"`
	PageUrl           string `json:"pageUrl"`
	OrderName         string `json:"orderName"`
	MerchantId        string `json:"merchantId"`
	Currency          string `json:"currency"`
	SourceIp          string `json:"sourceIp"`
	UserId            string `json:"userId"`
	JumpType          string `json:"jumpType"`
}
