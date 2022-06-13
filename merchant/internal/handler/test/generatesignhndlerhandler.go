package test

import (
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/logic/test"
	"com.copo/bo_service/merchant/internal/svc"
	"encoding/json"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
)

func GenerateSignHndlerHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		bodyBytes, err := io.ReadAll(r.Body)
		if  err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		l := test.NewGenerateSignHndlerLogic(r.Context(), ctx)
		resp, err := l.GenerateSignHndler(req)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
