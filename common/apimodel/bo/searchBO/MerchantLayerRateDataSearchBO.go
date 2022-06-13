package searchBO

type MerchantLayerRateDataSearchBO struct {
	MerchantCode string `json:"merchant_code"`
	AgentLayerNo string `json:"agent_layer_no"`
	PayTypeCode  string `json:"pay_type_code"`
}
