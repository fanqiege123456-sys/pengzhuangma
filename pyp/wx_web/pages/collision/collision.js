// pages/collision/collision.js
const api = require('../../utils/api.js')

Page({
  data: {
    inputTag: '',
    hotTags: [],
    autoFocus: false,
    inputShake: false  // 输入框震动状态
  },

  onLoad(options) {
    // 加载热门标签
    this.loadHotTags()
    
    // 检查是否从其他页面传入关键词
    if (options.keyword) {
      this.setData({
        inputTag: decodeURIComponent(options.keyword)
      })
    } else if (options.tag) {
      // 支持从搜索结果页面传入的tag参数
      this.setData({
        inputTag: decodeURIComponent(options.tag)
      })
    }
  },

  onShow() {
    this.loadHotTags()
  },

  onPullDownRefresh() {
    this.loadHotTags().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 加载热门标签
  async loadHotTags() {
    try {
      const res = await api.getHotTagsAll()
      if (res.data.code === 200) {
        this.setData({
          hotTags: res.data.data || []
        })
      }
    } catch (error) {
      console.error('加载热门标签失败', error)
    }
  },

  // 标签输入
  onTagInput(e) {
    this.setData({
      inputTag: e.detail.value
    })
  },

  // 选择热门标签
  selectHotTag(e) {
    const tag = e.currentTarget.dataset.tag
    
    // 震动反馈
    wx.vibrateShort({
      type: 'light'
    })
    
    // 输入框震动动画
    this.setData({
      inputShake: true
    })
    
    // 显示提示
    wx.showToast({
      title: `已选择「${tag}」`,
      icon: 'success',
      duration: 1500
    })
    
    this.setData({
      inputTag: tag
    })
    
    // 清除震动动画
    setTimeout(() => {
      this.setData({
        inputShake: false
      })
    }, 600)
  },

  // 开始碰撞
  async startCollision() {
    const app = getApp()
    
    // 检查登录状态
    if (!app.globalData.hasLogin) {
      wx.showModal({
        title: '需要登录',
        content: '请先登录后再发起碰撞',
        confirmText: '去登录',
        cancelText: '取消',
        success: (res) => {
          if (res.confirm) {
            app.requireLogin(true)
          }
        }
      })
      return
    }
    
    const { inputTag } = this.data
    
    // 验证关键词
    if (!inputTag.trim()) {
      wx.showToast({
        title: '请输入关键词',
        icon: 'none'
      })
      return
    }

    wx.showLoading({ title: '碰撞中...' })

    try {
      // 简化的请求参数，只需要关键词
      const params = {
        tag: inputTag.trim(),
        cost_coins: 10
      }

      const res = await api.submitCollisionCode(params)
      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '碰撞码已发布并参与匹配',
          icon: 'success',
          duration: 2000
        })
        this.setData({ inputTag: '' })
        setTimeout(() => {
          wx.navigateTo({
            url: `/pages/collision-result/collision-result?keyword=${encodeURIComponent(inputTag.trim())}`
          })
        }, 2000)
      } else {
        const errorMsg = res.data.msg || res.data.message || '碰撞失败'
        
        // 处理积分不足
        if (errorMsg.includes('不足') || errorMsg.includes('Insufficient')) {
          wx.showModal({
            title: '积分不足',
            content: errorMsg,
            confirmText: '去充值',
            cancelText: '取消',
            success: (modalRes) => {
              if (modalRes.confirm) {
                wx.navigateTo({
                  url: '/pages/recharge/recharge'
                })
              }
            }
          })
        } else {
          wx.showToast({
            title: errorMsg,
            icon: 'none'
          })
        }
      }
    } catch (error) {
      wx.hideLoading()
      console.error('碰撞失败', error)
      wx.showToast({
        title: '碰撞失败，请重试',
        icon: 'none'
      })
    }
  },

  // 分享
  onShareAppMessage() {
    return {
      title: '来碰撞火花，找到志同道合的人！',
      path: '/pages/collision/collision'
    }
  },

  onShareTimeline() {
    return {
      title: '碰撞火花 - 用关键词找到你的知己 🎯'
    }
  }
})
