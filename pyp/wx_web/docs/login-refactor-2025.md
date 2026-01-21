# 微信小程序登录重构文档（2025最新规范）

## 更新日期
2025-10-19

## 背景说明

根据微信官方文档（https://developers.weixin.qq.com/miniprogram/dev/platform-capabilities/miniapp/quickstart/auth.html），微信小程序的用户信息获取方式已经改变：

### ❌ 已废弃的方式
- `wx.getUserProfile()` - 已停止维护
- `wx.getUserInfo()` + `open-type="getUserInfo"` - 已停止维护

### ✅ 推荐的新方式
1. **头像昵称填写组件** - 用户主动填写（推荐）
2. **手机号快速验证组件** - 用户授权获取手机号
3. **仅使用 openid** - 不获取用户信息

## 当前实现方案

我们采用**简化登录流程**：

1. **仅使用 `wx.login()` 获取 code**
2. **后端通过 code 换取 openid**
3. **用户信息由用户后续主动填写**

### 优点
- ✅ 符合最新微信规范
- ✅ 登录流程简单快速
- ✅ 不需要用户授权（降低流失率）
- ✅ 尊重用户隐私

### 缺点
- ❌ 无法自动获取用户头像昵称
- ❌ 需要用户手动完善资料

## 代码实现

### 1. 前端登录流程（app.js）

```javascript
// 微信登录（新版本API）
login() {
  return new Promise((resolve, reject) => {
    console.log('开始微信登录流程')
    
    wx.showLoading({
      title: '登录中...'
    })
    
    // 仅调用wx.login获取code
    wx.login({
      success: (loginRes) => {
        if (loginRes.code) {
          // 仅使用code登录
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

// 使用code登录（不获取用户信息）
loginWithCode(code) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${this.globalData.apiUrl}/user/login`,
      method: 'POST',
      data: {
        code: code
        // 不再发送 userInfo
      },
      header: {
        'content-type': 'application/json'
      },
      success: (res) => {
        if (res.data.code === 200) {
          const { token, user } = res.data.data
          
          // 保存token和用户信息
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

### 2. 首页登录界面（index.wxml）

```html
<!-- 未登录提示 -->
<view class="card login-prompt" wx:if="{{!userInfo}}">
  <view class="prompt-title">🎯 欢迎使用碰撞小程序</view>
  <view class="prompt-desc">点击下方按钮即可快速登录，开启奇妙之旅</view>
  <button class="login-btn" bindtap="handleLogin" type="primary">
    快速登录
  </button>
  <view class="login-tip">登录后可以完善个人资料</view>
</view>

<!-- 已登录显示用户信息 -->
<view class="card welcome-card" wx:if="{{userInfo}}">
  <view class="welcome-message">{{welcomeMessage}}</view>
  <view class="user-info">
    <view class="info-item">
      <text class="label">用户昵称：</text>
      <text class="value">{{userInfo.nickname || '未设置'}}</text>
    </view>
    <view class="info-item">
      <text class="label">积分余额：</text>
      <text class="value">{{userInfo.coins || 0}}</text>
    </view>
  </view>
</view>
```

### 3. 后端登录逻辑（user.go）

后端已支持 `userInfo` 为可选参数：

```go
// 微信小程序登录
func (uc *UserController) WechatLogin(c *gin.Context) {
    var req struct {
        Code     string `json:"code" binding:"required"`
        UserInfo *struct {  // 注意：使用指针，可选
            NickName  string `json:"nickName"`
            AvatarUrl string `json:"avatarUrl"`
            // ...
        } `json:"userInfo"`
    }
    
    // ... 获取 openID
    
    // 查找或创建用户
    var user models.User
    if err := config.DB.Where("open_id = ?", openID).First(&user).Error; err != nil {
        // 创建新用户（使用默认值）
        nickname := "微信用户"
        avatar := "默认头像URL"
        
        // 如果提供了userInfo，使用用户提供的信息
        if req.UserInfo != nil {
            nickname = req.UserInfo.NickName
            avatar = req.UserInfo.AvatarUrl
        }
        
        user = models.User{
            OpenID:   openID,
            Nickname: nickname,
            Avatar:   avatar,
            WechatNo: "wx" + utils.GenerateRandomString(8),
            Coins:    1000,
        }
        
        config.DB.Create(&user)
    }
    
    // 返回 token 和用户信息
    token, _ := utils.GenerateToken(user.ID, "user")
    c.JSON(200, gin.H{
        "code": 200,
        "data": gin.H{
            "token": token,
            "user":  user,
        },
    })
}
```

## 用户体验流程

### 1. 首次登录
```
打开小程序
  ↓
显示登录卡片
  ↓
点击"快速登录"
  ↓
自动获取 code
  ↓
后端创建账户（默认昵称"微信用户"）
  ↓
登录成功，显示首页 ✓
```

### 2. 完善资料（可选）
```
登录后
  ↓
进入"个人中心"或"设置"
  ↓
使用头像昵称填写组件
  ↓
用户主动上传头像和填写昵称
  ↓
更新到后端 ✓
```

### 3. 再次使用
```
打开小程序
  ↓
检查本地 token
  ↓
token 有效
  ↓
直接进入首页 ✓
```

## 后续优化方案

### 方案1：头像昵称填写组件

在个人中心页面添加：

```html
<!-- 头像选择 -->
<button class="avatar-wrapper" open-type="chooseAvatar" bind:chooseavatar="onChooseAvatar">
  <image class="avatar" src="{{avatarUrl}}"></image>
</button>

<!-- 昵称输入 -->
<input type="nickname" class="nickname" placeholder="请输入昵称" bind:change="onNicknameChange"/>
```

```javascript
// 选择头像
onChooseAvatar(e) {
  const { avatarUrl } = e.detail
  // 上传到服务器
  this.uploadAvatar(avatarUrl)
}

// 输入昵称
onNicknameChange(e) {
  const { value } = e.detail
  // 更新昵称
  this.updateNickname(value)
}
```

### 方案2：手机号快速验证

```html
<button open-type="getPhoneNumber" bindgetphonenumber="getPhoneNumber">
  获取手机号
</button>
```

```javascript
getPhoneNumber(e) {
  const { code } = e.detail
  // 将 code 发送到后端换取手机号
  this.bindPhoneNumber(code)
}
```

## 关键改动总结

### 前端改动
1. **删除 `wx.getUserProfile()` 调用**
2. **删除 `loginWithCodeAndProfile()` 方法**
3. **简化为 `loginWithCode()` 方法**
4. **首页登录按钮改为"快速登录"**
5. **添加"登录后可以完善个人资料"提示**

### 后端改动
- ✅ 无需改动（已支持 userInfo 可选）

### 数据库
- ✅ 无需改动

## 测试步骤

1. **清除缓存**
   ```
   微信开发者工具 → 清缓存 → 清除数据缓存
   ```

2. **测试快速登录**
   ```
   编译 → 看到登录卡片 → 点击"快速登录" → 查看控制台
   ```

3. **验证登录状态**
   ```
   检查本地存储：token、userInfo
   检查页面显示：用户昵称、积分余额
   ```

4. **测试功能访问**
   ```
   点击"发起碰撞" → 正常进入
   点击"我的匹配" → 正常进入
   ```

## 环境配置

### 开发环境
```bash
# .env 文件
# 不配置 APPID，使用 code 的 MD5 hash
```

### 生产环境
```bash
# .env 文件
WECHAT_APPID=wx1234567890abcdef
WECHAT_SECRET=1234567890abcdef1234567890abcdef
```

## 注意事项

1. **用户昵称默认值**：新用户默认昵称为"微信用户"
2. **头像默认值**：使用默认头像 URL
3. **完善资料**：引导用户在个人中心完善资料
4. **隐私保护**：不强制获取用户信息，符合隐私规范
5. **用户体验**：登录流程简单快速，减少流失

## 相关文档

- [微信官方文档 - 小程序登录](https://developers.weixin.qq.com/miniprogram/dev/platform-capabilities/miniapp/quickstart/auth.html)
- [头像昵称填写](https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/userProfile.html)
- [手机号快速验证](https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/getPhoneNumber.html)

## 文件清单

### 修改的文件
- `wx_web/app.js` - 简化登录流程
- `wx_web/pages/index/index.wxml` - 更新登录界面
- `wx_web/pages/index/index.wxss` - 添加样式

### 文档
- `wx_web/docs/login-refactor-2025.md` - 本文档

---

**版本**：v3.0 (2025最新规范)  
**更新日期**：2025-10-19  
**状态**：✅ 符合微信最新规范
