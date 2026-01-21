package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// Aliyun DirectMail SMTP 简单封装
// 配置请通过环境变量或配置文件注入

type SMTPCfg struct {
	Host     string // e.g. smtpdm.aliyun.com
	Port     int    // 25/465/587
	Username string // 邮箱账号
	Password string // SMTP 密码
	From     string // 发件人姓名和地址 e.g. "标签碰撞 <no-reply@example.com>"
}

type EmailService struct {
	cfg SMTPCfg
}

func NewEmailService(cfg SMTPCfg) *EmailService {
	return &EmailService{cfg: cfg}
}

// send plain text email
func (s *EmailService) SendMail(to []string, subject, body string) error {
	host := s.cfg.Host
	addr := fmt.Sprintf("%s:%d", host, s.cfg.Port)

	header := make(map[string]string)
	header["From"] = s.cfg.From
	header["To"] = strings.Join(to, ",")
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=UTF-8"
	header["Date"] = time.Now().Format(time.RFC1123Z)

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, host)

	// 如果使用 TLS 端口(465)，需要建立 TLS 连接
	if s.cfg.Port == 465 {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         host,
		}

		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return fmt.Errorf("tls dial error: %w", err)
		}

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp new client error: %w", err)
		}

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth error: %w", err)
		}

		if err = client.Mail(s.cfg.Username); err != nil {
			return fmt.Errorf("mail from error: %w", err)
		}

		for _, rec := range to {
			if err = client.Rcpt(rec); err != nil {
				return fmt.Errorf("rcpt to error: %w", err)
			}
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("data error: %w", err)
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("write message error: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("close writer error: %w", err)
		}

		client.Quit()
		return nil
	}

	// 非 TLS 端口, 直接使用 smtp.SendMail
	if err := smtp.SendMail(addr, auth, s.cfg.Username, to, []byte(message)); err != nil {
		return fmt.Errorf("sendmail error: %w", err)
	}

	return nil
}
