// pages/index/index.js
const api = require('../../utils/api.js')

Page({
  data: {
    keyword: '',
    hotTags24h: [],
    hotTagsAll: [],
    userInfo: null,
    showLoginBtn: true,  // 显式控制登录按钮显示
    collisionList: [],   // 我的碰撞列表
    collisionResults: [], // 碰撞结果（最近30天）
    searchShake: false   // 搜索框震动状态
  },

  onLoad() {
    const app = getApp()
    this.setData({
      userInfo: app.globalData.userInfo,
      showLoginBtn: !app.globalData.userInfo  // 用户未登录时显示登录按钮
    })
    this.loadHotTags()
    if (app.globalData.hasLogin) {
      this.loadCollisionList()
      this.loadCollisionResults()
    }
    
    // 设置定时轮询，每隔5秒刷新一次热门标签数据
    this.interval = setInterval(() => {
      this.loadHotTags()
    }, 5000)
  },

  onHide() {
    // 页面隐藏时清除定时器，避免不必要的请求
    if (this.interval) {
      clearInterval(this.interval)
      this.interval = null
    }
  },

  onUnload() {
    // 页面卸载时清除定时器
    if (this.interval) {
      clearInterval(this.interval)
      this.interval = null
    }
  },

  onShow() {
    const app = getApp()
    this.setData({
      userInfo: app.globalData.userInfo,
      showLoginBtn: !app.globalData.userInfo  // 用户未登录时显示登录按钮
    })
    // 在页面显示时刷新热门标签数据
    this.loadHotTags()
    if (app.globalData.hasLogin) {
      this.loadCollisionList()
      this.loadCollisionResults()
    }
    
    // 如果定时器不存在，重新设置
    if (!this.interval) {
      this.interval = setInterval(() => {
        this.loadHotTags()
      }, 5000)
    }
  },

  // 加载热门标签
  async loadHotTags() {
    try {
      // 加载24小时热门
      const res24h = await api.getHotTags24h()
      if (res24h.data.code === 200) {
        this.setData({
          hotTags24h: res24h.data.data || []
        })
      }

      // 加载总榜
      const resAll = await api.getHotTagsAll()
      if (resAll.data.code === 200) {
        this.setData({
          hotTagsAll: resAll.data.data || []
        })
      }
    } catch (error) {
      console.error('加载热门标签失败', error)
    }
  },

  // 加载我的碰撞列表
  async loadCollisionList() {
    try {
      // 调用获取用户所有碰撞码的API
      const res = await api.getMyCollisionCodes()
      
      if (res.data.code === 200) {
        let list = []
        // 检查返回数据格式，确保是数组
        if (res.data.data && res.data.data.codes && Array.isArray(res.data.data.codes)) {
          list = res.data.data.codes
        }
        // 格式化过期时间
        list = list.map(item => {
          let expireDisplay = ''
          let statusClass = ''
          let statusText = ''

          if (item.is_expired) {
            expireDisplay = '已过期'
            statusClass = 'expired'
            statusText = '已过期'
          } else if (item.collision_status === '已碰撞' || item.is_matched) {
            expireDisplay = item.time_left || this.formatDate(item.expires_at)
            statusClass = 'matched'
            statusText = '已碰撞'
          } else {
            expireDisplay = item.time_left || this.formatDate(item.expires_at)
            statusClass = 'active'
            statusText = '进行中'
          }

          return {
            ...item,
            keyword: item.tag || item.keyword || '',
            expireDisplay,
            statusClass,
            statusText
          }
        })
        this.setData({ collisionList: list })
      }
    } catch (error) {
      console.error('加载碰撞列表失败', error)
    }
  },

  // 加载碰撞结果
  async loadCollisionResults() {
    try {
      const res = await api.getCollisionResults(30)
      if (res.data.code === 200) {
        let results = []
        if (Array.isArray(res.data.data)) {
          results = res.data.data
        } else if (res.data.data && Array.isArray(res.data.data.list)) {
          results = res.data.data.list
        }
        // 处理数据，从id中提取关键词
        results = results.map(item => {
          let keyword = item.keyword || item.tag || ''
          // 如果keyword为空，从id中提取，格式如 "2026-01-12_测试"
          if (!keyword && item.id) {
            const parts = item.id.split('_')
            if (parts.length > 1) {
              keyword = parts.slice(1).join('_')
            }
          }
          return {
            ...item,
            keyword: keyword,
            match_count: item.total || item.match_count || 0
          }
        })
        this.setData({ collisionResults: results })
      }
    } catch (error) {
      console.error('加载碰撞结果失败', error)
    }
  },

  // 格式化日期
  formatDate(dateStr) {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  },

  // 修改碰撞列表项
  editCollisionItem(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/collision/collision?id=${id}&edit=true`
    })
  },

  // 删除碰撞列表项
  deleteCollisionItem(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '删除确认',
      content: '删除后将退还扣除当天的积分，确定删除吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            const result = await api.deleteCollisionList(id)
            if (result.data.code === 200) {
              wx.showToast({ title: '删除成功', icon: 'success' })
              this.loadCollisionList()
            } else {
              wx.showToast({ title: result.data.message || '删除失败', icon: 'none' })
            }
          } catch (error) {
            wx.showToast({ title: '操作失败', icon: 'none' })
          }
        }
      }
    })
  },

  // 查看碰撞结果详情
  viewResultDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/collision-result/collision-result?resultId=${id}`
    })
  },

  // 输入关键词
  onKeywordInput(e) {
    this.setData({
      keyword: e.detail.value
    })
  },

  // 搜索关键词
  searchKeyword(e) {
    const keyword = e.detail.value.trim()
    if (!keyword) return
    
    wx.navigateTo({
      url: `/pages/collision-result/collision-result?keyword=${encodeURIComponent(keyword)}`
    })
  },

  // 搜索按钮点击
  onSearchTap() {
    const keyword = this.data.keyword.trim()
    if (!keyword) {
      wx.showToast({
        title: '请输入搜索关键词',
        icon: 'none'
      })
      return
    }
    
    wx.navigateTo({
      url: `/pages/collision-result/collision-result?keyword=${encodeURIComponent(keyword)}`
    })
  },

  // 选择热门标签
  async selectHotTag(e) {
    const keyword = e.currentTarget.dataset.keyword
    
    // 震动反馈
    wx.vibrateShort({
      type: 'light'
    })
    
    // 搜索框震动动画
    this.setData({
      searchShake: true
    })
    
    // 显示提示
    wx.showToast({
      title: `已选择「${keyword}」`,
      icon: 'success',
      duration: 1000
    })
    
    this.setData({
      keyword: keyword
    })
    
    // 清除震动动画
    setTimeout(() => {
      this.setData({
        searchShake: false
      })
    }, 600)
    
    // 移除点击标签时的计数增加，只在搜索时增加计数
  },

  // 开始碰撞
  startCollision() {
    const app = getApp()
    
    // 检查登录
    if (!app.globalData.hasLogin) {
      wx.showModal({
        title: '需要登录',
        content: '请先登录后再使用碰撞功能',
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

    const { keyword } = this.data
    if (!keyword.trim()) {
      wx.showToast({
        title: '请输入关键词',
        icon: 'none'
      })
      return
    }

    // 跳转到碰撞结果页
    wx.navigateTo({
      url: `/pages/collision-result/collision-result?keyword=${encodeURIComponent(keyword)}`
    })
  },

  // 手动登录
  handleLogin() {
    const app = getApp()
    app.requireLogin(true)
  },

  // ========== 分享功能 ==========
  
  // 分享给朋友
  onShareAppMessage() {
    return {
      title: `标签碰撞 - 发现志同道合的朋友`,
      path: '/pages/index/index',
      imageUrl: '/images/share-default.png'
    }
  },

  // 分享到朋友圈
  onShareTimeline() {
    return {
      title: `标签碰撞，让志同道合的人相遇 🎯`
    }
  },

  // 下拉刷新
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    const app = getApp()
    try {
      await this.loadHotTags()
      if (app.globalData.hasLogin) {
        await this.loadCollisionList()
        await this.loadCollisionResults()
      }
      console.log('下拉刷新完成')
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  }
})
