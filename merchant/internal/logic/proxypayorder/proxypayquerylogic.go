package proxypayorder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"errors"
	"github.com/gioco-play/easy-i18n/i18n"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyPayQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayQueryLogic {
	return ProxyPayQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayQueryLogic) ProxyPayQuery(merReq *types.ProxyPayOrderQueryRequestX) (*types.ProxyPayOrderQueryResponse, error) {
	logx.Info("Enter proxy-query:", merReq)
	resp := &types.ProxyPayOrderQueryResponse{}
	// 1. 檢查白名單、商户号，單號是否存在
	merchantKey, txOrder, errWhite := l.CheckMerAndWhiteList(merReq)
	if errWhite != nil {
		logx.Error("商戶號及白名單檢查錯誤: ", i18n.Sprintf(errWhite.Error()))
		resp.RespCode = errWhite.Error()
		resp.RespMsg = i18n.Sprintf(errWhite.Error())
		return resp, errWhite
	}

	// 檢查簽名
	checkSign := utils.VerifySign(merReq.Sign, merReq.ProxyPayOrderQueryRequest, merchantKey)
	if !checkSign {
		return nil, errorz.New(response.INVALID_SIGN)
	}

	// 1. 查copo DB 訂單狀態
	//var resp = types.ProxyPayOrderQueryResponse{} //返回商戶查單物件
	var status string
	if txOrder.Status == "0" {
		status = "0"
	} else if txOrder.Status == "1" { //(0:待處理 1:處理中 2:交易中 20:成功 30:失敗 31:凍結)
		status = "4"
	} else if txOrder.Status == "2" {
		status = "3"
	} else if txOrder.Status == "20" {
		status = "1"
	} else if txOrder.Status == "30" || txOrder.Status == "31" {
		status = "2"
	}

	//返回給商戶查詢物件
	i18n.SetLang(language.English)
	resp.RespCode = response.API_SUCCESS
	resp.RespMsg = i18n.Sprintf(response.API_SUCCESS) //固定回商戶成功
	resp.MerchantId = merReq.MerchantId
	resp.OrderNo = txOrder.MerchantOrderNo
	resp.LastUpdateTime = txOrder.UpdatedAt.Format("2006-01-02 15:04:05")
	resp.PayOrderNo = txOrder.OrderNo
	resp.OrderStatus = status //0.待处理、1.成功、2.失败、3:交易中、4.处理中
	resp.CallbackStatus = txOrder.IsMerchantCallback
	resp.OrderAmount = strconv.FormatFloat(txOrder.OrderAmount, 'f', 2, 64)
	resp.Fee = strconv.FormatFloat(txOrder.TransferHandlingFee, 'f', 2, 64)
	resp.Sign = utils.SortAndSign2(resp, merchantKey)

	return resp, nil

	// 2. call 渠道
	//依單號取得渠道網關地址
	//var apiUrl string
	//var err error
	//if apiUrl, err = l.GetChannelApiUrlByOrderNo(merReq.OrderNo); err != nil {
	//	logx.Error("取渠道網官網錯誤: " + err.Error())
	//	return nil, err
	//}
	//url := 	apiUrl + "/api/proxy-pay-query"
	//
	//var errCHN error
	//ProxyQueryRespVO := &vo.ProxyQueryRespVO{}
	//ProxyQueryRespVO, errCHN = ordersService.CallChannel_ProxyQuery(&l.ctx, &l.svcCtx.Config, url, txOrder.OrderNo)
	//

	//if errCHN != nil {
	//	logx.Errorf("代付提單: %s ，渠道返回錯誤: %s, %#v", merReq.OrderNo, errCHN.Error(), &ProxyQueryRespVO)
	//	resp.RespCode = response.CHANNEL_REPLY_ERROR
	//	resp.RespMsg = i18n.Sprintf(response.CHANNEL_REPLY_ERROR) + ": Code: " + ProxyQueryRespVO.Code + " Message: " + ProxyQueryRespVO.Message
	//}

}

//检查商户号是否存在以及IP是否为白名单，
//若无误则返回"商户密鑰"、"copo訂單號"
func (l *ProxyPayQueryLogic) CheckMerAndWhiteList(req *types.ProxyPayOrderQueryRequestX) (merchantKey string, orderNo *types.OrderX, err error) {
	merchant := &types.Merchant{}
	// 1.檢查白名單
	if err = l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", req.MerchantId).Take(merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errorz.New(response.DATA_NOT_FOUND, err.Error())
		} else if err == nil && merchant != nil && merchant.Status != constants.MerchantStatusEnable {
			return "", nil, errorz.New(response.MERCHANT_ACCOUNT_NOT_FOUND, "商户号:"+merchant.Code)
		} else {
			return "", nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	if isWhite := merchantsService.IPChecker(req.Ip, merchant.ApiIP); !isWhite {
		return "", nil, errorz.New(response.API_IP_DENIED, "IP: "+req.Ip)
	}

	//2.检查订单号是否存在
	order := &types.OrderX{}
	if order, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, "", req.OrderNo); err != nil || order == nil {
		return "", nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, "Merchant OrderNo: "+req.OrderNo)
	}

	return merchant.ScrectKey, order, nil
}

/*
	@param merOrderNo: 商戶訂單編號
*/
func (l *ProxyPayQueryLogic) GetChannelApiUrlByOrderNo(merOrderNo string) (apiUrl string, err error) {

	SelectColumn := &SelectX{}

	selectX := "c.api_url as api_url," +
		"tx.order_no order_no"

	if err = l.svcCtx.MyDB.Table("tx_orders tx ").Select(selectX).
		Joins("left join ch_channels c on tx.channel_code = c.code ").
		Where("tx.merchant_order_no = ?", merOrderNo).Find(SelectColumn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errorz.New(response.DATA_NOT_FOUND)
		}
		return "", errorz.New(response.DATABASE_FAILURE)
	}
	apiUrl = SelectColumn.ApiUrl

	return
}

type SelectX struct {
	ApiUrl  string `json:"api_url"`
	OrderNo string `json:"order_no"`
}
