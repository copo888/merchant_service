package response

var (
	WITHDRAW_ORDER              = "WITHDRAW_ORDER"              //"下发单"
	INTERNAL_CHARGE_ORDER       = "INTERNAL_CHARGE_ORDER"       //"内充单"
	REGISTRATION_TEMPORARY_DATA = "REGISTRATION_TEMPORARY_DATA" //"暂存注册资料"
	PROXY_PERSON_PROCESS        = "PERSON_PROCESS"              //"代付人工还款"
	COMMISSION_WITHDRAW_ORDER   = "COMMISSION_WITHDRAW_ORDER"   //"佣金提单"

	//错误讯息
	//INVALID_PARAMETER       =  "1100001" //"无效参数"
	NEGATIVE_NUMBER         = "1100002" //"负数计数错误"
	REDIS_INVALID_PARAMETER = "1100003" //"redis无效参数"

	//EEEE: 9000 ~ 9999,Desc: 网络层级错误
	SIGN_KEY_FAIL = "1109001" //"加签错误，请确认加签规则"
)
