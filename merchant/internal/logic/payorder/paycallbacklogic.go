package payorder

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
	"errors"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/gioco-play/gozzle"
	"github.com/neccoys/go-zero-extension/redislock"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayCallBackLogic {
	return PayCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayCallBackLogic) PayCallBack(req types.PayCallBackRequest) (resp *types.PayCallBackResponse, err error) {

	// 只能回調成功/失敗
	if req.OrderStatus != "20" && req.OrderStatus != "30" {
		return nil, errorz.New(response.TRANSACTION_FAILURE, fmt.Sprintf("(req OrderStatus): %s", req.OrderStatus))
	}

	//检查单号是否存在
	orderX := &types.OrderX{}
	if req.PayOrderNo == "" {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST)
	} else if orderX, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.PayOrderNo, ""); err != nil && orderX == nil {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, "Copo OrderNo: "+req.PayOrderNo)
	}

	redisKey := fmt.Sprintf("%s-%s", orderX.MerchantCode, orderX.OrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "pay-call-back:")
	redisLock.SetExpire(5)
	if isOK, _ := redisLock.Acquire(); isOK {
		defer redisLock.Release()
		return l.DoPayCallBack(req)
	} else {
		logx.Infof(i18n.Sprintf(response.TRANSACTION_PROCESSING))
		return nil, errorz.New(response.TRANSACTION_PROCESSING)
	}
	return
}

func (l *PayCallBackLogic) DoPayCallBack(req types.PayCallBackRequest) (resp *types.PayCallBackResponse, err error) {

	// CALL transactionc PayOrderTranaction
	rpc := transactionclient.NewTransaction(l.svcCtx.RpcService("transaction.rpc"))
	callBackResp, err2 := rpc.PayCallBackTranaction(l.ctx, &transaction.PayCallBackRequest{
		CallbackTime:   req.CallbackTime,
		ChannelOrderNo: req.ChannelOrderNo,
		OrderAmount:    req.OrderAmount,
		OrderStatus:    req.OrderStatus,
		PayOrderNo:     req.PayOrderNo,
	})

	if err2 != nil {
		return nil, err2
	} else if callBackResp == nil {
		return nil, errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "PayCallBackTranaction callBackResp is nil")
	} else if callBackResp.Code != response.API_SUCCESS {
		return nil, errorz.New(callBackResp.Code, callBackResp.Message)
	}

	logx.Info("PayCallBackTranaction return:", callBackResp)

	// 只有成功單 且 有回掉網址 才回調
	if req.OrderStatus == "20" && len(callBackResp.NotifyUrl) > 0 {
		respString, err3 := l.callNoticeURL(callBackResp)
		if err3 == nil && respString == "success" {
			if err4 := l.svcCtx.MyDB.Table("tx_orders").
				Where("order_no = ?", req.PayOrderNo).
				Updates(map[string]interface{}{"is_merchant_callback": "1"}).Error; err4 != nil {
				logx.Error("回調成功,但更改回調狀態失敗")
			}
		} else {
			logx.Errorf("回調商戶失敗: %s, %#v", respString, err3)
		}
	}

	return
}

func (l *PayCallBackLogic) callNoticeURL(callBackResp *transaction.PayCallBackResponse) (respString string, err error) {

	var merchant *types.Merchant
	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", callBackResp.MerchantCode).
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return "", errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	payCallBackVO := vo.PayCallBackVO{
		AccessType:   "1",
		Language:     "zh-CN",
		MerchantId:   callBackResp.MerchantCode,
		OrderNo:      callBackResp.MerchantOrderNo,
		OrderTime:    callBackResp.OrderTime,
		PayOrderTime: callBackResp.PayOrderTime,
		Fee:          fmt.Sprintf("%.2f", callBackResp.TransferHandlingFee),
		PayOrderId:   callBackResp.OrderNo,
	}

	// 若有實際金額則回覆實際
	if callBackResp.ActualAmount > 0 {
		payCallBackVO.OrderAmount = fmt.Sprintf("%.2f", callBackResp.ActualAmount)
	} else {
		payCallBackVO.OrderAmount = fmt.Sprintf("%.2f", callBackResp.OrderAmount)
	}

	// API 支付状态 0：处理中，1：成功，2：失败，3：成功(人工确认)
	payCallBackVO.OrderStatus = "0"
	if callBackResp.Status == constants.SUCCESS {
		payCallBackVO.OrderStatus = "1"
	} else if callBackResp.Status == constants.FAIL {
		payCallBackVO.OrderStatus = "2"
	}

	payCallBackVO.Sign = utils.SortAndSign2(payCallBackVO, merchant.ScrectKey)

	// TODO: 通知商戶
	span := trace.SpanFromContext(l.ctx)
	res, errx := gozzle.Post(callBackResp.NotifyUrl).Timeout(10).Trace(span).JSON(payCallBackVO)
	if errx != nil {
		return "", errorz.New(response.GENERAL_EXCEPTION, err.Error())
	} else if res.Status() != 200 {
		logx.Errorf("call NotifyUrl httpStatus:%d", res.Status())
		return "", errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("call NotifyUrl httpStatus:%d", res.Status()))
	}
	respString = string(res.Body()[:])

	return
}
