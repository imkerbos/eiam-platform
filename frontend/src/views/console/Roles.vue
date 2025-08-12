<template>
  <div class="roles-page">
    <a-card title="Role Management" :bordered="false">
      <template #extra>
        <a-button type="primary" @click="showAddRoleModal">
          <template #icon>
            <PlusOutlined />
          </template>
          Add Role
        </a-button>
      </template>

      <a-table
        :columns="columns"
        :data-source="roles"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            <a-tag :color="record.type === 'system' ? 'red' : 'blue'">
              {{ record.type }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'scope'">
            <a-tag :color="getScopeColor(record.scope)">
              {{ record.scope }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="editRole(record)">
                Edit
              </a-button>
              <a-button type="link" size="small" @click="managePermissions(record)">
                Permissions
              </a-button>
              <a-popconfirm
                v-if="record.type !== 'system'"
                title="Are you sure you want to delete this role?"
                @confirm="deleteRole(record.id)"
              >
                <a-button type="link" size="small" danger>
                  Delete
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- Add/Edit Role Modal -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        layout="vertical"
      >
        <a-form-item label="Name" name="name">
          <a-input v-model:value="formData.name" />
        </a-form-item>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="formData.description" :rows="3" />
        </a-form-item>
        <a-form-item label="Type" name="type">
          <a-select v-model:value="formData.type">
            <a-select-option value="custom">Custom</a-select-option>
            <a-select-option value="system">System</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Scope" name="scope">
          <a-select v-model:value="formData.scope">
            <a-select-option value="global">Global</a-select-option>
            <a-select-option value="organization">Organization</a-select-option>
            <a-select-option value="application">Application</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item v-if="formData.scope === 'organization'" label="Organization" name="scopeId">
          <a-select v-model:value="formData.scopeId" placeholder="Select organization">
            <a-select-option v-for="org in organizations" :key="org.id" :value="org.id">
              {{ org.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item v-if="formData.scope === 'application'" label="Application" name="scopeId">
          <a-select v-model:value="formData.scopeId" placeholder="Select application">
            <a-select-option v-for="app in applications" :key="app.id" :value="app.id">
              {{ app.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Permissions Modal -->
    <a-modal
      v-model:open="permissionsModalVisible"
      title="Manage Permissions"
      width="800px"
      @ok="handlePermissionsOk"
      @cancel="handlePermissionsCancel"
    >
      <a-tree
        v-model:checkedKeys="checkedPermissions"
        :tree-data="permissionsTree"
        checkable
        :default-expand-all="true"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import type { Role, Pagination } from '@/types/api'

// Data
const loading = ref(false)
const roles = ref<Role[]>([])
const organizations = ref([])
const applications = ref([])
const permissionsTree = ref([])
const modalVisible = ref(false)
const permissionsModalVisible = ref(false)
const modalTitle = ref('Add Role')
const formRef = ref()
const editingRole = ref<Role | null>(null)
const currentRole = ref<Role | null>(null)
const checkedPermissions = ref<string[]>([])

const formData = reactive({
  name: '',
  description: '',
  type: 'custom',
  scope: 'global',
  scopeId: undefined
})

const formRules = {
  name: [{ required: true, message: 'Please input role name!' }],
  type: [{ required: true, message: 'Please select role type!' }],
  scope: [{ required: true, message: 'Please select scope!' }]
}

const pagination = reactive<Pagination>({
  current: 1,
  pageSize: 10,
  total: 0,
  total_pages: 0
})

const columns = [
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: 'Description',
    dataIndex: 'description',
    key: 'description'
  },
  {
    title: 'Type',
    dataIndex: 'type',
    key: 'type'
  },
  {
    title: 'Scope',
    dataIndex: 'scope',
    key: 'scope'
  },
  {
    title: 'Scope Name',
    dataIndex: 'scopeName',
    key: 'scopeName'
  },
  {
    title: 'Created At',
    dataIndex: 'createdAt',
    key: 'createdAt'
  },
  {
    title: 'Action',
    key: 'action'
  }
]

// Methods
const loadRoles = async () => {
  loading.value = true
  try {
    // Mock data for now
    roles.value = [
      {
        id: '1',
        name: 'Administrator',
        description: 'Full system administrator',
        type: 'system',
        scope: 'global',
        scopeName: '-',
        createdAt: '2024-01-01T00:00:00Z'
      },
      {
        id: '2',
        name: 'User Manager',
        description: 'Can manage users in organization',
        type: 'custom',
        scope: 'organization',
        scopeName: 'Headquarters',
        createdAt: '2024-01-01T00:00:00Z'
      }
    ]
    pagination.total = 2
    pagination.total_pages = 1
  } catch (error) {
    message.error('Failed to load roles')
  } finally {
    loading.value = false
  }
}

const loadOrganizations = async () => {
  try {
    // Mock data for now
    organizations.value = [
      { id: '1', name: 'Headquarters' },
      { id: '2', name: 'Branch Office' }
    ]
  } catch (error) {
    message.error('Failed to load organizations')
  }
}

const loadApplications = async () => {
  try {
    // Mock data for now
    applications.value = [
      { id: '1', name: 'HR System' },
      { id: '2', name: 'CRM System' }
    ]
  } catch (error) {
    message.error('Failed to load applications')
  }
}

const loadPermissions = async () => {
  try {
    // Mock permissions tree data
    permissionsTree.value = [
      {
        title: 'User Management',
        key: 'user',
        children: [
          { title: 'View Users', key: 'user:read' },
          { title: 'Create Users', key: 'user:create' },
          { title: 'Update Users', key: 'user:update' },
          { title: 'Delete Users', key: 'user:delete' }
        ]
      },
      {
        title: 'Organization Management',
        key: 'organization',
        children: [
          { title: 'View Organizations', key: 'organization:read' },
          { title: 'Create Organizations', key: 'organization:create' },
          { title: 'Update Organizations', key: 'organization:update' },
          { title: 'Delete Organizations', key: 'organization:delete' }
        ]
      },
      {
        title: 'Role Management',
        key: 'role',
        children: [
          { title: 'View Roles', key: 'role:read' },
          { title: 'Create Roles', key: 'role:create' },
          { title: 'Update Roles', key: 'role:update' },
          { title: 'Delete Roles', key: 'role:delete' }
        ]
      }
    ]
  } catch (error) {
    message.error('Failed to load permissions')
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadRoles()
}

const getScopeColor = (scope: string) => {
  const colors = {
    global: 'blue',
    organization: 'green',
    application: 'orange'
  }
  return colors[scope as keyof typeof colors] || 'default'
}

const showAddRoleModal = () => {
  modalTitle.value = 'Add Role'
  editingRole.value = null
  resetForm()
  modalVisible.value = true
}

const editRole = (role: Role) => {
  modalTitle.value = 'Edit Role'
  editingRole.value = role
  Object.assign(formData, {
    name: role.name,
    description: role.description,
    type: role.type,
    scope: role.scope,
    scopeId: role.scopeId
  })
  modalVisible.value = true
}

const resetForm = () => {
  Object.assign(formData, {
    name: '',
    description: '',
    type: 'custom',
    scope: 'global',
    scopeId: undefined
  })
  formRef.value?.resetFields()
}

const handleModalOk = async () => {
  try {
    await formRef.value?.validate()
    // TODO: Implement API call
    message.success(editingRole.value ? 'Role updated successfully' : 'Role created successfully')
    modalVisible.value = false
    loadRoles()
  } catch (error) {
    message.error('Please check the form')
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
  resetForm()
}

const managePermissions = (role: Role) => {
  currentRole.value = role
  // TODO: Load current role permissions
  checkedPermissions.value = ['user:read', 'organization:read']
  permissionsModalVisible.value = true
}

const handlePermissionsOk = async () => {
  try {
    // TODO: Implement API call to update role permissions
    message.success('Permissions updated successfully')
    permissionsModalVisible.value = false
  } catch (error) {
    message.error('Failed to update permissions')
  }
}

const handlePermissionsCancel = () => {
  permissionsModalVisible.value = false
  currentRole.value = null
  checkedPermissions.value = []
}

const deleteRole = async (roleId: string) => {
  try {
    // TODO: Implement API call
    message.success('Role deleted successfully')
    loadRoles()
  } catch (error) {
    message.error('Failed to delete role')
  }
}

// Watchers
watch(() => formData.scope, () => {
  formData.scopeId = undefined
})

// Lifecycle
onMounted(() => {
  loadRoles()
  loadOrganizations()
  loadApplications()
  loadPermissions()
})
</script>

<style scoped>
.roles-page {
  padding: 24px;
}
</style>
