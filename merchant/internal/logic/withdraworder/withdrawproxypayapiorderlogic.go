package withdraworder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	"context"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/neccoys/go-zero-extension/redislock"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"regexp"
	"strconv"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawProxyPayApiOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawProxyPayApiOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawProxyPayApiOrderLogic {
	return WithdrawProxyPayApiOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawProxyPayApiOrderLogic) WithdrawProxyPayApiOrder(req *types.ProxyPayRequestX) (resp *types.WithdrawApiOrderResponse, err error) {
	logx.Infof("Enter withdraw-order-proxy:", req)
	db := l.svcCtx.MyDB
	redisKey := fmt.Sprintf("%s-%s", req.MerchantId, req.OrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "withdraw:")
	redisLock.SetExpire(5)

	if isOK, _ := redisLock.Acquire(); isOK {
		defer redisLock.Release()
		var merchant types.Merchant
		var orderWithdrawCreateResp *types.OrderWithdrawCreateResponse
		if err = db.Table("mc_merchants").Where("code = ?", req.MerchantId).Take(&merchant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorz.New(response.DATA_NOT_FOUND, err.Error())
			} else {
				return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}
		// 檢查參數
		b, err := l.checkParams(db, req, merchant)
		if !b {
			return nil, err
		}
		var orderAmount float64
		if s, ok := req.OrderAmount.(string); ok {
			if s, err := strconv.ParseFloat(s, 64); err == nil {
				orderAmount = s
			}
		} else if f, ok := req.OrderAmount.(float64); ok {
			orderAmount = f
		}
		var withdrawOrders []types.OrderWithdrawCreateRequestX
		var withdrawOrder types.OrderWithdrawCreateRequestX
		withdrawOrder.Type = "XF"
		withdrawOrder.MerchantAccountName = req.DefrayName
		withdrawOrder.MerchantBankName = req.BankName
		withdrawOrder.MerchantBankProvince = req.BankProvince
		withdrawOrder.MerchantBankCity = req.BankCity
		withdrawOrder.MerchantBankAccount = req.BankNo
		withdrawOrder.CurrencyCode = req.Currency
		withdrawOrder.OrderAmount = orderAmount
		withdrawOrder.Source = constants.API
		withdrawOrder.MerchantCode = req.MerchantId
		withdrawOrder.UserAccount = req.MerchantId
		withdrawOrder.NotifyUrl = req.NotifyUrl
		withdrawOrder.MerchantOrderNo = req.OrderNo
		withdrawOrder.ChangeType = "1" // 1:下發轉代付

		withdrawOrders = append(withdrawOrders, withdrawOrder)

		orderWithdrawCreateResp, err = ordersService.WithdrawOrderCreate(db, withdrawOrders, constants.API, l.ctx, l.svcCtx)
		if err != nil {
			logx.Error("下發api提單失敗: ", err.Error())
			return nil, err
		}

		i18n.SetLang(language.English)
		newData := make(map[string]string)
		newData["withdrawOrderNo"] = orderWithdrawCreateResp.OrderNo
		newSign := utils.SortAndSign(newData, merchant.ScrectKey)

		respData := types.RespData{
			WithdrawOrderNo: orderWithdrawCreateResp.OrderNo,
			Sign:            newSign,
		}
		resp = &types.WithdrawApiOrderResponse{
			RespCode: response.API_SUCCESS,
			RespMsg:  i18n.Sprintf(response.API_SUCCESS), //固定回商戶成功
			RespData: respData,
		}
		return resp, nil
	} else {
		return nil, errorz.New(response.API_GENERAL_ERROR)
	}
}

func (l *WithdrawProxyPayApiOrderLogic) checkParams(db *gorm.DB, req *types.ProxyPayRequestX, merchant types.Merchant) (b bool, err error) {
	var merchantCurrency *types.MerchantCurrency
	// 检查币别
	if err = db.Table("mc_merchant_currencies").Where("merchant_code = ? AND currency_code = ?", req.MerchantId, req.Currency).
		Take(&merchantCurrency).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errorz.New(response.MERCHANT_CURRENCY_NOT_SET, err.Error())
		} else {
			return false, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	// 檢查白名單
	if isWhite := merchantsService.IPChecker(req.Ip, merchant.ApiIP); !isWhite {
		return false, errorz.New(response.API_IP_DENIED, "IP: "+req.Ip)
	}

	//验证银行卡号(必填)(必须为数字)(长度必须在10~22码)
	isMatch2, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.BankNo)
	currencyCode := req.Currency
	if currencyCode == constants.CURRENCY_THB {
		if req.BankNo == "" || len(req.BankNo) < 10 || len(req.BankNo) > 16 || !isMatch2 {
			logx.Error("銀行卡號檢查錯誤，需10-16碼內：", req.BankNo)
			return false,errorz.New(response.INVALID_BANK_NO, "BankNo: " + req.BankNo)
		}
	}else if currencyCode == constants.CURRENCY_CNY {
		if req.BankNo == "" || len(req.BankNo) < 13 || len(req.BankNo) > 22 || !isMatch2 {
			logx.Error("銀行卡號檢查錯誤，需13-22碼內：", req.BankNo)
			return false,errorz.New(response.INVALID_BANK_NO, "BankNo: " + req.BankNo)
		}
	}

	var orderAmount float64
	if s, ok := req.OrderAmount.(string); ok {
		if s, err := strconv.ParseFloat(s, 64); err == nil {
			orderAmount = s
		} else {
			return false, errorz.New(response.API_INVALID_PARAMETER, err.Error())
		}
	} else if f, ok := req.OrderAmount.(float64); ok {
		orderAmount = f
	} else {
		s := fmt.Sprintf("OrderAmount err: %#v", req.OrderAmount)
		logx.Errorf(s)
		return false, errorz.New(response.API_INVALID_PARAMETER, s)
	}
	req.ProxyPayOrderRequest.OrderAmount = orderAmount
	// 驗簽檢查
	checkSign := utils.VerifySign(req.Sign, req.ProxyPayOrderRequest, merchant.ScrectKey)
	if !checkSign {
		return false, errorz.New(response.INVALID_SIGN)
	}

	//确认是否重复订单
	var isExist bool
	if err = db.Table("tx_orders").
		Select("count(*) > 0 ").
		Where("merchant_code = ? AND merchant_order_no = ?", req.MerchantId, req.OrderNo).
		Find(&isExist).Error; err != nil {
		return false, errorz.New(response.GENERAL_EXCEPTION)
	}
	if isExist {
		return false, errorz.New(response.REPEAT_ORDER_NO)
	}
	return true, nil
}
