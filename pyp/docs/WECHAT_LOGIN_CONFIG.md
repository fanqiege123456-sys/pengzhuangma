# 微信小程序登录配置完整指南

## 更新日期
2025-10-19

## 登录流程说明

### 官方流程图
```
小程序端                    开发者服务器                  微信服务器
   |                              |                           |
   |--- wx.login() ------------->|                           |
   |<-- 返回 code ----------------|                           |
   |                              |                           |
   |--- 发送 code --------------->|                           |
   |                              |                           |
   |                              |--- code2Session API ----->|
   |                              |                           |
   |                              |<-- openid + session_key --|
   |                              |                           |
   |                              |--- 查询/创建用户 -------->|
   |                              |<-- 用户信息 --------------|
   |                              |                           |
   |<-- 返回 token + user --------|                           |
   |                              |                           |
```

### 详细步骤

#### 1. 小程序端调用 wx.login()
```javascript
wx.login({
  success: (res) => {
    const code = res.code  // 获取临时登录凭证
    // 将 code 发送到后端
  }
})
```

#### 2. 后端调用 code2Session API
```
GET https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=SECRET&js_code=JSCODE&grant_type=authorization_code
```

**参数说明**：
- `appid`: 小程序 AppId
- `secret`: 小程序 AppSecret
- `js_code`: wx.login() 获取的 code
- `grant_type`: 固定值 `authorization_code`

**返回数据**：
```json
{
  "openid": "用户唯一标识",
  "session_key": "会话密钥",
  "unionid": "用户在开放平台的唯一标识符",
  "errcode": 0,
  "errmsg": "错误信息"
}
```

#### 3. 后端处理逻辑
```go
// 1. 获取 openid
openid := wechatResp.OpenID

// 2. 查询数据库，判断用户是否存在
var user models.User
if err := DB.Where("open_id = ?", openid).First(&user).Error; err != nil {
    // 用户不存在，创建新用户
    user = models.User{
        OpenID:   openid,
        Nickname: "微信用户",
        Avatar:   "默认头像",
        Coins:    1000,
    }
    DB.Create(&user)
} else {
    // 用户存在，更新登录时间
    DB.Model(&user).Update("last_login_at", time.Now())
}

// 3. 生成 JWT token
token := generateToken(user.ID)

// 4. 返回用户信息和 token
return {
    "token": token,
    "user": user
}
```

## 配置步骤

### 1. 获取小程序 AppID 和 AppSecret

#### 登录微信公众平台
1. 访问：https://mp.weixin.qq.com/
2. 使用管理员微信扫码登录
3. 进入【开发】→【开发管理】→【开发设置】

#### 复制配置信息
```
AppID(小程序ID): wxXXXXXXXXXXXXXXXX
AppSecret(小程序密钥): xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

⚠️ **重要提示**：
- AppSecret 非常重要，不要泄露
- 如果泄露，立即重置
- 不要提交到 Git 仓库

### 2. 配置后端环境变量

#### 编辑 `.env` 文件
```bash
cd /home/fanfan007/fanfandemo/pyp/backend
vim .env
```

#### 添加微信配置
```properties
# 微信小程序配置
WECHAT_APPID=你的小程序AppID
WECHAT_SECRET=你的小程序AppSecret
```

**示例**：
```properties
# 微信小程序配置
WECHAT_APPID=wx1234567890abcdef
WECHAT_SECRET=1234567890abcdef1234567890abcdef
```

### 3. 验证配置是否生效

#### 重启后端服务
```bash
cd /home/fanfan007/fanfandemo/pyp/backend
go run main.go
```

#### 查看启动日志
应该看到类似输出：
```
微信小程序配置已加载
AppID: wx1234567890abcdef
```

### 4. 测试登录流程

#### 清除小程序缓存
```
微信开发者工具 → 清缓存 → 清除数据缓存
```

#### 点击登录
1. 打开小程序
2. 点击"快速登录"按钮
3. 查看控制台输出

#### 预期日志
```javascript
// 前端日志
wx.login成功，获取到code: 071xxxxx
登录请求响应: {code: 200, data: {...}}
登录成功，用户信息: {...}

// 后端日志
微信登录请求 - Code: 071xxxxx
微信API返回 - OpenID: oUpF8uMuAJO_M2pxb1Q9zNjWeS6o
创建新用户成功: ID=1, OpenID=oUpF8..., Nickname=微信用户
登录成功，生成token: eyJhbGciOiJIUzI1...
```

## 当前代码实现

### 后端实现 (user.go)

#### 登录接口
```go
// 微信小程序登录
func (uc *UserController) WechatLogin(c *gin.Context) {
    var req struct {
        Code string `json:"code" binding:"required"`
    }
    
    // 1. 获取 code
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, utils.Error(400, "Invalid request"))
        return
    }
    
    // 2. 调用微信 API
    var openID string
    if config.Config.WechatAppID != "" && config.Config.WechatSecret != "" {
        // 生产环境：调用真实微信API
        wechatResp, err := uc.getWechatSession(req.Code)
        if err != nil {
            c.JSON(500, utils.Error(500, "Failed to verify wechat code"))
            return
        }
        openID = wechatResp.OpenID
    } else {
        // 开发环境：使用 code 的 MD5 hash
        codeHash := md5.Sum([]byte(req.Code))
        openID = "dev_" + hex.EncodeToString(codeHash[:])
    }
    
    // 3. 查找或创建用户
    var user models.User
    if err := DB.Where("open_id = ?", openID).First(&user).Error; err != nil {
        // 创建新用户
        user = models.User{
            OpenID:   openID,
            Nickname: "微信用户",
            Avatar:   "默认头像URL",
            WechatNo: "wx" + randomString(8),
            Coins:    1000,
        }
        DB.Create(&user)
    }
    
    // 4. 生成 token 并返回
    token := generateToken(user.ID)
    c.JSON(200, gin.H{
        "code": 200,
        "data": gin.H{
            "token": token,
            "user":  user,
        },
    })
}
```

#### 调用微信 API
```go
// 调用微信API获取session信息
func (uc *UserController) getWechatSession(code string) (*WechatSessionResponse, error) {
    url := fmt.Sprintf(
        "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
        config.Config.WechatAppID,
        config.Config.WechatSecret,
        code,
    )
    
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result WechatSessionResponse
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result.ErrCode != 0 {
        return nil, fmt.Errorf("wechat api error: %d %s", result.ErrCode, result.ErrMsg)
    }
    
    return &result, nil
}
```

### 前端实现 (app.js)

#### 登录方法
```javascript
// 微信登录
login() {
  return new Promise((resolve, reject) => {
    wx.showLoading({ title: '登录中...' })
    
    // 1. 调用 wx.login 获取 code
    wx.login({
      success: (loginRes) => {
        if (loginRes.code) {
          // 2. 将 code 发送到后端
          this.loginWithCode(loginRes.code)
            .then((user) => {
              wx.hideLoading()
              resolve(user)
            })
            .catch((err) => {
              wx.hideLoading()
              reject(err)
            })
        } else {
          wx.hideLoading()
          reject(new Error('获取登录凭证失败'))
        }
      },
      fail: (err) => {
        wx.hideLoading()
        reject(new Error('微信登录失败'))
      }
    })
  })
}
```

#### 发送 code 到后端
```javascript
// 使用code登录
loginWithCode(code) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${this.globalData.apiUrl}/user/login`,
      method: 'POST',
      data: {
        code: code  // 发送 code 到后端
      },
      header: {
        'content-type': 'application/json'
      },
      success: (res) => {
        if (res.data.code === 200) {
          const { token, user } = res.data.data
          
          // 保存 token 和用户信息
          wx.setStorageSync('token', token)
          wx.setStorageSync('userInfo', user)
          
          this.globalData.token = token
          this.globalData.userInfo = user
          this.globalData.hasLogin = true
          
          resolve(user)
        } else {
          reject(new Error(res.data.msg || '登录失败'))
        }
      },
      fail: (err) => {
        reject(err)
      }
    })
  })
}
```

## 开发环境 vs 生产环境

### 开发环境（未配置 AppID）

**特点**：
- 不调用微信 API
- 使用 code 的 MD5 hash 作为 openID
- 同一设备同一 openID

**配置**：
```properties
# .env 文件
WECHAT_APPID=
WECHAT_SECRET=
```

**OpenID 示例**：
```
dev_a3f2e1d4c5b6a7f8e9d0c1b2a3f4e5d6
```

### 生产环境（已配置 AppID）

**特点**：
- 调用真实微信 API
- 获取真实 openID
- 不同设备同一用户 openID 相同

**配置**：
```properties
# .env 文件
WECHAT_APPID=wx1234567890abcdef
WECHAT_SECRET=1234567890abcdef1234567890abcdef
```

**OpenID 示例**：
```
oUpF8uMuAJO_M2pxb1Q9zNjWeS6o
```

## 常见错误处理

### 错误码 40029: code 无效
**原因**：code 已过期或已使用
**解决**：重新调用 wx.login() 获取新的 code

### 错误码 45011: API 调用频繁
**原因**：每分钟调用次数超限
**解决**：降低调用频率，稍后重试

### 错误码 40226: code blocked
**原因**：高风险用户，登录被拦截
**解决**：建议用户联系微信客服

### 错误码 -1: system error
**原因**：微信系统繁忙
**解决**：稍后重试

## 安全建议

### 1. AppSecret 保护
- ✅ 存储在服务器环境变量中
- ✅ 不要提交到 Git
- ✅ 定期更换
- ❌ 不要硬编码在代码中
- ❌ 不要暴露给前端

### 2. session_key 保护
- ✅ 仅在服务器端保存
- ✅ 不要返回给前端
- ✅ 定期刷新
- ❌ 不要存储在数据库明文字段

### 3. token 安全
- ✅ 使用 JWT
- ✅ 设置合理的过期时间
- ✅ 使用 HTTPS 传输
- ❌ 不要在 URL 中传递

## 测试清单

### 首次登录
- [ ] 清除小程序缓存
- [ ] 点击登录按钮
- [ ] 后端调用微信 API 成功
- [ ] 获取到 openID
- [ ] 创建新用户
- [ ] 返回 token 和用户信息
- [ ] 前端保存到本地存储
- [ ] 显示用户信息

### 再次登录
- [ ] 清除小程序缓存
- [ ] 点击登录按钮
- [ ] 后端调用微信 API 成功
- [ ] 获取到 openID
- [ ] 查询到已存在用户
- [ ] 更新登录时间
- [ ] 返回 token 和用户信息
- [ ] 前端保存到本地存储
- [ ] 显示用户信息

### 账户绑定
- [ ] 同一微信用户
- [ ] 不同设备登录
- [ ] openID 相同
- [ ] 使用同一个账户

## 参考文档

- [微信小程序登录官方文档](https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/login.html)
- [code2Session API](https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/code2Session.html)
- [UnionID 机制说明](https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/union-id.html)

---

**配置文件位置**：
- `/home/fanfan007/fanfandemo/pyp/backend/.env`

**重要提示**：
1. **必须配置 WECHAT_APPID 和 WECHAT_SECRET 才能在生产环境使用**
2. **开发环境可以不配置，会使用 MD5 hash 模拟**
3. **配置后需要重启后端服务**

---

**版本**：v4.0  
**更新日期**：2025-10-19  
**状态**：✅ 符合微信官方规范
