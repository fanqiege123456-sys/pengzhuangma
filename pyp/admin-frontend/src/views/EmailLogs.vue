<template>
  <div class="email-logs-page">
    <el-card>
      <div class="page-header">
        <h2>邮件日志</h2>
      </div>
      
      <!-- 筛选区域 -->
      <div class="filter-section">
        <div class="filter-row">
          <el-input
            v-model="keyword"
            placeholder="搜索主题或收件人"
            clearable
            style="width: 220px"
            @keyup.enter="handleSearch"
          />
          <el-select v-model="statusFilter" placeholder="状态筛选" clearable style="width: 120px">
            <el-option label="全部" value="" />
            <el-option label="已发送" value="sent" />
            <el-option label="待发送" value="pending" />
            <el-option label="发送失败" value="failed" />
          </el-select>
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            clearable
          />
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="resetFilter">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </div>
      </div>
      
      <el-table :data="logs" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="to_email" label="收件人" min-width="180" />
        <el-table-column prop="subject" label="主题" min-width="250" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag
              :type="scope.row.status === 'sent' ? 'success' : 
                     scope.row.status === 'failed' ? 'danger' : 'warning'"
              size="small"
            >
              {{ scope.row.status === 'sent' ? '已发送' : 
                 scope.row.status === 'failed' ? '发送失败' : '待发送' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="scope">
            {{ scope.row.type === 'collision' ? '碰撞通知' : scope.row.type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="sent_at" label="发送时间" width="170">
          <template #default="scope">
            {{ formatDateTime(scope.row.sent_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="error_msg" label="错误信息" min-width="150" show-overflow-tooltip>
          <template #default="scope">
            <span class="error-text">{{ scope.row.error_msg || '-' }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import api from '@/utils/api'

const route = useRoute()

const logs = ref([])
const loading = ref(false)
const keyword = ref('')
const statusFilter = ref('')
const dateRange = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
let refreshTimer = null

const loadLogs = async () => {
  try {
    loading.value = true
    
    const params = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    
    if (keyword.value) {
      params.keyword = keyword.value
    }
    if (statusFilter.value) {
      params.status = statusFilter.value
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    
    const response = await api.get('/admin/api/email/logs', { params })
    
    if (response.data.code === 200) {
      logs.value = response.data.data?.list || []
      total.value = response.data.data?.pagination?.total || 0
    }
  } catch (error) {
    console.error('Failed to load email logs:', error)
    ElMessage.error('加载邮件日志失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadLogs()
}

const handleSizeChange = () => {
  currentPage.value = 1
  loadLogs()
}

const handlePageChange = () => {
  loadLogs()
}

const resetFilter = () => {
  keyword.value = ''
  statusFilter.value = ''
  dateRange.value = null
  currentPage.value = 1
  loadLogs()
}

const formatDateTime = (value) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}年${month}月${day}日 ${hours}:${minutes}:${seconds}`
}

// 页面可见性变化监听
const handleVisibilityChange = () => {
  if (!document.hidden) {
    loadLogs()
  }
}

// 监听路由变化
watch(() => route.path, (newPath) => {
  if (newPath === '/email-logs') {
    loadLogs()
  }
})

onMounted(() => {
  loadLogs()
  // 每60秒自动刷新
  refreshTimer = setInterval(loadLogs, 60000)
  // 监听页面可见性变化
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.email-logs-page {
  padding: 0;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
}

.filter-section {
  margin-bottom: 20px;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.error-text {
  color: #F56C6C;
  font-size: 12px;
}
</style>
