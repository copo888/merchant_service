package proxypayorder

import (
	"context"

	"com.copo/bo_service/merchant/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyPayRepaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayRepaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayRepaymentLogic {
	return ProxyPayRepaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayRepaymentLogic) ProxyPayRepayment() error {
	// todo: add your logic here and delete this line

	return nil
}
