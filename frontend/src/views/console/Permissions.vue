<template>
  <div class="permissions-page">
    <div class="page-header">
      <h2>Permissions Management</h2>
      <p>Manage system permissions and access controls</p>
    </div>

    <!-- Permission Categories -->
    <a-tabs v-model:activeKey="activeTab" type="card" @change="handleTabChange">
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
        <a-form-item label="Role Code" name="code">
          <a-input v-model:value="roleForm.code" placeholder="Enter role code" />
        </a-form-item>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="roleForm.description" placeholder="Enter role description" />
        </a-form-item>
        <a-form-item label="Type" name="type">
          <a-select v-model:value="roleForm.type">
            <a-select-option value="system">System</a-select-option>
            <a-select-option value="custom">Custom</a-select-option>
            <a-select-option value="application">Application</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Scope" name="scope">
          <a-select v-model:value="roleForm.scope">
            <a-select-option value="global">Global</a-select-option>
            <a-select-option value="organization">Organization</a-select-option>
            <a-select-option value="application">Application</a-select-option>
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
        <a-form-item label="Permission Code" name="code">
          <a-input v-model:value="permissionForm.code" placeholder="Enter permission code" />
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
            <a-select-option value="execute">Execute</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Category" name="category">
          <a-select v-model:value="permissionForm.category">
            <a-select-option value="system">System</a-select-option>
            <a-select-option value="application">Application</a-select-option>
            <a-select-option value="data">Data</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="permissionForm.description" placeholder="Enter permission description" />
        </a-form-item>
        <a-form-item label="Is System Permission" name="is_system">
          <a-switch v-model:checked="permissionForm.is_system" />
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
        <a-form-item label="User" name="user_id">
          <a-select
            v-model:value="assignmentForm.user_id"
            placeholder="Select user"
            show-search
            :filter-option="filterUserOption"
          >
            <a-select-option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.display_name }} ({{ user.username }})
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Role" name="role_id">
          <a-select
            v-model:value="assignmentForm.role_id"
            placeholder="Select role"
            show-search
            :filter-option="filterRoleOption"
          >
            <a-select-option v-for="role in roles" :key="role.id" :value="role.id">
              {{ role.name }} ({{ role.code }})
            </a-select-option>
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
import { permissionApi, type Permission, type Role, type RoleAssignment, type CreatePermissionRequest, type CreateRoleRequest, type AssignRoleRequest } from '@/api/permissions'
import { userApi } from '@/api/users'
import type { User } from '@/types/api'

// Data
const activeTab = ref('roles')
const loading = ref(false)
const roleModalVisible = ref(false)
const permissionModalVisible = ref(false)
const assignmentModalVisible = ref(false)
const editingRole = ref(false)
const editingPermission = ref(false)
const editingAssignment = ref(false)

// Real data
const roles = ref<Role[]>([])
const permissions = ref<Permission[]>([])
const assignments = ref<RoleAssignment[]>([])
const users = ref<User[]>([])

// Table columns
const roleColumns = [
  { title: 'Role Name', dataIndex: 'name', key: 'name' },
  { title: 'Code', dataIndex: 'code', key: 'code' },
  { title: 'Description', dataIndex: 'description', key: 'description' },
  { title: 'Type', dataIndex: 'type', key: 'type' },
  { title: 'Scope', dataIndex: 'scope', key: 'scope' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Created', dataIndex: 'created_at', key: 'created_at' },
  { title: 'Actions', key: 'actions' }
]

const permissionColumns = [
  { title: 'Permission Name', dataIndex: 'name', key: 'name' },
  { title: 'Code', dataIndex: 'code', key: 'code' },
  { title: 'Resource', dataIndex: 'resource', key: 'resource' },
  { title: 'Action', dataIndex: 'action', key: 'action' },
  { title: 'Category', dataIndex: 'category', key: 'category' },
  { title: 'Description', dataIndex: 'description', key: 'description' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Created', dataIndex: 'created_at', key: 'created_at' },
  { title: 'Actions', key: 'actions' }
]

const assignmentColumns = [
  { title: 'User', dataIndex: 'display_name', key: 'display_name' },
  { title: 'Username', dataIndex: 'username', key: 'username' },
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Role', dataIndex: 'role_name', key: 'role_name' },
  { title: 'Role Code', dataIndex: 'role_code', key: 'role_code' },
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
const roleForm = reactive<CreateRoleRequest & { status?: string }>({
  name: '',
  code: '',
  description: '',
  type: 'custom',
  scope: 'global',
  status: 'active'
})

const permissionForm = reactive<CreatePermissionRequest & { status?: string }>({
  name: '',
  code: '',
  resource: '',
  action: 'read',
  description: '',
  category: 'system',
  is_system: false,
  status: 'active'
})

const assignmentForm = reactive<AssignRoleRequest & { status?: string }>({
  user_id: '',
  role_id: '',
  status: 'active'
})

// Form rules
const roleRules = {
  name: [{ required: true, message: 'Please enter role name' }],
  code: [{ required: true, message: 'Please enter role code' }],
  description: [{ required: true, message: 'Please enter role description' }],
  type: [{ required: true, message: 'Please select role type' }],
  scope: [{ required: true, message: 'Please select role scope' }]
}

const permissionRules = {
  name: [{ required: true, message: 'Please enter permission name' }],
  code: [{ required: true, message: 'Please enter permission code' }],
  resource: [{ required: true, message: 'Please enter resource' }],
  action: [{ required: true, message: 'Please select action' }],
  category: [{ required: true, message: 'Please select category' }]
}

const assignmentRules = {
  user_id: [{ required: true, message: 'Please select user' }],
  role_id: [{ required: true, message: 'Please select role' }]
}

// Methods
const loadData = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'roles') {
      await loadRoles()
    } else if (activeTab.value === 'permissions') {
      await loadPermissions()
    } else if (activeTab.value === 'assignments') {
      await loadAssignments()
    }
  } catch (error) {
    message.error('Failed to load data')
  } finally {
    loading.value = false
  }
}

const loadRoles = async () => {
  try {
    const response = await permissionApi.getRoles({
      page: pagination.current,
      page_size: pagination.pageSize
    })
    roles.value = response.items
    pagination.total = response.total
  } catch (error) {
    message.error('Failed to load roles')
  }
}

const loadPermissions = async () => {
  try {
    const response = await permissionApi.getPermissions({
      page: pagination.current,
      page_size: pagination.pageSize
    })
    permissions.value = response.items
    pagination.total = response.total
  } catch (error) {
    message.error('Failed to load permissions')
  }
}

const loadAssignments = async () => {
  try {
    const response = await permissionApi.getRoleAssignments({
      page: pagination.current,
      page_size: pagination.pageSize
    })
    assignments.value = response.items
    pagination.total = response.total
  } catch (error) {
    message.error('Failed to load role assignments')
  }
}

const loadUsers = async () => {
  try {
    const response = await userApi.getUsers({ page: 1, page_size: 1000 })
    users.value = response.items
  } catch (error) {
    message.error('Failed to load users')
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
    code: '',
    description: '',
    type: 'custom',
    scope: 'global',
    status: 'active'
  })
  roleModalVisible.value = true
}

const editRole = (role: Role) => {
  editingRole.value = true
  Object.assign(roleForm, {
    name: role.name,
    code: role.code,
    description: role.description,
    type: role.type,
    scope: role.scope,
    status: role.status
  })
  roleModalVisible.value = true
}

const handleRoleSubmit = async () => {
  try {
    if (editingRole.value) {
      // Find the role ID from the current roles list
      const currentRole = roles.value.find(r => r.code === roleForm.code)
      if (currentRole) {
        await permissionApi.updateRole(currentRole.id, roleForm)
        message.success('Role updated successfully')
      }
    } else {
      await permissionApi.createRole(roleForm)
      message.success('Role created successfully')
    }
    roleModalVisible.value = false
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Failed to save role')
  }
}

const handleRoleCancel = () => {
  roleModalVisible.value = false
}

const deleteRole = async (id: string) => {
  try {
    await permissionApi.deleteRole(id)
    message.success('Role deleted successfully')
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Failed to delete role')
  }
}

// Permission methods
const showPermissionModal = () => {
  editingPermission.value = false
  Object.assign(permissionForm, {
    name: '',
    code: '',
    resource: '',
    action: 'read',
    description: '',
    category: 'system',
    is_system: false,
    status: 'active'
  })
  permissionModalVisible.value = true
}

const editPermission = (permission: Permission) => {
  editingPermission.value = true
  Object.assign(permissionForm, {
    name: permission.name,
    code: permission.code,
    resource: permission.resource,
    action: permission.action,
    description: permission.description,
    category: permission.category,
    is_system: permission.is_system,
    status: permission.status
  })
  permissionModalVisible.value = true
}

const handlePermissionSubmit = async () => {
  try {
    if (editingPermission.value) {
      // Find the permission ID from the current permissions list
      const currentPermission = permissions.value.find(p => p.code === permissionForm.code)
      if (currentPermission) {
        await permissionApi.updatePermission(currentPermission.id, permissionForm)
        message.success('Permission updated successfully')
      }
    } else {
      await permissionApi.createPermission(permissionForm)
      message.success('Permission created successfully')
    }
    permissionModalVisible.value = false
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Failed to save permission')
  }
}

const handlePermissionCancel = () => {
  permissionModalVisible.value = false
}

const deletePermission = async (id: string) => {
  try {
    await permissionApi.deletePermission(id)
    message.success('Permission deleted successfully')
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Failed to delete permission')
  }
}

// Assignment methods
const showAssignmentModal = () => {
  editingAssignment.value = false
  Object.assign(assignmentForm, {
    user_id: '',
    role_id: '',
    status: 'active'
  })
  assignmentModalVisible.value = true
  loadUsers() // Load users when opening the modal
}

const editAssignment = (assignment: RoleAssignment) => {
  editingAssignment.value = true
  Object.assign(assignmentForm, {
    user_id: assignment.user_id,
    role_id: assignment.role_id,
    status: assignment.status
  })
  assignmentModalVisible.value = true
}

const handleAssignmentSubmit = async () => {
  try {
    await permissionApi.assignRole(assignmentForm)
    message.success('Role assigned successfully')
    assignmentModalVisible.value = false
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Failed to assign role')
  }
}

const handleAssignmentCancel = () => {
  assignmentModalVisible.value = false
}

const removeAssignment = async (assignment: RoleAssignment) => {
  try {
    await permissionApi.removeRole(assignment.user_id, assignment.role_id)
    message.success('Role assignment removed successfully')
    loadData()
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Failed to remove role assignment')
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
const viewRolePermissions = (role: Role) => {
  message.info(`Viewing permissions for role: ${role.name}`)
  // TODO: Implement permission view modal
}

// Watch for tab changes
const handleTabChange = (key: string) => {
  activeTab.value = key
  pagination.current = 1
  loadData()
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
