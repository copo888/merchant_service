package ordersService

import (
	"com.copo/bo_service/common/apimodel/bo"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants/redisKey"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/service/merchantchannelrateservice"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

func VerifyPayOrder(db *gorm.DB, req types.PayOrderRequestX, merchant *types.Merchant) (err error) {

	// 開啟多選商戶 必需給指定代碼
	if merchant.BillLadingType == "1" && len(req.PayTypeNo) == 0 {
		return errorz.New(response.NO_CHANNEL_SET, "")
	}

	// 支付提單API 不可給DF代碼
	if strings.EqualFold(req.PayType, "DF") {
		return errorz.New(response.PAY_TYPE_INVALID, "")
	}

	// USDT 限制PayType
	if strings.EqualFold(req.Currency, "USDT") && !utils.Contain(req.PayType, []string{"UT", "UE", "UU"}) {
		return errorz.New(response.INVALID_USDT_TYPE, fmt.Sprintf("(payType): %s", req.PayType))
	}

	// 確認商戶訂單號重複
	if isExist, err := model.NewOrder(db).IsExistByMerchantOrderNo(merchant.Code, req.OrderNo); isExist {
		return errorz.New(response.ORDER_NUMBER_EXIST, "")
	} else if err != nil {
		return errorz.New(response.SYSTEM_ERROR, err.Error())
	}

	return nil
}

func CallChannelForPay(db *gorm.DB, req types.PayOrderRequestX, merchant *types.Merchant, orderNo string, ctx context.Context, svcCtx *svc.ServiceContext) (payReplyVO *vo.PayReplyVO, correspondMerChnRate *types.CorrespondMerChnRate, err error) {
    orderAmount := ""
	if s, ok := req.OrderAmount.(string); ok {
		orderAmount = s
	} else if f, ok := req.OrderAmount.(float64); ok {
		orderAmount = fmt.Sprintf("%f", f)
	} else {
		s := fmt.Sprintf("OrderAmount err: %#v", req.OrderAmount)
		logx.Errorf(s)
		return nil, nil, errorz.New(response.API_INVALID_PARAMETER, s)
	}

	// 取得支付渠道資訊
	if correspondMerChnRate, err = merchantchannelrateservice.GetDesignationMerChnRate(db, req.MerchantId, req.PayType, req.Currency, req.PayTypeNo, merchant.BillLadingType); err != nil {
		return
	}

	if correspondMerChnRate.Fee < correspondMerChnRate.ChFee ||
		correspondMerChnRate.HandlingFee < correspondMerChnRate.ChHandlingFee {
		return nil, nil, errorz.New(response.RATE_SETTING_ERROR)
	}

	// 確認支付金額上下限
	var amount float64
	if amount, err = strconv.ParseFloat(orderAmount, 64); err != nil {
		return nil, nil, errorz.New(response.INVALID_AMOUNT, fmt.Sprintf("(orderAmount): %s", req.OrderAmount))
	}
	if amount < 0 {
		return nil, nil, errorz.New(response.ORDER_AMOUNT_INVALID, fmt.Sprintf("(orderAmount): %f", req.OrderAmount))
	}
	if amount > correspondMerChnRate.SingleMaxCharge {
		return nil, nil, errorz.New(response.ORDER_AMOUNT_LIMIT_MAX, fmt.Sprintf("(orderAmount): %f", req.OrderAmount))
	}
	if amount < correspondMerChnRate.SingleMinCharge {
		return nil, nil, errorz.New(response.ORDER_AMOUNT_LIMIT_MIN, fmt.Sprintf("(orderAmount): %f", req.OrderAmount))
	}

	// 組成請求json
	payBO := bo.PayBO{
		OrderNo:           orderNo,
		PayType:           correspondMerChnRate.PayTypeCode,
		ChannelPayType:    correspondMerChnRate.MapCode,
		TransactionAmount: orderAmount,
		BankCode:          req.BankCode,
		PageUrl:           req.PageUrl,
		OrderName:         req.OrderName,
		MerchantId:        req.MerchantId,
		Currency:          req.Currency,
		SourceIp:          req.UserIp,
		UserId:            req.UserId,
		JumpType:          req.JumpType,
	}

	// call 渠道app
	span := trace.SpanFromContext(ctx)
	payKey, errk := utils.MicroServiceEncrypt(svcCtx.Config.ApiKey.PayKey, svcCtx.Config.ApiKey.PublicKey)
	if errk != nil {
		return nil, nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	url := fmt.Sprintf("%s:%s/api/pay", svcCtx.Config.Server, correspondMerChnRate.ChannelPort)
	res, errx := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationPaykey", payKey).JSON(payBO)
	if errx != nil {
		return nil, nil, errorz.New(response.GENERAL_EXCEPTION, errx.Error())
	} else if res.Status() != 200 {
		return nil, nil, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("call channelApp httpStatus:%d", res.Status()))
	}

	// 處理res
	channelRespBodyVO := vo.PayReplBodyVO{}
	if err = res.DecodeJSON(&channelRespBodyVO); err != nil {
		return nil, nil, errorz.New(response.CHANNEL_REPLY_ERROR, err.Error())
	}
	if channelRespBodyVO.Code != "0" {
		return nil, nil, errorz.New(channelRespBodyVO.Code, channelRespBodyVO.Message)
	}
	payReplyVO = &channelRespBodyVO.Data

	return
}

func GetPayOrderResponse(req types.PayOrderRequestX, replyVO vo.PayReplyVO, orderNo string, ctx context.Context, svcCtx *svc.ServiceContext) (resp *types.PayOrderResponse, err error) {

	resp = &types.PayOrderResponse{}
	// 預設url
	info := replyVO.PayPageInfo

	// PayPageType 非url 非json 就跑 html
	if !strings.EqualFold(replyVO.PayPageType, "url") && !strings.EqualFold(replyVO.PayPageType, "json") {
		if !strings.EqualFold(replyVO.PayPageType, "html") {
			logx.Error(fmt.Sprintf("Channel Reply Type:%s error", replyVO.PayPageType))
		}
		// TODO: 實作包HTML功能
	}
	if strings.EqualFold(replyVO.PayPageType, "json") && replyVO.IsCheckOutMer {
		// 存渠道銀行卡訊息至 Redis 6分鐘
		if err = svcCtx.RedisClient.Set(ctx, redisKey.CACHE_PAY_ORDER_CHANNEL_BANK+orderNo, replyVO.PayPageInfo, 6*time.Minute).Err(); err != nil {
			return nil, errorz.New(response.GENERAL_EXCEPTION)
		}
		// 返回自組收銀台 URL
		info = fmt.Sprintf("%s/#/checkoutMer?id=%s", svcCtx.Config.FrontEndDomain, orderNo)
	}
	resp.BankCode = req.BankCode
	resp.Info = info
	resp.PayOrderNo = orderNo
	resp.Type = replyVO.PayPageType

	return
}
