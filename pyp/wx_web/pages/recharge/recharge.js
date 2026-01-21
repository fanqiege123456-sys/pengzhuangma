// pages/recharge/recharge.js
const api = require('../../utils/api.js')

Page({
  data: {
    currentBalance: 0,
    selectedAmount: 0,
    customAmount: '',
    finalAmount: 0,
    submitting: false,
    consumeRecords: [], // 消费记录
    amountOptions: [
      { value: 10, desc: '' },
      { value: 20, desc: '' },
      { value: 50, desc: '推荐' },
      { value: 100, desc: '热门' },
      { value: 200, desc: '' },
      { value: 500, desc: '' }
    ]
  },

  onLoad() {
    this.loadBalance()
    this.loadConsumeRecords()
  },

  onShow() {
    // 刷新余额和消费记录
    this.loadBalance()
    this.loadConsumeRecords()
  },

  // 加载消费记录
  async loadConsumeRecords() {
    try {
      const res = await api.getConsumeRecords()
      
      if (res.data.code === 200) {
        let data = res.data.data || {}
        let records = []
        
        // 处理新的数据格式：data.records 是数组
        if (data.records && Array.isArray(data.records)) {
          records = data.records
        } else if (Array.isArray(data)) {
          // 兼容旧格式：data 直接是数组
          records = data
        } else {
          console.log('API返回的数据格式:', data)
          records = []
        }
        
        // 格式化数据
        records = records.map(item => {
          return {
            ...item,
            type_display: item.type_display || this.getTypeDisplay(item.type),
            created_at_display: item.created_at_display || this.formatDateTime(item.created_at),
            // 处理金额显示：负数表示消费，正数表示收入
            amount: item.coins || item.amount || 0,
            amount_display: this.formatAmount(item.coins || item.amount || 0, item.type)
          }
        })
        
        this.setData({
          consumeRecords: records
        })
      } else {
        console.error('获取消费记录失败:', res.data.msg)
        this.setData({
          consumeRecords: []
        })
      }
    } catch (error) {
      console.error('加载消费记录失败', error)
      this.setData({
        consumeRecords: []
      })
    }
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

  // 获取类型显示文本
  getTypeDisplay(type) {
    const typeMap = {
      'collision': '碰撞提交',
      'collision_submit': '碰撞提交',
      'renew_collision': '续期碰撞',
      'collision_match': '匹配奖励',
      'match_reward': '匹配奖励',
      'force_add': '强制添加',
      'haidilao': '海底捞',
      'send_email': '发送邮件',
      'recharge': '充值',
      'refund': '退款',
      'system': '系统调整'
    }
    return typeMap[type] || type
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
  async loadBalance() {
    try {
      const res = await api.getBalance()
      
      if (res.data.code === 200) {
        this.setData({
          currentBalance: (res.data.data.balance / 100).toFixed(2) // 分转元
        })
      }
    } catch (error) {
      console.error('加载余额失败', error)
    }
  },

  // 选择充值金额
  selectAmount(e) {
    const amount = parseFloat(e.currentTarget.dataset.amount)
    
    this.setData({
      selectedAmount: amount,
      customAmount: '',
      finalAmount: amount
    })
  },

  // 输入自定义金额
  onCustomAmountInput(e) {
    const value = e.detail.value
    
    this.setData({
      customAmount: value,
      selectedAmount: 0
    })
    
    // 实时更新最终金额
    const amount = parseFloat(value)
    if (!isNaN(amount) && amount > 0) {
      this.setData({
        finalAmount: amount
      })
    } else {
      this.setData({
        finalAmount: 0
      })
    }
  },

  // 自定义金额失焦验证
  onCustomAmountBlur(e) {
    const value = parseFloat(e.detail.value)
    
    if (isNaN(value) || value < 1) {
      this.setData({
        customAmount: '',
        finalAmount: 0
      })
      
      if (!isNaN(value) && value < 1) {
        wx.showToast({
          title: '充值金额最低1元',
          icon: 'none'
        })
      }
    } else if (value > 10000) {
      this.setData({
        customAmount: '10000',
        finalAmount: 10000
      })
      
      wx.showToast({
        title: '充值金额最高10000元',
        icon: 'none'
      })
    } else {
      // 保留两位小数
      const formattedValue = value.toFixed(2)
      this.setData({
        customAmount: formattedValue,
        finalAmount: parseFloat(formattedValue)
      })
    }
  },

  // 提交充值
  async submitRecharge() {
    const { finalAmount } = this.data
    
    if (!finalAmount || finalAmount < 1) {
      wx.showToast({
        title: '请选择充值金额',
        icon: 'none'
      })
      return
    }

    if (finalAmount > 10000) {
      wx.showToast({
        title: '充值金额不能超过10000元',
        icon: 'none'
      })
      return
    }

    this.setData({
      submitting: true
    })

    try {
      // 创建充值订单
      const orderRes = await api.createRechargeOrder(Math.round(finalAmount * 100)) // 元转分
      
      if (orderRes.data.code !== 200) {
        throw new Error(orderRes.data.msg || '创建订单失败')
      }

      const { order_id, prepay_id } = orderRes.data.data

      // 调起微信支付
      const payRes = await wx.requestPayment({
        timeStamp: Date.now().toString(),
        nonceStr: Math.random().toString(36).substr(2, 15),
        package: `prepay_id=${prepay_id}`,
        signType: 'MD5',
        paySign: 'mock_sign' // 这里需要后端生成真实的签名
      })

      // 支付成功
      wx.showToast({
        title: '充值成功',
        icon: 'success'
      })

      // 刷新余额和消费记录
      setTimeout(() => {
        this.loadBalance()
        this.loadConsumeRecords()
        // 重置选择状态
        this.setData({
          selectedAmount: 0,
          customAmount: '',
          finalAmount: 0
        })
      }, 1500)

    } catch (error) {
      console.error('充值失败', error)
      
      if (error.errMsg && error.errMsg.includes('cancel')) {
        wx.showToast({
          title: '支付已取消',
          icon: 'none'
        })
      } else {
        wx.showToast({
          title: error.message || '充值失败',
          icon: 'error'
        })
      }
    } finally {
      this.setData({
        submitting: false
      })
    }
  },

  // 下拉刷新
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    try {
      await this.loadBalance()
      await this.loadConsumeRecords()
      console.log('下拉刷新完成')
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  }
})
