package bo

//bo 代付请求channel app 的BO
type ProxyPayBO struct {
	OrderNo              string `json:"orderNo"`
	TransactionType      string `json:"transactionType"`
	TransactionAmount    string `json:"transactionAmount"`
	ReceiptAccountNumber string `json:"receiptAccountNumber"`
	ReceiptAccountName   string `json:"receiptAccountName"`
	ReceiptCardProvince  string `json:"receiptCardProvince"`
	ReceiptCardCity      string `json:"receiptCardCity"`
	ReceiptCardArea      string `json:"receiptCardArea"`
	ReceiptCardBranch    string `json:"receiptCardBranch"`
	ReceiptCardBankCode  string `json:"receiptCardBankCode"`
	ReceiptCardBankName  string `json:"receiptCardBankName"`
}

type ProxyQueryBO struct {
	OrderNo        string `json:"orderNo"`
	ChannelOrderNo string `json:"channelOrderNo"`
}

//商戶請求COPO 的代付請求BO
type ProxyPayOrderRequestBO struct {
	AccessType   string  `json:"accessType" validate:"max=2"`
	MerchantId   string  `json:"merchantId" valiate: "required"`
	Sign         string  `json:"sign" valiate: "required"`
	NotifyUrl    string  `json:"notifyUrl" validate: "required"`
	Language     string  `json:"language" valiate: "required"`
	OrderNo      string  `json:"orderNo" validate: "required"`
	BankId       string  `json:"bankId" validate: "required"`
	BankName     string  `json:"bankName" validate: "required"`
	BankProvince string  `json:"bankProvince" validate: "required"`
	BankCity     string  `json:"bankCity" validate: "required"`
	BranchName   string  `json:"branchName, optional"`
	BankNo       string  `json:"bankNo" valiate: "required"`
	OrderAmount  float64 `json:"orderAmount" validate: "required"` //到小數兩位
	DefrayName   string  `json:"defrayName" validate: "required"`
	DefrayId     string  `json:"defrayId, optional"`
	DefrayMobile string  `json:"defrayMobile, optional"`
	DefrayEmail  string  `json:"defrayEmail, optional"`
	Currency     string  `json:"currency" validate: "required"`
	PayTypeSubNo string  `json:"payTypeSubNo,optional"`
}

type ProxyPayOrderRequestBOX struct {
	ProxyPayOrderRequestBO
	Ip string `json:"ip, optional"`
}

type ProxyPayOrderQueryRequestBO struct {
	AccessType string `json:"accessType" valiate: "required"`
	MerchantId string `json:"merchantId" valiate: "required"`
	OrderNo    string `json:"orderNo" valiate: "required"`
	Sign       string `json:"sign" valiate: "required"`
	Language   string `json:"language" valiate: "required"`
}

type ProxyPayOrderQueryResponseVO struct {
	RespCode       string `json:"respCode"`
	RespMsg        string `json:"respMsg"`
	MerchantId     string `json:"merchantId"`
	OrderNo        string `json:"orderNo"`
	LastUpdateTime string `json:"lastUpdateTime"`
	PayOrderNo     string `json:"payOrderNo"`
	OrderStatus    string `json:"orderStatus"`
	CallbackStatus string `json:"callbackStatus"`
	OrderAmount    string `json:"orderAmount"`
	Fee            string `json:"fee"`
	Sign           string `json:"sign"`
}
