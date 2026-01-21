// pages/edit-collision-code/edit-collision-code.js
const api = require('../../utils/api.js')

Page({
  data: {
    collisionCode: {},
    tag: '',
    selectedValidityIndex: 0,
    validityOptions: [
      { days: 1, costCoins: 10, label: '1天 (10金币)' },
      { days: 3, costCoins: 25, label: '3天 (25金币)' },
      { days: 7, costCoins: 50, label: '7天 (50金币)' },
      { days: 14, costCoins: 90, label: '14天 (90金币)' },
      { days: 30, costCoins: 150, label: '30天 (150金币)' }
    ],
    canSubmit: false,
    submitting: false
  },

  onLoad(options) {
    const codeId = options.id
    if (codeId) {
      this.setData({ codeId })
      this.getCollisionCodeDetail(codeId)
    } else {
      wx.showToast({
        title: '参数错误',
        icon: 'none'
      })
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    }
  },

  onTagInput(e) {
    this.setData({
      tag: e.detail.value
    })
    this.checkSubmit()
  },

  onValidityChange(e) {
    this.setData({
      selectedValidityIndex: parseInt(e.detail.value)
    })
    this.checkSubmit()
  },

  checkSubmit() {
    const { tag, collisionCode } = this.data
    // 只要有关键词就可以提交（可以是原关键词或新关键词）
    const canSubmit = (tag && tag.trim()) || (collisionCode.tag && collisionCode.tag.trim())
    this.setData({ canSubmit })
  },

  async getCollisionCodeDetail(codeId) {
    wx.showLoading({ title: '加载中...' })
    
    try {
      const res = await api.getCollisionCode(codeId)
      
      if (res.data.code === 200) {
        const collisionCode = res.data.data
        this.setData({
          collisionCode,
          tag: collisionCode.tag || ''
        })
        this.checkSubmit()
      } else {
        wx.showToast({
          title: res.data.msg || '加载失败',
          icon: 'none'
        })
        setTimeout(() => {
          wx.navigateBack()
        }, 1500)
      }
    } catch (error) {
      console.error('获取碰撞码详情失败:', error)
      wx.showToast({
        title: '网络错误',
        icon: 'none'
      })
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    } finally {
      wx.hideLoading()
    }
  },

  async submitForm(e) {
    if (this.data.submitting) return
    
    const { tag, selectedValidityIndex, validityOptions, collisionCode } = this.data
    const selectedOption = validityOptions[selectedValidityIndex]
    
    // 验证关键词
    const finalTag = tag.trim() || collisionCode.tag
    if (!finalTag) {
      wx.showToast({
        title: '请输入关键词',
        icon: 'none'
      })
      return
    }

    // 检查是否有变化
    const hasTagChange = finalTag !== collisionCode.tag
    const hasValidityChange = selectedOption.days > 0
    
    if (!hasTagChange && !hasValidityChange) {
      wx.showToast({
        title: '没有修改内容',
        icon: 'none'
      })
      return
    }

    this.setData({ submitting: true })
    wx.showLoading({ title: '保存中...' })

    try {
      const res = await api.updateCollisionCode(collisionCode.id, {
        tag: finalTag,
        days: selectedOption.days,
        cost_coins: selectedOption.costCoins
      })

      if (res.data.code === 200) {
        wx.showToast({
          title: '修改成功',
          icon: 'success'
        })
        
        // 延迟返回，让用户看到成功提示
        setTimeout(() => {
          wx.navigateBack()
        }, 1500)
      } else {
        wx.showToast({
          title: res.data.msg || '修改失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('修改碰撞码失败:', error)
      wx.showToast({
        title: '网络错误',
        icon: 'none'
      })
    } finally {
      wx.hideLoading()
      this.setData({ submitting: false })
    }
  },

  // 页面分享
  onShareAppMessage() {
    return {
      title: '碰撞码 - 修改关键词',
      path: '/pages/index/index'
    }
  }
})