package payorder

import (
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/logic/payorder"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"encoding/json"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func PayOrderChannelBankHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PayOrderChannelBankRequest

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseJsonBody(r, &req); err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		if err := utils.MyValidator.Struct(req); err != nil {
			response.Json(w, r, response.INVALID_PARAMETER, nil, err)
			return
		}

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := payorder.NewPayOrderChannelBankLogic(r.Context(), ctx)
		resp, err := l.PayOrderChannelBank(&req)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
