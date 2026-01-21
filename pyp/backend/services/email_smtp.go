package services

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"collision-backend/config"
	"collision-backend/models"

	"gorm.io/gorm"
)

// SMTPEmailService SMTP邮件服务（用于阿里企业邮箱）
type SMTPEmailService struct {
	SMTPHost  string
	SMTPPort  int
	Username  string // 发信地址
	Password  string // SMTP密码
	FromAlias string // 显示的发件人昵称
	ReplyTo   string // 回信地址
	DB        *gorm.DB
}

// NewSMTPEmailService 创建SMTP邮件服务实例
func NewSMTPEmailService(db *gorm.DB) *SMTPEmailService {
	cfg := config.GetConfig()
	return &SMTPEmailService{
		SMTPHost:  cfg.SMTPHost,
		SMTPPort:  cfg.SMTPPort,
		Username:  cfg.SMTPUsername,
		Password:  cfg.SMTPPassword,
		FromAlias: cfg.SMTPFromAlias,
		ReplyTo:   cfg.SMTPReplyTo,
		DB:        db,
	}
}

// SendEmail 发送邮件
func (s *SMTPEmailService) SendEmail(userID uint64, toEmail, subject, htmlBody string, emailType string) error {
	// 创建邮件记录
	emailLog := &models.EmailLog{
		UserID:  userID,
		ToEmail: toEmail,
		Subject: subject,
		Content: htmlBody,
		Type:    emailType,
		Status:  "pending",
	}
	s.DB.Create(emailLog)

	// 构建邮件内容
	msg := s.buildMessage(toEmail, subject, htmlBody, []string{}, []string{}, []string{})

	// 建立SMTP连接
	addr := fmt.Sprintf("%s:%d", s.SMTPHost, s.SMTPPort)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.SMTPHost)

	// 发送邮件
	receivers := []string{toEmail}
	fmt.Println("开始发送邮件", addr, auth, s.Username, receivers)

	var err error
	if s.SMTPPort == 465 {
		// 使用SSL连接，465端口是SSL加密端口
		// 创建SSL连接
		conn, err := tls.Dial("tcp", addr, &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         s.SMTPHost,
		})
		if err != nil {
			fmt.Println("SSL连接失败", err)
			emailLog.Status = "failed"
			emailLog.ErrorMsg = err.Error()
			s.DB.Save(emailLog)
			return fmt.Errorf("SSL连接失败: %v", err)
		}
		defer conn.Close()

		// 创建SMTP客户端
		client, err := smtp.NewClient(conn, s.SMTPHost)
		if err != nil {
			fmt.Println("创建SMTP客户端失败", err)
			emailLog.Status = "failed"
			emailLog.ErrorMsg = err.Error()
			s.DB.Save(emailLog)
			return fmt.Errorf("创建SMTP客户端失败: %v", err)
		}

		// 认证
		if err := client.Auth(auth); err != nil {
			fmt.Println("SMTP认证失败", err)
			emailLog.Status = "failed"
			emailLog.ErrorMsg = err.Error()
			s.DB.Save(emailLog)
			return fmt.Errorf("SMTP认证失败: %v", err)
		}

		// 设置发件人
		if err := client.Mail(s.Username); err != nil {
			fmt.Println("设置发件人失败", err)
			emailLog.Status = "failed"
			emailLog.ErrorMsg = err.Error()
			s.DB.Save(emailLog)
			return fmt.Errorf("设置发件人失败: %v", err)
		}

		// 设置收件人
		for _, rec := range receivers {
			if err := client.Rcpt(rec); err != nil {
				fmt.Println("设置收件人失败", err)
				emailLog.Status = "failed"
				emailLog.ErrorMsg = err.Error()
				s.DB.Save(emailLog)
				return fmt.Errorf("设置收件人失败: %v", err)
			}
		}

		// 发送邮件内容
		wc, err := client.Data()
		if err != nil {
			fmt.Println("获取邮件数据写入器失败", err)
			emailLog.Status = "failed"
			emailLog.ErrorMsg = err.Error()
			s.DB.Save(emailLog)
			return fmt.Errorf("获取邮件数据写入器失败: %v", err)
		}

		_, err = wc.Write([]byte(msg))
		if err != nil {
			fmt.Println("写入邮件内容失败", err)
			emailLog.Status = "failed"
			emailLog.ErrorMsg = err.Error()
			s.DB.Save(emailLog)
			return fmt.Errorf("写入邮件内容失败: %v", err)
		}

		err = wc.Close()
		if err != nil {
			fmt.Println("关闭邮件数据写入器失败", err)
			emailLog.Status = "failed"
			emailLog.ErrorMsg = err.Error()
			s.DB.Save(emailLog)
			return fmt.Errorf("关闭邮件数据写入器失败: %v", err)
		}

		err = client.Quit()
		if err != nil {
			fmt.Println("关闭SMTP客户端失败", err)
			// 不返回错误，因为邮件已经发送成功
		}
	} else {
		// 使用普通连接，非SSL端口
		err = smtp.SendMail(addr, auth, s.Username, receivers, []byte(msg))
	}

	if err != nil {
		fmt.Println("发送邮件失败", err)
		emailLog.Status = "failed"
		emailLog.ErrorMsg = err.Error()
		s.DB.Save(emailLog)
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	// 更新邮件记录为已发送
	now := time.Now()
	// 更新邮件记录
	s.DB.Model(emailLog).Updates(map[string]interface{}{
		"status":    "sent",
		"sent_at":   now,
		"error_msg": "",
	})

	// 尝试关联碰撞匹配记录
	// 提取关键词
	keyword := extractKeywordFromSubject(subject)
	if keyword != "" {
		// 查询匹配记录
		var collisionResult models.CollisionResult
		s.DB.Where("user_id = ? AND keyword = ? AND matched_at > ?", userID, keyword, now.Add(-5*time.Minute)).
			First(&collisionResult)
		if collisionResult.ID > 0 {
			// 更新碰撞结果记录，标记邮件已发送
			s.DB.Model(&collisionResult).Updates(map[string]interface{}{
				"email_sent":    true,
				"email_sent_at": now,
			})
		}
	}

	return nil
}

// extractKeywordFromSubject 从邮件主题中提取关键词
func extractKeywordFromSubject(subject string) string {
	// 主题格式: "标签碰撞 - 您有新的碰撞匹配 [关键词]"
	start := strings.Index(subject, "[")
	end := strings.Index(subject, "]")
	if start >= 0 && end > start {
		return subject[start+1 : end]
	}
	return ""
}

// SendEmailWithCC 发送邮件（包含抄送和密送）
// toEmails: 收件人列表
// ccEmails: 抄送人列表
// bccEmails: 密送人列表
func (s *SMTPEmailService) SendEmailWithCC(userID uint64, subject, htmlBody string,
	toEmails []string, ccEmails []string, bccEmails []string, emailType string) error {

	if len(toEmails) == 0 {
		return fmt.Errorf("收件人列表不能为空")
	}

	// 创建邮件记录
	emailLog := &models.EmailLog{
		UserID:  userID,
		ToEmail: strings.Join(toEmails, ","),
		Subject: subject,
		Content: htmlBody,
		Type:    emailType,
		Status:  "pending",
	}
	s.DB.Create(emailLog)

	// 构建邮件内容
	msg := s.buildMessage(strings.Join(toEmails, ","), subject, htmlBody, ccEmails, bccEmails, toEmails)

	// 建立SMTP连接
	addr := fmt.Sprintf("%s:%d", s.SMTPHost, s.SMTPPort)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.SMTPHost)

	// 合并所有收件人（包括To、Cc、Bcc）
	receivers := append(toEmails, append(ccEmails, bccEmails...)...)
	fmt.Println("开始发送邮件")
	// 发送邮件
	err := smtp.SendMail(addr, auth, s.Username, receivers, []byte(msg))

	if err != nil {
		emailLog.Status = "failed"
		emailLog.ErrorMsg = err.Error()
		s.DB.Save(emailLog)
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	// 更新邮件记录为已发送
	now := time.Now()
	s.DB.Model(emailLog).Updates(map[string]interface{}{
		"status":    "sent",
		"sent_at":   now,
		"error_msg": "",
	})

	return nil
}

// buildMessage 构建MIME格式的邮件内容
func (s *SMTPEmailService) buildMessage(toAddresses, subject, htmlBody string,
	ccAddresses []string, bccAddresses []string, actualToAddresses []string) string {

	// 如果没有实际的To地址，使用传入的toAddresses
	if len(actualToAddresses) == 0 {
		actualToAddresses = strings.Split(toAddresses, ",")
	}

	// 构建邮件头
	headers := make(map[string]string)
	headers["Subject"] = subject

	// 构建From字段
	fromAddr := mail.Address{
		Name:    s.FromAlias,
		Address: s.Username,
	}
	headers["From"] = fromAddr.String()
	headers["To"] = toAddresses

	if len(ccAddresses) > 0 {
		headers["Cc"] = strings.Join(ccAddresses, ",")
	}

	// Reply-To 和 Return-Path
	// 如果没有设置 Reply-To，则使用发件人地址
	replyTo := s.ReplyTo
	if replyTo == "" {
		replyTo = s.Username
	}
	headers["Reply-To"] = replyTo
	headers["Return-Path"] = s.Username
	headers["Message-ID"] = fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), s.SMTPHost)
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""
	headers["Content-Transfer-Encoding"] = "base64"

	// 构建邮件体
	var msg strings.Builder

	// 添加邮件头
	for key, value := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	msg.WriteString("\r\n")

	// 添加邮件内容（Base64编码）
	encoded := base64.StdEncoding.EncodeToString([]byte(htmlBody))

	// Base64换行处理（RFC 2045要求每76个字符换行）
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		msg.WriteString(encoded[i:end])
		msg.WriteString("\r\n")
	}

	return msg.String()
}

// SendVerifyEmail 发送验证邮件
func (s *SMTPEmailService) SendVerifyEmail(userID uint64, toEmail, code string) error {
	subject := "邮箱验证码"
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body style="font-family: Arial, sans-serif; background-color: #f5f5f5; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
				<h2 style="color: #333; text-align: center;">邮箱验证</h2>
				<p style="font-size: 14px; color: #666;">亲爱的用户，</p>
				<p style="font-size: 14px; color: #666;">您的邮箱验证码是：</p>
				<div style="text-align: center; margin: 30px 0;">
					<span style="font-size: 32px; font-weight: bold; color: #1890ff; letter-spacing: 4px;">%s</span>
				</div>
				<p style="font-size: 12px; color: #999;">验证码有效期为10分钟，请勿分享给他人。</p>
				<p style="font-size: 12px; color: #999;">如果您没有进行此操作，请忽略此邮件。</p>
				<hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
				<p style="font-size: 12px; color: #999; text-align: center;">此邮件由系统自动发送，请勿直接回复</p>
			</div>
		</body>
		</html>
	`, code)
	fmt.Println("发送邮箱验证码邮件内容:", htmlBody, code)
	return s.SendEmail(userID, toEmail, subject, htmlBody, "verify")
}

// SendCollisionNotifyEmail 发送碰撞匹配通知邮件（单收件人）
func (s *SMTPEmailService) SendCollisionNotifyEmail(userID uint64, toEmail, matcherName, matcherEmail string) error {
	subject := "您有一个新的碰撞匹配"
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body style="font-family: Arial, sans-serif; background-color: #f5f5f5; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
				<h2 style="color: #333; text-align: center;">🎉 新的碰撞匹配</h2>
				<p style="font-size: 14px; color: #666;">亲爱的用户，</p>
				<p style="font-size: 14px; color: #666;">恭喜！您有一个新的碰撞匹配。</p>
				<div style="background-color: #f0f9ff; border-left: 4px solid #1890ff; padding: 15px; margin: 20px 0; border-radius: 4px;">
					<p style="margin: 10px 0; color: #333;">
						<strong>匹配用户：</strong> %s
					</p>
					<p style="margin: 10px 0; color: #333;">
						<strong>邮箱：</strong> %s
					</p>
				</div>
				<p style="font-size: 14px; color: #666;">请登录应用查看更多详情。</p>
				<hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
				<p style="font-size: 12px; color: #999; text-align: center;">此邮件由系统自动发送，请勿直接回复</p>
			</div>
		</body>
		</html>
	`, matcherName, matcherEmail)
	return s.SendEmail(userID, toEmail, subject, htmlBody, "collision")
}

// SendCollisionNotifyEmailWithPartner 发送碰撞匹配通知邮件（包含双方邮箱）
func (s *SMTPEmailService) SendCollisionNotifyEmailWithPartner(userID uint64, toEmail, partnerEmail, matcherName string) error {
	subject := "您有一个新的碰撞匹配"
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body style="font-family: Arial, sans-serif; background-color: #f5f5f5; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
				<h2 style="color: #333; text-align: center;">🎉 新的碰撞匹配</h2>
				<p style="font-size: 14px; color: #666;">亲爱的用户，</p>
				<p style="font-size: 14px; color: #666;">恭喜！您有一个新的碰撞匹配。</p>
				<div style="background-color: #f0f9ff; border-left: 4px solid #1890ff; padding: 15px; margin: 20px 0; border-radius: 4px;">
					<p style="margin: 10px 0; color: #333;">
						<strong>匹配用户：</strong> %s
					</p>
					<p style="margin: 10px 0; color: #333;">
						<strong>您的邮箱：</strong> %s
					</p>
					<p style="margin: 10px 0; color: #333;">
						<strong>对方邮箱：</strong> %s
					</p>
				</div>
				<p style="font-size: 14px; color: #666;">请登录应用查看更多详情。</p>
				<hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
				<p style="font-size: 12px; color: #999; text-align: center;">此邮件由系统自动发送，请勿直接回复</p>
			</div>
		</body>
		</html>
	`, matcherName, toEmail, partnerEmail)

	return s.SendEmail(userID, toEmail, subject, htmlBody, "collision")
}

// SendCollisionNotifyEmailWithPartner 重载版本：支持 Aliyun API 兼容的签名
// SendCollisionNotifyEmailWithPartnerCompat 发送碰撞匹配通知邮件(Aliyun 兼容版本)
func (s *SMTPEmailService) SendCollisionNotifyEmailWithPartnerCompat(userID uint64, toEmail, keyword string, matchCount int, partnerEmail string) error {
	subject := fmt.Sprintf("标签碰撞 - 您有新的碰撞匹配 [%s]", keyword)

	// 如果有对方邮箱，显示对方邮箱信息
	partnerInfo := ""
	if partnerEmail != "" {
		partnerInfo = fmt.Sprintf(`
            <div style="background: #fff3cd; padding: 20px; border-radius: 8px; margin-top: 20px; border-left: 4px solid #ffc107;">
                <p style="margin: 0; color: #856404; font-weight: bold;">📧 对方邮箱</p>
                <p style="margin: 10px 0 0 0; font-size: 20px; color: #333;">
                    <a href="mailto:%s" style="color: #667eea; text-decoration: none;">%s</a>
                </p>
                <p style="margin: 10px 0 0 0; color: #666; font-size: 14px;">点击邮箱可直接发送邮件联系对方~</p>
            </div>
`, partnerEmail, partnerEmail)
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #f5f7fa; padding: 40px 0; }
        .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 40px; text-align: center; }
        .header h1 { color: #fff; margin: 0; font-size: 28px; }
        .content { padding: 40px; }
        .highlight { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: #fff; padding: 30px; border-radius: 12px; text-align: center; margin: 30px 0; }
        .keyword { font-size: 32px; font-weight: bold; margin-bottom: 10px; }
        .count { font-size: 18px; opacity: 0.9; }
        .btn { display: inline-block; background: #667eea; color: #fff; padding: 15px 40px; border-radius: 30px; text-decoration: none; font-weight: bold; margin-top: 20px; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 碰撞成功</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>您的碰撞关键词有了新的匹配结果：</p>
            <div class="highlight">
                <div class="keyword">%s</div>
                <div class="count">碰撞到 %d 个新结果</div>
            </div>
            %s
            <p style="text-align: center; margin-top: 30px;">
                打开小程序查看更多详细匹配结果
            </p>
        </div>
        <div class="footer">
            标签碰撞 © 2024 All rights reserved.
        </div>
    </div>
</body>
</html>
`, keyword, matchCount, partnerInfo)

	return s.SendEmail(userID, toEmail, subject, htmlBody, "collision")
}
