<template>
  <div class="audit-page">
    <div class="page-header">
      <h2>Audit & Monitoring</h2>
      <p>Monitor system activities, user sessions, and security events</p>
    </div>

    <!-- Audit Categories -->
    <a-tabs v-model:activeKey="activeTab" type="card">
      <!-- Operation Audit -->
      <a-tab-pane key="operations" tab="Operation Audit">
        <div class="tab-content">
          <div class="filters-bar">
            <a-row :gutter="16">
              <a-col :span="6">
                <a-form-item label="User">
                  <a-select
                    v-model:value="operationFilters.user"
                    placeholder="All users"
                    allow-clear
                    show-search
                  >
                    <a-select-option v-for="user in users" :key="user.id" :value="user.username">
                      {{ user.display_name }}
                    </a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="6">
                <a-form-item label="Action">
                  <a-select
                    v-model:value="operationFilters.action"
                    placeholder="All actions"
                    allow-clear
                  >
                    <a-select-option value="create">Create</a-select-option>
                    <a-select-option value="read">Read</a-select-option>
                    <a-select-option value="update">Update</a-select-option>
                    <a-select-option value="delete">Delete</a-select-option>
                    <a-select-option value="login">Login</a-select-option>
                    <a-select-option value="logout">Logout</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="6">
                <a-form-item label="Resource">
                  <a-select
                    v-model:value="operationFilters.resource"
                    placeholder="All resources"
                    allow-clear
                  >
                    <a-select-option value="users">Users</a-select-option>
                    <a-select-option value="organizations">Organizations</a-select-option>
                    <a-select-option value="applications">Applications</a-select-option>
                    <a-select-option value="permissions">Permissions</a-select-option>
                    <a-select-option value="system">System</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="6">
                <a-form-item label="Date Range">
                  <a-range-picker
                    v-model:value="operationFilters.dateRange"
                    style="width: 100%"
                  />
                </a-form-item>
              </a-col>
            </a-row>
            <a-row>
              <a-col :span="24">
                <a-space>
                  <a-button type="primary" @click="searchOperations">
                    Search
                  </a-button>
                  <a-button @click="resetOperationFilters">
                    Reset
                  </a-button>
                  <a-button @click="exportOperations">
                    Export
                  </a-button>
                </a-space>
              </a-col>
            </a-row>
          </div>
          
          <a-table
            :columns="operationColumns"
            :data-source="operations"
            :loading="loading"
            :pagination="pagination"
            @change="handleTableChange"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <a-tag :color="getActionColor(record.action)">{{ record.action }}</a-tag>
              </template>
              <template v-else-if="column.key === 'resource'">
                <a-tag :color="getResourceColor(record.resource)">{{ record.resource }}</a-tag>
              </template>
              <template v-else-if="column.key === 'status'">
                <a-tag :color="record.status === 'success' ? 'green' : 'red'">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'details'">
                <a-button type="link" size="small" @click="viewOperationDetails(record)">
                  View Details
                </a-button>
              </template>
            </template>
          </a-table>
        </div>
      </a-tab-pane>

      <!-- Login Audit -->
      <a-tab-pane key="logins" tab="Login Audit">
        <div class="tab-content">
          <div class="filters-bar">
            <a-row :gutter="16">
              <a-col :span="6">
                <a-form-item label="User">
                  <a-select
                    v-model:value="loginFilters.user"
                    placeholder="All users"
                    allow-clear
                    show-search
                  >
                    <a-select-option v-for="user in users" :key="user.id" :value="user.username">
                      {{ user.display_name }}
                    </a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="6">
                <a-form-item label="Status">
                  <a-select
                    v-model:value="loginFilters.status"
                    placeholder="All statuses"
                    allow-clear
                  >
                    <a-select-option value="success">Success</a-select-option>
                    <a-select-option value="failed">Failed</a-select-option>
                    <a-select-option value="locked">Locked</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="6">
                <a-form-item label="IP Address">
                  <a-input v-model:value="loginFilters.ipAddress" placeholder="Enter IP address" />
                </a-form-item>
              </a-col>
              <a-col :span="6">
                <a-form-item label="Date Range">
                  <a-range-picker
                    v-model:value="loginFilters.dateRange"
                    style="width: 100%"
                  />
                </a-form-item>
              </a-col>
            </a-row>
            <a-row>
              <a-col :span="24">
                <a-space>
                  <a-button type="primary" @click="searchLogins">
                    Search
                  </a-button>
                  <a-button @click="resetLoginFilters">
                    Reset
                  </a-button>
                  <a-button @click="exportLogins">
                    Export
                  </a-button>
                </a-space>
              </a-col>
            </a-row>
          </div>
          
          <a-table
            :columns="loginColumns"
            :data-source="logins"
            :loading="loading"
            :pagination="pagination"
            @change="handleTableChange"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <a-tag :color="record.status === 'success' ? 'green' : 'red'">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'location'">
                <a-tag v-if="record.location">{{ record.location }}</a-tag>
                <span v-else>-</span>
              </template>
              <template v-else-if="column.key === 'details'">
                <a-button type="link" size="small" @click="viewLoginDetails(record)">
                  View Details
                </a-button>
              </template>
            </template>
          </a-table>
        </div>
      </a-tab-pane>

      <!-- Online Users -->
      <a-tab-pane key="sessions" tab="Online Users">
        <div class="tab-content">
          <div class="actions-bar">
            <a-space>
              <a-button type="primary" @click="refreshSessions">
                <ReloadOutlined />
                Refresh
              </a-button>
              <a-button @click="exportSessions">
                Export
              </a-button>
              <a-popconfirm
                title="Are you sure you want to terminate all sessions?"
                @confirm="terminateAllSessions"
              >
                <a-button danger>
                  Terminate All
                </a-button>
              </a-popconfirm>
            </a-space>
          </div>
          
          <a-table
            :columns="sessionColumns"
            :data-source="sessions"
            :loading="loading"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <a-tag :color="record.status === 'active' ? 'green' : 'orange'">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'location'">
                <a-tag v-if="record.location">{{ record.location }}</a-tag>
                <span v-else>-</span>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-button type="link" size="small" @click="viewSessionDetails(record)">
                    Details
                  </a-button>
                  <a-popconfirm
                    title="Are you sure you want to terminate this session?"
                    @confirm="terminateSession(record.id)"
                  >
                    <a-button type="link" size="small" danger>Terminate</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </a-tab-pane>
    </a-tabs>

    <!-- Operation Details Modal -->
    <a-modal
      v-model:open="operationDetailsVisible"
      title="Operation Details"
      width="800px"
      @cancel="operationDetailsVisible = false"
    >
      <div v-if="selectedOperation">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="User">
            {{ selectedOperation.user }}
          </a-descriptions-item>
          <a-descriptions-item label="Action">
            <a-tag :color="getActionColor(selectedOperation.action)">
              {{ selectedOperation.action }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Resource">
            <a-tag :color="getResourceColor(selectedOperation.resource)">
              {{ selectedOperation.resource }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Status">
            <a-tag :color="selectedOperation.status === 'success' ? 'green' : 'red'">
              {{ selectedOperation.status }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="IP Address">
            {{ selectedOperation.ipAddress }}
          </a-descriptions-item>
          <a-descriptions-item label="User Agent">
            {{ selectedOperation.userAgent }}
          </a-descriptions-item>
          <a-descriptions-item label="Timestamp">
            {{ selectedOperation.timestamp }}
          </a-descriptions-item>
          <a-descriptions-item label="Duration">
            {{ selectedOperation.duration }}ms
          </a-descriptions-item>
          <a-descriptions-item label="Description" :span="2">
            {{ selectedOperation.description }}
          </a-descriptions-item>
          <a-descriptions-item label="Request Data" :span="2">
            <pre>{{ JSON.stringify(selectedOperation.requestData, null, 2) }}</pre>
          </a-descriptions-item>
          <a-descriptions-item label="Response Data" :span="2">
            <pre>{{ JSON.stringify(selectedOperation.responseData, null, 2) }}</pre>
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>

    <!-- Login Details Modal -->
    <a-modal
      v-model:open="loginDetailsVisible"
      title="Login Details"
      width="600px"
      @cancel="loginDetailsVisible = false"
    >
      <div v-if="selectedLogin">
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="User">
            {{ selectedLogin.user }}
          </a-descriptions-item>
          <a-descriptions-item label="Status">
            <a-tag :color="selectedLogin.status === 'success' ? 'green' : 'red'">
              {{ selectedLogin.status }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="IP Address">
            {{ selectedLogin.ipAddress }}
          </a-descriptions-item>
          <a-descriptions-item label="Location">
            {{ selectedLogin.location }}
          </a-descriptions-item>
          <a-descriptions-item label="User Agent">
            {{ selectedLogin.userAgent }}
          </a-descriptions-item>
          <a-descriptions-item label="Timestamp">
            {{ selectedLogin.timestamp }}
          </a-descriptions-item>
          <a-descriptions-item label="Failure Reason" v-if="selectedLogin.failureReason">
            {{ selectedLogin.failureReason }}
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>

    <!-- Session Details Modal -->
    <a-modal
      v-model:open="sessionDetailsVisible"
      title="Session Details"
      width="600px"
      @cancel="sessionDetailsVisible = false"
    >
      <div v-if="selectedSession">
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="User">
            {{ selectedSession.user }}
          </a-descriptions-item>
          <a-descriptions-item label="Session ID">
            {{ selectedSession.sessionId }}
          </a-descriptions-item>
          <a-descriptions-item label="IP Address">
            {{ selectedSession.ipAddress }}
          </a-descriptions-item>
          <a-descriptions-item label="Location">
            {{ selectedSession.location }}
          </a-descriptions-item>
          <a-descriptions-item label="User Agent">
            {{ selectedSession.userAgent }}
          </a-descriptions-item>
          <a-descriptions-item label="Login Time">
            {{ selectedSession.loginTime }}
          </a-descriptions-item>
          <a-descriptions-item label="Last Activity">
            {{ selectedSession.lastActivity }}
          </a-descriptions-item>
          <a-descriptions-item label="Status">
            <a-tag :color="selectedSession.status === 'active' ? 'green' : 'orange'">
              {{ selectedSession.status }}
            </a-tag>
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import { auditApi } from '@/api/audit'

// Data
const activeTab = ref('operations')
const loading = ref(false)
const operationDetailsVisible = ref(false)
const loginDetailsVisible = ref(false)
const sessionDetailsVisible = ref(false)
const selectedOperation = ref(null)
const selectedLogin = ref(null)
const selectedSession = ref(null)

// Mock data
const users = ref([
  { id: '1', username: 'admin', display_name: 'Administrator' },
  { id: '2', username: 'user1', display_name: 'John Doe' },
  { id: '3', username: 'user2', display_name: 'Jane Smith' }
])

const operations = ref([
  {
    id: '1',
    user: 'admin',
    action: 'create',
    resource: 'users',
    status: 'success',
    ipAddress: '192.168.1.100',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    timestamp: '2024-01-01 10:00:00',
    duration: 150,
    description: 'Created new user: john.doe',
    requestData: { username: 'john.doe', email: 'john@example.com' },
    responseData: { id: '123', status: 'created' }
  },
  {
    id: '2',
    user: 'user1',
    action: 'update',
    resource: 'profile',
    status: 'success',
    ipAddress: '192.168.1.101',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    timestamp: '2024-01-01 09:30:00',
    duration: 80,
    description: 'Updated profile information',
    requestData: { displayName: 'John Doe Updated' },
    responseData: { status: 'updated' }
  }
])

const logins = ref([
  {
    id: '1',
    user: 'admin',
    status: 'success',
    ipAddress: '192.168.1.100',
    location: 'New York, US',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    timestamp: '2024-01-01 10:00:00',
    failureReason: null
  },
  {
    id: '2',
    user: 'user1',
    status: 'failed',
    ipAddress: '192.168.1.101',
    location: 'Los Angeles, US',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    timestamp: '2024-01-01 09:30:00',
    failureReason: 'Invalid password'
  }
])

const sessions = ref([
  {
    id: '1',
    user: 'admin',
    sessionId: 'sess_123456789',
    ipAddress: '192.168.1.100',
    location: 'New York, US',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    loginTime: '2024-01-01 10:00:00',
    lastActivity: '2024-01-01 10:15:00',
    status: 'active'
  },
  {
    id: '2',
    user: 'user1',
    sessionId: 'sess_987654321',
    ipAddress: '192.168.1.101',
    location: 'Los Angeles, US',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    loginTime: '2024-01-01 09:30:00',
    lastActivity: '2024-01-01 10:10:00',
    status: 'idle'
  }
])

// Filters
const operationFilters = reactive({
  user: '',
  action: '',
  resource: '',
  dateRange: []
})

const loginFilters = reactive({
  user: '',
  status: '',
  ipAddress: '',
  dateRange: []
})

// Table columns
const operationColumns = [
  { title: 'User', dataIndex: 'user', key: 'user' },
  { title: 'Action', dataIndex: 'action', key: 'action' },
  { title: 'Resource', dataIndex: 'resource', key: 'resource' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'IP Address', dataIndex: 'ipAddress', key: 'ipAddress' },
  { title: 'Timestamp', dataIndex: 'timestamp', key: 'timestamp' },
  { title: 'Duration', dataIndex: 'duration', key: 'duration', customRender: ({ text }: any) => `${text}ms` },
  { title: 'Details', key: 'details' }
]

const loginColumns = [
  { title: 'User', dataIndex: 'user', key: 'user' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'IP Address', dataIndex: 'ipAddress', key: 'ipAddress' },
  { title: 'Location', dataIndex: 'location', key: 'location' },
  { title: 'Timestamp', dataIndex: 'timestamp', key: 'timestamp' },
  { title: 'Details', key: 'details' }
]

const sessionColumns = [
  { title: 'User', dataIndex: 'user', key: 'user' },
  { title: 'IP Address', dataIndex: 'ipAddress', key: 'ipAddress' },
  { title: 'Location', dataIndex: 'location', key: 'location' },
  { title: 'Login Time', dataIndex: 'loginTime', key: 'loginTime' },
  { title: 'Last Activity', dataIndex: 'lastActivity', key: 'lastActivity' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Actions', key: 'actions' }
]

// Pagination
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  pageSizeOptions: ['10', '20', '50', '100'],
  showTotal: (total: number, range: [number, number]) => {
    return `显示 ${range[0]}-${range[1]} 条，共 ${total} 条`
  }
})

// Methods
const loadData = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'operations') {
      const response = await auditApi.getOperationLogs({
        page: pagination.current,
        page_size: pagination.pageSize,
        user_id: operationFilters.user,
        action: operationFilters.action,
        resource: operationFilters.resource,
        start_date: operationFilters.dateRange?.[0]?.format('YYYY-MM-DD'),
        end_date: operationFilters.dateRange?.[1]?.format('YYYY-MM-DD')
      })
      operations.value = response.items
      pagination.total = response.total
    } else if (activeTab.value === 'logins') {
      const response = await auditApi.getLoginLogs({
        page: pagination.current,
        page_size: pagination.pageSize,
        user_id: loginFilters.user,
        success: loginFilters.status === 'success' ? true : loginFilters.status === 'failed' ? false : undefined,
        login_type: loginFilters.loginType,
        start_date: loginFilters.dateRange?.[0]?.format('YYYY-MM-DD'),
        end_date: loginFilters.dateRange?.[1]?.format('YYYY-MM-DD')
      })
      logins.value = response.items
      pagination.total = response.total
    } else if (activeTab.value === 'sessions') {
      const response = await auditApi.getOnlineSessions({
        page: pagination.current,
        page_size: pagination.pageSize,
        user_id: sessionFilters.user,
        is_active: sessionFilters.status === 'active' ? true : sessionFilters.status === 'inactive' ? false : undefined
      })
      sessions.value = response.items
      pagination.total = response.total
    }
  } catch (error) {
    message.error('Failed to load data')
  } finally {
    loading.value = false
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadData()
}

// Color helpers
const getActionColor = (action: string) => {
  const colors = {
    create: 'green',
    read: 'blue',
    update: 'orange',
    delete: 'red',
    login: 'purple',
    logout: 'gray'
  }
  return colors[action as keyof typeof colors] || 'default'
}

const getResourceColor = (resource: string) => {
  const colors = {
    users: 'blue',
    organizations: 'green',
    applications: 'purple',
    permissions: 'orange',
    system: 'red'
  }
  return colors[resource as keyof typeof colors] || 'default'
}

// Operation methods
const searchOperations = () => {
  loadData()
}

const resetOperationFilters = () => {
  Object.assign(operationFilters, {
    user: '',
    action: '',
    resource: '',
    dateRange: []
  })
  loadData()
}

const exportOperations = () => {
  message.info('Exporting operation audit data...')
  // TODO: Implement export functionality
}

const viewOperationDetails = (operation: any) => {
  selectedOperation.value = operation
  operationDetailsVisible.value = true
}

// Login methods
const searchLogins = () => {
  loadData()
}

const resetLoginFilters = () => {
  Object.assign(loginFilters, {
    user: '',
    status: '',
    ipAddress: '',
    dateRange: []
  })
  loadData()
}

const exportLogins = () => {
  message.info('Exporting login audit data...')
  // TODO: Implement export functionality
}

const viewLoginDetails = (login: any) => {
  selectedLogin.value = login
  loginDetailsVisible.value = true
}

// Session methods
const refreshSessions = () => {
  loadData()
}

const exportSessions = () => {
  message.info('Exporting session data...')
  // TODO: Implement export functionality
}

const terminateAllSessions = async () => {
  try {
    // TODO: Implement API call
    message.success('All sessions terminated successfully')
    loadData()
  } catch (error) {
    message.error('Failed to terminate sessions')
  }
}

const viewSessionDetails = (session: any) => {
  selectedSession.value = session
  sessionDetailsVisible.value = true
}

const terminateSession = async (sessionId: string) => {
  try {
    // TODO: Implement API call
    message.success('Session terminated successfully')
    loadData()
  } catch (error) {
    message.error('Failed to terminate session')
  }
}

onMounted(() => {
  loadData()
})

// Watch active tab changes
watch(activeTab, () => {
  pagination.current = 1
  loadData()
})
</script>

<style scoped>
.audit-page {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h2 {
  font-size: 24px;
  font-weight: 600;
  color: #1890ff;
  margin: 0 0 8px 0;
}

.page-header p {
  color: #666;
  margin: 0;
}

.tab-content {
  background: #fff;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.filters-bar {
  margin-bottom: 24px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 6px;
}

.actions-bar {
  margin-bottom: 16px;
}

pre {
  background: #f5f5f5;
  padding: 8px;
  border-radius: 4px;
  font-size: 12px;
  overflow-x: auto;
}

/* Responsive Design */
@media (max-width: 768px) {
  .audit-page {
    padding: 16px;
  }
  
  .tab-content {
    padding: 16px;
  }
  
  .filters-bar {
    padding: 12px;
  }
}
</style>
