<template>
  <div class="keywords-page">
    <el-card>
      <div class="page-header">
        <h2>热门关键词管理</h2>
        <el-button type="primary" @click="dialogVisible = true">
          <el-icon><Plus /></el-icon>
          添加关键词
        </el-button>
      </div>
      
      <el-table :data="keywords" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="keyword" label="关键词" width="200" />
        <el-table-column prop="submit_count" label="提交次数" width="120" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="scope">
            <el-select
              v-model="scope.row.status"
              size="small"
              @change="updateStatus(scope.row)"
            >
              <el-option label="显示" value="show" />
              <el-option label="隐藏" value="hide" />
              <el-option label="黑洞" value="blackhole" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 添加关键词对话框 -->
    <el-dialog v-model="dialogVisible" title="添加关键词" width="400px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="关键词">
          <el-input v-model="form.keyword" placeholder="请输入关键词" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option label="显示" value="show" />
            <el-option label="隐藏" value="hide" />
            <el-option label="黑洞" value="blackhole" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import api from '@/utils/api'

const keywords = ref([])
const loading = ref(false)

const dialogVisible = ref(false)

const form = reactive({
  keyword: '',
  status: 'show'
})

const loadKeywords = async () => {
  try {
    loading.value = true
    const response = await api.get('/keywords')
    keywords.value = response.data.data || []
  } catch (error) {
    console.error('Failed to load keywords:', error)
    ElMessage.error('加载热门关键词失败')
  } finally {
    loading.value = false
  }
}

const updateStatus = async (row) => {
  try {
    await api.put(`/keywords/${row.id}/status`, { status: row.status })
    ElMessage.success(`关键词 "${row.keyword}" 状态已更新为 ${row.status}`)
  } catch (error) {
    console.error('Failed to update status:', error)
    ElMessage.error('更新状态失败')
    // 回滚状态
    loadKeywords()
  }
}

const handleSave = async () => {
  if (!form.keyword.trim()) {
    ElMessage.error('请输入关键词')
    return
  }
  
  try {
    await api.post('/keywords', {
      keyword: form.keyword,
      status: form.status
    })
    
    ElMessage.success('添加成功')
    form.keyword = ''
    form.status = 'show'
    dialogVisible.value = false
    
    // 重新加载数据
    loadKeywords()
  } catch (error) {
    console.error('Failed to create keyword:', error)
    ElMessage.error('添加关键词失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除关键词 "${row.keyword}" 吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await api.delete(`/keywords/${row.id}`)
    ElMessage.success('删除成功')
    
    // 重新加载数据
    loadKeywords()
  } catch (error) {
    if (error.message !== 'cancel') {
      console.error('Failed to delete keyword:', error)
      ElMessage.error('删除失败')
    }
  }
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

onMounted(() => {
  loadKeywords()
})
</script>

<style scoped>
.keywords-page {
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
}
</style>
