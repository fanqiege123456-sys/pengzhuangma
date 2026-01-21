<template>
  <el-container class="layout-container">
    <el-aside class="sidebar" :width="sidebarWidth">
      <div class="logo">
        <h3>碰撞码管理</h3>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
        :collapse="isCollapse"
        router
      >
        <el-menu-item index="/">
          <el-icon><House /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        
        <el-menu-item index="/collision-codes">
          <el-icon><List /></el-icon>
          <span>碰撞码管理</span>
        </el-menu-item>
        
        <el-menu-item index="/audit-settings">
          <el-icon><Check /></el-icon>
          <span>审核设置</span>
        </el-menu-item>
        
        <el-menu-item index="/audit-list">
          <el-icon><List /></el-icon>
          <span>待审核列表</span>
        </el-menu-item>
        
        <el-menu-item index="/keywords">
          <el-icon><Star /></el-icon>
          <span>热门关键词</span>
        </el-menu-item>

        <el-menu-item index="/forbidden-keywords">
          <el-icon><List /></el-icon>
          <span>违禁词</span>
        </el-menu-item>
        
        <el-menu-item index="/records">
          <el-icon><Document /></el-icon>
          <span>碰撞记录</span>
        </el-menu-item>
        
        <el-menu-item index="/email-logs">
          <el-icon><Message /></el-icon>
          <span>邮件日志</span>
        </el-menu-item>
        
        <el-menu-item index="/admins">
          <el-icon><Setting /></el-icon>
          <span>管理员管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-button
            type="text"
            size="large"
            @click="toggleSidebar"
          >
            <el-icon><Expand v-if="isCollapse" /><Fold v-else /></el-icon>
          </el-button>
        </div>
        
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="dropdown-link">
              <el-avatar :size="30" :src="avatarUrl" />
              <span style="margin-left: 8px">{{ authStore.admin.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <router-view />
        <div class="footer-beian">
          <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener noreferrer">
            津ICP备2024024482号
          </a>
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import {
  House,
  User,
  List,
  Star,
  Document,
  Setting,
  Expand,
  Fold,
  ArrowDown,
  Check,
  Message
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isCollapse = ref(false)
const avatarUrl = ref('https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png')

const activeMenu = computed(() => route.path)
const sidebarWidth = computed(() => isCollapse.value ? '64px' : '240px')

const toggleSidebar = () => {
  isCollapse.value = !isCollapse.value
}

const handleCommand = async (command) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确认退出登录？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      
      authStore.logout()
      ElMessage.success('退出成功')
      router.push('/login')
    } catch (error) {
      // 用户取消退出
    }
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  width: 100vw;
  min-width: 1200px;
}

.sidebar {
  background-color: #304156;
  transition: width 0.28s;
  height: 100vh;
  overflow: auto;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #2b3a4b;
  min-height: 60px;
}

.logo h3 {
  color: #fff;
  font-size: 16px;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
}

.sidebar-menu {
  border-right: none;
  height: calc(100vh - 60px);
}

.header {
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0,21,41,.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 60px;
  min-height: 60px;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.dropdown-link {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #606266;
  white-space: nowrap;
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
  height: calc(100vh - 60px);
  overflow: auto;
  min-height: calc(100vh - 60px);
  position: relative;
}

.footer-beian {
  position: fixed;
  bottom: 10px;
  right: 20px;
  z-index: 1000;
}

.footer-beian a {
  color: #999;
  text-decoration: none;
  font-size: 12px;
  padding: 4px 8px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 4px;
  transition: all 0.3s ease;
}

.footer-beian a:hover {
  color: #409eff;
  background: rgba(255, 255, 255, 0.95);
}

/* PC端响应式设计 */
@media (max-width: 1200px) {
  .layout-container {
    min-width: 1000px;
  }
  
  .main-content {
    padding: 15px;
  }
}

/* 确保侧边栏折叠时的样式 */
.sidebar.el-aside--collapse {
  width: 64px !important;
}

.sidebar.el-aside--collapse .logo h3 {
  display: none;
}
</style>
