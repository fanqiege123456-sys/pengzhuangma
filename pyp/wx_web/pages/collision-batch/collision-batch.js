// pages/collision-batch/collision-batch.js
const api = require('../../utils/api.js')

Page({
  data: {
    tag: '',
    gender: 0,
    ageMin: 18,
    ageMax: 35,
    locations: [],
    selectedLocations: [],
    allSelected: false,
    combinations: [],
    totalCost: 0,
    costPerCode: 10 // 每个碰撞码消耗10个碰撞币
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
    try {
      const res = await api.getLocations()
      if (res.data.code === 200) {
        this.setData({
          locations: res.data.data || []
        })
      }
    } catch (error) {
      console.error('加载地址失败', error)
      wx.showToast({
        title: '加载地址失败',
        icon: 'none'
      })
    }
  },

  // 标签输入
  onTagInput(e) {
    this.setData({
      tag: e.detail.value
    })
  },

  // 选择性别
  selectGender(e) {
    this.setData({
      gender: parseInt(e.currentTarget.dataset.gender)
    })
  },

  // 年龄输入
  onAgeMinInput(e) {
    this.setData({
      ageMin: parseInt(e.detail.value) || 18
    })
  },

  onAgeMaxInput(e) {
    this.setData({
      ageMax: parseInt(e.detail.value) || 35
    })
  },

  // 地址选择变化
  onLocationChange(e) {
    const values = e.detail.value
    this.setData({
      selectedLocations: values,
      allSelected: values.length === this.data.locations.length
    })
  },

  // 全选/取消全选
  toggleSelectAll(e) {
    const checked = e.detail.value.length > 0
    if (checked) {
      // 全选
      this.setData({
        selectedLocations: this.data.locations.map(loc => loc.id.toString()),
        allSelected: true
      })
    } else {
      // 取消全选
      this.setData({
        selectedLocations: [],
        allSelected: false
      })
    }
  },

  // 跳转到添加地址页面
  goToAddLocation() {
    wx.navigateTo({
      url: '/pages/location-edit/location-edit'
    })
  },

  // 生成组合列表
  generateCombinations() {
    const { tag, gender, ageMin, ageMax, locations, selectedLocations } = this.data

    if (selectedLocations.length === 0) {
      wx.showToast({
        title: '请至少选择一个地址',
        icon: 'none'
      })
      return
    }

    // 验证年龄范围
    if (ageMin > ageMax) {
      wx.showToast({
        title: '最小年龄不能大于最大年龄',
        icon: 'none'
      })
      return
    }

    // 生成组合
    const combinations = []
    selectedLocations.forEach(locationId => {
      const location = locations.find(loc => loc.id.toString() === locationId)
      if (location) {
        combinations.push({
          country: location.country,
          province: location.province,
          city: location.city,
          district: location.district,
          tag: tag.trim(),
          gender: gender,
          age_min: ageMin,
          age_max: ageMax,
          cost_coins: this.data.costPerCode
        })
      }
    })

    const totalCost = combinations.length * this.data.costPerCode

    this.setData({
      combinations,
      totalCost
    })

    wx.showToast({
      title: `已生成${combinations.length}个组合`,
      icon: 'success'
    })
  },

  // 删除组合
  deleteCombination(e) {
    const index = e.currentTarget.dataset.index
    const combinations = this.data.combinations
    combinations.splice(index, 1)

    const totalCost = combinations.length * this.data.costPerCode

    this.setData({
      combinations,
      totalCost
    })

    wx.showToast({
      title: '已删除',
      icon: 'success'
    })
  },

  // 批量提交
  async submitBatch() {
    const app = getApp()
    
    // 首先检查登录状态
    if (!app.globalData.hasLogin) {
      wx.showModal({
        title: '需要登录',
        content: '请先登录后再提交碰撞',
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
    
    const { combinations, totalCost } = this.data

    if (combinations.length === 0) {
      wx.showToast({
        title: '请先生成组合列表',
        icon: 'none'
      })
      return
    }

    // 确认提交
    wx.showModal({
      title: '确认提交',
      content: `将提交 ${combinations.length} 个碰撞码,消耗 ${totalCost} 碰撞币`,
      success: async (res) => {
        if (res.confirm) {
          wx.showLoading({
            title: '提交中...'
          })

          try {
            const response = await api.batchSubmitCollisionCodes({
              codes: combinations
            })

            wx.hideLoading()

            if (response.data.code === 200) {
              wx.showModal({
                title: '提交成功',
                content: response.data.data.message || `成功提交${combinations.length}个碰撞码`,
                showCancel: false,
                success: () => {
                  // 清空列表
                  this.setData({
                    combinations: [],
                    totalCost: 0,
                    selectedLocations: [],
                    allSelected: false
                  })
                  
                  // 可选:跳转到匹配列表
                  // wx.switchTab({ url: '/pages/collision/collision' })
                }
              })
            } else {
              wx.showToast({
                title: response.data.msg || '提交失败',
                icon: 'none'
              })
            }
          } catch (error) {
            wx.hideLoading()
            console.error('批量提交失败', error)
            
            const errorMsg = error.response?.data?.msg || '提交失败,请重试'
            wx.showToast({
              title: errorMsg,
              icon: 'none',
              duration: 3000
            })
          }
        }
      }
    })
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
