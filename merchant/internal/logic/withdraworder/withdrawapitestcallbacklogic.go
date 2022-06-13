package withdraworder

import (
	"context"

	"com.copo/bo_service/merchant/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawApiTestCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawApiTestCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawApiTestCallBackLogic {
	return WithdrawApiTestCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawApiTestCallBackLogic) WithdrawApiTestCallBack() (resp string, err error) {
	// todo: add your logic here and delete this line
	logx.Info("測試回調")
	resp = "success"
	return
	return
}
