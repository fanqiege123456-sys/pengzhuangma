// pages/chat/chat.js
Page({
  data: {
    userId: '',
    nickname: '',
    currentUserId: '',
    messages: [],
    inputText: '',
    inputFocus: false,
    scrollToView: ''
  },

  onLoad(options) {
    const { userId, nickname } = options
    const app = getApp()
    
    wx.setNavigationBarTitle({
      title: nickname || '聊天'
    })
    
    this.setData({
      userId,
      nickname,
      currentUserId: app.globalData.userInfo?.id
    })
    
    this.loadMessages()
  },

  // 加载聊天消息
  async loadMessages() {
    // 模拟聊天消息数据
    const mockMessages = [
      {
        id: 1,
        sender_id: this.data.userId,
        content: '你好！很高兴通过碰撞码认识你',
        avatar: '/images/default-avatar.png',
        time: '10:30',
        created_at: '2024-01-15 10:30:00'
      },
      {
        id: 2,
        sender_id: this.data.currentUserId,
        content: '我也是！你也在这个城市吗？',
        avatar: '/images/default-avatar.png',
        time: '10:31',
        created_at: '2024-01-15 10:31:00'
      },
      {
        id: 3,
        sender_id: this.data.userId,
        content: '是的，有空可以一起出来玩',
        avatar: '/images/default-avatar.png',
        time: '10:32',
        created_at: '2024-01-15 10:32:00'
      }
    ]

    this.setData({
      messages: mockMessages
    })

    // 滚动到最底部
    setTimeout(() => {
      this.scrollToBottom()
    }, 100)
  },

  // 输入变化
  onInputChange(e) {
    this.setData({
      inputText: e.detail.value
    })
  },

  // 发送消息
  sendMessage() {
    const { inputText } = this.data
    
    if (!inputText.trim()) {
      return
    }

    const newMessage = {
      id: Date.now(),
      sender_id: this.data.currentUserId,
      content: inputText.trim(),
      avatar: '/images/default-avatar.png',
      time: this.getCurrentTime(),
      created_at: new Date().toISOString()
    }

    this.setData({
      messages: [...this.data.messages, newMessage],
      inputText: ''
    })

    // 滚动到最底部
    setTimeout(() => {
      this.scrollToBottom()
    }, 100)

    // TODO: 发送消息到后端
    this.sendToServer(newMessage)
  },

  // 发送消息到服务器
  async sendToServer(message) {
    try {
      // const res = await api.sendMessage({
      //   receiver_id: this.data.userId,
      //   content: message.content
      // })
      
      // 模拟发送成功
      console.log('消息发送成功', message)
      
    } catch (error) {
      console.error('发送消息失败', error)
      wx.showToast({
        title: '发送失败',
        icon: 'error'
      })
    }
  },

  // 滚动到底部
  scrollToBottom() {
    const messages = this.data.messages
    if (messages.length > 0) {
      const lastMessageId = `msg-${messages[messages.length - 1].id}`
      this.setData({
        scrollToView: lastMessageId
      })
    }
  },

  // 获取当前时间
  getCurrentTime() {
    const now = new Date()
    const hours = now.getHours().toString().padStart(2, '0')
    const minutes = now.getMinutes().toString().padStart(2, '0')
    return `${hours}:${minutes}`
  },

  // 页面显示时聚焦输入框
  onShow() {
    setTimeout(() => {
      this.setData({
        inputFocus: true
      })
    }, 500)
  },

  // 下拉刷新
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    try {
      await this.loadMessages()
      console.log('下拉刷新完成')
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  }
})
