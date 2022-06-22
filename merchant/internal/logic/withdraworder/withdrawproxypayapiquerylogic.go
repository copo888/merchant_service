package withdraworder

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"context"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawProxyPayApiQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawProxyPayApiQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawProxyPayApiQueryLogic {
	return WithdrawProxyPayApiQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawProxyPayApiQueryLogic) WithdrawProxyPayApiQuery(req *types.ProxyPayOrderQueryRequestX) (resp *types.WithdrawApiQueryResponse, err error) {
	logx.Info("Enter withdraw-query-proxy:", req)
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
	if isWhite := merchantsService.IPChecker(req.Ip, merchant.ApiIP); !isWhite {
		logx.Error("白名單檢查錯誤: IP:", req.Ip)
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.Ip)
	}

	// 檢查簽名
	checkSign := utils.VerifySign(req.Sign, req.ProxyPayOrderQueryRequest, merchant.ScrectKey)
	if !checkSign {
		logx.Error("驗簽檢查錯誤: sign:", req.Sign)
		return nil, errorz.New(response.INVALID_SIGN)
	}

	if err = l.svcCtx.MyDB.Table("tx_orders").
		Where("merchant_code = ?", req.MerchantId).
		Where("merchant_order_no = ?", req.OrderNo).Take(&order).Error; err != nil {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, err.Error())
	}

	i18n.SetLang(language.English)
	resp = &types.WithdrawApiQueryResponse{
		RespCode:    response.API_SUCCESS,
		RespMsg:     i18n.Sprintf(response.API_SUCCESS),
		OrderStatus: "0",
		MerchantId:  order.MerchantCode,
		OrderAmount: fmt.Sprintf("%.2f", order.OrderAmount),
	}

	if order.Status == "20" {
		resp.OrderStatus = "1"
	} else if order.Status == "30" {
		resp.OrderStatus = "2"
	}

	resp.Sign = utils.SortAndSign2(*resp, merchant.ScrectKey)
	return
}
