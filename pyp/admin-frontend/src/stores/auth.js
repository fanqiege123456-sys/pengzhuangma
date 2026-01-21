import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/utils/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const admin = ref(JSON.parse(localStorage.getItem('admin') || '{}'))

  const isAuthenticated = computed(() => !!token.value)

  const login = async (username, password) => {
    try {
      const response = await api.post('/admin/login', {
        username,
        password
      })
      
      if (response.data.code === 200) {
        const { token: newToken, admin: adminInfo } = response.data.data
        token.value = newToken
        admin.value = adminInfo
        
        localStorage.setItem('token', newToken)
        localStorage.setItem('admin', JSON.stringify(adminInfo))
        
        return { success: true }
      } else {
        return { success: false, message: response.data.msg }
      }
    } catch (error) {
      console.error('Login error:', error)
      return { success: false, message: '登录失败' }
    }
  }

  const logout = () => {
    token.value = ''
    admin.value = {}
    localStorage.removeItem('token')
    localStorage.removeItem('admin')
  }

  return {
    token,
    admin,
    isAuthenticated,
    login,
    logout
  }
})
