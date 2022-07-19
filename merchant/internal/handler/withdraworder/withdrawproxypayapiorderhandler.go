package withdraworder

import (
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/logic/withdraworder"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"encoding/json"
	"github.com/thinkeridea/go-extend/exnet"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
)

func WithdrawProxyPayApiOrderHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProxyPayRequestX

		span := trace.SpanFromContext(r.Context())
		defer span.End()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		myIP := exnet.ClientIP(r)
		req.Ip = myIP

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

		l := withdraworder.NewWithdrawProxyPayApiOrderLogic(r.Context(), ctx)
		resp, err := l.WithdrawProxyPayApiOrder(&req)
		if err != nil {
			response.ApiErrorJson(w, r, err.Error(), err)
		} else {
			response.ApiJson(w, r, resp)
		}
	}
}
