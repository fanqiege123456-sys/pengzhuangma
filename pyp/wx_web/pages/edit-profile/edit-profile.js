// pages/edit-profile/edit-profile.js
const api = require('../../utils/api.js')

Page({
  data: {
    nickname: '',
    gender: '',
    age: null,
    ageIndex: 0,
    ageRange: [],
    bio: '',
    bioLength: 0,
    wechatNo: '',
    userInfo: null
  },

  onLoad() {
    const app = getApp()
    
    // 生成年龄范围 18-100
    const ageRange = []
    for (let i = 18; i <= 100; i++) {
      ageRange.push(i)
    }
    this.setData({ ageRange })
    
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
    
    this.loadUserInfo()
  },

  // 加载用户信息
  async loadUserInfo() {
    try {
      const res = await api.getUserInfo()
      
      if (res.data.code === 200) {
        const userInfo = res.data.data
        const age = userInfo.age || null
        const ageIndex = age ? age - 18 : 0
        
        this.setData({
          userInfo,
          nickname: userInfo.nickname || '',
          gender: userInfo.gender || 0,
          age: age,
          ageIndex: ageIndex,
          bio: userInfo.bio || '',
          bioLength: (userInfo.bio || '').length,
          wechatNo: userInfo.wechat_no || ''
        })
      }
    } catch (error) {
      console.error('加载用户信息失败', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    }
  },

  // 昵称输入
  onNicknameInput(e) {
    this.setData({
      nickname: e.detail.value
    })
  },

  // 选择性别
  selectGender(e) {
    const gender = parseInt(e.currentTarget.dataset.gender)
    this.setData({ gender })
  },

  // 年龄选择
  onAgeChange(e) {
    const index = e.detail.value
    const age = this.data.ageRange[index]
    this.setData({
      ageIndex: index,
      age: age
    })
  },

  // 简介输入
  onBioInput(e) {
    const bio = e.detail.value
    this.setData({
      bio: bio,
      bioLength: bio.length
    })
  },

  // 微信号输入
  onWechatNoInput(e) {
    this.setData({
      wechatNo: e.detail.value
    })
  },

  // 保存资料
  async saveProfile() {
    const { nickname, gender, age, bio, wechatNo } = this.data
    
    // 验证必填项
    if (!nickname.trim()) {
      wx.showToast({
        title: '请输入昵称',
        icon: 'none'
      })
      return
    }

    if (!gender || gender === 0) {
      wx.showToast({
        title: '请选择性别',
        icon: 'none'
      })
      return
    }

    wx.showLoading({
      title: '保存中...'
    })

    try {
      const updateData = {
        nickname: nickname.trim(),
        gender: gender, // gender已经是数字了
        age: age || 0,
        bio: bio.trim()
      }

      // 微信号可选
      if (wechatNo.trim()) {
        updateData.wechat_no = wechatNo.trim()
      }

      const res = await api.updateProfile(updateData)
      
      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '保存成功',
          icon: 'success',
          duration: 1500
        })

        // 更新全局用户信息
        const app = getApp()
        app.globalData.userInfo = res.data.data
        wx.setStorageSync('userInfo', res.data.data)

        // 延迟返回上一页
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
      console.error('保存资料失败', error)
      wx.showToast({
        title: '保存失败,请重试',
        icon: 'none'
      })
    }
  },

  // 下拉刷新
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    try {
      await this.loadUserInfo()
      console.log('下拉刷新完成')
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  }
})
