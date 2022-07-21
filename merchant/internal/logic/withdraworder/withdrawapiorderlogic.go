package withdraworder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/neccoys/go-zero-extension/redislock"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"regexp"
	"strconv"
)

type WithdrawApiOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawApiOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawApiOrderLogic {
	return WithdrawApiOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawApiOrderLogic) WithdrawApiOrder(req *types.WithdrawApiOrderRequestX) (resp *types.WithdrawApiOrderResponse, err error) {
	logx.Infof("Enter withdraw-order: %#v", req)
	db := l.svcCtx.MyDB
	redisKey := fmt.Sprintf("%s-%s", req.MerchantId, req.OrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "withdraw:")
	redisLock.SetExpire(5)

	if isOK, _ := redisLock.Acquire(); isOK {
		defer redisLock.Release()
		var orderWithdrawCreateResp *types.OrderWithdrawCreateResponse
		var merchant types.Merchant
		var merchantCurrency *types.MerchantCurrency
		// 检查币别
		if err = db.Table("mc_merchant_currencies").Where("merchant_code = ? AND currency_code = ?", req.MerchantId, req.Currency).
			Take(&merchantCurrency).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorz.New(response.MERCHANT_CURRENCY_NOT_SET, err.Error())
			} else {
				return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		// 檢查白名單
		if err = db.Table("mc_merchants").Where("code = ?", req.MerchantId).Take(&merchant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorz.New(response.DATA_NOT_FOUND, err.Error())
			} else {
				return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
			logx.Error("白名單檢查錯誤: ", req.MyIp)
			return nil, errorz.New(response.API_IP_DENIED, "IP: "+req.MyIp)
		}

		//验证银行卡号(必填)(必须为数字)(长度必须在10~22码)
		isMatch2, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.AccountNo)
		currencyCode := req.Currency
		if currencyCode == constants.CURRENCY_THB {
			if req.AccountNo == "" || len(req.AccountNo) < 10 || len(req.AccountNo) > 16 || !isMatch2 {
				logx.Error("銀行卡號檢查錯誤，需10-16碼內：", req.AccountNo)
				return nil, errorz.New(response.INVALID_BANK_NO, "AccountNo: "+req.AccountNo)
			}
		} else if currencyCode == constants.CURRENCY_CNY {
			if req.AccountNo == "" || len(req.AccountNo) < 13 || len(req.AccountNo) > 22 || !isMatch2 {
				logx.Error("銀行卡號檢查錯誤，需13-22碼內：", req.AccountNo)
				return nil, errorz.New(response.INVALID_BANK_NO, "AccountNo: "+req.AccountNo)
			}
		}

		// 驗簽檢查
		if isSameSign := utils.VerifySign(req.Sign, req.WithdrawApiOrderRequest, merchant.ScrectKey); !isSameSign {
			logx.Error("驗簽檢查錯誤: ", req.Sign)
			return nil, errorz.New(response.INVALID_SIGN)
		}

		orderAmount, errParse := strconv.ParseFloat(req.OrderAmount, 64)
		if errParse != nil {
			return nil, errorz.New(response.GENERAL_EXCEPTION, errParse.Error())
		}

		//确认是否重复订单
		var isExist bool
		if err = db.Table("tx_orders").
			Select("count(*) > 0 ").
			Where("merchant_code = ? AND merchant_order_no = ?", req.MerchantId, req.OrderNo).
			Find(&isExist).Error; err != nil {
			return nil, errorz.New(response.GENERAL_EXCEPTION)
		}
		if isExist {
			logx.Error("訂單號重複錯誤: ", req.OrderNo)
			return nil, errorz.New(response.REPEAT_ORDER_NO)
		}

		var withdrawOrders []types.OrderWithdrawCreateRequestX
		var withdrawOrder types.OrderWithdrawCreateRequestX
		withdrawOrder.Type = "XF"
		withdrawOrder.MerchantAccountName = req.WithdrawName
		withdrawOrder.MerchantBankName = req.BankName
		withdrawOrder.MerchantBankProvince = req.BankProvince
		withdrawOrder.MerchantBankCity = req.BankCity
		withdrawOrder.MerchantBankAccount = req.AccountNo
		withdrawOrder.CurrencyCode = req.Currency
		withdrawOrder.OrderAmount = orderAmount
		withdrawOrder.Source = constants.API
		withdrawOrder.MerchantCode = req.MerchantId
		withdrawOrder.UserAccount = req.MerchantId
		withdrawOrder.NotifyUrl = req.NotifyUrl
		withdrawOrder.MerchantOrderNo = req.OrderNo
		withdrawOrder.PageUrl = req.PageUrl

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

		return
	} else {
		return nil, errorz.New(response.TRANSACTION_PROCESSING)
	}
}
