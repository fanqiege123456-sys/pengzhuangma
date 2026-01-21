// pages/match/list.js
const api = require('../../utils/api.js')

Page({
  data: {
    matches: [],
    loading: false,
    hasMore: true
  },

  onLoad() {
    const app = getApp()
    // 如果已登录,加载数据
    if (app.globalData.hasLogin) {
      this.loadMatches()
    }
    // 不再强制登录,允许浏览
  },

  onShow() {
    const app = getApp()
    // 如果已登录,刷新数据
    if (app.globalData.hasLogin) {
      this.loadMatches()
    }
    // 不再自动弹出登录提示,允许用户浏览
  },

  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    try {
      await this.loadMatches()
      console.log('下拉刷新完成')
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  },

  // 加载匹配记录列表
  async loadMatches() {
    if (this.data.loading) return

    this.setData({ loading: true })

    try {
      const res = await api.getMatches()
      
      if (res.data.code === 200) {
        const matches = res.data.data || []
        const app = getApp()
        const currentUserId = app.globalData.userInfo?.id
        
        // 处理匹配数据
        const formattedMatches = matches.map(match => {
          // 判断当前用户是 user1 还是 user2
          const partner = match.user_id1 === currentUserId ? match.User2 : match.User1
          const isExpired = new Date(match.add_friend_deadline) < new Date()
          
          return {
            id: match.id,
            tag: match.tag || '未知标签',
            partner: {
              id: partner?.id,
              nickname: partner?.nickname || '未知用户',
              avatar: partner?.avatar || '/images/default-avatar.png',
              wechat_no: partner?.wechat_no
            },
            matchType: match.match_type, // district, city, province, country
            matchLocation: this.formatLocation(match),
            status: match.status, // matched, friend_added, missed
            createdAt: match.created_at,
            deadline: match.add_friend_deadline,
            timeLeft: isExpired ? '已过期' : this.getTimeLeft(match.add_friend_deadline),
            isExpired: isExpired,
            canForceAdd: isExpired && match.status === 'matched'
          }
        })
        
        this.setData({
          matches: formattedMatches,
          loading: false
        })
      } else {
        wx.showToast({
          title: res.data.msg || '加载失败',
          icon: 'none'
        })
        this.setData({ loading: false })
      }
    } catch (error) {
      console.error('加载匹配记录失败', error)
      wx.showToast({
        title: '网络错误，请重试',
        icon: 'none'
      })
      this.setData({ loading: false })
    }
  },

  // 格式化位置
  formatLocation(match) {
    const parts = []
    if (match.match_country) parts.push(match.match_country)
    if (match.match_province) parts.push(match.match_province)
    if (match.match_city) parts.push(match.match_city)
    if (match.match_district) parts.push(match.match_district)
    return parts.join(' · ') || '未知地区'
  },

  // 计算剩余时间
  getTimeLeft(deadline) {
    const now = new Date()
    const end = new Date(deadline)
    const diff = end - now
    
    if (diff <= 0) return '已过期'
    
    const hours = Math.floor(diff / (1000 * 60 * 60))
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
    
    if (hours > 0) {
      return `剩余 ${hours}小时${minutes}分钟`
    } else {
      return `剩余 ${minutes}分钟`
    }
  },

  // 格式化时间
  formatTime(dateString) {
    const date = new Date(dateString)
    const now = new Date()
    const diff = now - date
    
    // 小于1分钟
    if (diff < 60000) {
      return '刚刚'
    }
    // 小于1小时
    if (diff < 3600000) {
      return `${Math.floor(diff / 60000)}分钟前`
    }
    // 小于24小时
    if (diff < 86400000) {
      return `${Math.floor(diff / 3600000)}小时前`
    }
    // 大于24小时
    const month = date.getMonth() + 1
    const day = date.getDate()
    return `${month}月${day}日`
  },

  // 查看匹配详情
  viewDetail(e) {
    const matchId = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/match/detail?id=${matchId}`
    })
  },

  // 复制微信号
  copyWechat(e) {
    const wechatNo = e.currentTarget.dataset.wechat
    if (!wechatNo) {
      wx.showToast({
        title: '暂无微信号',
        icon: 'none'
      })
      return
    }
    
    wx.setClipboardData({
      data: wechatNo,
      success: () => {
        wx.showToast({
          title: '已复制微信号',
          icon: 'success'
        })
      }
    })
  },

  // 强制添加好友（24小时后）
  async forceAddFriend(e) {
    const matchId = e.currentTarget.dataset.id
    
    try {
      const confirmRes = await wx.showModal({
        title: '强制添加好友',
        content: '已过24小时期限，将消耗100积分强制添加对方为好友，是否继续？',
        confirmText: '确定',
        cancelText: '取消'
      })
      
      if (!confirmRes.confirm) return
      
      wx.showLoading({
        title: '添加中...'
      })
      
      // 默认消耗 100 积分
      const res = await api.forceAddFriend({ match_id: matchId, cost_coins: 100 })
      wx.hideLoading()
      
      if (res.data.code === 200) {
        wx.showToast({
          title: '添加成功',
          icon: 'success'
        })
        
        // 刷新列表
        setTimeout(() => {
          this.loadMatches()
        }, 1500)
      } else {
        wx.showToast({
          title: res.data.msg || '添加失败',
          icon: 'none'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('强制添加好友失败', error)
      wx.showToast({
        title: '网络错误，请重试',
        icon: 'none'
      })
    }
  },

  // 跳转到碰撞页面
  goToCollision() {
    wx.switchTab({
      url: '/pages/collision/collision'
    })
  }
})
