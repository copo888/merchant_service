package excelizeutil

import (
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/xuri/excelize/v2"
	"unicode/utf8"
)

// SetColWidthAuto 依据字数自动设定栏宽
func SetColWidthAuto(xlsx *excelize.File, sheetName string) error {
	cols, err := xlsx.GetCols(sheetName)
	var baseWidth float64 = 4 //基础加的宽度
	if err != nil {
		return err
	}
	for idx, col := range cols {
		largestWidth := 0
		for _, rowCell := range col {
			cellWidth := utf8.RuneCountInString(rowCell) + 2 // + 2 for margin
			if cellWidth > largestWidth {
				largestWidth = cellWidth
			}
		}
		name, err := excelize.ColumnNumberToName(idx + 1)
		if err != nil {
			return err
		}
		xlsx.SetColWidth(sheetName, name, name, float64(largestWidth)+baseWidth)
	}
	return nil
}

// GetTxOrderStatusName tx_order表 status栏位
func GetTxOrderStatusName(orderStatus string) string {
	// 訂單狀態(0:待處理 1:處理中 20:成功 30:失敗 31:凍結)
	name := ""
	switch orderStatus {
	case "0":
		name = i18n.Sprintf("Pending")
	case "1":
		name = i18n.Sprintf("Processing")
	case "2":
		name = i18n.Sprintf("In transaction")
	case "20":
		name = i18n.Sprintf("Success")
	case "30":
		name = i18n.Sprintf("Fail")
	case "31":
		name = i18n.Sprintf("Freeze")
	default:
		name = orderStatus
	}
	return name
}

// GetTxOrderTypeName tx_order表 type栏位
func GetTxOrderTypeName(orderType string) string {
	return i18n.Sprintf(orderType)
}

// GetTxMerchantCallbackName tx_order表 isMerchantCallback栏位
func GetTxMerchantCallbackName(merchantCallback string) string {
	// 是否已经回调商户(0：否、1:是、2:不需回调)(透过API需提供的资讯)
	name := ""
	switch merchantCallback {
	case "0":
		name = i18n.Sprintf("No")
	case "1":
		name = i18n.Sprintf("Yes")
	case "2":
		name = i18n.Sprintf("No need to callback")
	default:
		name = merchantCallback
	}
	return name
}

func GetTxOrderReasonType(reasonType string) string {
	// 原因类型(1=(补单)修改金额/2=(补单)重复支付/3=(补单)其它/11=追回)
	name := ""
	switch reasonType {
	case "1":
		name = i18n.Sprintf("Amendment amount")
	case "2":
		name = i18n.Sprintf("Repeat payment")
	case "3":
		name = i18n.Sprintf("Recover")
	case "11":
		name = i18n.Sprintf("Other")
	default:
		name = reasonType
	}
	return name
}

func GetTxOrderSourceName(source string) string {
	// 订单来源(1:平台 2:API)
	name := ""
	switch source {
	case "1":
		name = i18n.Sprintf("Platform")
	case "2":
		name = "API"
	default:
		name = source
	}
	return name
}

// GetBalanceRecordTransactionTypeName mc_merchant_balance_records表 transaction_type栏位
func GetBalanceRecordTransactionTypeName(transactionType string) string {
	//(1=收款 ; 2=解冻;  3=冲正 4=还款;  5=补单; 11=出款 ; 12=冻结 ; 13=追回; 20=调整)
	name := ""

	switch transactionType {
	case "1":
		name = i18n.Sprintf("Reload")
	case "2":
		name = i18n.Sprintf("Unfreeze")
	case "3":
		name = i18n.Sprintf("Offset")
	case "4":
		name = i18n.Sprintf("Repayment")
	case "5":
		name = i18n.Sprintf("Make up")
	case "6":
		name = i18n.Sprintf("Payout")
	case "11":
		name = i18n.Sprintf("Payment")
	case "12":
		name = i18n.Sprintf("Freeze")
	case "13":
		name = i18n.Sprintf("Trace back")
	case "14":
		name = i18n.Sprintf("Deduction")
	case "15":
		name = i18n.Sprintf("Withdraw")
	case "20":
		name = i18n.Sprintf("Adjustment")

	default:
		name = transactionType
	}
	return name
}

// GetBalanceType balance_type栏位
func GetBalanceType(balanceType string) string {
	//余额类型 (DFB=代付余额 XFB=下发余额 YJB=佣金余额)
	name := ""
	switch balanceType {
	case "DFB":
		name = i18n.Sprintf("Payout balance")
	case "XFB":
		name = i18n.Sprintf("Withdraw balance")
	case "YJB":
		name = i18n.Sprintf("Commission balance")

	default:
		name = balanceType
	}
	return name
}
