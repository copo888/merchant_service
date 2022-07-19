package test

import (
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/logic/payorder"
	"com.copo/bo_service/merchant/internal/model"
	"context"
	"github.com/jinzhu/copier"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestPayOrderHndlerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestPayOrderHndlerLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestPayOrderHndlerLogic {
	return TestPayOrderHndlerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestPayOrderHndlerLogic) TestPayOrderHndler(req *types.TestPayOrderRequest) (resp *types.PayOrderResponse, err error) {
	var payOrderReq types.PayOrderRequestX
	copier.Copy(&payOrderReq, &req)
	payOrderReq.PayOrderRequest.AccessType = "1"
	payOrderReq.NotifyUrl = "http://172.16.204.115:8083/dior/merchant-api/test_merchant_pay-call-back"
	payOrderReq.PageUrl = ""
	payOrderReq.Language = "ZH-CN"
	payOrderReq.OrderNo = model.GenerateOrderNo("TEST")
	payOrderReq.OrderName = "TEST"
	payOrderReq.UserId = "测试员"
	payOrderReq.PayOrderRequest.OrderAmount = payOrderReq.OrderAmount.String()
	payOrderReq.PayOrderRequest.AccessType = payOrderReq.AccessType.String()
	payOrderReq.PayOrderRequest.PayTypeNo = payOrderReq.PayTypeNo.String()
	payOrderReq.Sign = utils.SortAndSign2(payOrderReq.PayOrderRequest, req.MerchantKey)
	payOrderReq.MyIp = req.IP
	payOrderReq.BankCode = req.BankCode

	pl := payorder.NewPayOrderLogic(l.ctx, l.svcCtx)
	return pl.PayOrder(payOrderReq)
}
