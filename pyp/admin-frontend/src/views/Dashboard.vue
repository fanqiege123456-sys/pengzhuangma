<template>
  <div class="dashboard">
    <el-row :gutter="24" class="stats-row">
      <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
        <div class="stat-card">
          <div class="stat-icon user">
            <el-icon><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-title">用户总数</div>
            <div class="stat-value">{{ stats.userCount }}</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
        <div class="stat-card">
          <div class="stat-icon code">
            <el-icon><List /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-title">今日碰撞码</div>
            <div class="stat-value">{{ stats.todayCodeCount }}</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
        <div class="stat-card">
          <div class="stat-icon success">
            <el-icon><Star /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-title">今日成功匹配</div>
            <div class="stat-value">{{ stats.todaySuccessCount }}</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
        <div class="stat-card">
          <div class="stat-icon revenue">
            <el-icon><Money /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-title">今日收入</div>
            <div class="stat-value">¥{{ stats.todayRevenue }}</div>
          </div>
        </div>
      </el-col>
    </el-row>
    
    <el-row :gutter="24" style="margin-top: 30px;" class="charts-row">
      <el-col :xs="24" :md="12" :lg="12" :xl="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>用户注册趋势</span>
              <el-radio-group v-model="userTrendDays" size="small" @change="loadUserTrend">
                <el-radio-button :value="7">7天</el-radio-button>
                <el-radio-button :value="30">30天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container" v-loading="userTrendLoading">
            <v-chart :option="userTrendOption" autoresize />
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :md="12" :lg="12" :xl="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>碰撞成功率统计</span>
              <el-radio-group v-model="successRateDays" size="small" @change="loadSuccessRate">
                <el-radio-button :value="7">7天</el-radio-button>
                <el-radio-button :value="30">30天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container" v-loading="successRateLoading">
            <v-chart :option="successRateOption" autoresize />
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row style="margin-top: 30px;" class="table-row">
      <el-col :span="24">
        <el-card class="table-card">
          <template #header>
            <span>热门碰撞码</span>
          </template>
          <el-table :data="hotCodes" style="width: 100%" v-loading="loading">
            <el-table-column prop="code" label="碰撞码" width="200" />
            <el-table-column prop="count" label="使用次数" width="120" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag
                  :type="scope.row.status === 'show' ? 'success' : 
                         scope.row.status === 'hide' ? 'warning' : 'danger'"
                >
                  {{ scope.row.status === 'show' ? '显示' : 
                     scope.row.status === 'hide' ? '隐藏' : '黑洞' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间">
              <template #default="scope">
                {{ formatDateTime(scope.row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { User, List, Star, Money } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'

// 注册 ECharts 组件
use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent])

const route = useRoute()

const stats = ref({
  userCount: 0,
  todayCodeCount: 0,
  todaySuccessCount: 0,
  todayRevenue: 0
})

const hotCodes = ref([])
const loading = ref(false)

// 用户趋势图
const userTrendDays = ref(7)
const userTrendLoading = ref(false)
const userTrendData = ref({ dates: [], counts: [] })

// 碰撞成功率图
const successRateDays = ref(7)
const successRateLoading = ref(false)
const successRateData = ref({ dates: [], rates: [], totalCodes: [], successCodes: [] })

// 用户趋势图配置
const userTrendOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    backgroundColor: 'rgba(255, 255, 255, 0.95)',
    borderColor: '#eee',
    borderWidth: 1,
    textStyle: { color: '#333' },
    formatter: (params) => {
      const data = params[0]
      return `${data.name}<br/>新增用户: <b>${data.value}</b> 人`
    }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    top: '10%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    data: userTrendData.value.dates,
    axisLine: { lineStyle: { color: '#ddd' } },
    axisLabel: { color: '#666', fontSize: 12 }
  },
  yAxis: {
    type: 'value',
    axisLine: { show: false },
    axisTick: { show: false },
    splitLine: { lineStyle: { color: '#f0f0f0' } },
    axisLabel: { color: '#666', fontSize: 12 }
  },
  series: [{
    data: userTrendData.value.counts,
    type: 'line',
    smooth: true,
    symbol: 'circle',
    symbolSize: 8,
    lineStyle: { color: '#667eea', width: 3 },
    itemStyle: { color: '#667eea', borderWidth: 2, borderColor: '#fff' },
    areaStyle: {
      color: {
        type: 'linear',
        x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: 'rgba(102, 126, 234, 0.3)' },
          { offset: 1, color: 'rgba(102, 126, 234, 0.05)' }
        ]
      }
    }
  }]
}))

// 碰撞成功率图配置
const successRateOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    backgroundColor: 'rgba(255, 255, 255, 0.95)',
    borderColor: '#eee',
    borderWidth: 1,
    textStyle: { color: '#333' },
    formatter: (params) => {
      const idx = params[0].dataIndex
      const rate = successRateData.value.rates[idx]
      const total = successRateData.value.totalCodes[idx]
      const success = successRateData.value.successCodes[idx]
      return `${params[0].name}<br/>
        成功率: <b>${rate}%</b><br/>
        总碰撞: ${total} 次<br/>
        成功匹配: ${success} 次`
    }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    top: '10%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    data: successRateData.value.dates,
    axisLine: { lineStyle: { color: '#ddd' } },
    axisLabel: { color: '#666', fontSize: 12 }
  },
  yAxis: {
    type: 'value',
    max: 100,
    axisLine: { show: false },
    axisTick: { show: false },
    splitLine: { lineStyle: { color: '#f0f0f0' } },
    axisLabel: { color: '#666', fontSize: 12, formatter: '{value}%' }
  },
  series: [{
    data: successRateData.value.rates,
    type: 'bar',
    barWidth: '50%',
    itemStyle: {
      color: {
        type: 'linear',
        x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: '#4facfe' },
          { offset: 1, color: '#00f2fe' }
        ]
      },
      borderRadius: [4, 4, 0, 0]
    }
  }]
}))

const loadDashboardData = async () => {
  try {
    loading.value = true
    const statsResponse = await api.get('/dashboard/stats')
    stats.value = statsResponse.data.data || stats.value
    
    const hotCodesResponse = await api.get('/dashboard/hot-codes')
    hotCodes.value = hotCodesResponse.data.data || []
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
    ElMessage.error('加载仪表盘数据失败')
  } finally {
    loading.value = false
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

const loadUserTrend = async () => {
  try {
    userTrendLoading.value = true
    const response = await api.get(`/dashboard/user-trend?period=${userTrendDays.value}`)
    if (response.data.code === 200 && response.data.data) {
      // 适配后端返回格式
      const data = response.data.data
      const dates = data.map(item => {
        const d = new Date(item.date)
        return `${d.getMonth() + 1}/${d.getDate()}`
      })
      const counts = data.map(item => item.count)
      userTrendData.value = { dates, counts }
    }
  } catch (error) {
    console.error('Failed to load user trend:', error)
    // 使用模拟数据
    generateMockUserTrend()
  } finally {
    userTrendLoading.value = false
  }
}

const loadSuccessRate = async () => {
  try {
    successRateLoading.value = true
    const response = await api.get(`/dashboard/success-rate?period=${successRateDays.value}`)
    if (response.data.code === 200 && response.data.data) {
      // 适配后端返回格式
      const data = response.data.data
      const dates = data.map(item => {
        const d = new Date(item.date)
        return `${d.getMonth() + 1}/${d.getDate()}`
      })
      const rates = data.map(item => Math.round(item.successRate))
      const totalCodes = data.map(item => item.totalCount)
      const successCodes = data.map(item => item.successCount)
      successRateData.value = { dates, rates, totalCodes, successCodes }
    }
  } catch (error) {
    console.error('Failed to load success rate:', error)
    // 使用模拟数据
    generateMockSuccessRate()
  } finally {
    successRateLoading.value = false
  }
}

// 生成模拟数据（后端接口未实现时使用）
const generateMockUserTrend = () => {
  const dates = []
  const counts = []
  const days = userTrendDays.value
  for (let i = days - 1; i >= 0; i--) {
    const date = new Date()
    date.setDate(date.getDate() - i)
    dates.push(`${date.getMonth() + 1}/${date.getDate()}`)
    counts.push(Math.floor(Math.random() * 10) + 1)
  }
  userTrendData.value = { dates, counts }
}

const generateMockSuccessRate = () => {
  const dates = []
  const rates = []
  const totalCodes = []
  const successCodes = []
  const days = successRateDays.value
  for (let i = days - 1; i >= 0; i--) {
    const date = new Date()
    date.setDate(date.getDate() - i)
    dates.push(`${date.getMonth() + 1}/${date.getDate()}`)
    const total = Math.floor(Math.random() * 50) + 10
    const success = Math.floor(Math.random() * total * 0.8)
    totalCodes.push(total)
    successCodes.push(success)
    rates.push(Math.round((success / total) * 100))
  }
  successRateData.value = { dates, rates, totalCodes, successCodes }
}

// 页面可见性变化监听
const handleVisibilityChange = () => {
  if (!document.hidden) {
    loadDashboardData()
    loadUserTrend()
    loadSuccessRate()
  }
}

// 监听路由变化
watch(() => route.path, (newPath) => {
  if (newPath === '/' || newPath === '/dashboard') {
    loadDashboardData()
    loadUserTrend()
    loadSuccessRate()
  }
})

onMounted(() => {
  loadDashboardData()
  loadUserTrend()
  loadSuccessRate()
  // 监听页面可见性变化
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.dashboard {
  padding: 0;
  width: 100%;
  height: 100%;
}

.stats-row {
  margin-bottom: 0;
}

.charts-row, .table-row {
  margin-top: 30px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 25px;
  box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  min-height: 110px;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px 0 rgba(0, 0, 0, 0.15);
}

.chart-card, .table-card {
  border-radius: 12px;
  box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.08);
  transition: box-shadow 0.2s ease;
}

.chart-card:hover, .table-card:hover {
  box-shadow: 0 8px 25px 0 rgba(0, 0, 0, 0.12);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.stat-icon {
  width: 70px;
  height: 70px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 25px;
  font-size: 28px;
  color: white;
  flex-shrink: 0;
}

.stat-icon.user {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.code {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.success {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.revenue {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-title {
  color: #909399;
  font-size: 15px;
  margin-bottom: 10px;
  font-weight: 500;
}

.stat-value {
  color: #303133;
  font-size: 32px;
  font-weight: bold;
  line-height: 1;
}

.chart-container {
  height: 320px;
  width: 100%;
}

/* 响应式设计 */
@media (max-width: 1400px) {
  .stat-card {
    padding: 20px;
    min-height: 100px;
  }
  
  .stat-icon {
    width: 60px;
    height: 60px;
    font-size: 24px;
    margin-right: 20px;
  }
  
  .stat-value {
    font-size: 28px;
  }
}

@media (max-width: 992px) {
  .charts-row {
    margin-top: 20px;
  }
  
  .charts-row .el-col {
    margin-bottom: 20px;
  }
  
  .chart-container {
    height: 280px;
  }
}

@media (max-width: 768px) {
  .stats-row .el-col {
    margin-bottom: 15px;
  }
  
  .stat-card {
    padding: 15px;
    min-height: 90px;
    margin-bottom: 15px;
  }
  
  .stat-icon {
    width: 50px;
    height: 50px;
    font-size: 20px;
    margin-right: 15px;
  }
  
  .stat-title {
    font-size: 13px;
  }
  
  .stat-value {
    font-size: 24px;
  }
  
  .chart-container {
    height: 250px;
  }
  
  .charts-row, .table-row {
    margin-top: 20px;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
}
</style>
