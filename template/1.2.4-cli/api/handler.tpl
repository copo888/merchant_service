package {{.PkgName}}

import (
	"net/http"
    "com.copo/bo_service/common/response"
    "com.copo/bo_service/common/utils"
    "encoding/json"
	{{if .After1_1_10}}"github.com/zeromicro/go-zero/rest/httpx"{{end}}
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
	{{.ImportPackages}}
)

func {{.HandlerName}}(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}

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

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), ctx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}req{{end}})
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			{{if .HasResp}}response.Json(w, r, response.SUCCESS, resp, err){{else}}response.Json(w, r, response.SUCCESS, nil, err){{end}}
		}
	}
}
