# 登录系统强制授权更新文档

## 更新日期
2025-10-19

## 更新概述
实现**强制登录机制**，移除模拟登录降级方案，用户启动小程序后必须进行微信授权登录才能使用任何功能。

## 主要变更

### 1. app.js 核心逻辑调整

#### 1.1 移除自动登录
- ❌ 删除 `checkLogin()` 中的自动调用 `login()`
- ✅ 新增 `checkLoginStatus()` 仅检查登录状态，不自动登录
- ✅ 新增 `requireLogin()` 方法供各页面调用

```javascript
// 检查登录状态（仅检查，不自动登录）
checkLoginStatus() {
  const token = wx.getStorageSync('token')
  const userInfo = wx.getStorageSync('userInfo')
  
  if (token && userInfo) {
    this.globalData.token = token
    this.globalData.userInfo = userInfo
    this.globalData.hasLogin = true
  } else {
    this.globalData.hasLogin = false
  }
}

// 检查是否需要登录（供页面调用）
requireLogin() {
  if (!this.globalData.hasLogin) {
    wx.redirectTo({
      url: '/pages/login/login'
    })
    return false
  }
  return true
}
```

#### 1.2 移除模拟登录降级
删除了所有 `mockLogin()` 调用，改为返回明确错误：

- `wx.login` 失败 → 返回错误提示
- 匿名登录失败 → 返回错误提示
- 网络请求失败 → 返回错误提示

### 2. app.json 配置调整

```json
{
  "pages": [
    "pages/login/login",  // 登录页作为首页
    "pages/index/index",
    "pages/collision/collision",
    // ...其他页面
  ]
}
```

### 3. 各页面添加登录检查

所有主要页面（index、collision、match/list、profile、friends）都在 `onLoad` 和 `onShow` 中添加登录检查：

```javascript
onLoad() {
  const app = getApp()
  // 检查登录状态，未登录则跳转登录页
  if (!app.requireLogin()) {
    return
  }
  // 继续正常逻辑
  this.loadData()
}

onShow() {
  const app = getApp()
  // 每次显示时检查登录状态
  if (!app.globalData.hasLogin) {
    wx.redirectTo({
      url: '/pages/login/login'
    })
    return
  }
  // 继续正常逻辑
  this.refreshData()
}
```

### 4. 退出登录功能

#### 4.1 settings.js 添加 logout 方法

```javascript
// 退出登录
logout() {
  wx.showModal({
    title: '退出登录',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        const app = getApp()
        
        // 清除本地存储
        wx.removeStorageSync('token')
        wx.removeStorageSync('userInfo')
        
        // 清除全局数据
        app.globalData.token = null
        app.globalData.userInfo = null
        app.globalData.hasLogin = false
        
        wx.showToast({
          title: '已退出登录',
          icon: 'success'
        })
        
        // 跳转到登录页
        setTimeout(() => {
          wx.reLaunch({
            url: '/pages/login/login'
          })
        }, 1500)
      }
    }
  })
}
```

#### 4.2 settings.wxml 添加退出登录按钮

```html
<!-- 退出登录按钮 -->
<button class="logout-btn" bindtap="logout">
  退出登录
</button>
```

#### 4.3 settings.wxss 添加样式

```css
.logout-btn {
  width: 100%;
  height: 88rpx;
  background: #fff;
  color: #ff4d4f;
  border: 2rpx solid #ff4d4f;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: bold;
}
```

## 用户体验流程

### 首次使用
1. 打开小程序 → 自动进入登录页
2. 用户选择"微信授权登录"或"匿名登录"
3. 授权成功 → 跳转到首页
4. 可以正常使用所有功能

### 已登录用户
1. 打开小程序 → 自动检查本地 token
2. token 有效 → 直接进入首页
3. token 失效 → 跳转到登录页重新登录

### 退出登录
1. 进入"个人中心" → "设置"
2. 点击"退出登录"按钮
3. 确认对话框 → 确定
4. 清除本地数据 → 跳转到登录页

### 未登录访问限制
1. 用户尝试访问任何需要登录的页面
2. 自动检测到未登录状态
3. 立即跳转到登录页
4. 登录成功后可继续使用

## 登录方式

### 1. 微信授权登录（推荐）
- 获取用户微信头像、昵称等信息
- 调用 `wx.getUserProfile` → 获取 code + userInfo
- 发送到后端 `/api/user/login`
- 保存 token 和用户信息

### 2. 匿名登录（备选）
- 仅使用微信 code 登录
- 不获取用户头像昵称
- 发送到后端 `/api/user/login`（只传 code）
- 保存 token 和用户信息

### 3. ❌ 模拟登录（已移除）
- 之前用于开发测试的降级方案
- 会生成随机假账户
- 已完全移除

## 安全性增强

1. **强制登录验证**：所有页面都必须验证登录状态
2. **无降级方案**：登录失败不再自动生成假账户
3. **明确错误提示**：网络错误或登录失败会显示具体原因
4. **token 验证**：所有 API 请求都携带 token，后端验证身份

## 测试检查清单

- [ ] 清除小程序缓存后首次启动，显示登录页
- [ ] 点击"微信授权登录"，弹出授权框
- [ ] 授权成功后跳转到首页，显示用户信息
- [ ] 退出登录后，再次进入小程序显示登录页
- [ ] 未登录状态下访问任何 tabBar 页面，自动跳转登录页
- [ ] 登录后刷新页面，用户状态保持（不会重复跳转登录页）
- [ ] 登录失败时显示明确错误信息，不会生成假账户

## 注意事项

1. **清除缓存测试**：修改后需要清除小程序缓存重新测试
2. **后端接口**：确保后端 `/api/user/login` 接口正常工作
3. **Token 过期**：如果后端返回 401，前端应清除 token 并跳转登录页
4. **网络异常**：网络错误时应提示用户，不再自动生成假账户

## 相关文件清单

### 核心文件
- `/wx_web/app.js` - 应用主逻辑
- `/wx_web/app.json` - 页面配置

### 登录相关
- `/wx_web/pages/login/login.js` - 登录页逻辑
- `/wx_web/pages/login/login.wxml` - 登录页 UI
- `/wx_web/pages/login/login.wxss` - 登录页样式

### 设置相关
- `/wx_web/pages/settings/settings.js` - 添加退出登录
- `/wx_web/pages/settings/settings.wxml` - 添加退出按钮
- `/wx_web/pages/settings/settings.wxss` - 按钮样式

### 需要登录验证的页面
- `/wx_web/pages/index/index.js` - 首页
- `/wx_web/pages/collision/collision.js` - 碰撞页
- `/wx_web/pages/match/list.js` - 匹配列表
- `/wx_web/pages/profile/profile.js` - 个人中心
- `/wx_web/pages/friends/friends.js` - 好友列表

## 后续优化建议

1. **Token 自动刷新**：在 token 即将过期时自动刷新
2. **网络异常处理**：统一的网络错误处理机制
3. **登录状态监听**：实现全局登录状态变化监听
4. **记住登录状态**：支持"记住我"功能（可选）
5. **第三方登录**：支持其他登录方式（如手机号登录）
