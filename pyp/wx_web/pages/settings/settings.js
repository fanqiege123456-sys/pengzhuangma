// pages/settings/settings.js
const api = require('../../utils/api.js')

Page({
  data: {
    userInfo: {},
    locationInfo: {
      country: '中国',
      province: '',
      city: '',
      district: ''
    },
    allowUpperLevel: false,
    allowForceAdd: false,
    allowHaidilao: false,
    emailVisible: true,
    emailVerified: false
  },

  onLoad() {
    this.loadUserInfo()
    this.loadContactInfo()
  },

  onShow() {
    this.loadUserInfo()
    this.loadContactInfo()
  },

  // 加载用户信息
  async loadUserInfo() {
    try {
      const res = await api.getUserInfo()
      if (res.data.code === 200) {
        const user = res.data.data
        this.setData({
          userInfo: user,
          locationInfo: {
            country: user.country || '中国',
            province: user.province || '',
            city: user.city || '',
            district: user.district || ''
          },
          allowUpperLevel: user.allow_upper_level || false,
          allowForceAdd: user.allow_force_add || false,
          allowHaidilao: user.allow_haidilao || false
        })
      }
    } catch (error) {
      console.error('加载用户信息失败', error)
    }
  },

  // 加载联系方式信息
  async loadContactInfo() {
    try {
      const res = await api.getUserContacts()
      if (res.data.code === 200 && res.data.data) {
        this.setData({
          emailVerified: res.data.data.email_verified || false,
          emailVisible: res.data.data.email_visible !== false // 默认为 true
        })
      }
    } catch (error) {
      console.error('加载联系方式失败', error)
    }
  },

  // 地区选择
  onRegionChange(e) {
    const [province, city, district] = e.detail.value
    this.setData({
      'locationInfo.province': province,
      'locationInfo.city': city,
      'locationInfo.district': district
    })
  },

  // 切换允许上级匹配
  toggleUpperLevel(e) {
    this.setData({
      allowUpperLevel: e.detail.value
    })
  },

  // 切换允许强制添加好友
  async toggleForceAdd(e) {
    const value = e.detail.value
    const oldValue = this.data.allowForceAdd
    
    // 先更新 UI
    this.setData({
      allowForceAdd: value
    })

    try {
      const res = await api.updateProfile({
        allow_force_add: value
      })

      if (res.data.code === 200) {
        wx.showToast({
          title: '设置已更新',
          icon: 'success'
        })
      } else {
        // 恢复原状态
        this.setData({
          allowForceAdd: oldValue
        })
        wx.showToast({
          title: res.data.msg || '更新失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('更新设置失败', error)
      // 恢复原状态
      this.setData({
        allowForceAdd: oldValue
      })
      wx.showToast({
        title: '网络错误',
        icon: 'none'
      })
    }
  },

  // 切换允许被海底捞
  async toggleHaidilao(e) {
    const value = e.detail.value
    const oldValue = this.data.allowHaidilao

    // 先更新 UI
    this.setData({
      allowHaidilao: value
    })

    try {
      const res = await api.updateProfile({
        allow_haidilao: value
      })

      if (res.data.code === 200) {
        wx.showToast({
          title: '设置已更新',
          icon: 'success'
        })
      } else {
        // 恢复原状态
        this.setData({
          allowHaidilao: oldValue
        })
        wx.showToast({
          title: res.data.msg || '更新失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('更新设置失败', error)
      // 恢复原状态
      this.setData({
        allowHaidilao: oldValue
      })
      wx.showToast({
        title: '网络错误',
        icon: 'none'
      })
    }
  },

  // 切换邮箱显示
  async toggleEmailVisible(e) {
    const value = e.detail.value
    const oldValue = this.data.emailVisible

    // 先更新 UI
    this.setData({
      emailVisible: value
    })

    try {
      const res = await api.updateEmailVisibility({
        email_visible: value
      })

      if (res.data.code === 200) {
        wx.showToast({
          title: '设置已更新',
          icon: 'success'
        })
      } else {
        // 恢复原状态
        this.setData({
          emailVisible: oldValue
        })
        wx.showToast({
          title: res.data.msg || '更新失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('更新设置失败', error)
      // 恢复原状态
      this.setData({
        emailVisible: oldValue
      })
      wx.showToast({
        title: '网络错误',
        icon: 'none'
      })
    }
  },

  // 跳转到邮箱认证页面
  goBindEmail() {
    wx.navigateTo({
      url: '/pages/bind-email/bind-email'
    })
  },

  // 保存设置
  async saveSettings() {
    const { locationInfo, allowUpperLevel } = this.data

    if (!locationInfo.province || !locationInfo.city || !locationInfo.district) {
      wx.showToast({
        title: '请完善地址信息',
        icon: 'none'
      })
      return
    }

    wx.showLoading({
      title: '保存中...'
    })

    try {
      const res = await api.updateUserLocation({
        country: locationInfo.country,
        province: locationInfo.province,
        city: locationInfo.city,
        district: locationInfo.district,
        allow_upper_level: allowUpperLevel
      })

      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '保存成功',
          icon: 'success',
          duration: 1500
        })
        
        // 更新全局用户信息
        const app = getApp()
        if (app.globalData.userInfo) {
          Object.assign(app.globalData.userInfo, res.data.data)
        }

        // 保存成功后返回
        setTimeout(() => {
          wx.navigateBack()
        }, 1500)
      } else {
        wx.showToast({
          title: res.data.msg || '保存失败',
          icon: 'none'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('保存设置失败', error)
      wx.showToast({
        title: '网络错误，请重试',
        icon: 'none'
      })
    }
  },

  // 跳转到调试面板
  goToDebug() {
    wx.navigateTo({
      url: '/pages/debug/debug',
      success: () => {
        console.log('跳转到调试面板')
      },
      fail: (err) => {
        console.error('跳转调试面板失败:', err)
        wx.showToast({
          title: '调试面板不可用',
          icon: 'none'
        })
      }
    })
  },

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
          
          // 跳转到首页
          setTimeout(() => {
            wx.reLaunch({
              url: '/pages/index/index'
            })
          }, 1500)
        }
      }
    })
  },

  // 下拉刷新
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    try {
      await Promise.all([
        this.loadUserInfo(),
        this.loadContactInfo()
      ])
      console.log('下拉刷新完成')
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  }
})