// pages/collision-result/collision-result.js
const api = require('../../utils/api.js')

Page({
  data: {
    keyword: '',
    mode: 'search',
    resultId: '',
    results: [],
    matchResults: [],
    loading: false,
    showEmailForm: false,
    selectedUser: {},
    emailMessage: '',
    canSubmitEmail: false,
    submitting: false,
    expiredTip: '' // 过期提示
  },

  onLoad(options) {
    let keyword = options.keyword || ''
    const resultId = options.resultId || ''
    
    // 解码URL编码的关键词
    if (keyword) {
      try {
        keyword = decodeURIComponent(keyword)
      } catch (error) {
        console.error('解码关键词失败:', error)
        // 如果解码失败，使用原始值
      }
    }
    
    if (resultId) {
      this.setData({ mode: 'result', resultId })
      this.loadResultDetail(resultId)
      return
    }

    this.setData({ keyword })

    if (keyword) {
      this.searchCollisionCodes(keyword)
    } else {
      wx.showToast({
        title: '缺少搜索关键词',
        icon: 'none'
      })
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    }
  },

  // 加载匹配结果详情
  async loadResultDetail(resultId) {
    this.setData({ loading: true })

    try {
      const res = await api.getCollisionResultDetail(resultId, { offset: 0, limit: 50 })
      if (res.data.code === 200) {
        const list = res.data.data || []
        const formattedResults = list.map(item => {
          const displayName = item.remark && item.remark.trim()
            ? item.remark
            : `匹配用户${item.matched_user_id || ''}`
          const canNotify = !!item.matched_email || (item.matched_email && item.matched_email.indexOf('隐藏') !== -1)

          return {
            ...item,
            nickname: displayName,
            avatar: '/images/default-avatar.png',
            created_at_display: item.matched_at,
            status_text: item.email_sent ? '已发邮件' : '未发邮件',
            status_class: item.email_sent ? 'active' : 'expired',
            can_notify: canNotify
          }
        })

        this.setData({
          results: formattedResults,
          expiredTip: ''
        })
      } else {
        wx.showToast({
          title: res.data.message || '加载失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('加载匹配结果失败:', error)
      wx.showToast({
        title: '网络错误',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadMatchResultsByKeyword(keyword) {
    try {
      const res = await api.getCollisionResults(30, keyword)
      if (res.data.code !== 200) {
        return
      }
      const groups = res.data.data || []
      const matchResults = []

      groups.forEach(group => {
        let groupKeyword = group.keyword_raw || group.keyword || ''
        if (!groupKeyword && group.id) {
          const parts = String(group.id).split('_')
          if (parts.length > 1) {
            groupKeyword = parts.slice(1).join('_')
          }
        }
        const matches = group.matches || []
        matches.forEach(item => {
          const canNotify = !!item.matched_email
          matchResults.push({
            ...item,
            match_result_id: item.id,
            keyword: groupKeyword,
            matched_at_display: item.matched_at,
            status_text: item.email_sent ? '已发邮件' : '未发邮件',
            status_class: item.email_sent ? 'active' : 'expired',
            can_notify: canNotify
          })
        })
      })

      this.setData({ matchResults })
    } catch (error) {
      console.error('加载匹配结果失败:', error)
    }
  },

  // 搜索碰撞码
  async searchCollisionCodes(keyword) {
    this.setData({ loading: true })
    
    try {
      console.log('开始搜索关键词:', keyword)
      const res = await api.searchCollisionCodes(keyword)
      
      console.log('API返回完整响应:', res)
      console.log('响应状态码:', res.statusCode)
      console.log('响应数据:', res.data)
      
      if (res.data.code === 200) {
        const results = res.data.data || []
        
        console.log('搜索结果数组:', results)
        console.log('结果数量:', results.length)
        
        // 处理数据，格式化显示信息
        const formattedResults = results.map(item => {
          console.log('处理单个结果项:', item)
          
          // 判断碰撞码是否过期
          const isExpired = item.expires_at ? new Date(item.expires_at) < new Date() : false
          const canNotify = item.can_notify !== undefined
            ? item.can_notify
            : !!(item.email && item.email.trim())
          
          return {
            ...item,
            created_at_display: this.formatTime(item.created_at),
            has_email: canNotify,
            can_notify: canNotify,
            // 添加过期状态
            is_expired: isExpired,
            status_text: isExpired ? '已过期（仍可匹配）' : '活跃',
            status_class: isExpired ? 'expired' : 'active'
          }
        })
        
        console.log('格式化后的结果:', formattedResults)
        
        // 统计过期碰撞码数量
        const expiredCount = formattedResults.filter(user => user.is_expired).length
        let expiredTip = ''
        if (expiredCount > 0) {
          expiredTip = `搜索结果中有 ${expiredCount} 个碰撞码已过期，但仍然可以联系这些用户。`
        }
        
        this.setData({ 
          results: formattedResults,
          expiredTip: expiredTip
        })
      } else {
        console.error('API返回错误:', res.data.code, res.data.msg)
        wx.showToast({
          title: res.data.msg || '搜索失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('搜索碰撞码失败:', error)
      wx.showToast({
        title: '网络错误',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }

    this.loadMatchResultsByKeyword(keyword)
  },

  // 发起新碰撞
  createCollision() {
    const { keyword } = this.data
    wx.navigateTo({
      url: `/pages/collision/collision?tag=${encodeURIComponent(keyword)}`
    })
  },

  // 显示发送邮件表单
  showEmailForm(e) {
    const user = e.currentTarget.dataset.user
    
    if (!user.can_notify) {
      wx.showToast({
        title: '该用户暂不支持邮件通知',
        icon: 'none'
      })
      return
    }
    
    this.setData({
      showEmailForm: true,
      selectedUser: user,
      emailMessage: '',
      canSubmitEmail: false
    })
  },

  // 取消发送邮件
  cancelEmail() {
    this.setData({
      showEmailForm: false,
      selectedUser: {},
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
    
    const { selectedUser, keyword, emailMessage, mode, resultId } = this.data
    
    if (!emailMessage.trim()) {
      wx.showToast({
        title: '请输入邮件内容',
        icon: 'none'
      })
      return
    }

    this.setData({ submitting: true })
    
    try {
      let res
      if (mode === 'result' || selectedUser.match_result_id) {
        res = await api.sendEmailToMatch({
          result_id: selectedUser.match_result_id || selectedUser.id,
          content: emailMessage.trim()
        })
      } else {
        res = await api.sendEmailToMatchedUser(
          selectedUser.user_id,
          keyword,
          emailMessage.trim()
        )
      }

      if (res.data.code === 200) {
        wx.showToast({
          title: '邮件发送成功',
          icon: 'success'
        })
        this.cancelEmail()
        if (mode === 'result' && resultId) {
          this.loadResultDetail(resultId)
        }
      } else {
        wx.showToast({
          title: res.data.msg || '发送失败',
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

  // 格式化时间
  formatTime(timeStr) {
    if (!timeStr) return ''
    
    try {
      const date = new Date(timeStr)
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
      // 小于1天
      if (diff < 86400000) {
        return `${Math.floor(diff / 3600000)}小时前`
      }
      // 小于7天
      if (diff < 604800000) {
        return `${Math.floor(diff / 86400000)}天前`
      }
      
      // 超过7天显示具体日期
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      return `${month}-${day}`
    } catch (error) {
      return timeStr
    }
  },

  // 下拉刷新
  onPullDownRefresh() {
    if (this.data.keyword) {
      this.searchCollisionCodes(this.data.keyword)
    }
    wx.stopPullDownRefresh()
  },

  // 页面分享
  onShareAppMessage() {
    return {
      title: `搜索"${this.data.keyword}"的碰撞结果`,
      path: '/pages/index/index'
    }
  }
})
