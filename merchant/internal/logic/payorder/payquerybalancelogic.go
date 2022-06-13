package payorder

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"context"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"gorm.io/gorm"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayQueryBalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayQueryBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayQueryBalanceLogic {
	return PayQueryBalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayQueryBalanceLogic) PayQueryBalance(req types.PayQueryBalanceRequestX) (resp *types.PayQueryBalanceResponse, err error) {
	var merchant *types.Merchant
	var dfBalance *types.MerchantBalance
	var xfBalance *types.MerchantBalance
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
	if isSameSign := utils.VerifySign(req.Sign, req.PayQueryBalanceRequest, merchant.ScrectKey); !isSameSign {
		return nil, errorz.New(response.SIGN_KEY_FAIL)
	}

	if len(req.Currency) < 2 {
		req.Currency = "CNY"
	}

	// TODO: 兩種餘額兩筆資料
	if err = l.svcCtx.MyDB.Table("mc_merchant_balances").
		Where("merchant_code = ?", req.MerchantId).
		Where("currency_code = ?", req.Currency).
		Where("balance_type =  'DFB' ").
		Take(&dfBalance).Error; err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}
	if err = l.svcCtx.MyDB.Table("mc_merchant_balances").
		Where("merchant_code = ?", req.MerchantId).
		Where("currency_code = ?", req.Currency).
		Where("balance_type =  'XFB' ").
		Take(&xfBalance).Error; err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	resp = &types.PayQueryBalanceResponse{
		RespCode:        response.API_SUCCESS,
		RespMsg:         i18n.Sprintf(response.API_SUCCESS),
		FrozenAmount:    fmt.Sprintf("%.2f", utils.FloatAdd(xfBalance.FrozenAmount, dfBalance.FrozenAmount)),
		PayAmount:       fmt.Sprintf("%.2f", xfBalance.Balance),
		ProxyAmount:     fmt.Sprintf("%.2f", dfBalance.Balance),
		AvailableAmount: fmt.Sprintf("%.2f", utils.FloatAdd(xfBalance.Balance, dfBalance.Balance)),
	}
	resp.Sign = utils.SortAndSign2(*resp, merchant.ScrectKey)
	return
}
