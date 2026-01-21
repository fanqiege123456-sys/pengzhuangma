// pages/match/detail.js
const api = require('../../utils/api.js')

Page({
  data: {
    match: null,
    loading: false
  },

  onLoad(options) {
    const id = options.id
    if (id) {
      this.loadMatch(id)
    }
  },

  async loadMatch(id) {
    this.setData({ loading: true })
    try {
      const res = await api.getMatchDetail(id)
      if (res.data && res.data.code === 200) {
        const match = res.data.data
        // 格式化时间展示
        match.time_left_text = this.formatTimeLeft(match.add_friend_deadline)
        match.match_location_text = [match.match_location.country, match.match_location.province, match.match_location.city, match.match_location.district].filter(Boolean).join('、')
        this.setData({ match })
      } else {
        wx.showToast({ title: res.data?.msg || '加载失败', icon: 'none' })
      }
    } catch (error) {
      console.error('获取匹配详情失败', error)
      wx.showToast({ title: '网络错误', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  formatTimeLeft(deadline) {
    const now = new Date()
    const end = new Date(deadline)
    const diff = end - now
    if (diff <= 0) return '已过期'
    const hours = Math.floor(diff / (1000 * 60 * 60))
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
    if (hours > 0) return `剩余 ${hours}小时${minutes}分钟`
    return `剩余 ${minutes}分钟`
  },

  // 主动添加好友（24小时内）
  async handleAddFriend() {
    const matchId = this.data.match.id
    wx.showLoading({ title: '处理中...' })
    try {
      const res = await api.addFriend(matchId)
      wx.hideLoading()
      if (res.data && res.data.code === 200) {
        wx.showToast({ title: '已添加好友', icon: 'success' })
        // 返回上一页并刷新列表
        setTimeout(() => {
          wx.navigateBack()
        }, 1000)
      } else {
        wx.showToast({ title: res.data?.msg || '添加失败', icon: 'none' })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('添加好友失败', error)
      wx.showToast({ title: '网络错误', icon: 'none' })
    }
  },

  // 强制添加好友（24小时后）
  async handleForceAdd() {
    const matchId = this.data.match.id
    const confirm = await new Promise((resolve) => {
      wx.showModal({ title: '确认', content: '确定消耗100积分强制添加好友吗？', success: (res) => resolve(res.confirm) })
    })
    if (!confirm) return
    wx.showLoading({ title: '处理...' })
    try {
      const res = await api.forceAddFriend({ match_id: matchId, cost_coins: 100 })
      wx.hideLoading()
      if (res.data && res.data.code === 200) {
        wx.showToast({ title: '已添加好友', icon: 'success' })
        setTimeout(() => wx.navigateBack(), 800)
      } else {
        wx.showToast({ title: res.data?.msg || '操作失败', icon: 'none' })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('强制添加失败', error)
      wx.showToast({ title: '网络错误', icon: 'none' })
    }
  },

  // 复制微信号
  handleCopyWechat() {
    const wechat = this.data.match.partner.wechat_no
    if (!wechat) {
      wx.showToast({ title: '暂无微信号', icon: 'none' })
      return
    }
    wx.setClipboardData({ data: wechat, success: () => wx.showToast({ title: '已复制微信号', icon: 'success' }) })
  }
})