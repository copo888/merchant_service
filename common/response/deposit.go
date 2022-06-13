package response

var (
	/**
	 * 通用系统讯息码
	 */
	//SUCCESS="1040001" //"操作成功"

	/**
	 * 前端操作讯息码
	 */
	//SERVICE_ERROR="1042001" //"抱歉，您的操作出了问题，请联系客服，谢谢"
	//BUSINESS_ERROR="1042002" //"服务异常，请联系客服，谢谢"
	//BOSS_TOKEN_TIMEOUT="1042003" //"登录逾时，请重新登入"
	//APP_TOKEN_TIMEOUT="1042004" //"登入失败，请重新登入"
	//INVALID_PARAMETER="1042005" //"无效的参数"
	//PARAMETER_TYPE_ERROE="1042006" //"参数类型错误"
	//INSERT_FAIL="1042007" //"新增资料失败"
	REVIEW_REASON_ERROR = "1042008" //"请输入审核不通过原因"
	//INSERT_MER_ID_REPEAT="1042009" //"新增资料失败:商户号已存在"
	//DATE_RANGE_ERROR="1042010" //"无效的时间区间"

	/**
	 * 系统类型讯息码
	 */
	//ILLEGAL_REQUEST="1042101" //"非法请求"
	//ILLEGAL_PARAMETER="1042102" //"非法参数"
	//UN_ROLE_AUTHORIZE="1042103" //"无此应用服务使用授权"
	//DATA_NOT_FOUND="1042104" //"资料无法取得"
	//GENERAL_ERROR="1042105" //"通用错误"
	//CONNECT_SERVICE_FAILURE="1042106" //"服务连线失败"
	//UPDATE_DATABASE_FAILURE="1042107" //"数据更新失败"
	//UPDATE_DATABASE_REPEAT="1042108" //"数据重复更新"
	//INSERT_DATABASE_FAILURE="1042109" //"数据新增失败"
	//DELETE_DATABASE_FAILURE="1042110" //"数据删除失败"
	//DATABASE_FAILURE="1042111" //"数据库错误"

	CHANNEL_DEFEND = "1042112" //"渠道维护中"
	//RATE_NOT_CONFIGURED =  "1042113" //"未配置商户渠道费率"
	ALREADY_CALL_BACK = "1042114" //"此订单已回调完成，不需重复回调"
	CHARGE_AMT_EXCEED = "1042115" //"单笔充值金额超过最大限制"
	NO_DATA_TO_UPDATE = "1042116" //"无数据可更新或提单已是目前状态"
)
