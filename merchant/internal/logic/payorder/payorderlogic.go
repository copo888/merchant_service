package payorder

import (
	"com.copo/bo_service/common/constants/redisKey"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/jinzhu/copier"
	"github.com/neccoys/go-zero-extension/redislock"
	"golang.org/x/text/language"
	"time"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayOrderLogic {
	return PayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayOrderLogic) PayOrder(req types.PayOrderRequestX) (resp *types.PayOrderResponse, err error) {

	redisKey := fmt.Sprintf("%s-%s", req.MerchantId, req.OrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "pay-order:")
	redisLock.SetExpire(5)

	if isOK, _ := redisLock.Acquire(); isOK {
		if resp, err = l.DoPayOrder(req); err != nil {
			return
		}
		defer redisLock.Release()
	} else {
		return nil, errorz.New(response.TRANSACTION_PROCESSING)
	}

	return
}

func (l *PayOrderLogic) DoPayOrder(req types.PayOrderRequestX) (resp *types.PayOrderResponse, err error) {

	var merchant *types.Merchant

	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req.MerchantId).
		Where("status = ?", "1").
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	// 檢查白名單
	if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.MyIp)
	}

	// 檢查驗簽 TODO: 驗簽先拿掉
	//if isSameSign := utils.VerifySign(req.Sign, req.PayOrderRequest, merchant.ScrectKey); !isSameSign {
	//	return nil, errorz.New(response.INVALID_SIGN)
	//}

	// 资料验证
	if err = ordersService.VerifyPayOrder(l.svcCtx.MyDB, req, merchant); err != nil {
		return
	}

	// 產生訂單號
	orderNo := model.GenerateOrderNo("ZF")

	// 確認是否返回實名制UI畫面
	if req.JumpType == "UI" {
		resp, err = l.RequireUserIdPage(req, orderNo)
		return
	}

	// Call Channel
	payReplyVO, correspondMerChnRate, chaErr := ordersService.CallChannelForPay(l.svcCtx.MyDB, req, merchant, orderNo, l.ctx, l.svcCtx)
	if chaErr != nil {
		return nil, chaErr
	}

	// Call GRPC transaction_service
	var rpcPayOrder transaction.PayOrder
	var rpcRate transaction.CorrespondMerChnRate
	copier.Copy(&rpcPayOrder, &req)
	copier.Copy(&rpcRate, correspondMerChnRate)
	if s, ok := req.OrderAmount.(string); ok {
		rpcPayOrder.OrderAmount = s
	} else if f, ok := req.OrderAmount.(float64); ok {
		rpcPayOrder.OrderAmount = fmt.Sprintf("%.2f", f)
	} else {
		s := fmt.Sprintf("OrderAmount err: %#v", req.OrderAmount)
		logx.Errorf(s)
		return resp, errorz.New(response.API_INVALID_PARAMETER, s)
	}
	// CALL transactionc PayOrderTranaction
	rpc := transactionclient.NewTransaction(l.svcCtx.RpcService("transaction.rpc"))
	rpcResp, err2 := rpc.PayOrderTranaction(l.ctx, &transaction.PayOrderRequest{
		PayOrder:       &rpcPayOrder,
		Rate:           &rpcRate,
		OrderNo:        orderNo,
		ChannelOrderNo: payReplyVO.ChannelOrderNo,
	})
	if err2 != nil {
		logx.Errorf("PayOrderTranaction rpcResp error:%s", err2.Error())
		return nil, err2
	} else if rpcResp == nil {
		logx.Errorf("Code:%s, Message:%s", rpcResp.Code, rpcResp.Message)
		return nil, errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "PayOrderTranaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		logx.Errorf("Code:%s, Message:%s", rpcResp.Code, rpcResp.Message)
		return nil, errorz.New(rpcResp.Code, rpcResp.Message)
	}

	// 判斷返回格式 1.html, 2.json  3.url
	resp, err = ordersService.GetPayOrderResponse(req, *payReplyVO, orderNo, l.ctx, l.svcCtx)
	if err != nil {
		return
	}
	i18n.SetLang(language.English)
	resp.RespCode = response.API_SUCCESS
	resp.RespMsg = i18n.Sprintf(response.API_SUCCESS)
	resp.Status = 0
	resp.Sign = utils.SortAndSign2(*resp, merchant.ScrectKey)

	return
}

func (l *PayOrderLogic) RequireUserIdPage(req types.PayOrderRequestX, orderNo string) (*types.PayOrderResponse, error) {
	// 资料转JSON
	dataJson, err := json.Marshal(req)
	if err != nil {
		return nil, errorz.New(response.API_PARAMETER_TYPE_ERROE)
	}
	// 存 Redis
	if err = l.svcCtx.RedisClient.Set(l.ctx, redisKey.CACHE_ORDER_DATA + orderNo, dataJson, 30 * time.Minute).Err(); err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION)
	}

	url := fmt.Sprintf("%s/#/checkoutPlayer?id=%s&lang=%s", l.svcCtx.Config.FrontEndDomain, orderNo, req.Language)

	return &types.PayOrderResponse{
		Status:     0,
		PayOrderNo: orderNo,
		BankCode:   req.BankCode,
		Type:       "url",
		Info:       url,
	}, nil
}
