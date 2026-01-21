<template>
  <div class="audit-settings-page">
    <el-card class="settings-card">
      <div class="page-header">
        <h2>审核设置</h2>
        <el-button @click="goBack">返回碰撞码管理</el-button>
      </div>
      
      <!-- 审核开关 -->
      <div class="setting-section">
        <div class="setting-item">
          <div class="setting-left">
            <div class="setting-icon">
              <el-icon :size="24"><Check /></el-icon>
            </div>
            <div class="setting-info">
              <div class="setting-title">碰撞码审核</div>
              <div class="setting-desc">开启后，用户提交的碰撞码需要管理员审核通过后才能参与碰撞匹配</div>
            </div>
          </div>
          <el-switch
            v-model="enableCollisionAudit"
            :loading="switchLoading"
            @change="handleAuditChange"
            size="large"
            active-text="开"
            inactive-text="关"
          />
        </div>
      </div>
    </el-card>

    <!-- 审核统计 -->
    <el-card class="stats-card">
      <template #header>
        <div class="card-header">
          <span>审核统计</span>
          <el-button type="primary" text @click="loadStats">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <el-row :gutter="20" v-loading="statsLoading">
        <el-col :span="6">
          <div class="stat-card pending" @click="goToList('pending')">
            <div class="stat-icon">⏳</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.pending }}</div>
              <div class="stat-label">待审核</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card approved" @click="goToList('approved')">
            <div class="stat-icon">✅</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.approved }}</div>
              <div class="stat-label">已通过</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card rejected" @click="goToList('rejected')">
            <div class="stat-icon">❌</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.rejected }}</div>
              <div class="stat-label">已拒绝</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card total">
            <div class="stat-icon">📊</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total }}</div>
              <div class="stat-label">总计</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 快捷操作 -->
    <el-card class="actions-card">
      <template #header>
        <span>快捷操作</span>
      </template>
      <div class="quick-actions">
        <el-button type="warning" size="large" @click="goToList('pending')">
          <el-icon><Bell /></el-icon>
          处理待审核 ({{ stats.pending }})
        </el-button>
        <el-button type="success" size="large" @click="batchApproveAll" :disabled="stats.pending === 0">
          <el-icon><Check /></el-icon>
          全部通过
        </el-button>
        <el-button size="large" @click="goBack">
          <el-icon><List /></el-icon>
          查看全部碰撞码
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Refresh, Bell, List } from '@element-plus/icons-vue'
import api from '@/utils/api'

const router = useRouter()
const switchLoading = ref(false)
const statsLoading = ref(false)
const enableCollisionAudit = ref(false)

const stats = reactive({
  pending: 0,
  approved: 0,
  rejected: 0,
  total: 0
})

const loadSettings = async () => {
  try {
    const response = await api.get('/dashboard/audit-setting')
    if (response.data.code === 200) {
      enableCollisionAudit.value = response.data.data?.enableCollisionAudit || false
    }
  } catch (error) {
    console.error('加载设置失败:', error)
  }
}

const loadStats = async () => {
  statsLoading.value = true
  try {
    const response = await api.get('/dashboard/audit-stats')
    if (response.data.code === 200 && response.data.data) {
      stats.pending = response.data.data.pending || 0
      stats.approved = response.data.data.approved || 0
      stats.rejected = response.data.data.rejected || 0
      stats.total = response.data.data.total || 0
    }
  } catch (error) {
    console.error('加载统计失败:', error)
  } finally {
    statsLoading.value = false
  }
}

const handleAuditChange = async (value) => {
  switchLoading.value = true
  try {
    const response = await api.put('/dashboard/audit-setting', {
      enableCollisionAudit: value
    })
    
    if (response.data.code === 200) {
      ElMessage.success(value ? '已开启碰撞码审核' : '已关闭碰撞码审核')
    } else {
      enableCollisionAudit.value = !value
      ElMessage.error(response.data.message || '设置失败')
    }
  } catch (error) {
    enableCollisionAudit.value = !value
    console.error('设置失败:', error)
    ElMessage.error('设置失败')
  } finally {
    switchLoading.value = false
  }
}

// 全部通过
const batchApproveAll = async () => {
  if (stats.pending === 0) return
  
  try {
    await ElMessageBox.confirm(`确定将所有 ${stats.pending} 条待审核碰撞码全部通过？`, '确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const response = await api.put('/collisions/batch-approve-all')
    if (response.data.code === 200) {
      ElMessage.success('全部通过成功')
      loadStats()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量通过失败:', error)
      ElMessage.error('操作失败')
    }
  }
}

const goToList = (auditStatus) => {
  if (auditStatus === 'pending') {
    router.push('/audit-list')
  } else {
    router.push({ path: '/collision-codes', query: { audit_status: auditStatus } })
  }
}

const goBack = () => {
  router.push('/collision-codes')
}

onMounted(() => {
  loadSettings()
  loadStats()
})
</script>

<style scoped>
.audit-settings-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 18px;
  color: #303133;
}

/* 设置区域 */
.setting-section {
  padding: 10px 0;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e7ed 100%);
  border-radius: 12px;
}

.setting-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.setting-icon {
  width: 48px;
  height: 48px;
  background: #409EFF;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.setting-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
}

.setting-desc {
  font-size: 13px;
  color: #909399;
  max-width: 400px;
}

/* 统计卡片 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-card.pending {
  background: linear-gradient(135deg, #FFF8E6 0%, #FFF1CC 100%);
  border: 1px solid #FAECD8;
}

.stat-card.approved {
  background: linear-gradient(135deg, #E6F7E6 0%, #D4EDDA 100%);
  border: 1px solid #C3E6CB;
}

.stat-card.rejected {
  background: linear-gradient(135deg, #FDECEA 0%, #F8D7DA 100%);
  border: 1px solid #F5C6CB;
}

.stat-card.total {
  background: linear-gradient(135deg, #E8F4FD 0%, #D1E9FC 100%);
  border: 1px solid #B8DAFF;
  cursor: default;
}

.stat-card.total:hover {
  transform: none;
  box-shadow: none;
}

.stat-icon {
  font-size: 32px;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
}

.stat-card.pending .stat-value { color: #E6A23C; }
.stat-card.approved .stat-value { color: #67C23A; }
.stat-card.rejected .stat-value { color: #F56C6C; }
.stat-card.total .stat-value { color: #409EFF; }

.stat-label {
  font-size: 13px;
  color: #606266;
  margin-top: 4px;
}

/* 快捷操作 */
.quick-actions {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}
</style>
