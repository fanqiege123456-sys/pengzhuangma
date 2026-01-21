<template>
  <div class="users-page">
    <el-card>
      <div class="page-header">
        <h2>用户管理</h2>
        <div class="header-actions">
          <el-input
            v-model="searchQuery"
            placeholder="搜索用户"
            style="width: 200px; margin-right: 16px;"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            创建用户
          </el-button>
          <el-button type="primary" @click="loadUsers">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
      
      <el-table
        v-loading="loading"
        :data="users"
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="nickname" label="昵称" width="120">
          <template #default="scope">
            {{ scope.row.nickname || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="wechat_no" label="微信号" width="150" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="email" label="邮箱" min-width="200">
          <template #default="scope">
            {{ scope.row.email || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="age" label="年龄" width="80">
          <template #default="scope">
            {{ scope.row.age || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="gender" label="性别" width="80">
          <template #default="scope">
            {{ scope.row.gender === 1 ? '男' : scope.row.gender === 2 ? '女' : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="coins" label="碰撞币" width="100" />
        <el-table-column prop="total_recharge" label="充值总额" width="120">
          <template #default="scope">
            ¥{{ (scope.row.total_recharge / 100).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="allow_force_add" label="允许强制加友" width="140">
          <template #default="scope">
            <el-tag :type="scope.row.allow_force_add ? 'success' : 'info'">
              {{ scope.row.allow_force_add ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="200">
          <template #default="scope">
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(scope.row)"
            >
              编辑
            </el-button>
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
    
    <!-- 编辑用户对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'edit' ? '编辑用户' : '查看用户'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="昵称">
          <el-input v-model="form.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="微信号">
          <el-input v-model="form.wechat_no" placeholder="请输入微信号" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="年龄">
          <el-input-number v-model="form.age" :min="0" :max="100" placeholder="请输入年龄" />
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="form.gender" placeholder="请选择性别">
            <el-option label="未知" value="" />
            <el-option label="男" value="1" />
            <el-option label="女" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="碰撞币" prop="coins">
          <el-input-number v-model="form.coins" :min="0" />
        </el-form-item>
        <el-form-item label="允许强制加友">
          <el-switch v-model="form.allow_force_add" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSave">确定</el-button>
        </span>
      </template>
    </el-dialog>
    
    <!-- 创建用户对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建用户"
      width="500px"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="120px"
      >
        <el-form-item label="微信号" prop="wechat_no">
          <el-input v-model="createForm.wechat_no" placeholder="请输入微信号" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="createForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="createForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="碰撞币" prop="coins">
          <el-input-number v-model="createForm.coins" :min="0" :step="100" />
        </el-form-item>
        <el-form-item label="允许强制加友">
          <el-switch v-model="createForm.allow_force_add" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleCreateSave">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import api from '@/utils/api'

const route = useRoute()

const loading = ref(false)
const users = ref([])
const searchQuery = ref('')
const selectedUsers = ref([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const dialogVisible = ref(false)
const dialogMode = ref('edit')
const formRef = ref()

const form = reactive({
  id: null,
  nickname: '',
  wechat_no: '',
  phone: '',
  email: '',
  age: null,
  gender: null,
  coins: 0,
  allow_force_add: false
})

const rules = {
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ]
}

// 创建用户相关
const createDialogVisible = ref(false)
const createFormRef = ref()

const createForm = reactive({
  wechat_no: '',
  nickname: '',
  phone: '',
  email: '',
  coins: 1000,
  allow_force_add: false
})

const createRules = {
  wechat_no: [
    { required: true, message: '请输入微信号', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ]
}

// 处理创建用户
const handleCreate = () => {
  // 重置表单
  Object.assign(createForm, {
    wechat_no: '',
    nickname: '',
    phone: '',
    email: '',
    coins: 1000,
    allow_force_add: false
  })
  createDialogVisible.value = true
}

// 保存创建的用户
const handleCreateSave = async () => {
  if (!createFormRef.value) return
  
  await createFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const response = await api.post('/users', createForm)
        
        if (response.data.code === 200) {
          ElMessage.success('创建成功')
          createDialogVisible.value = false
          loadUsers()
        }
      } catch (error) {
        console.error('Failed to create user:', error)
      }
    }
  })
}

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const response = await api.get('/users', {
      params: {
        page: pagination.page,
        page_size: pagination.pageSize,
        search: searchQuery.value
      }
    })
    
    if (response.data.code === 200) {
      users.value = response.data.data.list
      pagination.total = response.data.data.pagination.total
    }
  } catch (error) {
    console.error('Failed to load users:', error)
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  loadUsers()
}

// 分页处理
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  loadUsers()
}

const handleCurrentChange = (page) => {
  pagination.page = page
  loadUsers()
}

// 选择处理
const handleSelectionChange = (selection) => {
  selectedUsers.value = selection
}

// 编辑用户
const handleEdit = (row) => {
  dialogMode.value = 'edit'
  form.id = row.id
  form.nickname = row.nickname || ''
  form.wechat_no = row.wechat_no
  form.phone = row.phone
  form.email = row.email || ''
  form.age = row.age || null
  form.gender = row.gender || null
  form.coins = row.coins
  form.allow_force_add = row.allow_force_add
  dialogVisible.value = true
}

// 保存用户
const handleSave = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const response = await api.put(`/users/${form.id}`, {
          nickname: form.nickname,
          wechat_no: form.wechat_no,
          phone: form.phone,
          email: form.email,
          age: form.age,
          gender: form.gender,
          coins: form.coins,
          allow_force_add: form.allow_force_add
        })
        
        if (response.data.code === 200) {
          ElMessage.success('更新成功')
          dialogVisible.value = false
          loadUsers()
        }
      } catch (error) {
        console.error('Failed to update user:', error)
      }
    }
  })
}

// 删除用户
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 ${row.wechat_no} 吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await api.delete(`/users/${row.id}`)
    if (response.data.code === 200) {
      ElMessage.success('删除成功')
      loadUsers()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete user:', error)
    }
  }
}

// 格式化日期
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
  loadUsers()
})

// 监听路由变化，确保每次进入页面都加载最新数据
watch(() => route.path, (newPath, oldPath) => {
  if (newPath === '/users' && newPath !== oldPath) {
    loadUsers()
  }
})

// 监听页面可见性变化，当页面重新可见时刷新数据
document.addEventListener('visibilitychange', () => {
  if (!document.hidden && route.path === '/users') {
    loadUsers()
  }
})
</script>

<style scoped>
.users-page {
  padding: 0;
  height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 0 4px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pagination-container {
  margin-top: 24px;
  text-align: right;
  padding: 16px 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 表格优化 */
:deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
}

:deep(.el-table th) {
  background-color: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
}

:deep(.el-table__row:hover > td) {
  background-color: #f5f7fa !important;
}

/* 卡片样式优化 */
:deep(.el-card) {
  border-radius: 12px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.08);
  border: none;
}

:deep(.el-card__body) {
  padding: 24px;
}

/* 搜索框样式 */
:deep(.el-input__wrapper) {
  border-radius: 8px;
}

/* 按钮样式优化 */
:deep(.el-button) {
  border-radius: 6px;
  padding: 8px 16px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }
  
  .header-actions {
    justify-content: space-between;
  }
  
  .pagination-container {
    text-align: center;
  }
}
</style>
