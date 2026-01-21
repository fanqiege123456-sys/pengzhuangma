// pages/normal-collision/normal-collision.js
const api = require('../../utils/api.js')

Page({
  data: {
    groups: [],
    total: 0,
    loading: false,
    showEmailForm: false,
    selectedMatch: null,
    emailMessage: '',
    canSubmitEmail: false,
    submitting: false
  },

  onLoad() {
    this.loadMatchResults()
  },

  onShow() {
    this.loadMatchResults()
  },

  onPullDownRefresh() {
    this.loadMatchResults().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 加载匹配结果（按关键词分组）
  async loadMatchResults() {
    if (this.data.loading) return

    this.setData({ loading: true })

    try {
      const res = await api.getCollisionResults(30)
      if (res.data.code === 200) {
        const rawGroups = res.data.data || []
        const groups = rawGroups.map(group => {
          let keyword = group.keyword_raw || group.keyword || ''
          if (!keyword && group.id) {
            const parts = String(group.id).split('_')
            if (parts.length > 1) {
              keyword = parts.slice(1).join('_')
            }
          }
          const matches = (group.matches || []).map(item => ({
            ...item,
            matched_at_display: item.matched_at || '',
            status_text: item.email_sent ? '已发邮件' : '未发邮件',
            status_class: item.email_sent ? 'sent' : 'pending'
          }))

          return {
            id: group.id,
            keyword,
            total: group.total || matches.length,
            date: group.date,
            matches,
            expanded: false
          }
        })

        const total = groups.reduce((sum, item) => sum + (item.total || 0), 0)

        this.setData({
          groups,
          total
        })
      }
    } catch (error) {
      console.error('加载匹配结果失败', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  // 展开/收起分组
  toggleGroup(e) {
    const id = e.currentTarget.dataset.id
    const groups = this.data.groups.map(group => {
      if (group.id === id) {
        return { ...group, expanded: !group.expanded }
      }
      return group
    })
    this.setData({ groups })
  },

  // 发起碰撞
  goToStartCollision() {
    const app = getApp()

    if (!app.globalData.hasLogin) {
      wx.showModal({
        title: '需要登录',
        content: '请先登录后再发起碰撞',
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

    wx.navigateTo({
      url: '/pages/collision/collision'
    })
  },

  // 显示发送邮件弹窗
  showEmailForm(e) {
    const item = e.currentTarget.dataset.item
    if (!item) return

    this.setData({
      showEmailForm: true,
      selectedMatch: item,
      emailMessage: '',
      canSubmitEmail: false
    })
  },

  // 取消发送邮件
  cancelEmail() {
    this.setData({
      showEmailForm: false,
      selectedMatch: null,
      emailMessage: '',
      canSubmitEmail: false
    })
  },

  // 输入邮件内容
  onEmailMessageInput(e) {
    const value = e.detail.value || ''
    this.setData({
      emailMessage: value,
      canSubmitEmail: !!value.trim()
    })
  },

  // 提交邮件
  async submitEmail() {
    if (this.data.submitting) return

    const { selectedMatch, emailMessage } = this.data
    if (!selectedMatch) return

    if (!emailMessage.trim()) {
      wx.showToast({
        title: '请输入邮件内容',
        icon: 'none'
      })
      return
    }

    this.setData({ submitting: true })

    try {
      const res = await api.sendEmailToMatch({
        result_id: selectedMatch.id,
        content: emailMessage.trim()
      })

      if (res.data.code === 200) {
        wx.showToast({
          title: '邮件发送成功',
          icon: 'success'
        })
        this.cancelEmail()
        this.loadMatchResults()
      } else {
        wx.showToast({
          title: res.data.message || '发送失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('发送邮件失败:', error)
      wx.showToast({
        title: '网络错误',
        icon: 'none'
      })
    } finally {
      this.setData({ submitting: false })
    }
  },

  // 分享给朋友
  onShareAppMessage() {
    return {
      title: '我的匹配结果',
      path: '/pages/normal-collision/normal-collision'
    }
  },

  // 分享到朋友圈
  onShareTimeline() {
    return {
      title: '我的匹配结果'
    }
  }
})
