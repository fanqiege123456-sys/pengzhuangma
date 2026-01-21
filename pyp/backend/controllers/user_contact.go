package controllers

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"collision-backend/config"
	"collision-backend/models"
	"collision-backend/services"

	"github.com/gin-gonic/gin"
)

// GetUserContacts 获取用户联系方式
func GetUserContacts(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var contact models.UserContact
	if err := config.DB.Where("user_id = ?", userID).First(&contact).Error; err != nil {
		// 不存在则返回空
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"email":          contact.Email,
			"email_verified": contact.EmailVerified,
			"email_visible":  contact.EmailVisible,
			"phone":          contact.Phone,
			"phone_verified": contact.PhoneVerified,
		},
	})
}

// UpdateEmailVisibility 更新邮箱显示设置
func UpdateEmailVisibility(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		EmailVisible *bool `json:"email_visible" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var contact models.UserContact
	if err := config.DB.Where("user_id = ?", userID).First(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请先绑定邮箱"})
		return
	}

	// 使用 map 更新确保 bool 零值能正确更新
	config.DB.Model(&contact).Updates(map[string]interface{}{
		"email_visible": *req.EmailVisible,
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设置成功",
		"data": gin.H{
			"email_visible": *req.EmailVisible,
		},
	})
}

// BindEmail 绑定邮箱(发送验证码)
func BindEmail(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "邮箱格式不正确"})
		return
	}

	// 生成6位随机验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 发送验证码邮件
	emailService := services.NewSMTPEmailService(config.DB)
	err := emailService.SendVerifyEmail(uint64(userID), req.Email, code)
	if err != nil {
		fmt.Println("发送邮件失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证码发送失败"})
		return
	}

	// 记录验证码到数据库，作为Redis的备选方案
	var contact models.UserContact
	// 计算过期时间
	expireTime := time.Now().Add(10 * time.Minute)
	if err := config.DB.Where("user_id = ?", userID).First(&contact).Error; err != nil {
		// 创建新记录
		contact = models.UserContact{
			UserID:            uint64(userID),
			Email:             req.Email,
			EmailVerifyCode:   code,
			EmailVerifyExpire: &expireTime,
			EmailVerified:     false,
			EmailVisible:      true,
		}
		config.DB.Create(&contact)
	} else {
		// 更新现有记录
		config.DB.Model(&contact).Updates(map[string]interface{}{
			"email":               req.Email,
			"email_verify_code":   code,
			"email_verify_expire": &expireTime,
			"email_verified":      false,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "验证码已发送",
	})
}

// VerifyEmail 验证邮箱
func VerifyEmail(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 从数据库获取验证码
	var contact models.UserContact
	if err := config.DB.Where("user_id = ?", userID).First(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码已过期或不存在"})
		return
	}

	// 检查验证码是否匹配
	if contact.Email != req.Email || contact.EmailVerifyCode != req.Code {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码错误"})
		return
	}

	// 检查验证码是否过期
	if contact.EmailVerifyExpire == nil || contact.EmailVerifyExpire.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码已过期"})
		return
	}

	// 验证成功，更新用户状态
	config.DB.Model(&contact).Updates(map[string]interface{}{
		"email":               req.Email,
		"email_verified":      true,
		"email_visible":       true,
		"email_verify_code":   "",  // 清空验证码
		"email_verify_expire": nil, // 清空过期时间
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "邮箱验证成功",
	})
}

// SendPhoneVerifyCode 发送手机验证码
func SendPhoneVerifyCode(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		Phone string `json:"phone" binding:"required,len=11"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号格式不正确"})
		return
	}

	// 生成6位随机验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	if err := services.SendSMSVerifyCode(req.Phone, code); err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": 501, "message": "短信服务未配置"})
		return
	}

	// 记录验证码到数据库
	var contact models.UserContact
	// 计算过期时间
	expireTime := time.Now().Add(10 * time.Minute)
	if err := config.DB.Where("user_id = ?", userID).First(&contact).Error; err != nil {
		// 创建新记录
		contact = models.UserContact{
			UserID:            uint64(userID),
			Phone:             req.Phone,
			EmailVerifyCode:   code, // 复用EmailVerifyCode字段存储手机验证码
			EmailVerifyExpire: &expireTime,
			PhoneVerified:     false,
		}
		config.DB.Create(&contact)
	} else {
		// 更新现有记录
		config.DB.Model(&contact).Updates(map[string]interface{}{
			"phone":               req.Phone,
			"email_verify_code":   code, // 复用EmailVerifyCode字段存储手机验证码
			"email_verify_expire": &expireTime,
			"phone_verified":      false,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "验证码已发送",
	})
}

// VerifyPhone 验证手机号
func VerifyPhone(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		Phone string `json:"phone" binding:"required,len=11"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 从数据库获取验证码
	var contact models.UserContact
	if err := config.DB.Where("user_id = ?", userID).First(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码已过期或不存在"})
		return
	}

	// 检查验证码是否匹配
	if contact.Phone != req.Phone || contact.EmailVerifyCode != req.Code {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码错误"})
		return
	}

	// 检查验证码是否过期
	if contact.EmailVerifyExpire == nil || contact.EmailVerifyExpire.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码已过期"})
		return
	}

	// 验证成功，更新用户状态
	config.DB.Model(&contact).Updates(map[string]interface{}{
		"phone":               req.Phone,
		"phone_verified":      true,
		"email_verify_code":   "",  // 清空验证码
		"email_verify_expire": nil, // 清空过期时间
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "手机号验证成功",
	})
}

// BindPhone 绑定手机号（微信小程序方式）
func BindPhone(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		Code          string `json:"code" binding:"required"`
		EncryptedData string `json:"encrypted_data"`
		IV            string `json:"iv"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

		var phone string
	if req.EncryptedData != "" && req.IV != "" {
		sessionKey, err := getWechatSessionKey(req.Code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "微信登录失败"})
			return
		}

		phone, err = decryptWechatPhone(sessionKey, req.EncryptedData, req.IV)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号解密失败"})
			return
		}
	} else {
		var err error
		phone, err = getWechatPhoneNumberByCode(req.Code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "获取手机号失败: " + err.Error()})
			return
		}
	}
// 查找或创建联系方式记录
	var contact models.UserContact
	if err := config.DB.Where("user_id = ?", userID).First(&contact).Error; err != nil {
		// 新建记录
		contact = models.UserContact{
			UserID:        uint64(userID),
			Phone:         phone,
			PhoneVerified: true,
		}
		config.DB.Create(&contact)
	} else {
		// 更新现有记录，使用 map 确保所有字段正确更新
		config.DB.Model(&contact).Updates(map[string]interface{}{
			"phone":          phone,
			"phone_verified": true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "手机号绑定成功",
		"data": gin.H{
			"phone":          phone,
			"phone_verified": true,
		},
	})
}

type wechatSessionResponse struct {
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type wechatAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type wechatPhoneNumberResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
	} `json:"phone_info"`
}

func getWechatAccessToken() (string, error) {
	if config.Config.WechatAppID == "" || config.Config.WechatSecret == "" {
		return "", errors.New("wechat config missing")
	}

	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(config.Config.WechatAppID),
		url.QueryEscape(config.Config.WechatSecret),
	)

	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("wechat response status %d: %s", resp.StatusCode, string(body))
	}

	var payload wechatAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		return "", fmt.Errorf("wechat error %d: %s", payload.ErrCode, payload.ErrMsg)
	}
	if payload.AccessToken == "" {
		return "", errors.New("wechat access_token empty")
	}

	return payload.AccessToken, nil
}

func getWechatPhoneNumberByCode(code string) (string, error) {
	accessToken, err := getWechatAccessToken()
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s",
		url.QueryEscape(accessToken),
	)

	body, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("wechat response status %d: %s", resp.StatusCode, string(respBody))
	}

	var payload wechatPhoneNumberResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		return "", fmt.Errorf("wechat error %d: %s", payload.ErrCode, payload.ErrMsg)
	}

	phone := payload.PhoneInfo.PhoneNumber
	if phone == "" {
		phone = payload.PhoneInfo.PurePhoneNumber
	}
	if phone == "" {
		return "", errors.New("wechat phone number empty")
	}

	return phone, nil
}

func getWechatSessionKey(code string) (string, error) {
	if config.Config.WechatAppID == "" || config.Config.WechatSecret == "" {
		return "", errors.New("wechat config missing")
	}

	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(config.Config.WechatAppID),
		url.QueryEscape(config.Config.WechatSecret),
		url.QueryEscape(code),
	)

	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("wechat response status %d: %s", resp.StatusCode, string(body))
	}

	var payload wechatSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		return "", fmt.Errorf("wechat error %d: %s", payload.ErrCode, payload.ErrMsg)
	}
	if payload.SessionKey == "" {
		return "", errors.New("wechat session_key empty")
	}

	return payload.SessionKey, nil
}

type wechatPhonePayload struct {
	PhoneNumber string `json:"phoneNumber"`
}

func decryptWechatPhone(sessionKey, encryptedData, iv string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return "", err
	}
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}
	if len(ivBytes) != aes.BlockSize {
		return "", errors.New("invalid iv size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(encryptedBytes)%aes.BlockSize != 0 {
		return "", errors.New("invalid encrypted data size")
	}

	decrypted := make([]byte, len(encryptedBytes))
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(decrypted, encryptedBytes)

	decrypted, err = pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		return "", err
	}

	var payload wechatPhonePayload
	if err := json.Unmarshal(decrypted, &payload); err != nil {
		return "", err
	}
	if payload.PhoneNumber == "" {
		return "", errors.New("phone number empty")
	}

	return payload.PhoneNumber, nil
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid padding size")
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, errors.New("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}
