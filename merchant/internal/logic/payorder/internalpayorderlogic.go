package payorder

import (
	"com.copo/bo_service/common/constants/redisKey"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	"context"
	"encoding/json"
	"errors"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/jinzhu/copier"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InternalPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInternalPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) InternalPayOrderLogic {
	return InternalPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InternalPayOrderLogic) InternalPayOrder(request *types.InternalPayOrderRequest) (resp *types.PayOrderResponse, err error) {

	var merchant *types.Merchant
	req := types.PayOrderRequestX{}
	orderNo := request.ID

	// 取得存在Redis的資料
	result, err := l.svcCtx.RedisClient.Get(l.ctx, redisKey.CACHE_ORDER_DATA + orderNo).Result()
	if err != nil {
		return nil, errorz.New(response.API_INVALID_PARAMETER, err.Error())
	}
	if err = json.Unmarshal([]byte(result), &req); err != nil {
		return nil, errorz.New(response.API_PARAMETER_TYPE_ERROE, err.Error())
	}
	req.JumpType = ""
	req.UserId = request.UserName


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

	// 资料验证
	if err = ordersService.VerifyPayOrder(l.svcCtx.MyDB, req, merchant); err != nil {
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
	// CALL transactionc PayOrderTranaction
	rpc := transactionclient.NewTransaction(l.svcCtx.RpcService("transaction.rpc"))
	rpcResp, err2 := rpc.PayOrderTranaction(l.ctx, &transaction.PayOrderRequest{
		PayOrder:       &rpcPayOrder,
		Rate:           &rpcRate,
		OrderNo:        orderNo,
		ChannelOrderNo: payReplyVO.ChannelOrderNo,
	})

	if err2 != nil {
		return nil, err2
	} else if rpcResp == nil {
		return nil, errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "PayOrderTranaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return nil, errorz.New(rpcResp.Code, rpcResp.Message)
	}

	// 判斷返回格式 1.html, 2.json  3.url
	resp, err = ordersService.GetPayOrderResponse(req, *payReplyVO, orderNo,l.ctx , l.svcCtx)
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
