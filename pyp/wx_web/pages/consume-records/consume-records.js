// pages/consume-records/consume-records.js
const api = require('../../utils/api')

Page({
  data: {
    currentTab: 'consume',
    // 消费记录
    consumeList: [],
    consumePage: 1,
    consumeHasMore: true,
    consumeLoading: false,
    // 充值记录
    rechargeList: [],
    rechargePage: 1,
    rechargeHasMore: true,
    rechargeLoading: false
  },

  onLoad() {
    this.loadConsumeRecords()
  },

  onPullDownRefresh() {
    if (this.data.currentTab === 'consume') {
      this.setData({ consumePage: 1, consumeList: [], consumeHasMore: true })
      this.loadConsumeRecords().finally(() => wx.stopPullDownRefresh())
    } else {
      this.setData({ rechargePage: 1, rechargeList: [], rechargeHasMore: true })
      this.loadRechargeRecords().finally(() => wx.stopPullDownRefresh())
    }
  },

  switchTab(e) {
    const tab = e.currentTarget.dataset.tab
    this.setData({ currentTab: tab })
    
    if (tab === 'consume' && this.data.consumeList.length === 0) {
      this.loadConsumeRecords()
    } else if (tab === 'recharge' && this.data.rechargeList.length === 0) {
      this.loadRechargeRecords()
    }
  },

  // 加载消费记录
  async loadConsumeRecords() {
    if (this.data.consumeLoading || !this.data.consumeHasMore) return

    this.setData({ consumeLoading: true })

    try {
      const res = await api.getConsumeRecords(this.data.consumePage, 20)
      
      if (res.data.code === 200 && res.data.data) {
        // 兼容多种返回格式
        const data = res.data.data
        const list = data.records || data.list || data.List || []
        const total = data.total || data.Total || (data.Pagination && data.Pagination.Total) || 0
        
        // 处理数据
        const processedList = list.map(item => ({
          ...item,
          created_at_display: this.formatTime(item.created_at),
          coins: item.coins || item.coin || item.amount || 0,
          reason: item.reason || item.description || item.remark || '消费',
          type_display: this.getTypeDisplay(item.type || item.consume_type || ''),
          // 处理金额显示：根据类型确定正负
          amount_display: this.formatAmount(item.coins || item.coin || item.amount || 0, item.type || item.consume_type || '')
        }))

        const newList = this.data.consumePage === 1 ? processedList : [...this.data.consumeList, ...processedList]
        const hasMore = total > 0 ? newList.length < total : false

        this.setData({
          consumeList: newList,
          consumeHasMore: hasMore,
          consumePage: this.data.consumePage + 1
        })
      }
    } catch (error) {
      console.error('加载消费记录失败', error)
      wx.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      this.setData({ consumeLoading: false })
    }
  },

  // 加载充值记录
  async loadRechargeRecords() {
    if (this.data.rechargeLoading || !this.data.rechargeHasMore) return

    this.setData({ rechargeLoading: true })

    try {
      const res = await api.getRechargeRecords(this.data.rechargePage, 20)
      
      if (res.data.code === 200 && res.data.data) {
        // 兼容新旧两种返回格式
        const data = res.data.data
        const list = data.records || data.List || []
        const total = data.total || (data.Pagination && data.Pagination.Total) || 0
        
        // 处理数据
        const processedList = list.map(item => ({
          ...item,
          created_at_display: this.formatTime(item.created_at),
          amount_display: (item.amount / 100).toFixed(2), // 分转元
          status_display: this.getStatusDisplay(item.status)
        }))

        const newList = this.data.rechargePage === 1 ? processedList : [...this.data.rechargeList, ...processedList]
        const hasMore = newList.length < total

        this.setData({
          rechargeList: newList,
          rechargeHasMore: hasMore,
          rechargePage: this.data.rechargePage + 1
        })
      }
    } catch (error) {
      console.error('加载充值记录失败', error)
      wx.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      this.setData({ rechargeLoading: false })
    }
  },

  loadMoreConsume() {
    this.loadConsumeRecords()
  },

  loadMoreRecharge() {
    this.loadRechargeRecords()
  },

  // 格式化金额显示
  formatAmount(amount, type) {
    // 消费类型（减少金额）
    const consumeTypes = ['collision', 'collision_submit', 'renew_collision', 'force_add', 'haidilao', 'send_email']
    // 收入类型（增加金额）
    const incomeTypes = ['recharge', 'refund', 'match_reward', 'system', 'collision_refund']
    
    if (consumeTypes.includes(type)) {
      // 消费显示为负数
      return -Math.abs(amount)
    } else if (incomeTypes.includes(type)) {
      // 收入显示为正数
      return Math.abs(amount)
    } else {
      // 其他情况按原值显示
      return amount
    }
  },

  // 格式化时间
  formatTime(timeStr) {
    if (!timeStr) return ''
    const date = new Date(timeStr)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hour = String(date.getHours()).padStart(2, '0')
    const minute = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day} ${hour}:${minute}`
  },

  // 消费类型显示
  getTypeDisplay(type) {
    const typeMap = {
      'collision': '碰撞提交',
      'collision_submit': '碰撞提交',
      'renew_collision': '续期碰撞',
      'collision_refund': '审核返还',
      'force_add': '强制添加',
      'match_reward': '匹配奖励',
      'haidilao': '海底捞',
      'send_email': '发送邮件',
      'recharge': '充值',
      'refund': '退款',
      'system': '系统调整'
    }
    return typeMap[type] || type
  },

  // 充值状态显示
  getStatusDisplay(status) {
    const statusMap = {
      'pending': '待支付',
      'success': '成功',
      'failed': '失败'
    }
    return statusMap[status] || status
  }
})
