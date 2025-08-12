<template>
  <div class="permissions-page">
    <div class="page-header">
      <h2>Permissions Management</h2>
      <p>Manage system permissions and access controls</p>
    </div>

    <!-- Permission Categories -->
    <a-tabs v-model:activeKey="activeTab" type="card">
      <a-tab-pane key="roles" tab="Roles">
        <div class="tab-content">
          <div class="actions-bar">
            <a-button type="primary" @click="showRoleModal">
              <PlusOutlined />
              Add Role
            </a-button>
          </div>
          
          <a-table
            :columns="roleColumns"
            :data-source="roles"
            :loading="loading"
            :pagination="pagination"
            @change="handleTableChange"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <a-tag :color="record.status === 'active' ? 'green' : 'red'">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'permissions'">
                <a-tag v-for="perm in record.permissions.slice(0, 3)" :key="perm" style="margin: 2px">
                  {{ perm }}
                </a-tag>
                <a-tag v-if="record.permissions.length > 3" color="blue">
                  +{{ record.permissions.length - 3 }} more
                </a-tag>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-button type="link" size="small" @click="editRole(record)">
                    Edit
                  </a-button>
                  <a-button type="link" size="small" @click="viewRolePermissions(record)">
                    Permissions
                  </a-button>
                  <a-popconfirm
                    title="Are you sure you want to delete this role?"
                    @confirm="deleteRole(record.id)"
                  >
                    <a-button type="link" size="small" danger>Delete</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </a-tab-pane>

      <a-tab-pane key="permissions" tab="Permissions">
        <div class="tab-content">
          <div class="actions-bar">
            <a-button type="primary" @click="showPermissionModal">
              <PlusOutlined />
              Add Permission
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
              <template v-if="column.key === 'status'">
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
                    title="Are you sure you want to delete this permission?"
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

      <a-tab-pane key="assignments" tab="Role Assignments">
        <div class="tab-content">
          <div class="actions-bar">
            <a-button type="primary" @click="showAssignmentModal">
              <PlusOutlined />
              Assign Role
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
              <template v-if="column.key === 'status'">
                <a-tag :color="record.status === 'active' ? 'green' : 'red'">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-button type="link" size="small" @click="editAssignment(record)">
                    Edit
                  </a-button>
                  <a-popconfirm
                    title="Are you sure you want to remove this assignment?"
                    @confirm="removeAssignment(record.id)"
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

    <!-- Role Modal -->
    <a-modal
      v-model:open="roleModalVisible"
      :title="editingRole ? 'Edit Role' : 'Add Role'"
      @ok="handleRoleSubmit"
      @cancel="handleRoleCancel"
    >
      <a-form
        ref="roleFormRef"
        :model="roleForm"
        :rules="roleRules"
        layout="vertical"
      >
        <a-form-item label="Role Name" name="name">
          <a-input v-model:value="roleForm.name" placeholder="Enter role name" />
        </a-form-item>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="roleForm.description" placeholder="Enter role description" />
        </a-form-item>
        <a-form-item label="Status" name="status">
          <a-select v-model:value="roleForm.status">
            <a-select-option value="active">Active</a-select-option>
            <a-select-option value="inactive">Inactive</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Permission Modal -->
    <a-modal
      v-model:open="permissionModalVisible"
      :title="editingPermission ? 'Edit Permission' : 'Add Permission'"
      @ok="handlePermissionSubmit"
      @cancel="handlePermissionCancel"
    >
      <a-form
        ref="permissionFormRef"
        :model="permissionForm"
        :rules="permissionRules"
        layout="vertical"
      >
        <a-form-item label="Permission Name" name="name">
          <a-input v-model:value="permissionForm.name" placeholder="Enter permission name" />
        </a-form-item>
        <a-form-item label="Resource" name="resource">
          <a-input v-model:value="permissionForm.resource" placeholder="Enter resource (e.g., users, organizations)" />
        </a-form-item>
        <a-form-item label="Action" name="action">
          <a-select v-model:value="permissionForm.action">
            <a-select-option value="create">Create</a-select-option>
            <a-select-option value="read">Read</a-select-option>
            <a-select-option value="update">Update</a-select-option>
            <a-select-option value="delete">Delete</a-select-option>
            <a-select-option value="manage">Manage</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="permissionForm.description" placeholder="Enter permission description" />
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
      :title="editingAssignment ? 'Edit Assignment' : 'Assign Role'"
      @ok="handleAssignmentSubmit"
      @cancel="handleAssignmentCancel"
    >
      <a-form
        ref="assignmentFormRef"
        :model="assignmentForm"
        :rules="assignmentRules"
        layout="vertical"
      >
        <a-form-item label="User" name="userId">
          <a-select
            v-model:value="assignmentForm.userId"
            placeholder="Select user"
            show-search
            :filter-option="filterUserOption"
          >
            <a-select-option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.display_name }} ({{ user.username }})
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Role" name="roleId">
          <a-select
            v-model:value="assignmentForm.roleId"
            placeholder="Select role"
            show-search
            :filter-option="filterRoleOption"
          >
            <a-select-option v-for="role in roles" :key="role.id" :value="role.id">
              {{ role.name }}
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
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'

// Data
const activeTab = ref('roles')
const loading = ref(false)
const roleModalVisible = ref(false)
const permissionModalVisible = ref(false)
const assignmentModalVisible = ref(false)
const editingRole = ref(false)
const editingPermission = ref(false)
const editingAssignment = ref(false)

// Mock data
const roles = ref([
  {
    id: '1',
    name: 'Super Admin',
    description: 'Full system access',
    status: 'active',
    permissions: ['users:manage', 'organizations:manage', 'applications:manage', 'system:manage'],
    created_at: '2024-01-01'
  },
  {
    id: '2',
    name: 'Organization Admin',
    description: 'Organization management access',
    status: 'active',
    permissions: ['users:read', 'organizations:manage', 'applications:read'],
    created_at: '2024-01-01'
  },
  {
    id: '3',
    name: 'User',
    description: 'Basic user access',
    status: 'active',
    permissions: ['profile:manage', 'applications:read'],
    created_at: '2024-01-01'
  }
])

const permissions = ref([
  {
    id: '1',
    name: 'Manage Users',
    resource: 'users',
    action: 'manage',
    description: 'Full user management permissions',
    status: 'active',
    created_at: '2024-01-01'
  },
  {
    id: '2',
    name: 'Read Users',
    resource: 'users',
    action: 'read',
    description: 'View user information',
    status: 'active',
    created_at: '2024-01-01'
  },
  {
    id: '3',
    name: 'Manage Organizations',
    resource: 'organizations',
    action: 'manage',
    description: 'Full organization management permissions',
    status: 'active',
    created_at: '2024-01-01'
  }
])

const assignments = ref([
  {
    id: '1',
    user: 'admin',
    userDisplayName: 'Administrator',
    role: 'Super Admin',
    status: 'active',
    assigned_at: '2024-01-01'
  },
  {
    id: '2',
    user: 'user1',
    userDisplayName: 'John Doe',
    role: 'User',
    status: 'active',
    assigned_at: '2024-01-01'
  }
])

const users = ref([
  { id: '1', username: 'admin', display_name: 'Administrator' },
  { id: '2', username: 'user1', display_name: 'John Doe' },
  { id: '3', username: 'user2', display_name: 'Jane Smith' }
])

// Table columns
const roleColumns = [
  { title: 'Role Name', dataIndex: 'name', key: 'name' },
  { title: 'Description', dataIndex: 'description', key: 'description' },
  { title: 'Permissions', dataIndex: 'permissions', key: 'permissions' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Created', dataIndex: 'created_at', key: 'created_at' },
  { title: 'Actions', key: 'actions' }
]

const permissionColumns = [
  { title: 'Permission Name', dataIndex: 'name', key: 'name' },
  { title: 'Resource', dataIndex: 'resource', key: 'resource' },
  { title: 'Action', dataIndex: 'action', key: 'action' },
  { title: 'Description', dataIndex: 'description', key: 'description' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Created', dataIndex: 'created_at', key: 'created_at' },
  { title: 'Actions', key: 'actions' }
]

const assignmentColumns = [
  { title: 'User', dataIndex: 'userDisplayName', key: 'userDisplayName' },
  { title: 'Username', dataIndex: 'user', key: 'user' },
  { title: 'Role', dataIndex: 'role', key: 'role' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Assigned', dataIndex: 'assigned_at', key: 'assigned_at' },
  { title: 'Actions', key: 'actions' }
]

// Pagination
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true
})

// Forms
const roleForm = reactive({
  name: '',
  description: '',
  status: 'active'
})

const permissionForm = reactive({
  name: '',
  resource: '',
  action: 'read',
  description: '',
  status: 'active'
})

const assignmentForm = reactive({
  userId: '',
  roleId: '',
  status: 'active'
})

// Form rules
const roleRules = {
  name: [{ required: true, message: 'Please enter role name' }],
  description: [{ required: true, message: 'Please enter role description' }],
  status: [{ required: true, message: 'Please select status' }]
}

const permissionRules = {
  name: [{ required: true, message: 'Please enter permission name' }],
  resource: [{ required: true, message: 'Please enter resource' }],
  action: [{ required: true, message: 'Please select action' }],
  description: [{ required: true, message: 'Please enter permission description' }],
  status: [{ required: true, message: 'Please select status' }]
}

const assignmentRules = {
  userId: [{ required: true, message: 'Please select user' }],
  roleId: [{ required: true, message: 'Please select role' }],
  status: [{ required: true, message: 'Please select status' }]
}

// Methods
const loadData = async () => {
  loading.value = true
  try {
    // TODO: Implement API calls
    // const response = await permissionApi.getRoles()
    // roles.value = response.data
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

// Role methods
const showRoleModal = () => {
  editingRole.value = false
  Object.assign(roleForm, {
    name: '',
    description: '',
    status: 'active'
  })
  roleModalVisible.value = true
}

const editRole = (role: any) => {
  editingRole.value = true
  Object.assign(roleForm, role)
  roleModalVisible.value = true
}

const handleRoleSubmit = async () => {
  try {
    // TODO: Implement API call
    message.success(editingRole.value ? 'Role updated successfully' : 'Role created successfully')
    roleModalVisible.value = false
    loadData()
  } catch (error) {
    message.error('Failed to save role')
  }
}

const handleRoleCancel = () => {
  roleModalVisible.value = false
}

const deleteRole = async (id: string) => {
  try {
    // TODO: Implement API call
    message.success('Role deleted successfully')
    loadData()
  } catch (error) {
    message.error('Failed to delete role')
  }
}

// Permission methods
const showPermissionModal = () => {
  editingPermission.value = false
  Object.assign(permissionForm, {
    name: '',
    resource: '',
    action: 'read',
    description: '',
    status: 'active'
  })
  permissionModalVisible.value = true
}

const editPermission = (permission: any) => {
  editingPermission.value = true
  Object.assign(permissionForm, permission)
  permissionModalVisible.value = true
}

const handlePermissionSubmit = async () => {
  try {
    // TODO: Implement API call
    message.success(editingPermission.value ? 'Permission updated successfully' : 'Permission created successfully')
    permissionModalVisible.value = false
    loadData()
  } catch (error) {
    message.error('Failed to save permission')
  }
}

const handlePermissionCancel = () => {
  permissionModalVisible.value = false
}

const deletePermission = async (id: string) => {
  try {
    // TODO: Implement API call
    message.success('Permission deleted successfully')
    loadData()
  } catch (error) {
    message.error('Failed to delete permission')
  }
}

// Assignment methods
const showAssignmentModal = () => {
  editingAssignment.value = false
  Object.assign(assignmentForm, {
    userId: '',
    roleId: '',
    status: 'active'
  })
  assignmentModalVisible.value = true
}

const editAssignment = (assignment: any) => {
  editingAssignment.value = true
  Object.assign(assignmentForm, assignment)
  assignmentModalVisible.value = true
}

const handleAssignmentSubmit = async () => {
  try {
    // TODO: Implement API call
    message.success(editingAssignment.value ? 'Assignment updated successfully' : 'Role assigned successfully')
    assignmentModalVisible.value = false
    loadData()
  } catch (error) {
    message.error('Failed to save assignment')
  }
}

const handleAssignmentCancel = () => {
  assignmentModalVisible.value = false
}

const removeAssignment = async (id: string) => {
  try {
    // TODO: Implement API call
    message.success('Assignment removed successfully')
    loadData()
  } catch (error) {
    message.error('Failed to remove assignment')
  }
}

// Filter methods
const filterUserOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

const filterRoleOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

// View role permissions
const viewRolePermissions = (role: any) => {
  message.info(`Viewing permissions for role: ${role.name}`)
  // TODO: Implement permission view modal
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.permissions-page {
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

.actions-bar {
  margin-bottom: 16px;
}

/* Responsive Design */
@media (max-width: 768px) {
  .permissions-page {
    padding: 16px;
  }
  
  .tab-content {
    padding: 16px;
  }
}
</style>
