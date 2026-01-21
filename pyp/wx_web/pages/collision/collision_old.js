// pages/collision/collision.js
const api = require('../../utils/api.js')

Page({
  data: {
    // 我的地区信息
    myLocationText: '未设置',
    locationVisible: true,      // 地区是否可见
    
    // 碰撞参数
    inputTag: '',              // 兴趣标签
    genderMale: false,         // 期望男性
    genderFemale: false,       // 期望女性
    searchRegion: [],          // 搜索地区 [省份, 城市, 区县]
    searchRegionText: '选择地区',
    
    // 其他
    hotTags: [],              // 热门标签列表
    showHaidilao: false,      // 是否显示海底捞按钮
    haidilaoCount: 0,         // 可海底捞用户数量
    currentTag: '',           // 当前碰撞标签
    
    // 用户信息
    userInfo: null
  },

  onLoad() {
    console.log('collision page loaded')
    this.loadUserInfo()
    this.loadHotTags()
    this.loadSavedSettings()
  },

  onShow() {
    this.loadUserInfo()
    this.loadHotTags()
  },

  // 加载用户信息
  async loadUserInfo() {
    try {
      const res = await api.getUserInfo()
      if (res.data.code === 200) {
        const user = res.data.data
        this.setData({
          userInfo: user,
          locationVisible: user.location_visible !== false,
          myLocationText: this.formatLocation(user)
        })
      }
    } catch (error) {
      console.error('加载用户信息失败', error)
    }
  },

  // 格式化地址显示
  formatLocation(user) {
    const parts = []
    if (user.country) parts.push(user.country)
    if (user.province) parts.push(user.province)
    if (user.city && user.city !== user.province) parts.push(user.city)
    if (user.district) parts.push(user.district)
    
    return parts.length > 0 ? parts.join('、') : '未设置'
  },

  // 切换地区可见性
  async toggleLocationVisible() {
    const newVisible = !this.data.locationVisible
    
    try {
      await api.updateUserInfo({
        location_visible: newVisible
      })
      
      this.setData({
        locationVisible: newVisible
      })
      
      wx.showToast({
        title: newVisible ? '地区已设为可见' : '地区已隐藏',
        icon: 'success'
      })
    } catch (error) {
      console.error('更新地区可见性失败', error)
      wx.showToast({
        title: '设置失败',
        icon: 'none'
      })
    }
  },

  // 跳转到设置页面
  goToSettings() {
    wx.navigateTo({
      url: '/pages/settings/settings'
    })
  },

  // 加载热门标签
  async loadHotTags() {
    try {
      const res = await api.getHotCodes()
      if (res.data.code === 200) {
        this.setData({
          hotTags: res.data.data || []
        })
      }
    } catch (error) {
      console.error('加载热门标签失败', error)
    }
  },

  // 加载保存的设置
  loadSavedSettings() {
    const savedSettings = wx.getStorageSync('collisionSettings') || {}
    this.setData({
      genderMale: savedSettings.genderMale || false,
      genderFemale: savedSettings.genderFemale || false,
      searchRegion: savedSettings.searchRegion || [],
      searchRegionText: this.formatSearchRegion(savedSettings.searchRegion || [])
    })
  },

  // 保存设置
  saveSettings() {
    const { genderMale, genderFemale, searchRegion } = this.data
    wx.setStorageSync('collisionSettings', {
      genderMale,
      genderFemale,
      searchRegion
    })
  },

  // 标签输入
  onTagInput(e) {
    this.setData({
      inputTag: e.detail.value
    })
  },

  // 切换男性选择
  toggleGenderMale() {
    const newValue = !this.data.genderMale
    this.setData({
      genderMale: newValue
    })
    this.saveSettings()
  },

  // 切换女性选择
  toggleGenderFemale() {
    const newValue = !this.data.genderFemale
    this.setData({
      genderFemale: newValue
    })
    this.saveSettings()
  },

  // 搜索地区选择
  onSearchRegionChange(e) {
    const region = e.detail.value
    this.setData({
      searchRegion: region,
      searchRegionText: this.formatSearchRegion(region)
    })
    this.saveSettings()
  },

  // 格式化搜索地区显示
  formatSearchRegion(region) {
    if (!region || region.length === 0) {
      return '选择地区'
    }
    return region.filter(r => r).join('、')
  },

  // 选择热门标签
  selectHotTag(e) {
    const tag = e.currentTarget.dataset.tag
    this.setData({
      inputTag: tag
    })
  },

  // 开始碰撞
  async startCollision() {
    const { inputTag, genderMale, genderFemale, searchRegion, userInfo } = this.data
    
    // 验证用户是否设置了地区
    if (!userInfo || !userInfo.country) {
      wx.showModal({
        title: '提示',
        content: '请先在个人设置中完善您的地址信息',
        confirmText: '去设置',
        success: (res) => {
          if (res.confirm) {
            wx.navigateTo({
              url: '/pages/settings/settings'
            })
          }
        }
      })
      return
    }

    // 验证标签
    if (!inputTag.trim()) {
      wx.showToast({
        title: '请输入兴趣标签',
        icon: 'none'
      })
      return
    }

    // 验证搜索地区
    if (searchRegion.length === 0) {
      wx.showToast({
        title: '请选择期望匹配地区',
        icon: 'none'
      })
      return
    }

    wx.showLoading({
      title: '提交中...'
    })

    try {
      // 构建请求参数
      const requestData = {
        tag: inputTag.trim(),
        country: '中国',
        cost_coins: 10
      }

      // 添加搜索地区参数
      if (searchRegion[0]) requestData.province = searchRegion[0]
      if (searchRegion[1]) requestData.city = searchRegion[1]
      if (searchRegion[2]) requestData.district = searchRegion[2]

      // 添加性别筛选
      if (genderMale && !genderFemale) {
        requestData.gender = 1  // 只选男
      } else if (!genderMale && genderFemale) {
        requestData.gender = 2  // 只选女
      } else {
        requestData.gender = 0  // 不限或都选
      }

      console.log('提交碰撞数据:', requestData)
      const res = await api.submitCollisionCode(requestData)

      wx.hideLoading()

      if (res.data.code === 200) {
        const result = res.data.data
        
        // 检查是否可以海底捞
        if (result.can_haidilao && result.haidilao_count > 0) {
          this.setData({
            showHaidilao: true,
            haidilaoCount: result.haidilao_count,
            currentTag: inputTag.trim()
          })
        }

        wx.showToast({
          title: '提交成功，等待匹配',
          icon: 'success'
        })
        
        // 重置标签输入
        this.setData({
          inputTag: ''
        })
      } else {
        wx.showToast({
          title: res.data.message || '提交失败',
          icon: 'none'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('提交碰撞失败', error)
      wx.showToast({
        title: '提交失败，请重试',
        icon: 'none'
      })
    }
  },

  // 海底捞
  async doHaidilao() {
    const { currentTag } = this.data
    
    if (!currentTag) {
      wx.showToast({
        title: '请先进行碰撞',
        icon: 'none'
      })
      return
    }

    // 确认对话框
    wx.showModal({
      title: '确认海底捞',
      content: `确认花费100积分海底捞一个"${currentTag}"标签的用户吗？`,
      confirmText: '确认',
      cancelText: '取消',
      success: async (modalRes) => {
        if (modalRes.confirm) {
          wx.showLoading({
            title: '海底捞中...'
          })

          try {
            const res = await api.haidilao({
              tag: currentTag,
              cost_coins: 100
            })

            wx.hideLoading()

            if (res.data.code === 200) {
              const result = res.data.data
              
              // 隐藏海底捞按钮
              this.setData({
                showHaidilao: false,
                haidilaoCount: 0,
                currentTag: ''
              })

              wx.showModal({
                title: '海底捞成功！',
                content: `已添加 ${result.friend.nickname} 为好友`,
                showCancel: false,
                success: () => {
                  // 跳转到好友页面
                  wx.switchTab({
                    url: '/pages/friends/friends'
                  })
                }
              })
            } else {
              wx.showToast({
                title: res.data.msg || '海底捞失败',
                icon: 'none'
              })
            }
          } catch (error) {
            wx.hideLoading()
            console.error('海底捞失败', error)
            wx.showToast({
              title: '网络错误，请重试',
              icon: 'none'
            })
          }
        }
      }
    })
  }
})
