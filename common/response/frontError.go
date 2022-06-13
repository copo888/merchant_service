package response

var (
	/**
	 * 前端操作讯息码
	 */
	SERVICE_ERROR                                                = "1062001" //"抱歉，您的操作出了问题，请联系客服，谢谢"
	BUSINESS_ERROR                                               = "1062002" //"服务异常，请联系客服，谢谢"
	BOSS_TOKEN_TIMEOUT                                           = "1062003" //"登录逾时，请重新登入"
	APP_TOKEN_TIMEOUT                                            = "1062004" //"登入失败，请重新登入"
	INVALID_PARAMETER                                            = "1062005" //"无效的参数"
	PARAMETER_TYPE_ERROE                                         = "1062006" //"参数类型错误"
	INSERT_FAIL                                                  = "1062007" //"新增资料失败"
	NO_MER_RATE_INFO                                             = "1062008" //"查无商户费率信息"
	DATA_SETTING_ERROR                                           = "1062009" //"资料设定错误"
	NO_CONFIGURABLE_CHANNELS                                     = "1062010" //"无可配置的渠道"
	CONFIGURATION_BANK_DATA_DUPLICATION                          = "1062011" //"配置银行资料重复"
	CONFIGURATION_PAY_METHOD_DUPLICATION                         = "1062012" //"渠道支付方式资料重复"
	CONFIGURATION_CHANNEL_DATA_DUPLICATION                       = "1062013" //"渠道资料重复"
	DATE_RANGE_ERROR                                             = "1062014" //"无效的时间区间"
	ONLY_NUMBERIC                                                = "1062015" //"仅能为数字"
	MAXIMUN_AMOUNT_OF_CHARGE_NULL_ERROR                          = "1062016" //"内充金额最大值，不得为空值"
	CHANNEL_RATE_OVER_MERCHANT_RATE_ERROR                        = "1062017" //"渠道费率不得高于已配置的商户费率"
	CHANNEL_CHARGE_OVER_MERCHANT_CHARGE_ERROR                    = "1062018" //"渠道手续费不得高于已配置的商户手续费"
	CHANNEL_WITHDRAW_CHARGE_OVER_MERCHANT_WITHDRAW_CHARGE_ERROR  = "1062019" //"渠道下发手续费不得高于已配置的商户手续费"
	REQUEST_FORMAT_ERROR                                         = "1062020" //"无效传递格式"
	MERCHANT_RATE_NO_CHANGE_ERROR                                = "1062021" //"商户费率资讯没有异动"
	CHANNEL_RATE_MIN_CHARGE_OVER_MERCHANT_RATE_MIN_CHARGE_ERROR  = "1062022" //"渠道费率低消不得高于已配置的商户费率低消"
	CHANNEL_RATE_MIN_CHARGE_NULL_TO_NUM_ERROR                    = "1062023" //"渠道费率低消不得从空值改为数字"
	CHANNEL_RATE_MIN_CHARGE_NUM_TO_NULL_ERROR                    = "1062024" //"渠道费率低消不得从数字改为空值"
	CHANNEL_PAY_CHARGE_OVER_MERCHANT_MIN_PAY_CHARGE_ERROR        = "1062025" //"渠道支付手续费不得高于已配置的商户支付手续费"
	REQUIRE_CURRENCY_NOT_SET                                     = "1062026" //"人民币等币别不可禁止"
	MERCHANT_WITHDRAW_CHARGE_LOWER_CHANNEL_WITHDRAW_CHARGE_ERROR = "1062027" //"商户下发手续费不得低于已配置的渠道下发手续费"
	ILLEGAL_IP                                                   = "1062028" //"IP格式错误"
	INSERT_IP_LIST_DUPLICATE                                     = "1062029" //"新增IP清单重复"
	CURRENCCODING_NULL_ERROR                                     = "1062030" //"搜寻条件币别不可为空"
	BANKNO_ALREADY_EXIST_ERROR                                   = "1062031" //"银行代码已存在"
)
