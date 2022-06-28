package proxypayorder

import (
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/service/merchantbalanceservice"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"errors"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/neccoys/go-zero-extension/redislock"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"strconv"
)

type ProxyPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayOrderLogic {
	return ProxyPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayOrderLogic) ProxyPayOrder(merReq *types.ProxyPayRequestX) (*types.ProxyPayOrderResponse, error) {
	var resp *types.ProxyPayOrderResponse
	var err error
	redisKey := fmt.Sprintf("%s-%s", merReq.MerchantId, merReq.OrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "proxy-pay:")
	redisLock.SetExpire(5)
	if isOK, _ := redisLock.Acquire(); isOK {
		if resp, err = l.internalProxyPayOrder(merReq); err != nil {
			return nil, err
		}
		defer redisLock.Release()
	} else {
		return nil, errorz.New(response.TRANSACTION_PROCESSING)
	}
	return resp, err
}

func (l *ProxyPayOrderLogic) internalProxyPayOrder(merReq *types.ProxyPayRequestX) (*types.ProxyPayOrderResponse, error) {
	logx.Info("Enter proxy-order:", merReq)

	// 1. 檢查白名單及商户号
	merchantKey, errWhite := l.CheckMerAndWhiteList(merReq)
	if errWhite != nil {
		logx.Error("商戶號及白名單檢查錯誤: ", errWhite.Error())
		return nil, errWhite
	}

	// 2. 處理商户提交参数、訂單驗證，並返回商戶費率
	rate, errCreate := ordersService.ProxyOrder(l.svcCtx.MyDB, merReq)
	if errCreate != nil {
		logx.Error("代付提單商户提交参数驗證錯誤: ", errCreate.Error())
		return nil, errCreate
	}

	balanceType, errBalance := merchantbalanceservice.GetBalanceType(l.svcCtx.MyDB, rate.ChannelCode, constants.ORDER_TYPE_DF)
	if errBalance != nil {
		return nil, errBalance
	}

	// 3. 依balanceType决定要呼叫哪种transaction方法
	//    呼叫 transaction rpc (merReq, rate) (ProxyOrderNo) 并产生订单

	//產生rpc 代付需要的請求的資料物件
	ProxyPayOrderRequest, rateRpc := generateRpcdata(merReq, rate)

	rpc := transactionclient.NewTransaction(l.svcCtx.RpcService("transaction.rpc"))

	var errRpc error
	var res *transaction.ProxyOrderResponse
	if balanceType == "DFB" {
		res, errRpc = rpc.ProxyOrderTranaction_DFB(l.ctx, &transaction.ProxyOrderRequest{
			Req:  ProxyPayOrderRequest,
			Rate: rateRpc,
		})
	} else if balanceType == "XFB" {
		res, errRpc = rpc.ProxyOrderTranaction_XFB(l.ctx, &transaction.ProxyOrderRequest{
			Req:  ProxyPayOrderRequest,
			Rate: rateRpc,
		})
	}

	if errRpc != nil {
		logx.Error("代付提單:", errRpc.Error())
		return nil, errorz.New(response.FAIL, errRpc.Error())
	} else {
		logx.Infof("代付交易rpc完成，%s 錢包扣款完成: %#v", balanceType, res)
	}

	var queryErr error
	var respOrder = &types.OrderX{}
	if respOrder, queryErr = model.QueryOrderByOrderNo(l.svcCtx.MyDB, res.ProxyOrderNo, ""); queryErr != nil {
		logx.Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
		return nil, errorz.New(response.FAIL, queryErr.Error())
	}

	// 4: call channel (不論是否有成功打到渠道，都要返回給商戶success，一渠道返回訂單狀態決定此訂單狀態(代處理/處理中))
	var errCHN error
	proxyPayRespVO := &vo.ProxyPayRespVO{}
	proxyPayRespVO, errCHN = ordersService.CallChannel_ProxyOrder(&l.ctx, &l.svcCtx.Config, merReq, respOrder, rate)
	//5. 返回給商戶物件
	var proxyResp = types.ProxyPayOrderResponse{}
	i18n.SetLang(language.English)
	if errCHN != nil || proxyPayRespVO.Code != "0" {
		logx.Errorf("代付提單: %s ，渠道返回錯誤: %s, %#v", respOrder.OrderNo, errCHN, proxyPayRespVO)
		proxyResp.RespCode = response.CHANNEL_REPLY_ERROR
		proxyResp.RespMsg = i18n.Sprintf(response.CHANNEL_REPLY_ERROR) + ": Code: " + proxyPayRespVO.Code + " Message: " + proxyPayRespVO.Message

		//将商户钱包加回 (merchantCode, orderNO)，更新狀態為失敗單
		var resRpc *transaction.ProxyPayFailResponse
		if balanceType == "DFB" {
			resRpc, errRpc = rpc.ProxyOrderTransactionFail_DFB(l.ctx, &transaction.ProxyPayFailRequest{
				MerchantCode: respOrder.MerchantCode,
				OrderNo:      respOrder.OrderNo,
			})
		} else if balanceType == "XFB" {
			resRpc, errRpc = rpc.ProxyOrderTransactionFail_XFB(l.ctx, &transaction.ProxyPayFailRequest{
				MerchantCode: respOrder.MerchantCode,
				OrderNo:      respOrder.OrderNo,
			})
		}
		//因在transaction_service 已更新過訂單，重新抓取訂單
		if respOrder, queryErr = model.QueryOrderByOrderNo(l.svcCtx.MyDB, res.ProxyOrderNo, ""); queryErr != nil {
			logx.Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
			return nil, errorz.New(response.FAIL, queryErr.Error())
		}

		respOrder.ErrorType = "1" //   1.渠道返回错误	2.渠道异常	3.商户参数错误	4.账户为黑名单	5.其他
		respOrder.ErrorNote = "渠道返回错误: " + proxyPayRespVO.Message

		if errRpc != nil {
			logx.Errorf("代付提单 %s 还款失败。 Err: %s", respOrder.OrderNo, errRpc.Error())
			respOrder.RepaymentStatus = constants.REPAYMENT_FAIL
			return nil, errorz.New(response.FAIL, errRpc.Error())
		} else {
			logx.Infof("代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
			respOrder.RepaymentStatus = constants.REPAYMENT_SUCCESS
		}

	} else {
		respOrder.ChannelOrderNo = proxyPayRespVO.Data.ChannelOrderNo
		//条整订单状态从"待处理" 到 "交易中"
		respOrder.Status = constants.TRANSACTION
		proxyResp.RespCode = response.API_SUCCESS
		proxyResp.RespMsg = i18n.Sprintf(response.API_SUCCESS) //固定回商戶成功
	}

	// 更新订单
	if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
		logx.Error("代付订单更新状态错误: ", errUpdate.Error())
	}
	// 依渠道返回给予订单状态
	var orderStatus string
	if respOrder.Status == constants.FAIL {
		orderStatus = "2"
	} else {
		orderStatus = "0"
	}

	proxyResp.MerchantId = respOrder.MerchantCode
	proxyResp.OrderNo = respOrder.MerchantOrderNo
	proxyResp.PayOrderNo = respOrder.OrderNo
	proxyResp.OrderStatus = orderStatus //渠道返回成功: "處理中" 失敗: "失敗"
	proxyResp.Sign = utils.SortAndSign2(proxyResp, merchantKey)

	return &proxyResp, nil
}

//检查商户号是否存在以及IP是否为白名单，若无误则返回"商户密鑰"
func (l *ProxyPayOrderLogic) CheckMerAndWhiteList(req *types.ProxyPayRequestX) (merchantKey string, err error) {
	merchant := &types.Merchant{}
	// 檢查白名單
	if err = l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", req.MerchantId).Take(merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errorz.New(response.DATA_NOT_FOUND, err.Error())
		} else if err == nil && merchant != nil && merchant.Status != constants.MerchantStatusEnable {
			return "", errorz.New(response.MERCHANT_ACCOUNT_NOT_FOUND, "商户号:"+merchant.Code)
		} else {
			return "", errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	if isWhite := merchantsService.IPChecker(req.Ip, merchant.ApiIP); !isWhite {
		return "", errorz.New(response.API_IP_DENIED, "IP: "+req.Ip)
	}

	if errCurrency := l.svcCtx.MyDB.Table("mc_merchant_currencies").Where("merchant_code = ? AND currency_code = ? AND status = ?", req.MerchantId, req.Currency, '1').Error; errCurrency != nil {
		return "", errorz.New(response.MERCHANT_CURRENCY_NOT_SET)
	}

	//TODO 檢查幣別啟用、商戶錢包是否啟用禁運
	return merchant.ScrectKey, nil
}

// 產生rpc 代付需要的請求的資料物件
func generateRpcdata(merReq *types.ProxyPayRequestX, rate *types.CorrespondMerChnRate) (*transaction.ProxyPayOrderRequest, *transaction.CorrespondMerChnRate) {

	var orderAmount float64
	if s, ok := merReq.OrderAmount.(string); ok {
		if s, err := strconv.ParseFloat(s, 64); err == nil {
			orderAmount = s
		}
	} else if f, ok := merReq.OrderAmount.(float64); ok {
		orderAmount = f
	}

	ProxyPayOrderRequest := &transaction.ProxyPayOrderRequest{
		AccessType:   merReq.AccessType,
		MerchantId:   merReq.MerchantId,
		Sign:         merReq.Sign,
		NotifyUrl:    merReq.NotifyUrl,
		Language:     merReq.Language,
		OrderNo:      merReq.OrderNo,
		BankId:       merReq.BankId,
		BankName:     merReq.BankName,
		BankProvince: merReq.BankProvince,
		BankCity:     merReq.BankCity,
		BranchName:   merReq.BranchName,
		BankNo:       merReq.BankNo,
		OrderAmount:  orderAmount,
		DefrayName:   merReq.DefrayName,
		DefrayId:     merReq.DefrayId,
		DefrayMobile: merReq.DefrayMobile,
		DefrayEmail:  merReq.DefrayEmail,
		Currency:     merReq.Currency,
		PayTypeSubNo: merReq.PayTypeSubNo,
	}
	rateRpc := &transaction.CorrespondMerChnRate{
		MerchantCode:        rate.MerchantCode,
		ChannelPayTypesCode: rate.ChannelPayTypesCode,
		ChannelCode:         rate.ChannelCode,
		PayTypeCode:         rate.PayTypeCode,
		Designation:         rate.Designation,
		DesignationNo:       rate.DesignationNo,
		Fee:                 rate.Fee,
		HandlingFee:         rate.HandlingFee,
		ChFee:               rate.ChFee,
		ChHandlingFee:       rate.ChHandlingFee,
		SingleMinCharge:     rate.SingleMinCharge,
		SingleMaxCharge:     rate.SingleMaxCharge,
		CurrencyCode:        rate.CurrencyCode,
		ApiUrl:              rate.ApiUrl,
	}

	return ProxyPayOrderRequest, rateRpc
}
