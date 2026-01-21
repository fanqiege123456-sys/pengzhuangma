// pages/friends/friends.js
const api = require('../../utils/api.js')

Page({
  data: {
    searchText: '',
    friendsList: [],
    loading: false,
    refreshing: false,
    showActionModal: false,
    selectedFriend: null
  },

  onLoad() {
    const app = getApp()
    // 如果已登录,加载数据
    if (app.globalData.hasLogin) {
      this.loadFriends()
    }
    // 不再强制登录,允许浏览
  },

  onShow() {
    const app = getApp()
    // 如果已登录,刷新数据
    if (app.globalData.hasLogin) {
      this.loadFriends()
    }
    // 不再自动弹出登录提示,允许用户浏览界面
  },

  // 加载好友列表
  async loadFriends() {
    if (this.data.loading) return

    this.setData({
      loading: true
    })

    try {
      // 改为获取匹配记录
      const res = await api.getMatches()

      if (res.data.code === 200) {
        const matches = res.data.data || []
        // 将匹配记录转换为好友格式显示
        const friendsList = matches.map(match => ({
          id: match.id,
          nickname: match.User1?.nickname || match.User2?.nickname || '未知用户',
          avatar: match.User1?.avatar || match.User2?.avatar || '/images/default-avatar.png',
          wechat_no: match.User1?.wechat_no || match.User2?.wechat_no || '',
          match_time: match.created_at,
          status: match.status,
          collision_code: match.code
        }))
        
        this.setData({
          friendsList,
          loading: false
        })
      } else {
        throw new Error(res.data.msg || '加载失败')
      }
    } catch (error) {
      console.error('加载匹配记录失败', error)
      this.setData({
        loading: false
      })
      wx.showToast({
        title: '加载失败',
        icon: 'error'
      })
    }
  },

  // 下拉刷新
  onRefresh() {
    this.setData({
      refreshing: true
    })
    
    this.loadFriends().finally(() => {
      this.setData({
        refreshing: false
      })
    })
  },

  // 搜索输入
  onSearchInput(e) {
    this.setData({
      searchText: e.detail.value
    })
    
    // 防抖搜索
    clearTimeout(this.searchTimer)
    this.searchTimer = setTimeout(() => {
      this.loadFriends()
    }, 500)
  },

  // 开始聊天
  goToChat(e) {
    const user = e.currentTarget.dataset.user
    wx.navigateTo({
      url: `/pages/chat/chat?userId=${user.id}&nickname=${user.nickname}`
    })
  },

  // 显示操作菜单
  showActions(e) {
    const friend = e.currentTarget.dataset.friend
    this.setData({
      selectedFriend: friend,
      showActionModal: true
    })
  },

  // 隐藏操作菜单
  hideActionModal() {
    this.setData({
      showActionModal: false,
      selectedFriend: null
    })
  },

  // 阻止事件冒泡
  stopPropagation() {
    // 阻止点击事件冒泡
  },

  // 开始聊天
  startChat() {
    const { selectedFriend } = this.data
    this.hideActionModal()
    
    wx.navigateTo({
      url: `/pages/chat/chat?userId=${selectedFriend.id}&nickname=${selectedFriend.nickname}`
    })
  },

  // 复制微信号
  copyWechat() {
    const { selectedFriend } = this.data
    
    wx.setClipboardData({
      data: selectedFriend.wechat_no,
      success: () => {
        wx.showToast({
          title: '微信号已复制',
          icon: 'success'
        })
      }
    })
    
    this.hideActionModal()
  },

  // 屏蔽好友
  async blockFriend() {
    const { selectedFriend } = this.data
    
    try {
      const modalRes = await wx.showModal({
        title: '确认屏蔽',
        content: `确定要屏蔽 ${selectedFriend.nickname} 吗？`
      })

      if (modalRes.confirm) {
        wx.showLoading({
          title: '处理中...'
        })

        const res = await api.blockFriend(selectedFriend.id)
        wx.hideLoading()

        if (res.data.code === 200) {
          wx.showToast({
            title: '已屏蔽',
            icon: 'success'
          })
          
          this.loadFriends()
        } else {
          wx.showToast({
            title: res.data.msg || '操作失败',
            icon: 'error'
          })
        }
      }
    } catch (error) {
      wx.hideLoading()
      console.error('屏蔽好友失败', error)
      wx.showToast({
        title: '网络错误',
        icon: 'error'
      })
    }
    
    this.hideActionModal()
  },

  // 取消屏蔽
  async unblockFriend() {
    const { selectedFriend } = this.data
    
    try {
      wx.showLoading({
        title: '处理中...'
      })

      const res = await api.unblockFriend(selectedFriend.id)
      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '已取消屏蔽',
          icon: 'success'
        })
        
        this.loadFriends()
      } else {
        wx.showToast({
          title: res.data.msg || '操作失败',
          icon: 'error'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('取消屏蔽失败', error)
      wx.showToast({
        title: '网络错误',
        icon: 'error'
      })
    }
    
    this.hideActionModal()
  },

  // 删除好友
  async deleteFriend() {
    const { selectedFriend } = this.data
    
    try {
      const modalRes = await wx.showModal({
        title: '确认删除',
        content: `确定要删除好友 ${selectedFriend.nickname} 吗？删除后无法恢复。`,
        confirmText: '删除',
        confirmColor: '#FF3B30'
      })

      if (modalRes.confirm) {
        wx.showLoading({
          title: '删除中...'
        })

        const res = await api.deleteFriend(selectedFriend.id)
        wx.hideLoading()

        if (res.data.code === 200) {
          wx.showToast({
            title: '已删除',
            icon: 'success'
          })
          
          this.loadFriends()
        } else {
          wx.showToast({
            title: res.data.msg || '删除失败',
            icon: 'error'
          })
        }
      }
    } catch (error) {
      wx.hideLoading()
      console.error('删除好友失败', error)
      wx.showToast({
        title: '网络错误',
        icon: 'error'
      })
    }
    
    this.hideActionModal()
  },

  // 去碰撞页面
  goToCollision() {
    wx.switchTab({
      url: '/pages/index/index'
    })
  },

  // 下拉刷新 - 微信小程序内置下拉刷新事件
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    const app = getApp()
    
    try {
      // 如果已登录，刷新好友列表
      if (app.globalData.hasLogin) {
        await this.loadFriends()
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
