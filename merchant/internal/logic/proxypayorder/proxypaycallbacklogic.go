package proxypayorder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/service/merchantbalanceservice"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/gioco-play/easy-i18n/i18n"
	"golang.org/x/text/language"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyPayCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayCallBackLogic {
	return ProxyPayCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayCallBackLogic) ProxyPayCallBack(req *types.ProxyPayOrderCallBackRequest) (resp *types.ProxyPayOrderCallBackResponse, err error) {
	logx.Infof("渠道回調請求參數: %#v", req)
	//检查单号是否存在
	orderX := &types.OrderX{}
	if req.ProxyPayOrderNo == "" && req.ChannelOrderNo == "" {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST)
	} else if orderX, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.ProxyPayOrderNo, ""); err != nil && orderX == nil {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, "Copo OrderNo: "+req.ProxyPayOrderNo)
	}

	ProxyPayOrderCallBackResp := &types.ProxyPayOrderCallBackResponse{}
	service := ordersService.NewOrdersService(l.ctx, l.svcCtx)
	if errCallBack := service.ChannelCallBackForProxy(req, orderX); errCallBack != nil {
		ProxyPayOrderCallBackResp.RespMsg = errCallBack.Error()
		return ProxyPayOrderCallBackResp, errorz.New(response.PROXY_PAY_CALLBACK_FAIL, errCallBack.Error())
	}

	i18n.SetLang(language.English)
	callBackResp := &types.ProxyPayOrderCallBackResponse{
		RespCode: response.API_SUCCESS,
		RespMsg:  i18n.Sprintf(response.API_SUCCESS),
	}
	return callBackResp, nil
}

/*
	渠道(代付)回调用，更新状态用(Service)
	目前仅提供 1.渠道回调  TODO 2.scheduled回调使用
	提单order_status为[1=成功][2=失败]时，不接受回调的变更
	TODO orderService裡面同樣此方法若測試成功要刪除
*/
func (l *ProxyPayCallBackLogic) channelCallBackForProxy(req *types.ProxyPayOrderCallBackRequest, orderX *types.OrderX) error {
	//订单状态为[成功]或[失败]，判定为已结单，以及[人工處理單]不接受回調
	if orderX.Status == constants.SUCCESS || orderX.Status == constants.FAIL || orderX.Status == constants.FROZEN {
		logx.Infof("代付订单：%s，提单状态为：%s，判定为已结单，不接受回调变更。", orderX.OrderNo, orderX.Status)
		//TODO 是否有需要回寫代付歷程?
		return errorz.New(response.PROXY_PAY_IS_CLOSE, "此提单目前已为结单状态")
	} else if orderX.PersonProcessStatus != constants.PERSON_PROCESS_STATUS_NO_ROCESSING {
		logx.Infof("代付订单：%s，人工处里状态为：%s，判定已进入人工处里阶段，不接受回调变更。", orderX.OrderNo, orderX.PersonProcessStatus)
		//TODO 是否有需要回寫代付歷程?
		return errorz.New(response.PROXY_PAY_IS_PERSON_PROCESS, "提单目前为人工处里阶段，不可回调变更")
	}
	/*更新訂單:
	1. 訂單狀態(依渠道回調決定)
	2. 還款狀態(預設[0]：不需还款)，若渠道回調失敗單，則[1]：待还款
	*/
	//订单预设还款状态为"不需还款"，更新為待还款
	orderX.Status = l.getOrderStatus(req.ChannelResultStatus)
	orderX.RepaymentStatus = constants.REPAYMENT_NOT //还款状态：([0]：不需还款、1：待还款、2：还款成功、3：还款失败)，预设不需还款
	if req.ChannelResultStatus == constants.CALL_BACK_STATUS_FAIL {
		orderX.RepaymentStatus = constants.REPAYMENT_WAIT //还款状态：(0：不需还款、[1]：待还款、2：还款成功、3：还款失败)，预设不需还款
		if req.ChannelResultNote != "" {
			orderX.ErrorNote = "渠道回调:-" + req.ChannelResultNote //失败时，写入渠道返还的讯息
		} else {
			orderX.ErrorNote = "渠道回调: 交易失败"
		}
	}
	//是否以回調給商戶
	if orderX.IsMerchantCallback == constants.MERCHANT_CALL_BACK_NO {
		orderX.IsMerchantCallback = constants.MERCHANT_CALL_BACK_YES
		orderX.MerchantCallBackAt = time.Now().UTC()
	}

	orderX.UpdatedBy = req.UpdatedBy
	orderX.UpdatedAt = time.Now().UTC()
	orderX.CallBackStatus = req.ChannelResultStatus
	orderX.ChannelCallBackAt = time.Now().UTC()

	if orderX.ChannelOrderNo == "" {
		orderX.ChannelOrderNo = req.ChannelOrderNo
	}
	if req.ChannelCharge != 0 {
		//渠道有回傳渠道手續費
	}

	// 更新订单
	if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(orderX).Error; errUpdate != nil {
		logx.Error("代付订单更新状态错误: ", errUpdate.Error())
	}

	var errRpc error
	if req.ChannelResultStatus == constants.CALL_BACK_STATUS_SUCCESS {
		logx.Infof("代付订单回调状态为[成功]，主动回调API订单：%s=======================================>", orderX.Order.OrderNo)
		//回調商戶
		if errPoseMer := merchantsService.PostCallbackToMerchant(l.svcCtx.MyDB, &l.ctx, orderX); errPoseMer != nil {
			//不拋錯
			logx.Error("回調商戶錯誤:", errPoseMer)
		}

	} else if req.ChannelResultStatus == constants.CALL_BACK_STATUS_FAIL { //===========還款=========START
		logx.Info("代付订单回调状态为[失败]，开始还款=======================================>", orderX.Order.OrderNo)
		//呼叫RPC
		rpc := transactionclient.NewTransaction(l.svcCtx.RpcService("transaction.rpc"))
		balanceType, errBalance := merchantbalanceservice.GetBalanceType(l.svcCtx.MyDB, orderX.ChannelCode, constants.ORDER_TYPE_DF)
		if errBalance != nil {
			return errBalance
		}

		//當訂單還款狀態為待还款
		if orderX.RepaymentStatus == constants.REPAYMENT_WAIT {
			//将商户钱包加回 (merchantCode, orderNO)，更新狀態為失敗單
			var resRpc *transaction.ProxyPayFailResponse
			if balanceType == "DFB" {
				resRpc, errRpc = rpc.ProxyOrderTransactionFail_DFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: orderX.MerchantCode,
					OrderNo:      orderX.OrderNo,
				})
			} else if balanceType == "XFB" {
				resRpc, errRpc = rpc.ProxyOrderTransactionFail_XFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: orderX.MerchantCode,
					OrderNo:      orderX.OrderNo,
				})
			}

			if errRpc != nil {
				logx.Errorf("代付提单回调 %s 还款失败。 Err: %s", orderX.OrderNo, errRpc.Error())
				orderX.RepaymentStatus = constants.REPAYMENT_FAIL
				return errorz.New(response.FAIL, errRpc.Error())
			} else {
				logx.Infof("代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
				orderX.RepaymentStatus = constants.REPAYMENT_SUCCESS
				//TODO 收支紀錄
			}

			// 更新订单
			if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(orderX).Error; errUpdate != nil {
				logx.Error("代付订单更新状态错误: ", errUpdate.Error())
			}
		}

		//若訂單單來源為API且尚未回調給商戶，進行回調給商戶
		if orderX.Source == constants.API && orderX.IsMerchantCallback == constants.IS_MERCHANT_CALLBACK_NO {
			logx.Infof("代付订单回调状态为[失敗]，主动回调API订单：%s=======================================>", orderX.OrderNo)
			if errPoseMer := merchantsService.PostCallbackToMerchant(l.svcCtx.MyDB, &l.ctx, orderX); errPoseMer != nil {
				//不拋錯
				logx.Error("回調商戶錯誤:", errPoseMer)
			}
		}

	} // ===========還款=========END

	return nil
}

func (l *ProxyPayCallBackLogic) getOrderStatus(channelResultStatus string) string {

	var orderStatus string
	switch channelResultStatus {
	case "0":
		orderStatus = constants.TRANSACTION
	case "1":
		orderStatus = constants.TRANSACTION
	case "2":
		orderStatus = constants.SUCCESS
	case "3":
		orderStatus = constants.FAIL
	default:
		orderStatus = constants.TRANSACTION
	}
	return orderStatus
}
