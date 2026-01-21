// pages/match/match.js
const api = require('../../utils/api.js')

Page({
  data: {
    matchId: null,
    matchInfo: null,
    matchTime: '',
    deadline: ''
  },

  onLoad(options) {
    const { id } = options
    if (id) {
      this.setData({
        matchId: id
      })
      this.loadMatchDetail()
    }
  },

  // 加载匹配详情
  async loadMatchDetail() {
    try {
      wx.showLoading({
        title: '加载中...'
      })

      const res = await api.getMatchDetail(this.data.matchId)
      wx.hideLoading()

      if (res.data.code === 200) {
        const matchInfo = res.data.data
        this.setData({
          matchInfo,
          matchTime: this.formatTime(matchInfo.created_at),
          deadline: this.formatTime(matchInfo.add_friend_deadline)
        })
      } else {
        wx.showToast({
          title: res.data.msg || '加载失败',
          icon: 'error'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('加载匹配详情失败', error)
      wx.showToast({
        title: '网络错误',
        icon: 'error'
      })
    }
  },

  // 添加好友
  async addFriend() {
    try {
      wx.showLoading({
        title: '添加中...'
      })

      const res = await api.addFriend(this.data.matchId)
      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '添加成功',
          icon: 'success'
        })
        
        // 刷新匹配详情
        setTimeout(() => {
          this.loadMatchDetail()
        }, 1500)
      } else {
        wx.showToast({
          title: res.data.msg || '添加失败',
          icon: 'error'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('添加好友失败', error)
      wx.showToast({
        title: '网络错误',
        icon: 'error'
      })
    }
  },

  // 跳过匹配
  async skipMatch() {
    try {
      const modalRes = await wx.showModal({
        title: '确认跳过',
        content: '跳过后将无法再添加该用户为好友，确定要跳过吗？',
        confirmText: '确定跳过',
        cancelText: '取消'
      })

      if (modalRes.confirm) {
        wx.showLoading({
          title: '处理中...'
        })

        const res = await api.skipMatch(this.data.matchId)
        wx.hideLoading()

        if (res.data.code === 200) {
          wx.showToast({
            title: '已跳过',
            icon: 'success'
          })
          
          // 刷新匹配详情
          setTimeout(() => {
            this.loadMatchDetail()
          }, 1500)
        } else {
          wx.showToast({
            title: res.data.msg || '操作失败',
            icon: 'error'
          })
        }
      }
    } catch (error) {
      wx.hideLoading()
      console.error('跳过匹配失败', error)
      wx.showToast({
        title: '网络错误',
        icon: 'error'
      })
    }
  },

  // 开始聊天
  startChat() {
    const { matchInfo } = this.data
    const app = getApp()
    const currentUserId = app.globalData.userInfo?.id
    
    // 确定聊天对象
    const chatUser = matchInfo.user1.id === currentUserId ? matchInfo.user2 : matchInfo.user1
    
    wx.navigateTo({
      url: `/pages/chat/chat?userId=${chatUser.id}&nickname=${chatUser.nickname}`
    })
  },

  // 复制微信号
  copyWechatNo(e) {
    const wechatNo = e.currentTarget.dataset.wechat
    wx.setClipboardData({
      data: wechatNo,
      success: () => {
        wx.showToast({
          title: '微信号已复制',
          icon: 'success'
        })
      }
    })
  },

  // 格式化时间
  formatTime(timeStr) {
    if (!timeStr) return ''
    
    const date = new Date(timeStr)
    const year = date.getFullYear()
    const month = (date.getMonth() + 1).toString().padStart(2, '0')
    const day = date.getDate().toString().padStart(2, '0')
    const hour = date.getHours().toString().padStart(2, '0')
    const minute = date.getMinutes().toString().padStart(2, '0')
    
    return `${year}-${month}-${day} ${hour}:${minute}`
  }
})
