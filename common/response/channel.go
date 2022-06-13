package response

var (
	/**
	 * YK渠道
	 */
	CREDIT_LESSTHAN_ORDER   = "1063001" // "卡片额度不能小於单笔限额"
	CARD_ENABLE_NOT_EDIT    = "1063002" // "已交易卡片未停用不能配置"
	CREDIT_LIMIT_EDIT_ERROR = "1063003" // "下修金额小额目前使用额度，不可修改"
	CREDIT_LESS_THAN_ORDER  = "1063004" // "卡片额度不能小於单笔限额"
	CREDIT_LIMIT_NOT_ENABLE = "1063005" // "该卡已达使用笔数/使用额度，不可启用"
	YK_ORDER_USER_BLACKLIST = "1063006" // "系统侦测恶意拉单，请会员联系客服谘询"
	BLACKLIST_NEW_DUPLICATE = "1063007" // "新增黑名单使用者已重覆"
	BANKCARD_USER_NOT_FOUND = "1063008" // "非本系统会员，请查明再进行设定"
	BANKCARD_NEW_DUPLICATE  = "1063009" // "卡号重复，请重新建置"
	PHONE_BIND_DUPLICATE    = "1063010" // "该门号已绑定其他银行卡，请重新输入"
	PHONE_BIND_EXIST        = "1063011" // "银行卡已绑定，请先解除再绑定新门号"
	PHONE_FORMAT_INVALID    = "1063012" // "请输入完整号码"

	//渠道相關
	PAY_TYPE_DUPLICATED        = "1063030" //  "支付類型重複"
	UPDATE_PAYTYPE_NUM_ERROR   = "1063031" //  "支付類型編碼格式錯誤"
	PAY_TYPE_NOT_EXIST         = "1063032" //  "支付類型不存在"
	CHANNEL_PAYTYPE_DUPLICATED = "1063033" //  "渠道重複配置支付類型"
	WHITE_LIST_DUPLICATED      = "1063034" //  "白名單重複"
	INVALID_WHITE_LIST         = "1063035" //  "白名單格式錯誤"
	INVALID_PAY_TYPE_MAP       = "1063036" //  "支付类型通道编码格式错误"
	CHANNEL_IS_NOT_EXIST       = "1063037" //  "渠道不存在"

	SETTING_CHANNEL_RATE_CHARGE_ERROR = "1063038" //配置渠道成本费率不可高于商戶配置費率
	SETTING_CHANNEL_FEE_CHARGE_ERROR  = "1063039" //配置渠道成本手續費不可高于商戶配置手續費

	SETTING_CHANNEL_BALANCE_NULL = "1063040" // 請擇一渠道更新餘額

)
