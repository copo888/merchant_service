package merchantsService

import (
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"net/url"
	"strconv"
)

/*
	回調-商戶代付結果(須注意Scheduled Project 也有一組再用，修改時要注意)
	@return
*/
func PostCallbackToMerchant(db *gorm.DB, context *context.Context, orderX *types.OrderX) (err error) {
	span := trace.SpanFromContext(*context)
	merchant := &types.Merchant{}
	if err = db.Table("mc_merchants").Where("code = ?", orderX.MerchantCode).Find(merchant).Error; err != nil {
		return
	}

	ProxyPayCallBackMerRespVO := url.Values{}
	ProxyPayCallBackMerRespVO.Set("MerchantId", orderX.MerchantCode)
	ProxyPayCallBackMerRespVO.Set("OrderNo", orderX.MerchantOrderNo)
	ProxyPayCallBackMerRespVO.Set("PayOrderNo", orderX.OrderNo)
	ProxyPayCallBackMerRespVO.Set("OrderStatus", orderX.Status)
	ProxyPayCallBackMerRespVO.Set("OrderAmount", strconv.FormatFloat(orderX.OrderAmount, 'f', 2, 64))
	ProxyPayCallBackMerRespVO.Set("Fee", strconv.FormatFloat(orderX.Fee, 'f', 2, 64))
	ProxyPayCallBackMerRespVO.Set("PayOrderTime", orderX.TransAt.Time().Format("200601021504"))
	//sign := utils.SortAndSign2(ProxyPayCallBackMerRespVO,merchant.ScrectKey)
	ProxyPayCallBackMerRespVO.Set("Sign", "djiocpnvpqnpcnvpqn")
	logx.Infof("代付提单 %s ，回调商户URL= %s，回调资讯= %#v", orderX.OrderNo, orderX.NotifyUrl, ProxyPayCallBackMerRespVO)

	//TODO retry post for 10 times and 2s between each reqeuest
	//TODO 內部測試，測完需移除
	merResp, merCallBackErr := gozzle.Post("http://172.16.204.115:8083/dior/merchant-api/merchant-call-back").Timeout(10).Trace(span).Form(ProxyPayCallBackMerRespVO)
	//merResp, merCallBackErr := gozzle.Post(orderX.NotifyUrl).Timeout(10).Trace(span).Form(ProxyPayCallBackMerRespVO)
	if merCallBackErr != nil || merResp.Status() != 200 {
		if merCallBackErr != nil {
			logx.Errorf("代付提单%s 回调商户异常，錯誤: %#v", ProxyPayCallBackMerRespVO.Get("OrderNo"), merCallBackErr)
		} else if merResp.Status() != 200 {
			logx.Errorf("响应状态 %d 错误", merResp.Status())
		}
	}
	logx.Infof("代付提单 %s ，回调商户請求參數 %#v，商戶返回: %#v", ProxyPayCallBackMerRespVO.Get("OrderNo"), ProxyPayCallBackMerRespVO, string(merResp.Body()))
	return
}
