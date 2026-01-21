<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h2>碰撞码管理后台</h2>
      </div>
      
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="rules"
        class="login-form"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>
        
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            size="large"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>
        
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            class="login-button"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    
    <!-- 备案信息底栏 -->
    <div class="footer-beian">
      <div class="beian-content">
        <div class="beian-links">
          <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener noreferrer">
            津ICP备2024024482号
          </a>
          <span class="separator">|</span>
          <span>© 2024 碰撞码管理系统</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const loginFormRef = ref()
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const result = await authStore.login(loginForm.username, loginForm.password)
        if (result.success) {
          ElMessage.success('登录成功')
          router.push('/')
        } else {
          ElMessage.error(result.message || '登录失败')
        }
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow: hidden;
  padding-bottom: 60px; /* 为底栏留出空间 */
}

.login-box {
  width: 420px;
  max-width: 90vw;
  padding: 45px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
}

.login-header {
  text-align: center;
  margin-bottom: 35px;
}

.login-header h2 {
  color: #333;
  font-weight: 600;
  font-size: 26px;
  margin: 0;
}

.login-form {
  width: 100%;
}

.login-button {
  width: 100%;
  height: 45px;
  font-size: 16px;
}

/* PC端优化 */
@media (min-width: 768px) {
  .login-box {
    width: 450px;
    padding: 50px;
  }
  
  .login-header h2 {
    font-size: 28px;
  }
}

/* 确保在小屏幕上也能正常显示 */
@media (max-width: 480px) {
  .login-box {
    width: 95vw;
    padding: 30px 25px;
  }
  
  .login-header h2 {
    font-size: 22px;
  }
}

/* 备案信息底栏 */
.footer-beian {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(10px);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  z-index: 1000;
}

.beian-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 12px 20px;
}

.beian-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
}

.beian-links a {
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  transition: color 0.3s ease;
}

.beian-links a:hover {
  color: #409eff;
}

.separator {
  color: rgba(255, 255, 255, 0.4);
  margin: 0 4px;
}

/* 移动端备案信息适配 */
@media (max-width: 768px) {
  .beian-content {
    padding: 10px 15px;
  }
  
  .beian-links {
    font-size: 11px;
    flex-direction: column;
    gap: 4px;
  }
  
  .separator {
    display: none;
  }
}
</style>
