// pages/login/login.js
Page({
  data: {
  },

  onLoad() {
    // 页面加载时的初始化逻辑（如果需要）
  },

    // 微信登录
  handleLogin() {
    const app = getApp()
    wx.showLoading({
      title: '登录中...'
    })

    app.login()
      .then(() => {
        wx.hideLoading()
        wx.showToast({
          title: '登录成功',
          icon: 'success'
        })
        setTimeout(() => {
          wx.switchTab({
            url: '/pages/index/index'
          })
        }, 1500)
      })
      .catch((err) => {
        wx.hideLoading()
        console.error('登录失败:', err)
        wx.showToast({
          title: err?.message || '登录失败',
          icon: 'none'
        })
      })
  },
})
