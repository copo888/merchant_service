package test

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"errors"
	"gorm.io/gorm"
	"strconv"

	"com.copo/bo_service/merchant/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateSignHndlerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateSignHndlerLogic(ctx context.Context, svcCtx *svc.ServiceContext) GenerateSignHndlerLogic {
	return GenerateSignHndlerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateSignHndlerLogic) GenerateSignHndler(req map[string]interface{}) (resp string, err error) {
	var merchant *types.Merchant
	m := make(map[string]string)

	for key, val := range req {
		if key == "sign" {
			continue
		}
		if f, ok := val.(float64); ok {
			valTrans := strconv.FormatFloat(f, 'f', 2, 64)
			m[key] = valTrans
		} else if s, ok := val.(string); ok {
			m[key] = s
		}
	}

	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req["merchantId"]).
		Where("status = ?", "1").
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return "", errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	//delete(req, "sign")

	source := utils.JoinStringsInASCII(m, "&", false, false, merchant.ScrectKey)
	resp = utils.GetSign(source)
	return
}
