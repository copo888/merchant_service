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

type Body struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Trace   string      `json:"trace"`
}

func Json(w http.ResponseWriter, r *http.Request, code string, resp interface{}, err error) {
	var body Body

	span := trace.SpanFromContext(r.Context())

	languageX := r.Header.Get("Accept-Language")
	if len(languageX) > 0 {
		i18n.SetLang(language.Make(languageX))
	} else {
		i18n.SetLang(language.English)
	}

	body.Code = code
	body.Message = i18n.Sprintf(code)
	if err != nil {
		if v, ok := err.(*errorz.Err); ok && v.GetMessage() != "" {
			span.RecordError(errors.New(fmt.Sprintf("(%s)%s", code, v.GetMessage())))
		} else {
			span.RecordError(errors.New(fmt.Sprintf("(%s)%s %s", code, body.Message, err.Error())))
		}
	} else {
		body.Data = resp
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

type State struct {
	Code string
	Desc string
}
