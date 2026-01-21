// pages/email-bind/email-bind.js
const api = require('../../utils/api.js')

Page({
  data: {
    email: '',
    verifyCode: '',
    showVerifyCode: false,
    countdown: 0,
    currentEmail: '',
    isVerified: false,
    timer: null
  },

  onLoad() {
    this.loadCurrentEmail()
  },

  onUnload() {
    // 清除定时器
    if (this.data.timer) {
      clearInterval(this.data.timer)
    }
  },

  // 加载当前邮箱
  async loadCurrentEmail() {
    try {
      const res = await api.getUserContacts()
      if (res.data.code === 200 && res.data.data) {
        const { email, email_verified } = res.data.data
        this.setData({
          currentEmail: email || '',
          isVerified: email_verified || false
        })
      }
    } catch (error) {
      console.error('加载邮箱信息失败', error)
    }
  },

  // 输入邮箱
  onEmailInput(e) {
    this.setData({
      email: e.detail.value
    })
  },

  // 输入验证码
  onVerifyCodeInput(e) {
    this.setData({
      verifyCode: e.detail.value
    })
  },

  // 验证邮箱格式
  validateEmail(email) {
    const reg = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
    return reg.test(email)
  },

  // 发送验证码
  async sendVerifyCode() {
    const { email } = this.data

    if (!email) {
      wx.showToast({
        title: '请输入邮箱地址',
        icon: 'none'
      })
      return
    }

    if (!this.validateEmail(email)) {
      wx.showToast({
        title: '邮箱格式不正确',
        icon: 'none'
      })
      return
    }

    try {
      wx.showLoading({ title: '发送中...' })

      const res = await api.bindEmail({ email })

      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '验证码已发送',
          icon: 'success'
        })
        
        this.setData({
          showVerifyCode: true
        })

        this.startCountdown()
      } else {
        wx.showToast({
          title: res.data.message || '发送失败',
          icon: 'none'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('发送验证码失败', error)
      wx.showToast({
        title: '发送失败',
        icon: 'none'
      })
    }
  },

  // 开始倒计时
  startCountdown() {
    this.setData({ countdown: 60 })
    
    const timer = setInterval(() => {
      const { countdown } = this.data
      
      if (countdown <= 1) {
        clearInterval(timer)
        this.setData({ 
          countdown: 0,
          timer: null
        })
      } else {
        this.setData({ 
          countdown: countdown - 1 
        })
      }
    }, 1000)

    this.setData({ timer })
  },

  // 获取验证码按钮
  handleBind() {
    this.sendVerifyCode()
  },

  // 验证并绑定
  async handleVerify() {
    const { email, verifyCode } = this.data

    if (!verifyCode) {
      wx.showToast({
        title: '请输入验证码',
        icon: 'none'
      })
      return
    }

    if (verifyCode.length !== 6) {
      wx.showToast({
        title: '验证码格式不正确',
        icon: 'none'
      })
      return
    }

    try {
      wx.showLoading({ title: '验证中...' })

      const res = await api.verifyEmail({
        email,
        code: verifyCode
      })

      wx.hideLoading()

      if (res.data.code === 200) {
        wx.showToast({
          title: '绑定成功',
          icon: 'success'
        })

        setTimeout(() => {
          wx.navigateBack()
        }, 1500)
      } else {
        wx.showToast({
          title: res.data.message || '验证失败',
          icon: 'none'
        })
      }
    } catch (error) {
      wx.hideLoading()
      console.error('验证失败', error)
      wx.showToast({
        title: '验证失败',
        icon: 'none'
      })
    }
  },

  // 更换邮箱
  handleChange() {
    wx.showModal({
      title: '更换邮箱',
      content: '确定要更换绑定的邮箱吗?',
      success: (res) => {
        if (res.confirm) {
          this.setData({
            currentEmail: '',
            isVerified: false,
            email: '',
            verifyCode: '',
            showVerifyCode: false
          })
        }
      }
    })
  },

  // 下拉刷新
  async onPullDownRefresh() {
    console.log('开始下拉刷新...')
    try {
      await this.loadCurrentEmail()
      console.log('下拉刷新完成')
    } catch (error) {
      console.error('下拉刷新失败:', error)
    } finally {
      wx.stopPullDownRefresh()
    }
  }
})
