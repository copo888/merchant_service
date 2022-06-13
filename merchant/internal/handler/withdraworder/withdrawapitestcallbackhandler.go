package withdraworder

import (
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/logic/withdraworder"
	"com.copo/bo_service/merchant/internal/svc"
	"net/http"
)

func WithdrawApiTestCallBackHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := withdraworder.NewWithdrawApiTestCallBackLogic(r.Context(), ctx)
		resp, err := l.WithdrawApiTestCallBack()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
