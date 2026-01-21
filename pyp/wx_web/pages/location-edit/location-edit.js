// pages/location-edit/location-edit.js
const api = require('../../utils/api.js')

Page({
  data: {
    locationId: null,
    label: 'home',
    labelOptions: [
      { value: 'home', text: '老家', icon: '🏠' },
      { value: 'school', text: '学校', icon: '🎓' },
      { value: 'work', text: '工作地', icon: '💼' },
      { value: 'other', text: '其他', icon: '📍' }
    ],
    region: [],
    regionText: '',
    country: '',
    province: '',
    city: '',
    district: '',
    isDefault: false
  },

  onLoad(options) {
    const app = getApp()
    
    // 检查登录状态
    if (!app.globalData.hasLogin) {
      wx.showToast({
        title: '请先登录',
        icon: 'none',
        duration: 2000
      })
      setTimeout(() => {
        app.requireLogin(true)
      }, 2000)
      return
    }

    if (options.id) {
      this.setData({ locationId: options.id })
      this.loadLocationDetail(options.id)
      wx.setNavigationBarTitle({
        title: '编辑地址'
      })
    } else {
      wx.setNavigationBarTitle({
        title: '添加地址'
      })
    }
  },

  // 加载地址详情
  async loadLocationDetail(id) {
    wx.showLoading({
      title: '加载中...'
    })

    try {
      const res = await api.getLocations()
      wx.hideLoading()

      if (res.data.code === 200) {
        const location = res.data.data.find(item => item.id == id)
        if (location) {
          const region = []
          if (location.province) region.push(location.province)
          if (location.city) region.push(location.city)
          if (location.district) region.push(location.district)

          const regionText = [location.province, location.city, location.district]
            .filter(Boolean)
            .join('')

          this.setData({
            label: location.label,
            region: region,
            regionText: regionText,
            country: location.country,
            province: location.province,
            city: location.city,
            district: location.district,
            isDefault: location.is_default
          })
        }
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

  // 选择标签
  selectLabel(e) {
    const label = e.currentTarget.dataset.value
    this.setData({ label })
  },

  // 地区选择
  onRegionChange(e) {
    const region = e.detail.value
    this.setData({
      region: region,
      regionText: region.join(''),
      province: region[0] || '',
      city: region[1] || '',
      district: region[2] || '',
      country: '中国' // 默认中国
    })
  },

  // 默认地址切换
  onDefaultChange(e) {
    this.setData({
      isDefault: e.detail.value
    })
  },

  // 保存地址
  async saveLocation() {
    const { locationId, label, province, city, isDefault } = this.data

    // 验证
    if (!province || !city) {
      wx.showToast({
        title: '请选择地区',
        icon: 'none'
      })
      return
    }

    wx.showLoading({
      title: '保存中...'
    })

    try {
      const data = {
        label: label,
        country: this.data.country,
        province: this.data.province,
        city: this.data.city,
        district: this.data.district,
        is_default: isDefault
      }

      let res
      if (locationId) {
        // 更新
        res = await api.updateLocation(locationId, data)
      } else {
        // 创建
        res = await api.createLocation(data)
      }

      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '保存成功',
          icon: 'success',
          duration: 1500
        })

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
      console.error('保存地址失败', error)
      wx.showToast({
        title: '保存失败',
        icon: 'none'
      })
    }
  }
})
