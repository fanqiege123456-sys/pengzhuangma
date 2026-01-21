<template>
  <div class="admins-page">
    <el-card>
      <div class="page-header">
        <h2>管理员管理</h2>
        <el-button type="primary" @click="showAddDialog">
          <el-icon><Plus /></el-icon>
          添加管理员
        </el-button>
      </div>
      
      <el-table v-loading="loading" :data="admins" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="role" label="角色" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
              {{ scope.row.status === 'active' ? '活跃' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button
              type="warning"
              size="small"
              @click="toggleStatus(scope.row)"
              v-if="scope.row.username !== 'admin'"
            >
              {{ scope.row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(scope.row)"
              v-if="scope.row.username !== 'admin'"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
    
    <!-- 添加管理员对话框 -->
    <el-dialog v-model="dialogVisible" title="添加管理员" width="500px">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role">
            <el-option label="管理员" value="admin" />
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

const loading = ref(false)
const admins = ref([])
const dialogVisible = ref(false)
const formRef = ref()

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  username: '',
  password: '',
  email: '',
  role: 'admin'
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为3-20位', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ]
}

const loadAdmins = async () => {
  loading.value = true
  try {
    const response = await api.get('/admin/list', {
      params: {
        page: pagination.page,
        page_size: pagination.pageSize
      }
    })
    
    if (response.data.code === 200) {
      admins.value = response.data.data.list
      pagination.total = response.data.data.pagination.total
    }
  } catch (error) {
    console.error('Failed to load admins:', error)
  } finally {
    loading.value = false
  }
}

const showAddDialog = () => {
  form.username = ''
  form.password = ''
  form.email = ''
  form.role = 'admin'
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const response = await api.post('/admin/create', form)
        if (response.data.code === 200) {
          ElMessage.success('添加成功')
          dialogVisible.value = false
          loadAdmins()
        }
      } catch (error) {
        console.error('Failed to create admin:', error)
      }
    }
  })
}

const toggleStatus = async (row) => {
  const newStatus = row.status === 'active' ? 'inactive' : 'active'
  
  try {
    const response = await api.put(`/admin/${row.id}/status`, {
      status: newStatus
    })
    
    if (response.data.code === 200) {
      row.status = newStatus
      ElMessage.success(`管理员 ${row.username} 已${newStatus === 'active' ? '启用' : '禁用'}`)
      loadAdmins()
    }
  } catch (error) {
    console.error('Failed to update admin status:', error)
    ElMessage.error('更新状态失败')
  }
}

const handleDelete = async (row) => {
  ElMessageBox.confirm(
    `确定要删除管理员 ${row.username} 吗?`,
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      const response = await api.delete(`/admin/${row.id}`)
      
      if (response.data.code === 200) {
        ElMessage.success('删除成功')
        loadAdmins()
      }
    } catch (error) {
      console.error('Failed to delete admin:', error)
      if (error.response?.data?.msg) {
        ElMessage.error(error.response.data.msg)
      } else {
        ElMessage.error('删除失败')
      }
    }
  }).catch(() => {
    // 取消删除
  })
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  loadAdmins()
}

const handleCurrentChange = (page) => {
  pagination.page = page
  loadAdmins()
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  if (Number.isNaN(date.getTime())) return dateString
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}年${month}月${day}日 ${hours}:${minutes}:${seconds}`
}

onMounted(() => {
  loadAdmins()
})
</script>

<style scoped>
.admins-page {
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

.pagination-container {
  margin-top: 20px;
  text-align: right;
}
</style>
