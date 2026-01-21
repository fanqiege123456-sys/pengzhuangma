// pages/my-codes/my-codes.js
const api = require('../../utils/api')

Page({
  data: {
    codesList: [],
    filteredList: [],
    currentTab: 'all',
    totalCount: 0,
    matchedCount: 0,
    pendingCount: 0,  // 进行中未匹配
    expiredCount: 0
  },

  onLoad() {
    this.loadMyCodes()
  },

  onShow() {
    this.loadMyCodes()
  },

  onPullDownRefresh() {
    this.loadMyCodes().then(() => {
      wx.stopPullDownRefresh()
    })
  },

  async loadMyCodes() {
    try {
      wx.showLoading({ title: '加载中...' })
      
      const res = await api.getMyCollisionCodes()
      
      if (res.data.code === 200 && res.data.data) {
        const { codes = [], total = 0 } = res.data.data
        
        // 统计数量
        let matchedCount = 0
        let pendingCount = 0
        let expiredCount = 0
        // 处理数据
        const processedCodes = codes.map(item => {
          // 统计各状态数量
          if (item.is_expired) {
            expiredCount++
          } else if (item.collision_status === '已碰撞' || item.is_matched) {
            matchedCount++
          } else {
            pendingCount++
          }

          // 计算显示状态
          let displayStatus = ''
          let statusClass = ''
          let btnStatus = ''
          
          if (item.is_expired) {
            displayStatus = '已过期（仍可匹配）'
            statusClass = 'expired'
            btnStatus = '续期'
          } else if (item.collision_status === '已碰撞' || item.is_matched) {
            displayStatus = '已碰撞'
            statusClass = 'matched'
            btnStatus = '查看详情'
          } else {
            displayStatus = '进行中'
            statusClass = 'active'
            btnStatus = '等待碰撞'
          }

          return {
            ...item,
            display_status: displayStatus,
            status_class: statusClass,
            btn_status: btnStatus
          }
        })

        this.setData({
          codesList: processedCodes,
          totalCount: total,
          matchedCount,
          pendingCount,
          expiredCount
        })

        this.filterList()
        
        // 检查是否有过期的碰撞码，显示续期提醒
        if (expiredCount > 0) {
          // 只在首次加载时显示提醒，避免频繁弹窗
          const hasShownTip = wx.getStorageSync('expiredTipShown_' + new Date().toDateString())
          if (!hasShownTip) {
            wx.showModal({
              title: '续期提醒',
              content: `您有 ${expiredCount} 个碰撞码已过期，虽然仍可参与匹配，但建议您及时续期以获得更好的匹配效果。`,
              confirmText: '去续期',
              cancelText: '稍后',
              success: (res) => {
                if (res.confirm) {
                  // 切换到过期标签页
                  this.setData({ currentTab: 'expired' })
                  this.filterList()
                }
                // 记录已显示提醒，当天不再显示
                wx.setStorageSync('expiredTipShown_' + new Date().toDateString(), true)
              }
            })
          }
        }
      }
      
    } catch (err) {
      console.error('加载碰撞码失败', err)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    } finally {
      wx.hideLoading()
    }
  },

  // 格式化日期时间
  formatDateTime(dateStr) {
    if (!dateStr) return ''
    try {
      const date = new Date(dateStr)
      if (isNaN(date.getTime())) return ''
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const minute = String(date.getMinutes()).padStart(2, '0')
      return `${month}-${day} ${hour}:${minute}`
    } catch {
      return ''
    }
  },

  switchTab(e) {
    const tab = e.currentTarget.dataset.tab
    this.setData({ currentTab: tab })
    this.filterList()
  },

  filterList() {
    const { codesList, currentTab } = this.data
    let filteredList = []

    switch (currentTab) {
      case 'matched':
        // 已碰撞
        filteredList = codesList.filter(item => 
          (item.collision_status === '已碰撞' || item.is_matched)
        )
        break
      case 'active':
        // 进行中
        filteredList = codesList.filter(item => 
          !item.is_expired &&
          item.collision_status !== '已碰撞' &&
          !item.is_matched
        )
        break
      case 'expired':
        // 已过期
        filteredList = codesList.filter(item => item.is_expired)
        break
      default:
        filteredList = codesList
    }

    this.setData({ filteredList })
  },

  // 查看碰撞详情
  viewDetail(e) {
    const id = e.currentTarget.dataset.id
    const item = this.data.codesList.find(c => c.id === id)
    
    if (item && (item.collision_status === '已碰撞' || item.is_matched)) {
      wx.navigateTo({
        url: `/pages/collision-result/collision-result?keyword=${encodeURIComponent(item.tag)}`
      })
    }
  },

  // 重新碰撞
  reCollision(e) {
    const id = e.currentTarget.dataset.id
    const item = this.data.codesList.find(c => c.id === id)
    
    if (!item) return
    
    wx.navigateTo({
      url: `/pages/collision/collision?keyword=${encodeURIComponent(item.tag)}&edit=true`
    })
  },

  // 续费（审核通过的或已过期的都可以续费）
  renewCode(e) {
    const id = e.currentTarget.dataset.id
    const item = this.data.codesList.find(c => c.id === id)
    
    if (!item) return
    
    wx.showModal({
      title: '续费确认',
      content: `续费「${item.tag}」需要消耗10积分，确定续费吗？`,
      success: async (res) => {
        if (res.confirm) {
          try {
            wx.showLoading({ title: '续费中...' })
            const result = await api.renewCollisionCode(id)
            wx.hideLoading()
            
            if (result.data.code === 200) {
              wx.showToast({ title: '续费成功', icon: 'success' })
              this.loadMyCodes()
            } else {
              wx.showToast({ title: result.data.message || '续费失败', icon: 'none' })
            }
          } catch (error) {
            wx.hideLoading()
            wx.showToast({ title: '续费失败', icon: 'none' })
          }
        }
      }
    })
  },

  // 修改碰撞码
  editCode(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/edit-collision-code/edit-collision-code?id=${id}`
    })
  },

  // 删除碰撞码
  deleteCode(e) {
    const id = e.currentTarget.dataset.id
    const item = this.data.codesList.find(c => c.id === id)
    
    if (!item) return
    
    wx.showModal({
      title: '删除确认',
      content: `确定要删除「${item.tag}」吗？删除后无法恢复。`,
      confirmText: '删除',
      confirmColor: '#F56C6C',
      cancelText: '取消',
      success: async (res) => {
        if (res.confirm) {
          try {
            wx.showLoading({ title: '删除中...' })
            const result = await api.deleteCollisionCode(id)
            wx.hideLoading()
            
            if (result.data.code === 200) {
              wx.showToast({ title: '删除成功', icon: 'success' })
              this.loadMyCodes()
            } else {
              wx.showToast({ title: result.data.message || '删除失败', icon: 'none' })
            }
          } catch (error) {
            wx.hideLoading()
            wx.showToast({ title: '删除失败', icon: 'none' })
          }
        }
      }
    })
  },

  goToCollision() {
    wx.switchTab({
      url: '/pages/normal-collision/normal-collision'
    })
  }
})
