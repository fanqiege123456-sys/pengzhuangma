// utils/api.js
const app = getApp()

const api = {
  // 用户相关
  login: (data) => app.request({
    url: '/user/login',
    method: 'POST',
    data
  }),

  getUserInfo: () => app.request({
    url: '/user/info'
  }),

  updateProfile: (data) => app.request({
    url: '/user/profile',
    method: 'PUT',
    data
  }),

  getBalance: () => app.request({
    url: '/user/balance'
  }),

  // 碰撞码相关
  submitCollisionCode: (data) => app.request({
    url: '/collision/submit',
    method: 'POST',
    data
  }),

  getMatches: () => app.request({
    url: '/collision/matches'
  }),

  getHotCodes: () => app.request({
    url: '/collision/hot-codes'
  }),

  // 获取我的碰撞码
  getMyCollisionCode: () => app.request({
    url: '/collision/my-code'
  }),

  // 搜索碰撞码
  searchCollisionCodes: (keyword) => {
    console.log('API调用 - 搜索碰撞码, 关键词:', keyword)
    const requestData = { keyword }
    console.log('请求数据:', requestData)
    
    return app.request({
      url: '/collision/search',
      method: 'POST',
      data: requestData
    })
  },

  // 发送邮件给匹配用户
  sendEmailToMatchedUser: (matchedUserID, keyword, message) => app.request({
    url: '/collision/send-email',
    method: 'POST',
    data: {
      matched_user_id: matchedUserID,
      keyword: keyword,
      message: message
    }
  }),

  // 批量提交碰撞码
  batchSubmitCollisionCodes: (data) => app.request({
    url: '/collision/batch-submit',
    method: 'POST',
    data
  }),

  // 获取用户所有碰撞码
  getMyCollisionCodes: () => app.request({
    url: '/collision/my-codes'
  }),

  // 续期碰撞码
  renewCollisionCode: (id) => app.request({
    url: `/collision/my-codes/${id}/renew`,
    method: 'POST'
  }),

  // 删除碰撞码
  deleteCollisionCode: (id) => app.request({
    url: `/collision/my-codes/${id}`,
    method: 'DELETE'
  }),

  // 获取碰撞码详情
  getCollisionCode: (codeId) => app.request({
    url: `/user/collision-codes/${codeId}`
  }),

  // 修改碰撞码
  updateCollisionCode: (codeId, data) => app.request({
    url: `/user/collision-codes/${codeId}`,
    method: 'PUT',
    data
  }),

  // 重新提交被拒绝的碰撞码
  resubmitCollisionCode: (id) => app.request({
    url: `/collision/my-codes/${id}/resubmit`,
    method: 'POST'
  }),

  // 海底捞
  haidilao: (data) => app.request({
    url: '/collision/haidilao',
    method: 'POST',
    data
  }),

  // 强制添加好友
  forceAddFriend: (data) => app.request({
    url: '/collision/force-add-friend',
    method: 'POST',
    data
  }),

  // 用户地址管理
  updateUserLocation: (data) => app.request({
    url: '/user/location',
    method: 'PUT',
    data
  }),

  // 多地址管理
  getLocations: () => app.request({
    url: '/locations'
  }),

  createLocation: (data) => app.request({
    url: '/locations',
    method: 'POST',
    data
  }),

  updateLocation: (id, data) => app.request({
    url: `/locations/${id}`,
    method: 'PUT',
    data
  }),

  deleteLocation: (id) => app.request({
    url: `/locations/${id}`,
    method: 'DELETE'
  }),

  setDefaultLocation: (id) => app.request({
    url: `/locations/${id}/default`,
    method: 'PUT'
  }),

  // ========== V3.0 旧代码注释开始 ==========
  // 匹配相关 (已废弃,改为邮件通知)
  // getMatchList: (params = {}) => app.request({
  //   url: '/collision/matches',
  //   data: params
  // }),

  // getMatchDetail: (id) => app.request({
  //   url: `/collision/matches/${id}`
  // }),

  // addFriend: (matchId) => app.request({
  //   url: `/collision/add-friend`,
  //   method: 'POST',
  //   data: { match_id: matchId }
  // }),

  // skipMatch: (matchId) => app.request({
  //   url: `/match/${matchId}/skip`,
  //   method: 'POST'
  // }),

  // 好友相关 (已废弃,改为邮件通知)
  // getFriends: (params = {}) => app.request({
  //   url: '/friends',
  //   data: params
  // }),

  // sendFriendRequest: (data) => app.request({
  //   url: '/collision/send-friend-request',
  //   method: 'POST',
  //   data
  // }),

  // blockFriend: (friendId) => app.request({
  //   url: `/friends/${friendId}/block`,
  //   method: 'PUT'
  // }),

  // unblockFriend: (friendId) => app.request({
  //   url: `/friends/${friendId}/unblock`,
  //   method: 'PUT'
  // }),

  // deleteFriend: (friendId) => app.request({
  //   url: `/friends/${friendId}`,
  //   method: 'DELETE'
  // }),
  // ========== V3.0 旧代码注释结束 ==========

  // ========== V3.0 新增 API ==========
  // 热门标签
  getHotTags24h: () => app.request({
    url: '/hot-tags/24h'
  }),

  getHotTagsAll: () => app.request({
    url: '/hot-tags/all'
  }),

  // 点击热门标签
  clickHotTag: (keyword) => app.request({
    url: '/hot-tags/click',
    method: 'POST',
    data: { keyword }
  }),

  // 碰撞列表管理
  createCollisionList: (data) => app.request({
    url: '/collision-lists',
    method: 'POST',
    data
  }),

  getCollisionLists: () => app.request({
    url: '/collision-lists'
  }),

  updateCollisionList: (id, data) => app.request({
    url: `/collision-lists/${id}`,
    method: 'PUT',
    data
  }),

  deleteCollisionList: (id) => app.request({
    url: `/collision-lists/${id}`,
    method: 'DELETE'
  }),

  // 碰撞结果
  getCollisionResults: (days, keyword) => app.request({
    url: `/collision-results?days=${days}${keyword ? '&keyword=' + encodeURIComponent(keyword) : ''}`
  }),

  markAsKnown: (id) => app.request({
    url: `/collision-results/${id}/mark-known`,
    method: 'POST'
  }),

  // 用户联系方式
  bindEmail: (data) => app.request({
    url: '/user/email/bind',
    method: 'POST',
    data
  }),

  verifyEmail: (data) => app.request({
    url: '/user/email/verify',
    method: 'POST',
    data
  }),

  getUserContacts: () => app.request({
    url: '/user/contacts'
  }),

  // 更新邮箱显示设置
  updateEmailVisibility: (data) => app.request({
    url: '/user/email/visibility',
    method: 'PUT',
    data
  }),

  bindPhone: (data) => app.request({
    url: '/user/phone/bind',
    method: 'POST',
    data
  }),

  // 碰撞结果详情(分页)
  getCollisionResultDetail: (id, params) => app.request({
    url: `/collision-results/${id}/detail`,
    data: params
  }),

  // 更新匹配备注
  updateMatchRemark: (id, data) => app.request({
    url: `/collision-results/${id}/remark`,
    method: 'PUT',
    data
  }),

  // 发送邮件给匹配用户
  sendEmailToMatch: (data) => app.request({
    url: '/collision-results/send-email',
    method: 'POST',
    data
  }),

  // 获取与指定用户的共同关键词
  getCommonKeywords: (matchedUserId) => app.request({
    url: '/collision-results/common-keywords',
    method: 'POST',
    data: { matched_user_id: matchedUserId }
  }),

  // 碰撞动态列表
  getCollisionSparks: (params = {}) => app.request({
    url: '/collision-sparks',
    data: params
  }),
  // ========== V3.0 新增 API 结束 ==========

  // 充值相关
  createRechargeOrder: (amount) => app.request({
    url: '/recharge/create',
    method: 'POST',
    data: { amount }
  }),

  // 消费记录
  getConsumeRecords: (page = 1, pageSize = 10) => app.request({
    url: `/user/consume-records?page=${page}&page_size=${pageSize}`
  }),

  // 充值记录
  getRechargeRecords: (page = 1, pageSize = 10) => app.request({
    url: `/user/recharge-records?page=${page}&page_size=${pageSize}`
  }),

  // ========== V3.0 聊天相关 (已废弃) ==========
  // sendMessage: (data) => app.request({
  //   url: '/chat/send',
  //   method: 'POST',
  //   data
  // }),

  // getChatHistory: (userId) => app.request({
  //   url: `/chat/history/${userId}`
  // })
  // ========== V3.0 聊天相关结束 ==========
}

module.exports = api
