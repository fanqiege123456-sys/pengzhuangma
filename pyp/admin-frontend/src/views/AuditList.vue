<template>
  <div class="audit-list-page">
    <el-card>
      <div class="page-header">
        <h2>待审核列表</h2>
        <div class="header-actions">
          <el-button @click="goToSettings">
            <el-icon><Setting /></el-icon>
            审核设置
          </el-button>
          <el-button type="primary" @click="loadPendingList">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>

      <!-- 搜索和批量操作 -->
      <div class="filter-section">
        <el-input 
          v-model="searchKeyword" 
          placeholder="搜索关键词" 
          clearable 
          style="width: 200px;" 
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="resetSearch">重置</el-button>
        
        <div class="batch-actions" v-if="selectedRows.length > 0">
          <span class="selected-info">已选 {{ selectedRows.length }} 项</span>
          <el-button type="success" size="small" @click="batchApprove">
            <el-icon><Check /></el-icon>
            批量通过
          </el-button>
          <el-button type="danger" size="small" @click="openBatchRejectDialog">
            <el-icon><Close /></el-icon>
            批量拒绝
          </el-button>
        </div>
      </div>

      <!-- 统计信息 -->
      <div class="stats-bar">
        <div class="stat-item">
          <span class="stat-label">待审核</span>
          <span class="stat-value pending">{{ stats.pending }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">今日通过</span>
          <span class="stat-value approved">{{ stats.todayApproved }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">今日拒绝</span>
          <span class="stat-value rejected">{{ stats.todayRejected }}</span>
        </div>
      </div>

      <!-- 待审核列表 -->
      <el-table 
        :data="pendingList" 
        style="width: 100%" 
        v-loading="loading"
        stripe
        border
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column label="关键词" min-width="140">
          <template #default="scope">
            <span class="keyword-text" :class="{ forbidden: scope.row.is_forbidden }">{{ scope.row.tag }}</span>
          </template>
        </el-table-column>
        <el-table-column label="发布者" min-width="130">
          <template #default="scope">
            <div class="user-info">
              <span>{{ scope.row.user?.nickname || scope.row.user?.wechat_no || '-' }}</span>
              <span class="user-id">ID: {{ scope.row.user_id }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="170" sortable prop="created_at">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="消耗积分" width="90" align="center">
          <template #default="scope">
            {{ scope.row.cost_coins }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="180" align="center">
          <template #default="scope">
            <div class="action-btns">
              <el-button type="success" size="small" @click="approveOne(scope.row)">
                <el-icon><Check /></el-icon>
                通过
              </el-button>
              <el-button type="danger" size="small" @click="openRejectDialog(scope.row)">
                <el-icon><Close /></el-icon>
                拒绝
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 单个拒绝弹窗 -->
    <el-dialog v-model="rejectDialogVisible" title="拒绝碰撞码" width="420px" :close-on-click-modal="false">
      <div class="reject-info">
        <p>关键词：<strong>{{ currentRow?.tag }}</strong></p>
        <p>发布者：{{ currentRow?.user?.nickname || currentRow?.user?.wechat_no || '-' }}</p>
        <p class="refund-notice">💰 拒绝后将返还用户 <strong>{{ currentRow?.cost_coins || 10 }}</strong> 碰撞币</p>
      </div>
      <el-form>
        <el-form-item label="拒绝原因" required>
          <el-input
            v-model="rejectReason"
            type="textarea"
            :rows="3"
            placeholder="请输入拒绝原因，将展示给用户"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmReject" :loading="rejectLoading">确认拒绝</el-button>
      </template>
    </el-dialog>

    <!-- 批量拒绝弹窗 -->
    <el-dialog v-model="batchRejectDialogVisible" title="批量拒绝" width="420px" :close-on-click-modal="false">
      <div class="reject-info">
        <p>即将拒绝 <strong>{{ selectedRows.length }}</strong> 条碰撞码</p>
        <p class="refund-notice">💰 拒绝后将返还用户消耗的碰撞币</p>
      </div>
      <el-form>
        <el-form-item label="统一拒绝原因" required>
          <el-input
            v-model="batchRejectReason"
            type="textarea"
            :rows="3"
            placeholder="请输入统一的拒绝原因"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchRejectDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmBatchReject" :loading="batchRejectLoading">确认拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Setting, Search, Check, Close } from '@element-plus/icons-vue'
import api from '@/utils/api'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const pendingList = ref([])
const selectedRows = ref([])
const searchKeyword = ref('')

// 统计
const stats = reactive({
  pending: 0,
  todayApproved: 0,
  todayRejected: 0
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 拒绝弹窗
const rejectDialogVisible = ref(false)
const rejectReason = ref('')
const rejectLoading = ref(false)
const currentRow = ref(null)

// 批量拒绝
const batchRejectDialogVisible = ref(false)
const batchRejectReason = ref('')
const batchRejectLoading = ref(false)

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return dateStr
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}年${month}月${day}日 ${hours}:${minutes}:${seconds}`
}

const loadPendingList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }

    const response = await api.get('/collisions/pending', { params })
    
    if (response.data.code === 200) {
      const data = response.data.data
      pendingList.value = data?.list || data?.records || data || []
      pagination.total = data?.pagination?.total || data?.total || pendingList.value.length
      stats.pending = data?.pending_count || pagination.total
      stats.todayApproved = data?.today_approved || 0
      stats.todayRejected = data?.today_rejected || 0
    }
  } catch (error) {
    console.error('加载待审核列表失败:', error)
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadPendingList()
}

const resetSearch = () => {
  searchKeyword.value = ''
  pagination.page = 1
  loadPendingList()
}

const handleSelectionChange = (rows) => {
  selectedRows.value = rows
}

const handleSizeChange = () => {
  pagination.page = 1
  loadPendingList()
}

const handlePageChange = () => {
  loadPendingList()
}

// 单个通过
const approveOne = async (row) => {
  try {
    await ElMessageBox.confirm(`确定通过「${row.tag}」？`, '确认', {
      type: 'success'
    })
    
    const response = await api.put(`/collisions/${row.id}/approve`)
    if (response.data.code === 200) {
      ElMessage.success('审核通过')
      loadPendingList()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 批量通过
const batchApprove = async () => {
  if (selectedRows.value.length === 0) return
  
  try {
    await ElMessageBox.confirm(`确定批量通过 ${selectedRows.value.length} 条？`, '确认', {
      type: 'success'
    })
    
    const ids = selectedRows.value.map(row => row.id)
    const response = await api.put('/collisions/batch-approve', { ids })
    
    if (response.data.code === 200) {
      ElMessage.success(`成功通过 ${ids.length} 条`)
      selectedRows.value = []
      loadPendingList()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 打开拒绝弹窗
const openRejectDialog = (row) => {
  currentRow.value = row
  rejectReason.value = ''
  rejectDialogVisible.value = true
}

// 确认拒绝
const confirmReject = async () => {
  if (!rejectReason.value.trim()) {
    ElMessage.warning('请输入拒绝原因')
    return
  }
  
  rejectLoading.value = true
  try {
    const response = await api.put(`/collisions/${currentRow.value.id}/reject`, {
      reject_reason: rejectReason.value
    })
    
    if (response.data.code === 200) {
      const data = response.data.data || {}
      const refundCoins = data.refund_coins || currentRow.value.cost_coins || 10
      ElMessage.success(`已拒绝，已返还用户 ${refundCoins} 碰撞币`)
      rejectDialogVisible.value = false
      loadPendingList()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    rejectLoading.value = false
  }
}

// 打开批量拒绝弹窗
const openBatchRejectDialog = () => {
  batchRejectReason.value = ''
  batchRejectDialogVisible.value = true
}

// 确认批量拒绝
const confirmBatchReject = async () => {
  if (!batchRejectReason.value.trim()) {
    ElMessage.warning('请输入拒绝原因')
    return
  }
  
  batchRejectLoading.value = true
  try {
    const ids = selectedRows.value.map(row => row.id)
    const response = await api.put('/collisions/batch-reject', {
      ids,
      reject_reason: batchRejectReason.value
    })
    
    if (response.data.code === 200) {
      ElMessage.success(`成功拒绝 ${ids.length} 条，已返还用户碰撞币`)
      batchRejectDialogVisible.value = false
      selectedRows.value = []
      loadPendingList()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    batchRejectLoading.value = false
  }
}

const goToSettings = () => {
  router.push('/audit-settings')
}

onMounted(() => {
  loadPendingList()
})

// 监听路由变化，确保每次进入页面都加载最新数据
watch(() => route.path, (newPath, oldPath) => {
  if (newPath === '/audit-list' && newPath !== oldPath) {
    loadPendingList()
  }
})

// 监听页面可见性变化，当页面重新可见时刷新数据
document.addEventListener('visibilitychange', () => {
  if (!document.hidden && route.path === '/audit-list') {
    loadPendingList()
  }
})
</script>

<style scoped>
.audit-list-page {
  padding: 0;
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

.header-actions {
  display: flex;
  gap: 10px;
}

.filter-section {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
  padding-left: 20px;
  border-left: 1px solid #DCDFE6;
}

.selected-info {
  color: #409EFF;
  font-weight: 500;
}

.stats-bar {
  display: flex;
  gap: 30px;
  padding: 16px 20px;
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-label {
  color: #606266;
  font-size: 14px;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
}

.stat-value.pending { color: #E6A23C; }
.stat-value.approved { color: #67C23A; }
.stat-value.rejected { color: #F56C6C; }

.keyword-text {
  font-weight: 600;
  color: #409EFF;
}

.keyword-text.forbidden {
  color: #F56C6C;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.user-id {
  font-size: 12px;
  color: #909399;
}

.action-btns {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.reject-info {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.reject-info p {
  margin: 0 0 8px 0;
  color: #606266;
}

.reject-info p:last-child {
  margin-bottom: 0;
}

.refund-notice {
  color: #67C23A;
  background: #f0f9eb;
  padding: 8px 12px;
  border-radius: 4px;
  margin-top: 8px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
