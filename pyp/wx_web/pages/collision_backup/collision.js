// pages/collision/collision.js
Page({
  data: {
    inputCode: '',
    currentCode: ''
  },

  onLoad() {
    console.log('collision page loaded')
  },

  onInputChange(e) {
    this.setData({
      inputCode: e.detail.value
    })
  },

  startCollision() {
    const { inputCode } = this.data
    if (!inputCode.trim()) {
      wx.showToast({
        title: '请输入碰撞码',
        icon: 'none'
      })
      return
    }
    
    wx.showToast({
      title: '碰撞成功!',
      icon: 'success'
    })
  },

  generateCode() {
    const code = 'COLL' + Math.random().toString(36).substr(2, 6).toUpperCase()
    this.setData({
      currentCode: code
    })
    wx.showToast({
      title: '生成成功!',
      icon: 'success'
    })
  },

  copyCode() {
    const { currentCode } = this.data
    wx.setClipboardData({
      data: currentCode,
      success: () => {
        wx.showToast({
          title: '已复制',
          icon: 'success'
        })
      }
    })
  }
})
