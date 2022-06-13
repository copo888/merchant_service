package proxypayorder

import (
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/logic/proxypayorder"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"encoding/json"
	"github.com/thinkeridea/go-extend/exnet"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func ProxyPayQueryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProxyPayOrderQueryRequestX

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseJsonBody(r, &req); err != nil {
			logx.Error("Parse Error:", err.Error())
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		if err := utils.MyValidator.Struct(req); err != nil {
			logx.Error("Validatation Error: ", err.Error())
			response.Json(w, r, response.INVALID_PARAMETER, nil, err)
			return
		}
		req.Ip = exnet.ClientIP(r)

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := proxypayorder.NewProxyPayQueryLogic(r.Context(), ctx)
		resp, err := l.ProxyPayQuery(&req)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
