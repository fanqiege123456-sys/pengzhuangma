// pages/profile/profile.js
const api = require('../../utils/api.js')

Page({
  data: {
    isLoggedIn: false,  // 显式登录状态标志
    userInfo: null,     // 改为 null 而不是 {}
    balance: 0,
    myCollisionCode: null,
    isBlackhole: false,
    collisionTab: 'active',  // 当前选中的tab: active/expired
    activeCollisions: [],    // 进行中的碰撞
    expiredCollisions: [],   // 已过期的碰撞
    userContact: {
      email: '',
      emailVerified: false,
      phone: '',
      phoneVerified: false
    }
  },

  onLoad() {
    const app = getApp()
    this.setData({
      isLoggedIn: app.globalData.hasLogin
    })
    // 如果已登录，加载数据
    if (app.globalData.hasLogin) {
      this.loadUserInfo()
      this.loadMyCollisionCode()
      this.loadUserContacts()
    }
  },

  onShow() {
    const app = getApp()
    this.setData({
      isLoggedIn: app.globalData.hasLogin
    })
    // 如果已登录，刷新数据
    if (app.globalData.hasLogin) {
      this.loadUserInfo()
      this.loadBalance()
      this.loadMyCollisionCode()
      this.loadMyCollisionList()
      this.loadUserContacts()
    }
  },

  // 手动触发登录
  handleLogin() {
    const app = getApp()
    app.requireLogin(true) // 显示登录弹窗
  },

  // 加载用户信息
  async loadUserInfo() {
    const app = getApp()
    
    try {
      // 从API获取最新用户信息
      const res = await api.getUserInfo()
      
      if (res.data.code === 200) {
        const userInfo = res.data.data
        this.setData({
          userInfo
        })
        
        // 更新全局用户信息
        app.globalData.userInfo = userInfo
        wx.setStorageSync('userInfo', userInfo)
      }
    } catch (error) {
      console.error('加载用户信息失败', error)
      
      // 使用本地缓存的用户信息
      const localUserInfo = app.globalData.userInfo || wx.getStorageSync('userInfo')
      if (localUserInfo) {
        this.setData({
          userInfo: localUserInfo
        })
      }
    }
  },

  // 加载余额
  async loadBalance() {
    try {
      const res = await api.getBalance()
      
      if (res.data.code === 200) {
        this.setData({
          balance: res.data.data.balance || 0 // 直接使用碰撞币数量
        })
      }
    } catch (error) {
      console.error('加载余额失败', error)
    }
  },

  // 加载我的碰撞码
  async loadMyCollisionCode() {
    try {
      const res = await api.getMyCollisionCode()
      
      if (res.data.code === 200 && res.data.data.has_code) {
        const code = res.data.data.code
        if (code && code.expires_at) {
          const expiresAt = new Date(code.expires_at)
          if (!Number.isNaN(expiresAt.getTime())) {
            code.status = expiresAt > new Date() ? 'active' : 'expired'
          }
        }
        this.setData({
          myCollisionCode: code,
          isBlackhole: res.data.data.is_blackhole
        })
      }
    } catch (error) {
      console.error('加载碰撞码失败', error)
    }
  },

  // 加载我的碰撞列表（按状态分类）
  async loadMyCollisionList() {
    try {
      const res = await api.getMyCollisionCodes()
      
      if (res.data.code === 200) {
        // 兼容不同的数据结构
        let allCodes = []
        if (Array.isArray(res.data.data)) {
          allCodes = res.data.data
        } else if (res.data.data && Array.isArray(res.data.data.codes)) {
          allCodes = res.data.data.codes
        } else if (res.data.data && Array.isArray(res.data.data.list)) {
          allCodes = res.data.data.list
        }
        
        const activeCollisions = []
        const expiredCollisions = []
        
        allCodes.forEach(item => {
          // 直接使用后端返回的 time_left 字段显示剩余时间
          item.expires_at_display = item.time_left || '已过期'
          
          // 格式化过期时间显示（只显示日期和时间，不显示完整时间戳）
          if (item.expires_at) {
            try {
              const expireDate = new Date(item.expires_at)
              if (!isNaN(expireDate.getTime())) {
                // 格式化为 MM-DD HH:mm
                const month = String(expireDate.getMonth() + 1).padStart(2, '0')
                const day = String(expireDate.getDate()).padStart(2, '0')
                const hour = String(expireDate.getHours()).padStart(2, '0')
                const minute = String(expireDate.getMinutes()).padStart(2, '0')
                item.expires_at_formatted = `${month}-${day} ${hour}:${minute}`
              } else {
                item.expires_at_formatted = '时间格式错误'
              }
            } catch (error) {
              item.expires_at_formatted = '时间解析失败'
            }
          } else {
            item.expires_at_formatted = '无过期时间'
          }
          
          if (item.is_expired) {
            item.display_status = '已过期'
            item.status_class = 'expired'
            expiredCollisions.push(item)
            return
          }
          
          if (item.collision_status === '已碰撞' || item.is_matched) {
            item.display_status = '已碰撞'
            item.status_class = 'matched'
            activeCollisions.push(item)
            return
          }
          
          item.display_status = '进行中'
          item.status_class = 'active'
          activeCollisions.push(item)
        })
        
        this.setData({
          activeCollisions,
          expiredCollisions
        })
      }
    } catch (error) {
      console.error('加载碰撞列表失败', error)
    }
  },

  // 切换碰撞记录tab
  switchCollisionTab(e) {
    const tab = e.currentTarget.dataset.tab
    this.setData({
      collisionTab: tab
    })
  },

  // 查看/编辑碰撞码详情
  openCollisionCode(e) {
    const id = e.currentTarget.dataset.id
    if (!id) return
    wx.navigateTo({
      url: `/pages/edit-collision-code/edit-collision-code?id=${id}`
    })
  },

  // 续期碰撞
  renewCollision(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '续期碰撞',
      content: '续期需要消耗10积分，是否继续？',
      success: async (res) => {
        if (res.confirm) {
          try {
            const result = await api.renewCollisionCode(id)
            if (result.data.code === 200) {
              wx.showToast({
                title: '续期成功',
                icon: 'success'
              })
              this.loadMyCollisionList()
              this.loadBalance()
            } else {
              wx.showToast({
                title: result.data.message || '续期失败',
                icon: 'none'
              })
            }
          } catch (error) {
            wx.showToast({
              title: '操作失败',
              icon: 'none'
            })
          }
        }
      }
    })
  },

  // 删除碰撞
  deleteCollision(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '删除碰撞',
      content: '确定要删除这条碰撞记录吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            const result = await api.deleteCollisionCode(id)
            if (result.data.code === 200) {
              wx.showToast({
                title: '删除成功',
                icon: 'success'
              })
              this.loadMyCollisionList()
            } else {
              wx.showToast({
                title: result.data.message || '删除失败',
                icon: 'none'
              })
            }
          } catch (error) {
            wx.showToast({
              title: '操作失败',
              icon: 'none'
            })
          }
        }
      }
    })
  },

  // 重新提交被拒绝的碰撞码
  resubmitCollision(e) {
    const item = e.currentTarget.dataset.item
    if (!item || !item.id) return
    
    wx.showModal({
      title: '重新提交',
      content: `确定要重新提交「${item.tag}」吗？重新提交仅影响首页展示审核。`,
      confirmText: '确定',
      cancelText: '取消',
      success: async (res) => {
        if (res.confirm) {
          try {
            wx.showLoading({ title: '提交中...' })
            const result = await api.resubmitCollisionCode(item.id)
            wx.hideLoading()
            
            if (result.data.code === 200) {
              wx.showToast({
                title: '已重新提交，首页展示待审核',
                icon: 'success',
                duration: 2000
              })
              this.loadMyCollisionList()
            } else {
              wx.showToast({
                title: result.data.message || '提交失败',
                icon: 'none'
              })
            }
          } catch (error) {
            wx.hideLoading()
            wx.showToast({
              title: '操作失败',
              icon: 'none'
            })
          }
        }
      }
    })
  },

  // 反转黑洞状态 - 重新提交碰撞
  reverseBlackhole(e) {
    const id = e?.currentTarget?.dataset?.id
    wx.showModal({
      title: '提示',
      content: '当前标签处于黑洞状态,其他用户无法看到你。是否重新发起碰撞?',
      confirmText: '重新碰撞',
      success: (res) => {
        if (res.confirm) {
          wx.navigateTo({
            url: '/pages/collision/collision'
          })
        }
      }
    })
  },

  // 更换头像
  changeAvatar() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath
        
        wx.showLoading({
          title: '上传中...'
        })
        
        // 直接使用图片临时URL作为头像地址，发送JSON请求
        api.updateProfile({ avatar: tempFilePath })
          .then((res) => {
            wx.hideLoading()
            if (res.data.code === 200) {
              wx.showToast({
                title: '头像更新成功',
                icon: 'success'
              })
              
              // 更新头像
              this.setData({
                'userInfo.avatar': res.data.data.avatar
              })
              
              // 更新全局用户信息
              const app = getApp()
              app.globalData.userInfo = res.data.data
              wx.setStorageSync('userInfo', res.data.data)
            } else {
              wx.showToast({
                title: res.data.msg || '头像更新失败',
                icon: 'none'
              })
            }
          })
          .catch((error) => {
            wx.hideLoading()
            console.error('头像更新失败', error)
            wx.showToast({
              title: '头像更新失败',
              icon: 'none'
            })
          })
      }
    })
  },

  // 编辑资料
  editProfile() {
    wx.navigateTo({
      url: '/pages/edit-profile/edit-profile'
    })
  },

  // 充值
  goRecharge() {
    wx.navigateTo({
      url: '/pages/recharge/recharge'
    })
  },

  // 我的好友
  goToFriends() {
    wx.switchTab({
      url: '/pages/friends/friends'
    })
  },

  // 碰撞记录
  goToCollisionHistory() {
    wx.switchTab({
      url: '/pages/collision/collision'
    })
  },

  // 充值记录
  goToRechargeHistory() {
    wx.navigateTo({
      url: '/pages/recharge-history/recharge-history'
    })
  },

  // 设置匹配条件
  setFriendConditions() {
    wx.navigateTo({
      url: '/pages/friend-conditions/friend-conditions'
    })
  },

  // 切换强制添加好友设置
  async toggleForceAdd(e) {
    const value = e.detail.value
    
    try {
      wx.showLoading({
        title: '更新中...'
      })

      const res = await api.updateProfile({
        allow_force_add: value
      })

      wx.hideLoading()

      if (res.data.code === 200) {
        this.setData({
          'userInfo.allow_force_add': value
        })
        
        wx.showToast({
          title: '设置已更新',
          icon: 'success'
        })
      } else {
        wx.showToast({
          title: res.data.msg || '更新失败',
          icon: 'error'
        })
        
        // 恢复开关状态
        this.setData({
          'userInfo.allow_force_add': !value
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('更新设置失败', error)
      wx.showToast({
        title: '网络错误',
        icon: 'error'
      })
      
      // 恢复开关状态
      this.setData({
        'userInfo.allow_force_add': !value
      })
    }
  },

  // 退出登录
  async logout() {
    try {
      const res = await wx.showModal({
        title: '确认退出',
        content: '确定要退出登录吗？'
      })

      if (res.confirm) {
        const app = getApp()
        app.logout()
      }
    } catch (error) {
      console.error('退出登录失败', error)
    }
  },

  // 我的匹配
  goMyMatches() {
    wx.switchTab({
      url: '/pages/normal-collision/normal-collision'
    })
  },

  // 地址管理
  goLocations() {
    wx.navigateTo({
      url: '/pages/locations/locations'
    })
  },

  // 设置
  goSettings() {
    wx.navigateTo({
      url: '/pages/settings/settings'
    })
  },

  // 帮助与反馈
  goHelp() {
    wx.showToast({
      title: '功能开发中',
      icon: 'none'
    })
  },

  // 我的碰撞码列表
  goMyCollisionCodes() {
    wx.navigateTo({
      url: '/pages/my-codes/my-codes'
    })
  },

  // 消费记录
  goConsumeRecords() {
    wx.navigateTo({
      url: '/pages/consume-records/consume-records'
    })
  },

  // 关于我们
  goAbout() {
    wx.showModal({
      title: '关于标签碰撞',
      content: '版本: V3.0\n一款基于关键词碰撞的社交小程序',
      showCancel: false
    })
  },

  // ========== V3.0 联系方式相关 ==========
  
  // 加载用户联系方式
  async loadUserContacts() {
    try {
      const res = await api.getUserContacts()
      if (res.data.code === 200 && res.data.data) {
        const { email, email_verified, phone, phone_verified } = res.data.data
        this.setData({
          userContact: {
            email: email || '',
            emailVerified: email_verified || false,
            phone: phone || '',
            phoneVerified: phone_verified || false
          }
        })
      }
    } catch (error) {
      console.error('加载联系方式失败', error)
    }
  },

  // 绑定邮箱
  handleEmailBind() {
    wx.navigateTo({
      url: '/pages/email-bind/email-bind'
    })
  },

  // 绑定手机号
  handlePhoneBind() {
    // 弹出获取手机号授权弹窗
    wx.showModal({
      title: '绑定手机号',
      content: '需要获取您的手机号进行绑定，是否授权？',
      success: (res) => {
        if (res.confirm) {
          // 调用微信获取手机号API
          wx.getUserProfile({
            desc: '用于绑定手机号',
            success: (userRes) => {
              // 这里需要调用微信手机号授权API，在实际项目中应该使用<button open-type="getPhoneNumber">
              // 为了测试，我们直接跳转到手机号绑定页面（如果有）
              wx.showToast({
                title: '手机号绑定功能开发中',
                icon: 'none'
              })
            }
          })
        }
      }
    })
  },

  // 获取手机号
  getPhoneNumber(e) {
    if (e.detail.errMsg !== 'getPhoneNumber:ok') {
      wx.showToast({
        title: '未授权手机号',
        icon: 'none'
      })
      return
    }

    if (!e.detail.code) {
      wx.showToast({
        title: '获取手机号失败',
        icon: 'none'
      })
      return
    }

    const payload = {
      code: e.detail.code
    }
    if (e.detail.encryptedData && e.detail.iv) {
      payload.encrypted_data = e.detail.encryptedData
      payload.iv = e.detail.iv
    }

    api.bindPhone(payload).then(res => {
      if (res.data.code === 200) {
        wx.showToast({
          title: '绑定成功',
          icon: 'success'
        })
        this.loadUserContacts()
      } else {
        wx.showToast({
          title: res.data.message || '绑定失败',
          icon: 'none'
        })
      }
    }).catch(err => {
      console.error('绑定手机号失败', err)
      wx.showToast({
        title: '绑定失败',
        icon: 'none'
      })
    })
  },

  // ========== 分享功能 ==========
  
  // 分享给朋友
  onShareAppMessage() {
    return {
      title: `${this.data.userInfo?.nickname || '用户'}的碰撞主页`,
      path: '/pages/profile/profile',
      imageUrl: this.data.userInfo?.avatar || '/images/share-default.png'
    }
  },

  // 分享到朋友圈
  onShareTimeline() {
    return {
      title: `来看我的碰撞主页 - 标签碰撞`,
      query: ''
    }
  },

  // 下拉刷新
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    const app = getApp()
    
    try {
      // 如果已登录，刷新所有数据
      if (app.globalData.hasLogin) {
        await Promise.all([
          this.loadUserInfo(),
          this.loadBalance(),
          this.loadMyCollisionCode(),
          this.loadMyCollisionList(),
          this.loadUserContacts()
        ])
        console.log('下拉刷新完成')
      } else {
        // 未登录状态，直接停止刷新
        console.log('未登录状态，跳过数据刷新')
      }
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  }
})
