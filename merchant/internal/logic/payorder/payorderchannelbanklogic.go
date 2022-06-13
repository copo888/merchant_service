package payorder

import (
	"com.copo/bo_service/common/constants/redisKey"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"encoding/json"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayOrderChannelBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayOrderChannelBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayOrderChannelBankLogic {
	return PayOrderChannelBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayOrderChannelBankLogic) PayOrderChannelBank(req *types.PayOrderChannelBankRequest) (resp *types.PayOrderChannelBankResponse, err error) {
	// 取得存在Redis的資料

	redisKey := redisKey.CACHE_PAY_ORDER_CHANNEL_BANK + req.OrderNo
	result, err := l.svcCtx.RedisClient.Get(l.ctx, redisKey).Result()
	if err != nil {
		return nil, errorz.New(response.INVALID_ORDER_NO, err.Error())
	}
	expiration, err := l.svcCtx.RedisClient.TTL(l.ctx, redisKey).Result()
	if err != nil {
		return nil, errorz.New(response.INVALID_ORDER_NO, err.Error())
	}
	if err = json.Unmarshal([]byte(result), &resp); err != nil {
		return nil, errorz.New(response.API_PARAMETER_TYPE_ERROE, err.Error())
	}
	resp.Expiration = expiration.Seconds()

	return
}
