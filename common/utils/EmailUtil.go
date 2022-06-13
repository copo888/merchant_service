package utils

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/gomail.v2"
)

func SendEmail(mailService *gomail.Dialer, from, email, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", "copoepay@copoonline.com")
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	if err := mailService.DialAndSend(msg); err != nil {
		logx.Error(">>>>>>>", err)
		return errorz.New(response.SEND_MAIL_FAIL, err.Error())
	}
	return nil
}

