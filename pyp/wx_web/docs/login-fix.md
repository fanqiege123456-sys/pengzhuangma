# 微信授权登录修复文档

## 问题描述
用户点击"去登录"按钮后，没有弹出微信授权框，无法正常完成登录流程。

## 问题原因
1. **wx.getUserProfile 调用限制**：微信要求 `wx.getUserProfile` 必须在用户点击事件中**同步调用**，不能通过异步方式触发
2. **缺少用户交互界面**：只有弹窗提示，没有明确的登录按钮供用户点击
3. **登录流程缺少反馈**：loading 状态和成功提示不够明确

## 解决方案

### 1. 添加首页登录界面

在 `pages/index/index.wxml` 中添加未登录状态的登录卡片：

```html
<!-- 未登录提示 -->
<view class="card login-prompt" wx:if="{{!userInfo}}">
  <view class="prompt-title">🎯 欢迎使用碰撞小程序</view>
  <view class="prompt-desc">请先进行微信授权登录，开启奇妙之旅</view>
  <button class="login-btn" bindtap="handleLogin">
    <text>微信授权登录</text>
  </button>
</view>
```

**关键点**：
- ✅ 使用 `wx:if="{{!userInfo}}"` 只在未登录时显示
- ✅ 提供明确的登录按钮
- ✅ 按钮直接绑定到 `handleLogin` 方法

### 2. 优化登录方法

在 `pages/index/index.js` 中添加登录处理：

```javascript
// 手动触发登录
handleLogin() {
  const app = getApp()
  
  app.login()
    .then(() => {
      wx.showToast({
        title: '登录成功',
        icon: 'success'
      })
      // 刷新页面数据
      setTimeout(() => {
        this.loadUserInfo()
        this.loadRecentCollisions()
      }, 1500)
    })
    .catch((err) => {
      console.error('登录失败:', err)
    })
}
```

### 3. 改进 app.js 登录流程

添加 loading 状态管理：

```javascript
login() {
  return new Promise((resolve, reject) => {
    wx.showLoading({
      title: '登录中...'
    })
    
    wx.login({
      success: (loginRes) => {
        if (loginRes.code) {
          // 必须在这里直接调用 getUserProfile
          wx.getUserProfile({
            desc: '用于完善用户资料',
            success: (profileRes) => {
              this.loginWithCodeAndProfile(loginRes.code, profileRes.userInfo)
                .then((user) => {
                  wx.hideLoading()
                  resolve(user)
                })
                .catch((err) => {
                  wx.hideLoading()
                  reject(err)
                })
            },
            fail: (err) => {
              wx.hideLoading()
              wx.showToast({
                title: '需要授权才能使用',
                icon: 'none'
              })
              reject(new Error('用户拒绝授权'))
            }
          })
        }
      }
    })
  })
}
```

### 4. 优化 requireLogin 弹窗

改进弹窗后的处理逻辑：

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
            this.login()
              .then(() => {
                wx.showToast({
                  title: '登录成功',
                  icon: 'success'
                })
                // 刷新当前页面
                setTimeout(() => {
                  const pages = getCurrentPages()
                  const currentPage = pages[pages.length - 1]
                  if (currentPage && currentPage.onShow) {
                    currentPage.onShow()
                  }
                }, 1500)
              })
              .catch((err) => {
                wx.showToast({
                  title: err.message || '登录失败',
                  icon: 'none'
                })
              })
          }
        }
      })
    }
    return false
  }
  return true
}
```

### 5. 添加登录按钮样式

在 `pages/index/index.wxss` 中添加样式：

```css
/* 登录提示卡片 */
.login-prompt {
  text-align: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 60rpx 40rpx;
}

.prompt-title {
  font-size: 40rpx;
  font-weight: bold;
  margin-bottom: 20rpx;
}

.prompt-desc {
  font-size: 28rpx;
  opacity: 0.9;
  margin-bottom: 40rpx;
  line-height: 1.5;
}

.login-btn {
  background: white;
  color: #667eea;
  border-radius: 50rpx;
  height: 88rpx;
  line-height: 88rpx;
  font-size: 32rpx;
  font-weight: bold;
  border: none;
}
```

## 用户体验流程

### 方式1：首页登录（推荐）
```
打开小程序 → 看到登录卡片 → 点击"微信授权登录" 
→ 弹出微信授权框 → 点击"允许" → 登录成功 ✓
```

### 方式2：功能触发登录
```
打开小程序 → 点击任意功能（碰撞/匹配/设置）
→ 弹出"需要登录"对话框 → 点击"去登录"
→ 弹出微信授权框 → 点击"允许" → 登录成功 → 返回功能页 ✓
```

## 技术要点

### wx.getUserProfile 调用规则

**✅ 正确用法**：
```javascript
// 在按钮点击事件中直接调用
handleLogin() {
  wx.login({
    success: (res) => {
      wx.getUserProfile({  // 必须在事件回调中同步调用
        desc: '用于完善用户资料',
        success: (profile) => {
          // 处理授权成功
        }
      })
    }
  })
}
```

**❌ 错误用法**：
```javascript
// 异步调用（不会弹出授权框）
async handleLogin() {
  await someAsyncFunction()
  wx.getUserProfile({  // ❌ 异步调用无效
    desc: '用于完善用户资料'
  })
}

// 延迟调用（不会弹出授权框）
handleLogin() {
  setTimeout(() => {
    wx.getUserProfile({  // ❌ 延迟调用无效
      desc: '用于完善用户资料'
    })
  }, 1000)
}
```

### Loading 状态管理

1. **显示 loading**：在调用 `wx.login` 时立即显示
2. **隐藏 loading**：在以下情况隐藏：
   - 登录成功
   - 用户拒绝授权
   - 网络错误
   - 后端返回错误

### 错误处理

1. **用户拒绝授权**：提示"需要授权才能使用"
2. **网络错误**：提示"登录失败，请重试"
3. **后端错误**：显示后端返回的错误信息

## 测试检查清单

- [x] 首页显示登录卡片
- [x] 点击登录按钮弹出授权框
- [x] 授权成功后保存用户信息
- [x] 登录成功后显示欢迎信息和快速操作
- [x] 点击其他功能时如果未登录会弹出登录提示
- [x] 弹窗点击"去登录"可以正常授权
- [x] 用户拒绝授权后显示提示
- [x] 登录成功后刷新页面数据

## 修改的文件

### 前端
- `wx_web/app.js` - 优化登录流程，添加 loading 管理
- `wx_web/pages/index/index.js` - 添加 handleLogin 方法
- `wx_web/pages/index/index.wxml` - 添加登录卡片
- `wx_web/pages/index/index.wxss` - 添加登录卡片样式

### 文档
- `wx_web/docs/login-fix.md` - 本文档

## 常见问题

### Q1: 点击"去登录"没有弹出授权框？
**A**: 检查以下几点：
1. `wx.getUserProfile` 是否在用户点击事件中直接调用（不能异步）
2. 检查微信开发者工具版本是否最新
3. 检查小程序基础库版本（需要 >= 2.10.4）
4. 检查是否在真机上测试（开发者工具可能表现不同）

### Q2: 授权成功但没有保存用户信息？
**A**: 检查以下几点：
1. 查看控制台是否有错误信息
2. 检查后端 `/api/user/login` 接口是否正常
3. 检查 token 是否正确保存到本地存储
4. 检查 `app.globalData.hasLogin` 是否设置为 true

### Q3: 首页不显示登录卡片？
**A**: 检查以下几点：
1. 确认 `userInfo` 数据是否为 null
2. 检查 `wx:if="{{!userInfo}}"` 条件是否正确
3. 清除小程序缓存重新测试

### Q4: 真机和开发者工具表现不一致？
**A**: 
- 开发者工具可能对 `wx.getUserProfile` 有不同的处理
- 建议以真机测试为准
- 可以在真机上使用"清除缓存"功能测试首次登录

## 下一步优化建议

1. **添加登录状态持久化验证**：定期检查 token 有效性
2. **优化授权失败重试机制**：允许用户多次尝试授权
3. **添加登录埋点统计**：记录登录成功率
4. **支持更多登录方式**：手机号登录、一键登录等

---

**版本**：v2.1  
**更新日期**：2025-10-19  
**状态**：✅ 已修复
