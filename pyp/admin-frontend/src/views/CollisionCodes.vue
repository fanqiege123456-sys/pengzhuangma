<template>
  <div class="collision-codes-page">
    <el-card>
      <div class="page-header">
        <h2>碰撞码管理</h2>
        <div class="header-actions">
          <el-button @click="goToAuditSettings">
            <el-icon><Setting /></el-icon>
            审核设置
          </el-button>
          <el-button type="primary" @click="handleRefresh">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
      
      <!-- 筛选区域 -->
      <div class="filter-section">
        <el-select v-model="statusFilter" placeholder="状态筛选" clearable @change="handleFilterChange" style="width: 120px;">
          <el-option label="全部状态" value="" />
          <el-option label="活跃" value="active" />
          <el-option label="过期" value="expired" />
          <el-option label="黑洞" value="blackhole" />
        </el-select>
        <el-select v-model="auditFilter" placeholder="审核状态" clearable @change="handleFilterChange" style="width: 120px;">
          <el-option label="全部审核" value="" />
          <el-option label="待审核" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-input v-model="keywordSearch" placeholder="搜索关键词" clearable style="width: 180px;" @keyup.enter="handleFilterChange" />
        <el-button type="primary" @click="handleFilterChange">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="resetFilter">重置</el-button>
      </div>

      <!-- 待审核快捷入口 + 批量操作 -->
      <div class="action-bar">
        <div class="left-actions">
          <el-alert v-if="pendingCount > 0" type="warning" show-icon :closable="false" style="padding: 8px 16px;">
            <template #title>
              <span>有 <strong>{{ pendingCount }}</strong> 条待审核</span>
              <el-button type="warning" size="small" link @click="showPendingOnly" style="margin-left: 8px;">立即处理</el-button>
            </template>
          </el-alert>
        </div>
        <div class="right-actions" v-if="selectedRows.length > 0">
          <span class="selected-count">已选 {{ selectedRows.length }} 项</span>
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
      
      <el-table 
        :data="codes" 
        style="width: 100%" 
        v-loading="loading" 
        stripe 
        border
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" :selectable="canSelect" />
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column label="关键词" min-width="120">
          <template #default="scope">
            <span class="keyword-text" :class="{ forbidden: scope.row.is_forbidden }">
              {{ scope.row.tag || '-' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="用户" min-width="130">
          <template #default="scope">
            {{ scope.row.user?.wechat_no || scope.row.user?.phone || scope.row.user?.openid?.substring(0, 10) || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170" sortable>
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审核" width="80" align="center">
          <template #default="scope">
            <el-tag :type="getAuditType(scope.row.audit_status)" size="small">
              {{ getAuditText(scope.row.audit_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="拒绝原因" min-width="120">
          <template #default="scope">
            <el-tooltip v-if="scope.row.reject_reason" :content="scope.row.reject_reason" placement="top">
              <span class="reject-reason">{{ scope.row.reject_reason }}</span>
            </el-tooltip>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="cost_coins" label="消耗" width="70" align="center" />
        <el-table-column label="操作" fixed="right" width="200" align="center">
          <template #default="scope">
            <div class="action-btns">
              <!-- 待审核时显示审核按钮 -->
              <template v-if="scope.row.audit_status === 'pending' || !scope.row.audit_status">
                <el-button type="success" size="small" @click="approveCode(scope.row)">通过</el-button>
                <el-button type="danger" size="small" @click="openRejectDialog(scope.row)">拒绝</el-button>
              </template>
              <!-- 其他操作 -->
              <el-dropdown trigger="click">
                <el-button size="small">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="scope.row.status !== 'blackhole'" @click="updateStatus(scope.row, 'blackhole')">
                      设为黑洞
                    </el-dropdown-item>
                    <el-dropdown-item @click="handleDelete(scope.row)" divided>
                      <span style="color: #F56C6C;">删除</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 拒绝原因弹窗 -->
    <el-dialog v-model="rejectDialogVisible" title="拒绝碰撞码" width="420px" :close-on-click-modal="false">
      <div class="reject-info">
        <p>关键词：<strong>{{ currentRejectRow?.tag }}</strong></p>
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
      </div>
      <el-form>
        <el-form-item label="拒绝原因" required>
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
import { Refresh, Setting, Search, ArrowDown, Check, Close } from '@element-plus/icons-vue'
import api from '@/utils/api'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const codes = ref([])
const pendingCount = ref(0)
const selectedRows = ref([])

// 筛选
const statusFilter = ref('')
const auditFilter = ref('')
const keywordSearch = ref('')

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
const currentRejectRow = ref(null)

// 批量拒绝弹窗
const batchRejectDialogVisible = ref(false)
const batchRejectReason = ref('')
const batchRejectLoading = ref(false)

// 格式化日期
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

// 状态相关
const getStatusType = (status) => {
  const map = { active: 'success', expired: 'warning', blackhole: 'danger' }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = { active: '活跃', expired: '过期', blackhole: '黑洞' }
  return map[status] || status || '-'
}

const getAuditType = (status) => {
  const map = { approved: 'success', rejected: 'danger', pending: 'warning' }
  return map[status] || 'warning'
}

const getAuditText = (status) => {
  const map = { approved: '通过', rejected: '拒绝', pending: '待审' }
  return map[status] || '待审'
}

// 是否可选（只有待审核的可以批量操作）
const canSelect = (row) => {
  return row.audit_status === 'pending' || !row.audit_status
}

// 选择变化
const handleSelectionChange = (rows) => {
  selectedRows.value = rows
}

const loadCodes = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    
    if (statusFilter.value) params.status = statusFilter.value
    if (auditFilter.value) params.audit_status = auditFilter.value
    if (keywordSearch.value) params.keyword = keywordSearch.value
    
    const response = await api.get('/collisions', { params })
    
    if (response.data.code === 200) {
      const data = response.data.data
      codes.value = data?.list || data?.records || data || []
      pagination.total = data?.pagination?.total || data?.total || codes.value.length
      pendingCount.value = data?.pending_count || 0
    } else {
      codes.value = []
      pagination.total = 0
    }
  } catch (error) {
    console.error('加载碰撞码失败:', error)
    ElMessage.error('加载碰撞码失败')
    codes.value = []
  } finally {
    loading.value = false
  }
}

const handleRefresh = () => {
  loadCodes()
}

const handleFilterChange = () => {
  pagination.page = 1
  loadCodes()
}

const resetFilter = () => {
  statusFilter.value = ''
  auditFilter.value = ''
  keywordSearch.value = ''
  pagination.page = 1
  loadCodes()
}

const showPendingOnly = () => {
  auditFilter.value = 'pending'
  statusFilter.value = ''
  keywordSearch.value = ''
  pagination.page = 1
  loadCodes()
}

// 审核通过
const approveCode = async (row) => {
  try {
    await ElMessageBox.confirm(`确定通过「${row.tag}」的审核？`, '确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'success'
    })
    
    const response = await api.put(`/collisions/${row.id}/approve`)
    if (response.data.code === 200) {
      ElMessage.success('审核通过')
      loadCodes()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('审核失败:', error)
      ElMessage.error('操作失败')
    }
  }
}

// 批量通过
const batchApprove = async () => {
  if (selectedRows.value.length === 0) return
  
  try {
    await ElMessageBox.confirm(`确定批量通过 ${selectedRows.value.length} 条碰撞码？`, '确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'success'
    })
    
    const ids = selectedRows.value.map(row => row.id)
    const response = await api.put('/collisions/batch-approve', { ids })
    
    if (response.data.code === 200) {
      ElMessage.success(`成功通过 ${ids.length} 条`)
      selectedRows.value = []
      loadCodes()
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

// 打开拒绝弹窗
const openRejectDialog = (row) => {
  currentRejectRow.value = row
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
    const response = await api.put(`/collisions/${currentRejectRow.value.id}/reject`, {
      reject_reason: rejectReason.value
    })
    
    if (response.data.code === 200) {
      ElMessage.success('已拒绝')
      rejectDialogVisible.value = false
      loadCodes()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error) {
    console.error('拒绝失败:', error)
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
      ElMessage.success(`成功拒绝 ${ids.length} 条`)
      batchRejectDialogVisible.value = false
      selectedRows.value = []
      loadCodes()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error) {
    console.error('批量拒绝失败:', error)
    ElMessage.error('操作失败')
  } finally {
    batchRejectLoading.value = false
  }
}

const updateStatus = async (row, status) => {
  try {
    const response = await api.put(`/collisions/${row.id}/status`, { status })
    if (response.data.code === 200) {
      ElMessage.success('状态更新成功')
      loadCodes()
    }
  } catch (error) {
    console.error('更新状态失败:', error)
    ElMessage.error('操作失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除「${row.tag}」？`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const response = await api.delete(`/collisions/${row.id}`)
    if (response.data.code === 200) {
      ElMessage.success('删除成功')
      loadCodes()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

const handleSizeChange = () => {
  pagination.page = 1
  loadCodes()
}

const handleCurrentChange = () => {
  loadCodes()
}

const goToAuditSettings = () => {
  router.push('/audit-settings')
}

onMounted(() => {
  // 检查URL参数
  if (route.query.audit_status) {
    auditFilter.value = route.query.audit_status
  }
  loadCodes()
})

// 监听路由变化，确保每次进入页面都加载最新数据
watch(() => route.path, (newPath, oldPath) => {
  if (newPath === '/collision-codes' && newPath !== oldPath) {
    loadCodes()
  }
})

// 监听页面可见性变化，当页面重新可见时刷新数据
document.addEventListener('visibilitychange', () => {
  if (!document.hidden && route.path === '/collision-codes') {
    loadCodes()
  }
})
</script>

<style scoped>
.collision-codes-page {
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
  color: #303133;
  font-size: 18px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.filter-section {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  align-items: center;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  min-height: 40px;
}

.right-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.selected-count {
  color: #409EFF;
  font-weight: 500;
}

.keyword-text {
  font-weight: 600;
  color: #409EFF;
}

.keyword-text.forbidden {
  color: #F56C6C;
}

.reject-reason {
  color: #F56C6C;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  cursor: pointer;
}

.text-muted {
  color: #C0C4CC;
}

.action-btns {
  display: flex;
  gap: 6px;
  justify-content: center;
  flex-wrap: wrap;
}

.reject-info {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.reject-info p {
  margin: 0;
  color: #606266;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
