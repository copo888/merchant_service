package payorder

import (
	"context"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestMerchantPayCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestMerchantPayCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestMerchantPayCallBackLogic {
	return TestMerchantPayCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestMerchantPayCallBackLogic) TestMerchantPayCallBack(req *types.TestMerchantPayCallBackRequest) (resp string, err error) {
	// todo: add your logic here and delete this line
	logx.Infof("測試回調 %#v", req)
	resp = "success"
	return
}
