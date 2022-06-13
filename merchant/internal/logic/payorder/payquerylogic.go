package payorder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type PayQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayQueryLogic {
	return PayQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayQueryLogic) PayQuery(req types.PayQueryRequestX) (resp *types.PayQueryResponse, err error) {
	var merchant *types.Merchant
	var order *types.OrderX

	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req.MerchantId).
		Where("status = ?", "1").
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
		}
	}

	// 檢查白名單
	if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.MyIp)
	}

	// 檢查驗簽
	if isSameSign := utils.VerifySign(req.Sign, req.PayQueryRequest, merchant.ScrectKey); !isSameSign {
		return nil, errorz.New(response.SIGN_KEY_FAIL)
	}

	if err = l.svcCtx.MyDB.Table("tx_orders").
		Where("merchant_code = ?", req.MerchantId).
		Where("merchant_order_no = ?", req.OrderNo).Take(&order).Error; err != nil {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, err.Error())
	}


	i18n.SetLang(language.English)
	resp = &types.PayQueryResponse{
		RespCode:    response.API_SUCCESS,
		RespMsg:     i18n.Sprintf(response.API_SUCCESS),
		Language:    "zh-CN",
		BankCode:    order.MerchantBankNo,
		Fee:         fmt.Sprintf("%.2f", order.TransferHandlingFee),
		MerchantId:  order.MerchantCode,
		OrderNo:     order.MerchantOrderNo,
		OrderTime:   order.CreatedAt.Format("200601021504"),
		PayOrderId:  order.OrderNo,
	}

	// 若有實際金額則回覆實際
	if order.ActualAmount > 0 {
		resp.OrderAmount = fmt.Sprintf("%.2f", order.ActualAmount)
	} else {
		resp.OrderAmount = fmt.Sprintf("%.2f", order.OrderAmount)
	}

	// API 支付状态 0：处理中，1：成功，2：失败，3：成功(人工确认)
	resp.OrderStatus = "0"
	if order.Status == constants.SUCCESS {
		resp.OrderStatus = "1"
	} else if order.Status == constants.FAIL {
		resp.OrderStatus = "2"
	}

	if !order.TransAt.Time().IsZero() {
		resp.PayOrderTime = order.TransAt.Time().Format("200601021504")
	}

	resp.Sign = utils.SortAndSign2(*resp, merchant.ScrectKey)

	return
}
