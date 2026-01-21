App({
  globalData: {
    userInfo: null,
    token: null,
    apiUrl: 'https://p.chediaodu.com/api',
    hasLogin: false,
    balance: 0
  },

  onLaunch() {
    this.checkAndAutoLogin()
  },
  setApiUrl(url) {
    if (!url) return
    this.globalData.apiUrl = url
    wx.setStorageSync('apiUrl', url)
    console.log('API 地址已切换为:', url)
  },
  checkAndAutoLogin() {
    const token = wx.getStorageSync('token')
    const userInfo = wx.getStorageSync('userInfo')

    if (token && userInfo) {
      this.globalData.token = token
      this.globalData.userInfo = userInfo
      this.globalData.hasLogin = true
      console.log('使用本地缓存登录:', userInfo)

      this.validateToken().catch(() => {
        console.log('Token 已失效，需要重新登录')
        this.globalData.hasLogin = false
      })
    } else {
      this.globalData.hasLogin = false
      console.log('未登录，等待用户授权')
    }
  },
  validateToken() {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${this.globalData.apiUrl}/user/info`,
        method: 'GET',
        header: {
          'Authorization': `Bearer ${this.globalData.token}`
        },
        success: (res) => {
          if (res.data.code === 200) {
            console.log('Token 验证成功')
            resolve(res.data.data)
          } else {
            console.log('Token 已失效，需要重新登录')
            wx.removeStorageSync('token')
            wx.removeStorageSync('userInfo')
            this.globalData.token = null
            this.globalData.userInfo = null
            this.globalData.hasLogin = false
            reject()
          }
        },
        fail: reject
      })
    })
  },

  requireLogin(showModal = false) {
    if (!this.globalData.hasLogin) {
      if (showModal) {
        wx.showModal({
          title: '需要登录',
          content: '该功能需要登录后才能使用哦~',
          confirmText: '去登录',
          cancelText: '稍后再说',
          success: (res) => {
            if (res.confirm) {
              wx.showLoading({
                title: '登录中...',
                mask: true
              })

              this.login(true)
                .then(() => {
                  wx.hideLoading()

                  wx.showToast({
                    title: '登录成功',
                    icon: 'success',
                    duration: 1000
                  })

                  setTimeout(() => {
                    const pages = getCurrentPages()
                    const currentPage = pages[pages.length - 1]
                    if (currentPage && currentPage.onShow) {
                      currentPage.onShow()
                    }
                  }, 1000)
                })
                .catch((err) => {
                  wx.hideLoading()

                  console.error('登录失败:', err)
                  wx.showToast({
                    title: err.message || '登录失败',
                    icon: 'none',
                    duration: 2000
                  })
                })
            }
          }
        })
      }
      return false
    }
    return true
  },

  login(withProfile = true) {
    return new Promise((resolve, reject) => {
      console.log('开始微信登录流程')

      wx.showLoading({
        title: '登录中...'
      })

      const doLogin = (userInfo = null) => {
        wx.login({
          success: (loginRes) => {
            console.log('wx.login 成功，获取到 code:', loginRes.code)

            if (loginRes.code) {
              this.loginWithCode(loginRes.code, userInfo)
                .then((user) => {
                  wx.hideLoading()
                  resolve(user)
                })
                .catch((err) => {
                  wx.hideLoading()
                  reject(err)
                })
            } else {
              wx.hideLoading()
              console.error('wx.login 失败，未获取到 code:', loginRes.errMsg)
              reject(new Error('获取登录凭证失败'))
            }
          },
          fail: (err) => {
            wx.hideLoading()
            console.error('wx.login 调用失败:', err)
            wx.showToast({
              title: '登录失败',
              icon: 'none'
            })
            reject(new Error('微信登录失败'))
          }
        })
      }

      if (withProfile) {
        wx.getUserProfile({
          desc: '用于完善昵称和头像',
          success: (profileRes) => {
            doLogin(profileRes.userInfo || null)
          },
          fail: () => {
            wx.hideLoading()
            wx.showToast({
              title: '登录失败',
              icon: 'none'
            })
            reject(new Error('用户未授权昵称和头像'))
          }
        })
      } else {
        doLogin()
      }
    })
  },

  loginWithCode(code, userInfo = null) {
    return new Promise((resolve, reject) => {
      const payload = { code }
      if (userInfo) {
        payload.userInfo = userInfo
      }

      wx.request({
        url: `${this.globalData.apiUrl}/user/login`,
        method: 'POST',
        data: payload,
        header: {
          'content-type': 'application/json'
        },
        success: (res) => {
          console.log('登录请求响应:', res.data)

          if (res.data.code === 200) {
            const { token, user } = res.data.data

            wx.setStorageSync('token', token)
            wx.setStorageSync('userInfo', user)

            this.globalData.token = token
            this.globalData.userInfo = user
            this.globalData.hasLogin = true

            console.log('登录成功，用户信息:', user)
            resolve(user)
          } else {
            console.error('后端登录失败:', res.data.msg)
            reject(new Error(res.data.msg || '登录失败'))
          }
        },
        fail: (err) => {
          console.error('登录请求网络失败:', err)
          reject(err)
        }
      })
    })
  },
  request(options) {
    const { url, method = 'GET', data = {}, header = {} } = options

    const fullUrl = url.startsWith('http') ? url : `${this.globalData.apiUrl}${url}`
    console.log('发起请求:', {
      url: fullUrl,
      method,
      data,
      hasToken: !!this.globalData.token
    })

    return new Promise((resolve, reject) => {
      wx.request({
        url: fullUrl,
        method,
        data,
        header: {
          'content-type': 'application/json',
          'Authorization': this.globalData.token ? `Bearer ${this.globalData.token}` : '',
          ...header
        },
        success: (res) => {
          console.log('请求响应:', {
            url: fullUrl,
            statusCode: res.statusCode,
            data: res.data
          })

          if (res.data.code === 401) {
            this.login(false).then(() => {
              this.request(options).then(resolve).catch(reject)
            }).catch(reject)
          } else {
            resolve(res)
          }
        },
        fail: (error) => {
          console.error('请求失败:', {
            url: fullUrl,
            error
          })
          reject(error)
        }
      })
    })
  },

  logout() {
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
    this.globalData.token = null
    this.globalData.userInfo = null
    this.globalData.hasLogin = false
    
    wx.reLaunch({
      url: '/pages/index/index'
    })
  }
})










