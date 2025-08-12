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
          <a-card title="Recent Activities" :bordered="false">
            <a-list :data-source="recentActivities" size="small">
              <template #renderItem="{ item }">
                <a-list-item>
                  <a-list-item-meta
                    :title="item.title"
                    :description="item.description"
                  >
                    <template #avatar>
                      <a-avatar :icon="item.icon" />
                    </template>
                  </a-list-item-meta>
                  <div class="activity-time">{{ item.time }}</div>
                </a-list-item>
              </template>
            </a-list>
          </a-card>
        </a-col>
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
      </a-row>

      <!-- Debug section -->
      <a-divider />
      <a-card title="Debug Info" :bordered="false">
        <a-space>
          <a-button @click="checkToken">Check Token</a-button>
          <a-button @click="clearToken">Clear Token</a-button>
          <a-button @click="testAPI">Test API</a-button>
        </a-space>
        <div v-if="debugInfo" style="margin-top: 16px; padding: 16px; background: #f5f5f5; border-radius: 4px;">
          <pre>{{ debugInfo }}</pre>
        </div>
      </a-card>
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

const userStore = useUserStore()
const debugInfo = ref('')

// Dashboard data
const stats = reactive({
  totalUsers: 0,
  totalOrganizations: 0,
  activeSessions: 0,
  totalApplications: 0
})

const recentActivities = ref([
  {
    title: 'User Login',
    description: 'Admin user logged in',
    icon: 'user',
    time: '2 minutes ago'
  },
  {
    title: 'Organization Created',
    description: 'New organization "Test Branch" created',
    icon: 'team',
    time: '1 hour ago'
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

onMounted(() => {
  // Load dashboard data
  stats.totalUsers = 2
  stats.totalOrganizations = 2
  stats.activeSessions = 1
  stats.totalApplications = 0
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
</style>
