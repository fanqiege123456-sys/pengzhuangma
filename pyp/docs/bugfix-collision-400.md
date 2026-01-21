# 🐛 碰撞接口400错误修复

## 问题描述
```
2025/10/01 09:56:45 碰撞请求参数错误: Key: 'Tag' Error:Field validation for 'Tag' failed on the 'required' tag
[GIN] 2025/10/01 - 09:56:45 | 400 | 773.541µs | 127.0.0.1 | POST "/api/collision/submit"
```

## 问题原因

### 后端期望的参数格式：
```go
{
  "tag": "兴趣标签",        // required - 必填
  "country": "中国",
  "province": "省份",
  "city": "城市",
  "district": "区县",
  "gender": 0,
  "cost_coins": 10
}
```

### 前端发送的参数（错误）：
首页 `index.js` 发送的是：
```javascript
{
  "code": "碰撞码"  // ❌ 字段名错误，应该是 tag
}
```

## 修复方案

### 1. 修复首页碰撞提交逻辑

**文件**: `/wx_web/pages/index/index.js`

#### 修改点：
1. **字段名修正**: `code` → `tag`
2. **添加地址信息**: 从全局用户信息中获取
3. **添加必要参数**: `gender`, `cost_coins`
4. **地址验证**: 提交前检查用户是否已设置地址

#### 修改后的代码：
```javascript
async submitCollisionCode() {
  const { inputCode } = this.data
  
  // 验证输入
  if (!inputCode.trim()) {
    wx.showToast({ title: '请输入碰撞码', icon: 'none' })
    return
  }

  // 获取用户信息
  const app = getApp()
  const userInfo = app.globalData.userInfo

  // 检查地址信息
  if (!userInfo || !userInfo.city) {
    wx.showModal({
      title: '提示',
      content: '请先在设置页面完善地址信息',
      confirmText: '去设置',
      success: (res) => {
        if (res.confirm) {
          wx.navigateTo({ url: '/pages/settings/settings' })
        }
      }
    })
    return
  }

  // 提交正确格式的数据
  const res = await api.submitCollisionCode({
    tag: inputCode.trim(),              // ✅ 使用tag字段
    country: userInfo.country || '中国',
    province: userInfo.province || '',
    city: userInfo.city || '',
    district: userInfo.district || '',
    gender: 0,                          // 不限性别
    cost_coins: 10                      // 默认消耗10积分
  })
  
  // 处理返回结果
  if (res.data.code === 200) {
    const result = res.data.data
    
    if (result.matched) {
      // 立即匹配成功
      wx.showToast({ title: '碰撞成功！找到匹配', icon: 'success' })
      setTimeout(() => {
        wx.switchTab({ url: '/pages/friends/friends' })
      }, 1500)
    } else {
      // 发布成功但未立即匹配
      wx.showToast({ title: '碰撞码发布成功', icon: 'success' })
      
      // 如果可以海底捞，提示用户
      if (result.can_haidilao) {
        setTimeout(() => {
          wx.showModal({
            title: '可以海底捞',
            content: `有${result.haidilao_count}人使用过该标签，可花费100积分海底捞`,
            confirmText: '去碰撞页面',
            cancelText: '暂不'
          })
        }, 1500)
      }
    }
  }
}
```

### 2. 添加后端调试日志

**文件**: `/backend/controllers/collision.go`

#### 添加导入：
```go
import (
  "log"  // ✅ 新增
  // ...其他导入
)
```

#### 添加详细日志：
```go
if err := c.ShouldBindJSON(&req); err != nil {
  log.Printf("碰撞请求参数错误: %v", err)  // ✅ 打印详细错误
  c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request: "+err.Error()))
  return
}

log.Printf("收到碰撞请求 - UserID: %v, Tag: %s, Location: %s/%s/%s/%s, Gender: %d, CostCoins: %d",
  userID, req.Tag, req.Country, req.Province, req.City, req.District, req.Gender, req.CostCoins)
```

### 3. 碰撞页面参数补充

**文件**: `/wx_web/pages/collision/collision.js`

添加 `cost_coins` 参数：
```javascript
const requestData = {
  tag: inputTag.trim(),
  gender: selectedGender,
  cost_coins: 10  // ✅ 添加默认消耗
}
```

## 测试方法

### 使用 go run 快速测试（无需编译）
```bash
cd /home/fanfan007/fanfandemo/pyp/backend
go run main.go
```

### 测试步骤：
1. ✅ 启动后端服务
2. ✅ 在设置页面完善地址信息
3. ✅ 在首页输入碰撞码并提交
4. ✅ 检查后端日志，应该看到：
   ```
   收到碰撞请求 - UserID: 12, Tag: 测试标签, Location: 中国/广东省/深圳市/南山区, Gender: 0, CostCoins: 10
   ```
5. ✅ 前端应该提示"碰撞码发布成功"或"碰撞成功！找到匹配"

## 修复效果

### 修复前：
- ❌ 400 Bad Request
- ❌ Tag字段验证失败
- ❌ 缺少必要参数

### 修复后：
- ✅ 请求参数格式正确
- ✅ 包含所有必要字段
- ✅ 支持立即匹配和海底捞提示
- ✅ 地址验证，确保用户已设置地址
- ✅ 详细的后端日志，便于调试

## 相关文件

### 前端修改：
- `/wx_web/pages/index/index.js` - 首页碰撞提交逻辑
- `/wx_web/pages/collision/collision.js` - 碰撞页面参数补充

### 后端修改：
- `/backend/controllers/collision.go` - 添加日志和错误信息

## 注意事项

1. **用户必须先设置地址**：碰撞功能依赖用户的地址信息，首次使用需引导用户完善地址
2. **积分消耗**：每次碰撞默认消耗10积分
3. **开发测试**：使用 `go run main.go` 可以快速测试，无需每次编译

