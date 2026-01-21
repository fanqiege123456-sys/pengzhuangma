// pages/locations/locations.js
const api = require('../../utils/api.js')

Page({
  data: {
    locations: [],
    labelIcons: {
      home: '🏠',
      school: '🎓',
      work: '💼',
      other: '📍'
    },
    labelTexts: {
      home: '老家',
      school: '学校',
      work: '工作地',
      other: '其他'
    }
  },

  onLoad() {
    const app = getApp()
    // 如果已登录,加载数据
    if (app.globalData.hasLogin) {
      this.loadLocations()
    }
    // 不再强制登录,允许浏览
  },

  onShow() {
    const app = getApp()
    // 如果已登录,刷新数据
    if (app.globalData.hasLogin) {
      this.loadLocations()
    }
    // 不再自动弹出登录提示,允许用户浏览
  },

  // 加载地址列表
  async loadLocations() {
    wx.showLoading({
      title: '加载中...'
    })

    try {
      const res = await api.getLocations()
      wx.hideLoading()

      if (res.data.code === 200) {
        this.setData({
          locations: res.data.data || []
        })
      } else {
        wx.showToast({
          title: res.data.msg || '加载失败',
          icon: 'none'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('加载地址失败', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    }
  },

  // 添加地址
  addLocation() {
    wx.navigateTo({
      url: '/pages/location-edit/location-edit'
    })
  },

  // 编辑地址
  editLocation(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/location-edit/location-edit?id=${id}`
    })
  },

  // 确认删除
  confirmDelete(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认删除',
      content: '确定要删除这个地址吗?',
      confirmColor: '#FF3B30',
      success: (res) => {
        if (res.confirm) {
          this.deleteLocation(id)
        }
      }
    })
  },

  // 删除地址
  async deleteLocation(id) {
    wx.showLoading({
      title: '删除中...'
    })

    try {
      const res = await api.deleteLocation(id)
      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '删除成功',
          icon: 'success'
        })
        this.loadLocations()
      } else {
        wx.showToast({
          title: res.data.msg || '删除失败',
          icon: 'none'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('删除地址失败', error)
      wx.showToast({
        title: '删除失败',
        icon: 'none'
      })
    }
  },

  // 设为默认
  async setDefault(e) {
    const id = e.currentTarget.dataset.id
    
    wx.showLoading({
      title: '设置中...'
    })

    try {
      const res = await api.setDefaultLocation(id)
      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '设置成功',
          icon: 'success'
        })
        this.loadLocations()
      } else {
        wx.showToast({
          title: res.data.msg || '设置失败',
          icon: 'none'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('设置默认地址失败', error)
      wx.showToast({
        title: '设置失败',
        icon: 'none'
      })
    }
  },

  // 下拉刷新
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    const app = getApp()
    
    try {
      // 如果已登录，刷新地址列表
      if (app.globalData.hasLogin) {
        await this.loadLocations()
        console.log('下拉刷新完成')
      } else {
        // 未登录状态，跳过数据刷新
        console.log('未登录状态，跳过数据刷新')
      }
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  }
})
