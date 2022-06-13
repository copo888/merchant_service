package ordersService

import (
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/gioco-play/gozzle"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strings"
)

func WithdrawOrderCreate(db *gorm.DB, req []types.OrderWithdrawCreateRequestX, orderSource string, ctx context.Context, svcCtx *svc.ServiceContext) (resp *types.OrderWithdrawCreateResponse, err error) {
	var orders = req

	systemRate := types.SystemRate{}
	merchantCode := orders[0].MerchantCode
	userAccount := orders[0].UserAccount
	var currency = orders[0].CurrencyCode
	var terms []string

	terms = append(terms, fmt.Sprintf("currency_code = '%s'", currency))
	term := strings.Join(terms, " AND ")
	// 取得商户下发上下限资料
	if err = db.Table("bs_system_rate").Where(term).Find(&systemRate).Error; err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}
	// 判断有无设定下发上下限与手续费
	if systemRate.MinWithdrawCharge <= 0 {
		return nil, errorz.New(response.MER_WITHDRAW_MIN_LIMIT_NOT_SET)
	} else if systemRate.MaxWithdrawCharge <= 0 {
		return nil, errorz.New(response.MER_WITHDRAW_MAX_LIMIT_NOT_SET)
	} else if systemRate.WithdrawHandlingFee <= 0 {
		return nil, errorz.New(response.MER_WITHDRAW_CHARGE_NOT_SET)
	}

	var idxs []string
	var errs []string

	order := orders[0]
	transAmount := utils.FloatAdd(order.OrderAmount, systemRate.WithdrawHandlingFee)
	if systemRate.MaxWithdrawCharge < transAmount {
		//下发金额超过上限
		return nil, errorz.New(response.WITHDRAW_AMT_EXCEED_MAX_LIMIT)
	} else if systemRate.MinWithdrawCharge > transAmount {
		//下发金额未达下限
		return nil, errorz.New(response.WITHDRAW_AMT_NOT_REACH_MIN_LIMIT)
	}

	var errRpc error
	var res *transaction.WithdrawOrderResponse
	rpc := transactionclient.NewTransaction(svcCtx.RpcService("transaction.rpc"))

	res, errRpc = rpc.WithdrawOrderTransaction(ctx, &transaction.WithdrawOrderRequest{
		MerchantCode: merchantCode,
		UserAccount: userAccount,
		MerchantAccountName: order.MerchantAccountName,
		MerchantBankeAccount: order.MerchantBankAccount,
		MerchantBankNo: order.MerchantBankNo,
		MerchantBankName: order.MerchantBankName,
		MerchantBankProvince: order.MerchantBankProvince,
		MerchantBankCity: order.MerchantBankCity,
		CurrencyCode: order.CurrencyCode,
		OrderAmount: order.OrderAmount,
		OrderNo: model.GenerateOrderNo("XF"),
		HandlingFee: systemRate.WithdrawHandlingFee,
		Source: constants.API,
		MerchantOrderNo: order.MerchantOrderNo,
		NotifyUrl: order.NotifyUrl,
		PageUrl: order.PageUrl,
	})

	if errRpc != nil {
		logx.Error("下发提單:", errRpc.Error())
		return nil, errorz.New(response.FAIL, errRpc.Error())
	} else {
		logx.Infof("下发提单rpc完成，单号: %v", "XFB", res)
	}

	resp = &types.OrderWithdrawCreateResponse{
		OrderNo: res.OrderNo,
		Index:  idxs,
		Errs:   errs,
	}

	return
}

func WithdrawApiCallBack(db *gorm.DB, req types.OrderX) error {
	var orderX types.OrderX
	var merchant types.Merchant
	// 確認單號是否存在
	if err := db.Table("tx_orders").Where("merchant_order_no = ?", req.MerchantCode).Take(&orderX).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logx.Errorf("下发回调错误: 查无订单。商户订单号: %v", req.MerchantOrderNo)
			return errorz.New(response.INVALID_ORDER_NO, err.Error())
		} else {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	// 取得商戶密鑰
	if err := db.Table("mc_merchants").Where("code = ?", req.MerchantCode).Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logx.Errorf("下发回调错误: 查无商戶。商户号: %v", req.MerchantCode)
			return errorz.New(response.INVALID_MERCHANT_ID, err.Error())
		} else {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	// 状态 0：处理中，1：成功，2：失败，3：成功(人工确认)
	var orderStatus = "0"
	// 訂單狀態(0:待處理 1:處理中 20:成功 30:失敗 31:凍結)
	if orderX.Status == "20" {
		orderStatus = "2"
	} else if orderX.Status == "30" {
		orderStatus = "3"
	}

	resp := vo.WithdrawCallBackVO{
		MerchantId:  orderX.MerchantCode,
		OrderNo:     orderX.MerchantOrderNo,
		OrderAmount: fmt.Sprintf("%.2f", orderX.OrderAmount),
		OrderTime:   orderX.CreatedAt.Format("20060102150405"),
		ReviewTime:  orderX.TransAt.Time().Format("20060102150405"),
		Fee:         fmt.Sprintf("%.2f", orderX.TransferHandlingFee),
		OrderStatus: orderStatus,
		DiorOrderNo: orderX.OrderNo,
	}

	resp.Sign = utils.SortAndSign2(resp, merchant.ScrectKey)

	// 通知商戶
	res, err := gozzle.Post(orderX.NotifyUrl).JSON(resp)
	logx.Errorf("回調商戶，商戶返回 : %v", res)
	if res.String() != "success" || err != nil {
		logx.Errorf("錯誤: 回調商戶失敗，err : %v， res : %v", err, res.String())
		orderX.IsMerchantCallback = constants.IS_MERCHANT_CALLBACK_NO
	} else if res.String() == "success" {
		orderX.IsMerchantCallback = constants.IS_MERCHANT_CALLBACK_YES
	}
	if err1 := db.Table("tx_orders").Updates(orderX).Error; err1 != nil {
		db.Rollback()
		return errorz.New(response.DATABASE_FAILURE, err1.Error())
	}
	return nil
}
