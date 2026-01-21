# 微信账户绑定机制更新文档

## 更新日期
2025-10-19

## 更新概述
实现**微信账户绑定机制**，确保每次登录都是同一个微信账户，而不是每次都创建新账户。移除登录页面和匿名登录功能。

## 核心问题
之前的实现存在以下问题：
1. ❌ 后端使用 `mock_openid_` + code 生成随机 openID，导致每次登录都创建新用户
2. ❌ 有独立的登录页面，用户体验不佳
3. ❌ 支持匿名登录，导致用户身份不唯一
4. ❌ 没有正确调用微信 API 获取真实 openID

## 解决方案

### 1. 后端修改 - 真实 OpenID 绑定

#### 1.1 user.go 登录逻辑优化

```go
// 调用微信API验证code，获取openid
var openID string

if config.Config.WechatAppID != "" && config.Config.WechatSecret != "" {
    // 生产模式：调用微信API
    wechatResp, err := uc.getWechatSession(req.Code)
    if err != nil {
        return error
    }
    openID = wechatResp.OpenID
} else {
    // 开发模式：使用code的MD5 hash作为固定openID
    codeHash := md5.Sum([]byte(req.Code))
    openID = "dev_" + hex.EncodeToString(codeHash[:])
}
```

**关键点**：
- ✅ 生产环境使用真实微信 API 获取 openID
- ✅ 开发环境使用 code 的 MD5 hash，保证同一设备同一 openID
- ✅ 通过 openID 查找用户，存在则更新信息，不存在则创建

#### 1.2 新增微信 API 调用函数

```go
func (uc *UserController) getWechatSession(code string) (*WechatSessionResponse, error) {
    url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
        config.Config.WechatAppID,
        config.Config.WechatSecret,
        code,
    )
    
    resp, err := http.Get(url)
    // ... 解析返回的 openid 和 session_key
}
```

### 2. 前端修改 - 移除登录页

#### 2.1 app.json 配置

```json
{
  "pages": [
    "pages/index/index",  // 首页作为启动页
    // 移除 "pages/login/login"
    // ...其他页面
  ]
}
```

#### 2.2 app.js 登录流程

**启动时自动检查登录状态**：

```javascript
onLaunch() {
  // 自动尝试静默登录
  this.checkAndAutoLogin()
}

checkAndAutoLogin() {
  const token = wx.getStorageSync('token')
  const userInfo = wx.getStorageSync('userInfo')
  
  if (token && userInfo) {
    // 使用本地缓存
    this.globalData.token = token
    this.globalData.userInfo = userInfo
    this.globalData.hasLogin = true
    
    // 验证token有效性
    this.validateToken()
  } else {
    this.globalData.hasLogin = false
  }
}
```

**需要登录时弹窗提示**：

```javascript
requireLogin(showModal = true) {
  if (!this.globalData.hasLogin) {
    if (showModal) {
      wx.showModal({
        title: '需要登录',
        content: '请先进行微信授权登录',
        confirmText: '去登录',
        success: (res) => {
          if (res.confirm) {
            this.login()  // 触发登录流程
          }
        }
      })
    }
    return false
  }
  return true
}
```

**强制微信授权登录**：

```javascript
login() {
  return new Promise((resolve, reject) => {
    // 第一步：获取code
    wx.login({
      success: (loginRes) => {
        // 第二步：必须获取用户授权
        wx.getUserProfile({
          desc: '用于完善用户资料',
          success: (profileRes) => {
            // 第三步：发送到后端
            this.loginWithCodeAndProfile(loginRes.code, profileRes.userInfo)
              .then(resolve)
              .catch(reject)
          },
          fail: (err) => {
            wx.showToast({
              title: '需要授权才能使用',
              icon: 'none'
            })
            reject(new Error('用户拒绝授权'))
          }
        })
      }
    })
  })
}
```

**移除的功能**：
- ❌ 删除 `loginWithCode()` 匿名登录函数
- ❌ 删除 `mockLogin()` 模拟登录函数
- ❌ 删除 `getUserProfile()` 单独的函数（合并到 login 中）

#### 2.3 页面onShow检查

所有主要页面的 `onShow` 方法：

```javascript
onShow() {
  const app = getApp()
  if (!app.globalData.hasLogin) {
    app.requireLogin()  // 弹窗提示登录
    return
  }
  // 继续正常逻辑
  this.loadData()
}
```

### 3. 用户体验流程

#### 首次使用
```
打开小程序
  ↓
进入首页
  ↓
检测到未登录
  ↓
弹出授权提示框："需要登录 - 请先进行微信授权登录"
  ↓
用户点击"去登录"
  ↓
弹出微信授权框："获取你的昵称、头像等信息"
  ↓
用户点击"允许"
  ↓
后端创建账户（绑定微信openID）
  ↓
保存token + userInfo到本地
  ↓
可以正常使用所有功能 ✓
```

#### 再次使用（已登录）
```
打开小程序
  ↓
检查本地token
  ↓
token有效
  ↓
直接进入首页 ✓
```

#### 退出登录后
```
设置 → 退出登录
  ↓
清除本地token + userInfo
  ↓
跳转到首页
  ↓
弹出授权提示框
  ↓
用户重新授权
  ↓
使用同一个微信账户登录 ✓
```

## 开发与生产环境

### 开发环境（未配置微信APPID）
```env
# .env 文件为空或不存在
WECHAT_APPID=
WECHAT_SECRET=
```

**行为**：
- 使用 code 的 MD5 hash 作为固定 openID
- 同一个设备的同一个微信账户，每次登录都是同一个 openID
- 示例：`dev_a3f2e1d4c5b6a7f8e9d0c1b2a3f4e5d6`

### 生产环境（配置微信APPID）
```env
# .env 文件
WECHAT_APPID=wx1234567890abcdef
WECHAT_SECRET=1234567890abcdef1234567890abcdef
```

**行为**：
- 调用微信 API `https://api.weixin.qq.com/sns/jscode2session`
- 获取真实的 openID 和 session_key
- 通过 openID 唯一标识用户
- 示例：`oUpF8uMuAJO_M2pxb1Q9zNjWeS6o`

## 账户绑定机制

### 数据库设计
```sql
CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  open_id VARCHAR(255) UNIQUE NOT NULL,  -- 微信openID，唯一标识
  nickname VARCHAR(255),
  avatar VARCHAR(255),
  wechat_no VARCHAR(50),
  // ...其他字段
  INDEX idx_open_id (open_id)
);
```

### 登录流程
1. 前端调用 `wx.login()` 获取 code
2. 前端调用 `wx.getUserProfile()` 获取用户信息
3. 前端发送 code + userInfo 到后端 `/api/user/login`
4. 后端使用 code 换取 openID（生产）或生成固定hash（开发）
5. 后端通过 openID 查询数据库：
   - 存在 → 更新用户信息（昵称、头像等）
   - 不存在 → 创建新用户
6. 后端生成 JWT token 返回
7. 前端保存 token + userInfo 到本地存储

### 关键代码

**后端查找或创建用户**：
```go
var user models.User
if err := config.DB.Where("open_id = ?", openID).First(&user).Error; err != nil {
    // 用户不存在，创建新用户
    user = models.User{
        OpenID:   openID,
        Nickname: nickname,
        Avatar:   avatar,
        WechatNo: "wx" + utils.GenerateRandomString(8),
        Coins:    1000,
    }
    config.DB.Create(&user)
} else {
    // 用户已存在，更新信息
    user.Nickname = nickname
    user.Avatar = avatar
    config.DB.Save(&user)
}
```

## 安全性

1. **OpenID 唯一性**：每个微信用户在小程序中的 openID 是唯一的
2. **Session_key 保密**：不返回给前端，仅后端使用
3. **JWT Token**：用户登录后使用 token 进行身份验证
4. **Token 验证**：可选实现 token 有效性验证（避免过期token）

## 测试检查清单

### 开发环境测试
- [ ] 首次登录：创建新账户
- [ ] 再次登录：使用同一个账户（检查user_id一致）
- [ ] 退出登录后重新登录：使用同一个账户
- [ ] 不同设备登录：创建不同账户（因为code不同）

### 生产环境测试（需要配置APPID）
- [ ] 首次登录：创建新账户
- [ ] 再次登录：使用同一个账户
- [ ] 不同设备登录：使用同一个账户（因为openID相同）
- [ ] 卸载重装：使用同一个账户（openID不变）

### 功能测试
- [ ] 清除小程序缓存后，提示登录
- [ ] 用户拒绝授权后，显示提示
- [ ] 授权成功后，保存用户信息
- [ ] 退出登录后，清除本地数据
- [ ] 未登录状态访问任何页面，弹出登录提示

## 相关文件

### 后端
- `/backend/controllers/user.go` - 登录逻辑，微信API调用
- `/backend/config/config.go` - 微信APPID配置
- `/backend/.env` - 环境变量（需手动配置）

### 前端
- `/wx_web/app.js` - 登录流程，token管理
- `/wx_web/app.json` - 移除登录页
- `/wx_web/pages/*/index.js` - 所有页面添加登录检查
- `/wx_web/pages/settings/settings.js` - 退出登录功能

## 环境变量配置

### 开发环境
```bash
# .env 文件
# 不配置APPID，使用开发模式
```

### 生产环境
```bash
# .env 文件
WECHAT_APPID=你的小程序APPID
WECHAT_SECRET=你的小程序SECRET

# 获取方式：
# 1. 登录微信公众平台 https://mp.weixin.qq.com/
# 2. 开发 → 开发管理 → 开发设置
# 3. 复制 AppID(小程序ID) 和 AppSecret(小程序密钥)
```

## 常见问题

### Q1: 为什么每次登录都是新账户？
**A**: 检查后端是否正确使用 openID 查询用户。开发环境下确保使用 code 的 hash 而不是随机字符串。

### Q2: 如何测试账户绑定？
**A**: 
1. 清除小程序缓存
2. 登录一次，记录user_id（在控制台查看）
3. 清除缓存后再次登录
4. 检查user_id是否一致

### Q3: 生产环境如何配置？
**A**: 
1. 在微信公众平台获取APPID和SECRET
2. 在服务器 `.env` 文件中配置
3. 重启后端服务
4. 前端会自动调用真实API

### Q4: 用户拒绝授权怎么办？
**A**: 用户点击"拒绝"后，会显示提示"需要授权才能使用"。用户可以：
- 再次点击需要登录的功能，重新弹出授权提示
- 或者到"设置"中重新授权

## 后续优化建议

1. **Token 自动刷新**：在 token 即将过期时自动刷新
2. **登录状态监听**：实现全局登录状态变化监听
3. **多端登录**：支持同一账户在多个设备登录
4. **手机号登录**：添加手机号绑定功能
5. **UnionID 支持**：如果有公众号，可以使用 UnionID 打通多个应用
