package response

import (
	"com.copo/bo_service/common/errorz"
	_ "com.copo/bo_service/locales"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/text/language"
	"net/http"
)

type ApiErrorBody struct {
	RespCode string `json:"respCode"`
	RespMsg  string `json:"respMsg"`
	Status   int64  `json:"status"`
	Trace    string `json:"trace"`
}

func ApiErrorJson(w http.ResponseWriter, r *http.Request, code string, err error) {
	var body ApiErrorBody

	span := trace.SpanFromContext(r.Context())

	i18n.SetLang(language.English)

	body.RespCode = code
	body.RespMsg = i18n.Sprintf(code)
	if err != nil {
		body.Status = 1
		if v, ok := err.(*errorz.Err); ok && v.GetMessage() != "" {
			body.RespMsg += ": " + v.GetMessage()
			span.RecordError(errors.New(fmt.Sprintf("(%s)%s", code, v.GetMessage())))
		} else {
			span.RecordError(errors.New(fmt.Sprintf("(%s)%s %s", code, body.RespMsg, err.Error())))
		}
	}
	body.Trace = span.SpanContext().TraceID().String()

	if responseBytes, err := json.Marshal(body); err == nil {
		span.SetAttributes(attribute.KeyValue{
			Key:   "response",
			Value: attribute.StringValue(string(responseBytes)),
		})
	}

	httpx.OkJson(w, body)
}

func ApiJson(w http.ResponseWriter, r *http.Request, resp interface{}) {

	span := trace.SpanFromContext(r.Context())

	if responseBytes, err := json.Marshal(resp); err == nil {
		span.SetAttributes(attribute.KeyValue{
			Key:   "response",
			Value: attribute.StringValue(string(responseBytes)),
		})
	}

	httpx.OkJson(w, resp)
}
