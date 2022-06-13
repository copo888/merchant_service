package ordersService

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/types"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

func GetMerchantChannelRate(db *gorm.DB, merchantCode string, currencyCode string, orderType string) (resp *types.MerchantOrderRateListViewX, err error) {
	//var merchantOrderRateListViews []*types.MerchantOrderRateListViewX
	var terms []string
	terms = append(terms, fmt.Sprintf("merchant_code = '%s'", merchantCode))
	terms = append(terms, "merchnrate_status = '1'")           // 0:禁用 1:啟用
	terms = append(terms, fmt.Sprintf("pay_type_code = 'DF'")) // 內充看代付
	terms = append(terms, "designation = '1'")
	terms = append(terms, "chn_status = '1'")
	terms = append(terms, "chnpaytype_status = '1'")
	terms = append(terms, fmt.Sprintf("currency_code = '%s'", currencyCode))
	if orderType == "NC" {
		terms = append(terms, fmt.Sprintf("chn_is_proxy = '0'")) //支援支轉代(0:不支援 1:支援)
	}

	term := strings.Join(terms, " AND ")

	// 查询商户拥有的代付渠道
	// TODO 需要判斷商戶可用幣別的代付餘額下限值??(V1在系统常量)

	if err = db.Table("merchant_order_rate_list_view").Where(term).Order("designation_no").Limit(1).Take(&resp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED)
		}
		return nil, errorz.New(response.DATABASE_FAILURE, "数据库错误: "+err.Error())
	}

	return resp, nil
}
