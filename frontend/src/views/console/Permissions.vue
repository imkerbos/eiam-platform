<template>
  <div class="permissions-page">
    <div class="page-header">
      <h2>Permissions Management</h2>
      <p>Manage permission routes and access assignments for applications</p>
    </div>

    <!-- Permission Categories -->
    <a-tabs v-model:activeKey="activeTab" type="card" @change="handleTabChange">
      <a-tab-pane key="permissions" tab="Permission Routes">
        <div class="tab-content">
          <div class="actions-bar">
            <a-button type="primary" @click="showPermissionModal">
              <PlusOutlined />
              Create Permission Route
            </a-button>
          </div>
          
          <a-table
            :columns="permissionColumns"
            :data-source="permissions"
            :loading="loading"
            :pagination="pagination"
            @change="handleTableChange"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'applications'">
                <a-tag v-for="app in record.applications" :key="app" color="blue">
                  {{ app }}
                </a-tag>
                <span v-if="!record.applications || record.applications.length === 0">-</span>
              </template>
              <template v-else-if="column.key === 'groups'">
                <a-tag v-for="group in record.application_groups" :key="group" color="green">
                  {{ group }}
                </a-tag>
                <span v-if="!record.application_groups || record.application_groups.length === 0">-</span>
              </template>
              <template v-else-if="column.key === 'status'">
                <a-tag :color="record.status === 'active' ? 'green' : 'red'">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-button type="link" size="small" @click="editPermission(record)">
                    Edit
                  </a-button>
                  <a-popconfirm
                    title="Are you sure you want to delete this permission route?"
                    @confirm="deletePermission(record.id)"
                  >
                    <a-button type="link" size="small" danger>Delete</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </a-tab-pane>

      <a-tab-pane key="assignments" tab="Permission Assignments">
        <div class="tab-content">
          <div class="actions-bar">
            <a-button type="primary" @click="showAssignmentModal">
              <PlusOutlined />
              Assign Permission
            </a-button>
          </div>
          
          <a-table
            :columns="assignmentColumns"
            :data-source="assignments"
            :loading="loading"
            :pagination="pagination"
            @change="handleTableChange"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'assignee_type'">
                <a-tag :color="record.assignee_type === 'user' ? 'blue' : 'green'">
                  {{ record.assignee_type === 'user' ? 'User' : 'Organization' }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'status'">
                <a-tag :color="record.status === 'active' ? 'green' : 'red'">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-popconfirm
                    title="Are you sure you want to remove this assignment?"
                    @confirm="removeAssignment(record)"
                  >
                    <a-button type="link" size="small" danger>Remove</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </a-tab-pane>
    </a-tabs>

    <!-- Permission Modal -->
    <a-modal
      v-model:open="permissionModalVisible"
      :title="editingPermission ? 'Edit Permission Route' : 'Create Permission Route'"
      width="800px"
      @ok="handlePermissionSubmit"
      @cancel="handlePermissionCancel"
    >
      <a-form
        ref="permissionFormRef"
        :model="permissionForm"
        :rules="permissionRules"
        layout="vertical"
      >
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Route Name" name="name">
              <a-input 
                v-model:value="permissionForm.name" 
                placeholder="e.g., Finance Access, HR Management, Sales Report"
                :maxlength="100"
                @input="handleNameInput"
              />
              <div class="form-help-text">
                Display name for the permission route. Route Code will be auto-generated.
              </div>
            </a-form-item>
          </a-col>
                  <a-col :span="12">
                    <a-form-item label="Route Code" name="code">
                      <a-input 
                        v-model:value="permissionForm.code" 
                        placeholder="e.g., FINANCE_ACCESS, HR_MANAGEMENT, SALES_REPORT"
                        :maxlength="50"
                        @input="handleCodeInput"
                      />
                      <div class="form-help-text">
                        Auto-generated from Route Name. Format: UPPERCASE_WITH_UNDERSCORES
                      </div>
                    </a-form-item>
                  </a-col>
        </a-row>
        
        <a-form-item label="Description" name="description">
          <a-textarea 
            v-model:value="permissionForm.description" 
            :rows="3" 
            placeholder="Describe what this permission route allows access to"
            :maxlength="500"
          />
          <div class="form-help-text">
            Describe what this permission route allows access to, e.g., access to finance system, view sales reports
          </div>
        </a-form-item>
        
        <a-form-item label="Accessible Applications" name="applications">
          <a-select 
            v-model:value="permissionForm.applications" 
            mode="multiple" 
            placeholder="Select applications this route can access"
            :options="applicationOptions"
          />
          <div class="form-help-text">
            Select specific applications that this permission route can access
          </div>
        </a-form-item>
        
        <a-form-item label="Accessible Application Groups" name="groups">
          <a-select 
            v-model:value="permissionForm.groups" 
            mode="multiple" 
            placeholder="Select application groups this route can access"
            :options="groupOptions"
          />
          <div class="form-help-text">
            Select application groups (containing multiple applications) that this permission route can access
          </div>
        </a-form-item>
        
        <a-form-item label="Status" name="status">
          <a-select v-model:value="permissionForm.status">
            <a-select-option value="active">Active</a-select-option>
            <a-select-option value="inactive">Inactive</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Assignment Modal -->
    <a-modal
      v-model:open="assignmentModalVisible"
      :title="editingAssignment ? 'Edit Permission Assignment' : 'Assign Permission'"
      width="600px"
      @ok="handleAssignmentSubmit"
      @cancel="handleAssignmentCancel"
    >
      <a-form
        ref="assignmentFormRef"
        :model="assignmentForm"
        :rules="assignmentRules"
        layout="vertical"
      >
        <a-form-item label="Assignee Type" name="assigneeType">
          <a-radio-group v-model:value="assignmentForm.assigneeType" @change="handleAssigneeTypeChange">
            <a-radio value="user">Individual User</a-radio>
            <a-radio value="organization">Organization/Department</a-radio>
          </a-radio-group>
        </a-form-item>
        
        <a-form-item v-if="assignmentForm.assigneeType === 'user'" label="User" name="userId" :rules="assignmentRules.userId">
          <a-select v-model:value="assignmentForm.userId" placeholder="Select user">
            <a-select-option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.display_name || user.username }} ({{ user.email }})
            </a-select-option>
          </a-select>
        </a-form-item>
        
        <a-form-item v-if="assignmentForm.assigneeType === 'organization'" label="Organization/Department" name="organizationId" :rules="assignmentRules.organizationId">
          <a-select v-model:value="assignmentForm.organizationId" placeholder="Select organization">
            <a-select-option v-for="org in organizations" :key="org.id" :value="org.id">
              {{ org.name }} ({{ org.code }})
            </a-select-option>
          </a-select>
        </a-form-item>
        
        <a-form-item label="Permission Route" name="permissionId">
          <a-select v-model:value="assignmentForm.permissionId" placeholder="Select permission route">
            <a-select-option v-for="permission in permissions" :key="permission.id" :value="permission.id">
              {{ permission.name }} ({{ permission.code }})
            </a-select-option>
          </a-select>
        </a-form-item>
        
        <a-form-item label="Status" name="status">
          <a-select v-model:value="assignmentForm.status">
            <a-select-option value="active">Active</a-select-option>
            <a-select-option value="inactive">Inactive</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { permissionApi } from '@/api/permissions'
import { userApi } from '@/api/users'
import { organizationApi } from '@/api/organizations'
import { applicationApi } from '@/api/applications'

// Data
const loading = ref(false)
const activeTab = ref('permissions')
const permissions = ref<any[]>([])
const assignments = ref<any[]>([])
const users = ref<any[]>([])
const organizations = ref<any[]>([])
const applications = ref<any[]>([])
const applicationGroups = ref<any[]>([])
const permissionModalVisible = ref(false)
const assignmentModalVisible = ref(false)
const editingPermission = ref<any>(null)
const editingAssignment = ref<any>(null)
const permissionFormRef = ref()
const assignmentFormRef = ref()

const permissionForm = reactive({
  name: '',
  code: '',
  description: '',
  applications: [],
  groups: [],
  status: 'active'
})

const assignmentForm = reactive({
  assigneeType: 'user',
  userId: '',
  organizationId: '',
  permissionId: '',
  status: 'active'
})

const permissionRules = {
  name: [{ required: true, message: 'Please input permission route name!' }],
  code: [
    { required: true, message: 'Please input permission route code!' },
    { 
      pattern: /^[A-Z][A-Z0-9_]*$/, 
      message: 'Route code must start with uppercase letter and contain only uppercase letters, numbers, and underscores!' 
    },
    { min: 3, max: 50, message: 'Route code must be between 3 and 50 characters!' }
  ],
  status: [{ required: true, message: 'Please select status!' }]
}

const assignmentRules = computed(() => ({
  assigneeType: [{ required: true, message: 'Please select assignee type!' }],
  userId: assignmentForm.assigneeType === 'user' ? [{ required: true, message: 'Please select user!' }] : [],
  organizationId: assignmentForm.assigneeType === 'organization' ? [{ required: true, message: 'Please select organization!' }] : [],
  permissionId: [{ required: true, message: 'Please select permission route!' }],
  status: [{ required: true, message: 'Please select status!' }]
}))

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  total_pages: 0
})

const permissionColumns = [
  {
    title: 'Route Name',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: 'Route Code',
    dataIndex: 'code',
    key: 'code'
  },
  {
    title: 'Accessible Applications',
    dataIndex: 'applications',
    key: 'applications'
  },
  {
    title: 'Accessible Groups',
    dataIndex: 'groups',
    key: 'groups'
  },
  {
    title: 'Description',
    dataIndex: 'description',
    key: 'description'
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status'
  },
  {
    title: 'Actions',
    key: 'actions'
  }
]

const assignmentColumns = [
  {
    title: 'Assignee Type',
    dataIndex: 'assignee_type',
    key: 'assignee_type'
  },
  {
    title: 'Assignee Name',
    dataIndex: 'assignee_name',
    key: 'assignee_name'
  },
  {
    title: 'Permission Route',
    dataIndex: 'permission_name',
    key: 'permission_name'
  },
  {
    title: 'Route Code',
    dataIndex: 'permission_code',
    key: 'permission_code'
  },
  {
    title: 'Assigned At',
    dataIndex: 'assigned_at',
    key: 'assigned_at'
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status'
  },
  {
    title: 'Actions',
    key: 'actions'
  }
]

// Computed properties for options
const applicationOptions = computed(() => {
  return applications.value.map(app => ({
    label: app.name,
    value: app.id
  }))
})

const groupOptions = computed(() => {
  return applicationGroups.value.map(group => ({
    label: group.name,
    value: group.id
  }))
})

// Methods
const loadData = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'permissions') {
      await loadPermissions()
    } else if (activeTab.value === 'assignments') {
      await loadAssignments()
    }
  } catch (error: any) {
    console.error('Failed to load data:', error)
    message.error(error.response?.data?.message || 'Failed to load data')
  } finally {
    loading.value = false
  }
}

const loadPermissions = async () => {
  try {
    const response = await permissionApi.getPermissionRoutes({
      page: pagination.current,
      page_size: pagination.pageSize
    })
    permissions.value = response.items || []
    pagination.total = response.total
    pagination.total_pages = response.total_pages
  } catch (error: any) {
    console.error('Failed to load permission routes:', error)
    message.error(error.response?.data?.message || 'Failed to load permission routes')
  }
}

const loadAssignments = async () => {
  try {
    const response = await permissionApi.getPermissionRouteAssignments({
      page: pagination.current,
      page_size: pagination.pageSize
    })
    assignments.value = response.items || []
    pagination.total = response.total
    pagination.total_pages = response.total_pages
  } catch (error: any) {
    console.error('Failed to load permission route assignments:', error)
    message.error(error.response?.data?.message || 'Failed to load permission route assignments')
  }
}

const loadUsers = async () => {
  try {
    const response = await userApi.getUsers({
      page: 1,
      page_size: 1000
    })
    users.value = response.items || []
  } catch (error: any) {
    console.error('Failed to load users:', error)
  }
}

const loadOrganizations = async () => {
  try {
    const response = await organizationApi.getOrganizations({
      page: 1,
      page_size: 1000
    })
    organizations.value = response.items || []
  } catch (error: any) {
    console.error('Failed to load organizations:', error)
  }
}

const loadApplications = async () => {
  try {
    const response = await applicationApi.getApplications({
      page: 1,
      page_size: 1000
    })
    applications.value = response.items || []
  } catch (error: any) {
    console.error('Failed to load applications:', error)
  }
}

const loadApplicationGroups = async () => {
  try {
    const response = await applicationApi.getApplicationGroups({
      page: 1,
      page_size: 1000
    })
    applicationGroups.value = response.items || []
  } catch (error: any) {
    console.error('Failed to load application groups:', error)
  }
}

const handleTabChange = (key: string) => {
  activeTab.value = key
  loadData()
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadData()
}

// Permission methods
const showPermissionModal = () => {
  editingPermission.value = null
  Object.assign(permissionForm, {
    name: '',
    code: '',
    description: '',
    applications: [],
    groups: [],
    status: 'active'
  })
  permissionModalVisible.value = true
}

const editPermission = (permission: any) => {
  editingPermission.value = permission
  Object.assign(permissionForm, {
    name: permission.name,
    code: permission.code,
    description: permission.description,
    applications: permission.applications || [],
    groups: permission.application_groups || [],
    status: permission.status
  })
  permissionModalVisible.value = true
}

const handlePermissionSubmit = async () => {
  try {
    await permissionFormRef.value?.validate()
    if (editingPermission.value) {
      await permissionApi.updatePermissionRoute(editingPermission.value.id, {
        name: permissionForm.name,
        description: permissionForm.description,
        applications: permissionForm.applications,
        application_groups: permissionForm.groups,
        status: permissionForm.status
      })
      message.success('Permission route updated successfully')
    } else {
      await permissionApi.createPermissionRoute({
        name: permissionForm.name,
        code: permissionForm.code,
        description: permissionForm.description,
        applications: permissionForm.applications,
        application_groups: permissionForm.groups,
        status: permissionForm.status
      })
      message.success('Permission route created successfully')
    }
    permissionModalVisible.value = false
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Please check the form')
  }
}

const handlePermissionCancel = () => {
  permissionModalVisible.value = false
}

const deletePermission = async (permissionId: string) => {
  try {
    await permissionApi.deletePermissionRoute(permissionId)
    message.success('Permission route deleted successfully')
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Failed to delete permission route')
  }
}

// Assignment methods
const showAssignmentModal = () => {
  editingAssignment.value = null
  Object.assign(assignmentForm, {
    assigneeType: 'user',
    userId: '',
    organizationId: '',
    permissionId: '',
    status: 'active'
  })
  assignmentModalVisible.value = true
}


const handleAssignmentSubmit = async () => {
  try {
    await assignmentFormRef.value?.validate()
    await permissionApi.assignPermissionRoute({
      permission_route_id: assignmentForm.permissionId,
      assignee_type: assignmentForm.assigneeType,
      assignee_id: assignmentForm.assigneeType === 'user' ? assignmentForm.userId : assignmentForm.organizationId,
      status: assignmentForm.status
    })
    message.success('Permission route assigned successfully')
    assignmentModalVisible.value = false
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Please check the form')
  }
}

const handleAssignmentCancel = () => {
  assignmentModalVisible.value = false
}

const handleAssigneeTypeChange = () => {
  // 清空相关字段
  assignmentForm.userId = ''
  assignmentForm.organizationId = ''
}

const handleCodeInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  const value = target.value
  // 自动转换为大写，只允许字母、数字和下划线
  const formattedValue = value.toUpperCase().replace(/[^A-Z0-9_]/g, '')
  permissionForm.code = formattedValue
}

const handleNameInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  const value = target.value
  // 自动生成Route Code：转换为大写，空格替换为下划线，移除特殊字符
  const generatedCode = value
    .toUpperCase()
    .replace(/\s+/g, '_')
    .replace(/[^A-Z0-9_]/g, '')
    .replace(/_+/g, '_')
    .replace(/^_|_$/g, '')
  
  // 只有当Route Code为空或者与生成的代码不同时才自动更新
  if (!permissionForm.code || permissionForm.code === '') {
    permissionForm.code = generatedCode
  }
}

const removeAssignment = async (assignment: any) => {
  try {
    await permissionApi.removePermissionRouteAssignment(
      assignment.assignee_type,
      assignment.assignee_id,
      assignment.permission_route_id
    )
    message.success('Permission route assignment removed successfully')
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Failed to remove assignment')
  }
}

// Lifecycle
onMounted(() => {
  loadData()
  loadUsers()
  loadOrganizations()
  loadApplications()
  loadApplicationGroups()
})
</script>

<style scoped>
.permissions-page {
  padding: 24px;
}

.form-help-text {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
  line-height: 1.4;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.page-header p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.tab-content {
  background: #fff;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.actions-bar {
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>