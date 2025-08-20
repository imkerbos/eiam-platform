<template>
  <div class="dashboard-page">
    <a-card title="Dashboard" :bordered="false">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <template #title>
              <UserOutlined />
              Total Users
            </template>
            <div class="stat-number">{{ stats.totalUsers }}</div>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <template #title>
              <TeamOutlined />
              Total Organizations
            </template>
            <div class="stat-number">{{ stats.totalOrganizations }}</div>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <template #title>
              <SafetyOutlined />
              Active Sessions
            </template>
            <div class="stat-number">{{ stats.activeSessions }}</div>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <template #title>
              <AppstoreOutlined />
              Total Applications
            </template>
            <div class="stat-number">{{ stats.totalApplications }}</div>
          </a-card>
        </a-col>
      </a-row>

      <a-divider />

      <a-row :gutter="16">
        <a-col :span="12">
          <a-card title="Top 10 Login Users" :bordered="false">
            <a-list :data-source="topLoginUsers" size="small">
              <template #renderItem="{ item, index }">
                <a-list-item>
                  <a-list-item-meta
                    :title="item.username"
                    :description="item.last_login_time"
                  >
                    <template #avatar>
                      <a-avatar :style="{ backgroundColor: getRankColor(index + 1) }">
                        {{ index + 1 }}
                      </a-avatar>
                    </template>
                  </a-list-item-meta>
                  <div class="login-count">{{ item.login_count }} logins</div>
                </a-list-item>
              </template>
            </a-list>
          </a-card>
        </a-col>
        <a-col :span="12">
          <a-card title="Top 10 Login Applications" :bordered="false">
            <a-list :data-source="topLoginApplications" size="small">
              <template #renderItem="{ item, index }">
                <a-list-item>
                  <a-list-item-meta
                    :title="item.name"
                    :description="item.description"
                  >
                    <template #avatar>
                      <a-avatar :style="{ backgroundColor: getRankColor(index + 1) }">
                        {{ index + 1 }}
                      </a-avatar>
                    </template>
                  </a-list-item-meta>
                  <div class="login-count">{{ item.access_count }} accesses</div>
                </a-list-item>
              </template>
            </a-list>
          </a-card>
        </a-col>
      </a-row>

      <!-- System Status and Debug section -->
      <a-divider />
      <a-row :gutter="16">
        <a-col :span="12">
          <a-card title="System Status" :bordered="false">
            <a-list :data-source="systemStatus" size="small">
              <template #renderItem="{ item }">
                <a-list-item>
                  <a-list-item-meta :title="item.name">
                    <template #avatar>
                      <a-badge :status="item.status" />
                    </template>
                  </a-list-item-meta>
                  <div>{{ item.value }}</div>
                </a-list-item>
              </template>
            </a-list>
          </a-card>
        </a-col>
        <a-col :span="12">
          <a-card title="Debug Info" :bordered="false">
            <a-space>
              <a-button @click="checkToken">Check Token</a-button>
              <a-button @click="clearToken">Clear Token</a-button>
              <a-button @click="testAPI">Test API</a-button>
              <a-button @click="testDashboardAPI">Test Dashboard API</a-button>
            </a-space>
            <div v-if="debugInfo" style="margin-top: 16px; padding: 16px; background: #f5f5f5; border-radius: 4px;">
              <pre>{{ debugInfo }}</pre>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  UserOutlined,
  TeamOutlined,
  SafetyOutlined,
  AppstoreOutlined
} from '@ant-design/icons-vue'
import { useUserStore } from '@/stores/user'
import { systemApi } from '@/api/system'

const userStore = useUserStore()
const debugInfo = ref('')

// Dashboard data
const stats = reactive({
  totalUsers: 0,
  totalOrganizations: 0,
  activeSessions: 0,
  totalApplications: 0
})

const topLoginUsers = ref([
  {
    username: 'admin',
    last_login_time: '2 minutes ago',
    login_count: 15
  },
  {
    username: 'user1',
    last_login_time: '5 minutes ago',
    login_count: 12
  },
  {
    username: 'user2',
    last_login_time: '10 minutes ago',
    login_count: 8
  },
  {
    username: 'user3',
    last_login_time: '15 minutes ago',
    login_count: 6
  },
  {
    username: 'user4',
    last_login_time: '20 minutes ago',
    login_count: 5
  }
])

const topLoginApplications = ref([
  {
    name: 'HR System',
    description: 'Human Resources Management',
    access_count: 25
  },
  {
    name: 'CRM System',
    description: 'Customer Relationship Management',
    access_count: 18
  },
  {
    name: 'Email System',
    description: 'Corporate Email System',
    access_count: 12
  },
  {
    name: 'File Manager',
    description: 'Document Management System',
    access_count: 8
  },
  {
    name: 'Project Tracker',
    description: 'Project Management Tool',
    access_count: 6
  }
])

const systemStatus = ref([
  {
    name: 'Database',
    status: 'success',
    value: 'Connected'
  },
  {
    name: 'Redis',
    status: 'success',
    value: 'Connected'
  },
  {
    name: 'API Server',
    status: 'success',
    value: 'Running'
  }
])

// Debug functions
const checkToken = () => {
  const token = localStorage.getItem('access_token')
  const refreshToken = localStorage.getItem('refresh_token')
  const user = userStore.user
  
  debugInfo.value = JSON.stringify({
    hasToken: !!token,
    tokenLength: token?.length || 0,
    hasRefreshToken: !!refreshToken,
    refreshTokenLength: refreshToken?.length || 0,
    user: user,
    isLoggedIn: userStore.isLoggedIn
  }, null, 2)
}

const clearToken = () => {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  userStore.clearAuth()
  message.success('Token cleared')
  checkToken()
}

const testAPI = async () => {
  try {
    const response = await fetch('/health')
    const data = await response.json()
    debugInfo.value = JSON.stringify(data, null, 2)
  } catch (error: any) {
    debugInfo.value = JSON.stringify({ error: error.message }, null, 2)
  }
}

const testDashboardAPI = async () => {
  try {
    const response = await systemApi.getDashboardData()
    debugInfo.value = JSON.stringify(response, null, 2)
  } catch (error: any) {
    debugInfo.value = JSON.stringify({ error: error.message }, null, 2)
  }
}

// Get rank color based on position
const getRankColor = (rank: number) => {
  const colors = ['#f5222d', '#fa8c16', '#faad14', '#52c41a', '#1890ff', '#722ed1', '#eb2f96', '#13c2c2', '#52c41a', '#faad14']
  return colors[rank - 1] || '#d9d9d9'
}

onMounted(async () => {
  try {
    // Load dashboard data from API
    const dashboardData = await systemApi.getDashboardData()
    
    // Update stats
    Object.assign(stats, dashboardData.stats)
    
    // Update system status
    systemStatus.value = dashboardData.systemStatus
    
    // Load top login users and applications from API
    try {
      const [topUsers, topApps] = await Promise.all([
        systemApi.getTopLoginUsers(),
        systemApi.getTopLoginApplications()
      ])
      
      topLoginUsers.value = topUsers || []
      topLoginApplications.value = topApps || []
    } catch (error) {
      console.error('Failed to load top users/applications:', error)
      // Keep demo data if API fails
    }
  } catch (error: any) {
    console.error('Failed to load dashboard data:', error)
    message.error('Failed to load dashboard data')
    
    // Fallback to demo data
    stats.totalUsers = 2
    stats.totalOrganizations = 2
    stats.activeSessions = 1
    stats.totalApplications = 0
  }
})
</script>

<style scoped>
.dashboard-page {
  padding: 24px;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  color: #1890ff;
}

.activity-time {
  color: #999;
  font-size: 12px;
}

.login-count {
  color: #1890ff;
  font-size: 12px;
  font-weight: 500;
}
</style>
