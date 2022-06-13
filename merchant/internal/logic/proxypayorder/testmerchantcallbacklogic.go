package proxypayorder

import (
	"context"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestMerchantCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestMerchantCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestMerchantCallBackLogic {
	return TestMerchantCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestMerchantCallBackLogic) TestMerchantCallBack(req *types.MerchantCallBackReqeuest) (resp string, err error) {
	// todo: add your logic here and delete this line
	return "success", nil
}
