package services

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"collision-backend/config"
	"collision-backend/models"

	"gorm.io/gorm"
)

// AliyunEmailService 阿里云邮件服务
type AliyunEmailService struct {
	AccessKeyID     string
	AccessKeySecret string
	AccountName     string
	FromAlias       string
	Region          string
	DB              *gorm.DB
}

// NewAliyunEmailService 创建阿里云邮件服务实例
func NewAliyunEmailService(db *gorm.DB) *AliyunEmailService {
	cfg := config.GetConfig()
	return &AliyunEmailService{
		AccessKeyID:     cfg.AliyunDMAccessKey,
		AccessKeySecret: cfg.AliyunDMAccessSecret,
		AccountName:     cfg.AliyunDMAccount,
		FromAlias:       cfg.AliyunDMAccountName,
		Region:          cfg.AliyunDMRegion,
		DB:              db,
	}
}

// SendEmail 发送邮件
func (s *AliyunEmailService) SendEmail(userID uint64, toEmail, subject, htmlBody string, emailType string) error {
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

	// 构建请求参数
	params := map[string]string{
		"Action":           "SingleSendMail",
		"AccountName":      s.AccountName,
		"AddressType":      "1",
		"FromAlias":        s.FromAlias,
		"ReplyToAddress":   "true",
		"ToAddress":        toEmail,
		"Subject":          subject,
		"HtmlBody":         htmlBody,
		"Format":           "JSON",
		"Version":          "2015-11-23",
		"AccessKeyId":      s.AccessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"RegionId":         s.Region,
	}

	// 计算签名
	signature := s.computeSignature(params)
	params["Signature"] = signature

	// 构建请求URL
	endpoint := fmt.Sprintf("https://dm.%s.aliyuncs.com/", s.Region)

	// 发送请求
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	resp, err := http.PostForm(endpoint, values)
	if err != nil {
		emailLog.Status = "failed"
		emailLog.ErrorMsg = err.Error()
		s.DB.Save(emailLog)
		return fmt.Errorf("发送邮件失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// 检查响应
	if resp.StatusCode != 200 {
		emailLog.Status = "failed"
		emailLog.ErrorMsg = string(body)
		s.DB.Save(emailLog)
		return fmt.Errorf("发送邮件失败: %s", string(body))
	}

	// 解析响应
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if _, ok := result["EnvId"]; ok {
		// 发送成功
		now := time.Now()
		emailLog.Status = "sent"
		emailLog.SentAt = &now
		s.DB.Save(emailLog)
		return nil
	}

	// 发送失败
	emailLog.Status = "failed"
	emailLog.ErrorMsg = string(body)
	s.DB.Save(emailLog)
	return fmt.Errorf("发送邮件失败: %s", string(body))
}

// computeSignature 计算阿里云签名
func (s *AliyunEmailService) computeSignature(params map[string]string) string {
	// 排序参数
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建待签名字符串
	var canonicalizedQueryString string
	for _, k := range keys {
		canonicalizedQueryString += "&" + percentEncode(k) + "=" + percentEncode(params[k])
	}
	canonicalizedQueryString = canonicalizedQueryString[1:]

	stringToSign := "POST&%2F&" + percentEncode(canonicalizedQueryString)

	// HMAC-SHA1签名
	mac := hmac.New(sha1.New, []byte(s.AccessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}

// percentEncode URL编码
func percentEncode(s string) string {
	s = url.QueryEscape(s)
	s = strings.ReplaceAll(s, "+", "%20")
	s = strings.ReplaceAll(s, "*", "%2A")
	s = strings.ReplaceAll(s, "%7E", "~")
	return s
}

// SendVerifyEmail 发送验证码邮件
func (s *AliyunEmailService) SendVerifyEmail(userID uint64, toEmail, code string) error {
	subject := "标签碰撞 - 邮箱验证码"
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
        .code-box { background: #f8f9fa; border-radius: 8px; padding: 30px; text-align: center; margin: 30px 0; }
        .code { font-size: 42px; font-weight: bold; color: #667eea; letter-spacing: 8px; }
        .tip { color: #666; font-size: 14px; line-height: 1.8; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📧 邮箱验证</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>您正在绑定邮箱，验证码如下：</p>
            <div class="code-box">
                <div class="code">%s</div>
            </div>
            <p class="tip">
                • 验证码有效期为10分钟<br>
                • 如非本人操作，请忽略此邮件<br>
                • 请勿将验证码告知他人
            </p>
        </div>
        <div class="footer">
            标签碰撞 © 2024 All rights reserved.
        </div>
    </div>
</body>
</html>
`, code)

	return s.SendEmail(userID, toEmail, subject, htmlBody, "verify")
}

// SendCollisionNotifyEmail 发送碰撞匹配通知邮件
func (s *AliyunEmailService) SendCollisionNotifyEmail(userID uint64, toEmail, keyword string, matchCount int) error {
	return s.SendCollisionNotifyEmailWithPartner(userID, toEmail, keyword, matchCount, "")
}

// SendCollisionNotifyEmailWithPartner 发送碰撞匹配通知邮件(包含对方邮箱)
func (s *AliyunEmailService) SendCollisionNotifyEmailWithPartner(userID uint64, toEmail, keyword string, matchCount int, partnerEmail string) error {
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
