<template>
  <div class="records-page">
    <el-card>
      <div class="page-header">
        <h2>碰撞记录</h2>
      </div>
      
      <!-- 筛选区域 -->
      <div class="filter-section">
        <!-- 快捷日期按钮 -->
        <div class="quick-date-btns">
          <el-button 
            v-for="item in quickDateOptions" 
            :key="item.value"
            :type="activeQuickDate === item.value ? 'primary' : 'default'"
            @click="selectQuickDate(item.value)"
            size="default"
          >
            {{ item.label }}
          </el-button>
        </div>
        
        <!-- 自定义日期和其他筛选 -->
        <div class="custom-filter">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY年MM月DD日"
            value-format="YYYY-MM-DD"
            @change="handleCustomDateChange"
            clearable
            :disabled-date="disabledDate"
          />
          <el-select v-model="statusFilter" placeholder="状态筛选" clearable @change="handleFilterChange" style="width: 140px;">
            <el-option label="全部状态" value="" />
            <el-option label="已匹配" value="matched" />
            <el-option label="已加好友" value="friend_added" />
            <el-option label="已错过" value="missed" />
          </el-select>
          <el-button type="primary" @click="loadRecords">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="resetFilter">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </div>
      </div>
      
      <el-table :data="records" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="碰撞码" width="150" />
        <el-table-column label="用户1" min-width="180">
          <template #default="scope">
            <div>{{ scope.row.user1?.wechat_no || scope.row.user1?.nickname || '-' }}</div>
            <div class="email-text">{{ scope.row.user1_email || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="用户2" min-width="180">
          <template #default="scope">
            <div>{{ scope.row.user2?.wechat_no || scope.row.user2?.nickname || '-' }}</div>
            <div class="email-text">{{ scope.row.user2_email || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag
              :type="scope.row.status === 'matched' ? 'success' : 
                     scope.row.status === 'friend_added' ? 'primary' : 'warning'"
              size="small"
            >
              {{ scope.row.status === 'matched' ? '已匹配' : 
                 scope.row.status === 'friend_added' ? '已加好友' : '已错过' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="邮件状态" width="100">
          <template #default="scope">
            <el-tag
              v-if="scope.row.email_sent"
              :type="scope.row.email_status === 'sent' ? 'success' : 
                     scope.row.email_status === 'failed' ? 'danger' : 'warning'"
              size="small"
            >
              {{ scope.row.email_status === 'sent' ? '已发送' : 
                 scope.row.email_status === 'failed' ? '发送失败' : '待处理' }}
            </el-tag>
            <el-tag v-else type="info" size="small">未发送</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="email_sent_at" label="邮件发送时间" width="170">
          <template #default="scope">
            {{ formatDateTime(scope.row.email_sent_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
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

const records = ref([])
const loading = ref(false)
const dateRange = ref(null)
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const activeQuickDate = ref('3days') // 默认最近3天

// 快捷日期选项
const quickDateOptions = [
  { label: '今天', value: 'today' },
  { label: '最近3天', value: '3days' },
  { label: '最近7天', value: '7days' },
  { label: '最近30天', value: '30days' },
  { label: '本月', value: 'month' }
]

// 格式化日期为 YYYY-MM-DD
const formatDate = (date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
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

// 根据快捷选项计算日期范围
const getDateRangeByQuick = (type) => {
  const end = new Date()
  const start = new Date()
  
  switch (type) {
    case 'today':
      break
    case '3days':
      start.setDate(start.getDate() - 2)
      break
    case '7days':
      start.setDate(start.getDate() - 6)
      break
    case '30days':
      start.setDate(start.getDate() - 29)
      break
    case 'month':
      start.setDate(1)
      break
    default:
      start.setDate(start.getDate() - 2)
  }
  
  return [formatDate(start), formatDate(end)]
}

// 选择快捷日期
const selectQuickDate = (type) => {
  activeQuickDate.value = type
  const range = getDateRangeByQuick(type)
  dateRange.value = range
  currentPage.value = 1
  loadRecords()
}

// 自定义日期变化
const handleCustomDateChange = () => {
  activeQuickDate.value = '' // 清除快捷选中状态
  currentPage.value = 1
  loadRecords()
}

// 禁用未来日期
const disabledDate = (time) => {
  return time.getTime() > Date.now()
}

const loadRecords = async () => {
  try {
    loading.value = true
    
    // 构建查询参数
    const params = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    
    if (statusFilter.value) {
      params.status = statusFilter.value
    }
    
    const response = await api.get('/records', { params })
    
    if (response.data.code === 200) {
      records.value = response.data.data?.list || response.data.data || []
      total.value = response.data.data?.total || records.value.length
    } else {
      records.value = response.data.data || []
      total.value = records.value.length
    }
  } catch (error) {
    console.error('Failed to load records:', error)
    ElMessage.error('加载碰撞记录失败')
  } finally {
    loading.value = false
  }
}

const handleFilterChange = () => {
  currentPage.value = 1
  loadRecords()
}

const handleSizeChange = () => {
  currentPage.value = 1
  loadRecords()
}

const handlePageChange = () => {
  loadRecords()
}

const resetFilter = () => {
  activeQuickDate.value = '3days'
  dateRange.value = getDateRangeByQuick('3days')
  statusFilter.value = ''
  currentPage.value = 1
  loadRecords()
}

// 页面可见性变化监听
const handleVisibilityChange = () => {
  if (!document.hidden) {
    loadRecords()
  }
}

// 监听路由变化
watch(() => route.path, (newPath) => {
  if (newPath === '/records') {
    loadRecords()
  }
})

onMounted(() => {
  // 默认加载最近3天的数据
  dateRange.value = getDateRangeByQuick('3days')
  loadRecords()
  // 监听页面可见性变化
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.records-page {
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

.quick-date-btns {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.custom-filter {
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

@media (max-width: 768px) {
  .quick-date-btns {
    gap: 8px;
  }
  
  .custom-filter {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
}

.email-text {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
