package services

import (
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
	"skis-admin-backend/global"
)

func SendEmail(to, subject, body string) error {
	// 创建邮件对象
	m := gomail.NewMessage()

	// 设置发件人
	m.SetHeader("From", global.Config.Email.From)

	// 设置收件人
	m.SetHeader("To", to)

	// 设置主题
	m.SetHeader("Subject", subject)

	// 设置邮件正文
	m.SetBody("text/html", body)

	// 发送邮件
	d := gomail.NewDialer(global.Config.Email.Host, global.Config.Email.Port, global.Config.Email.UserName, global.Config.Email.Password)
	if err := d.DialAndSend(m); err != nil {
		global.Lg.Error("发送邮件失败", zap.Error(err))
		return err
	}

	return nil
}
