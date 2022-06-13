package response

var (
	SUCCESS = "0"     //"操作成功"
	FAIL    = "EX001" //"Fail"

	//EX000 = response{Code: "EX000" //"不合法請求"}
	//EX001 = response{Code: "EX001" //"参数不合法"}
	//EX002 = response{Code: "EX002", Desc: "登入失败"}
	//EX003 = response{Code: "EX003", Desc: "账号不存在或密码错误"}
	//EX004 = response{Code: "EX004", Desc: "账号已存在"}
	//
	//ME001 = response{Code: "ME001", Desc: "商戶餘額不足"}
	//ME002 = response{Code: "ME002", Desc: "交易单号重复"}
	//ME003 = response{Code: "ME003", Desc: "交易单号不存在"}
	//
	//CH001 = response{Code: "CH001", Desc: "渠道不存在"}
	//CH002 = response{Code: "CH002", Desc: "请求错误 %v"}
	//
	//S001 = response{Code: "S001", Desc: "系统出现异常"}
	//S002 = response{Code: "S002", Desc: "数据库连接失败"}
	//S003 = response{Code: "S003", Desc: "请求过于频繁"}
	//S004 = response{Code: "S004", Desc: "IP限制 %s"}
	//
	//M001 = response{Code: "M001", Desc: "系统维护中"}
)
