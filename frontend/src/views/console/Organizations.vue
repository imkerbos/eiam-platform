<template>
  <div class="organizations-page">
    <a-card title="Organization Management" :bordered="false">
      <template #extra>
        <a-space>
          <a-radio-group v-model:value="viewMode" button-style="solid">
            <a-radio-button value="list">List View</a-radio-button>
            <a-radio-button value="tree">Tree View</a-radio-button>
          </a-radio-group>
          <a-button type="primary" @click="showAddOrgModal">
            <template #icon>
              <PlusOutlined />
            </template>
            Add Organization
          </a-button>
        </a-space>
      </template>

      <!-- List View -->
      <a-table
        v-if="viewMode === 'list'"
        :columns="columns"
        :data-source="organizations"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            <a-tag :color="getTypeColor(record.type)">
              {{ getTypeName(record.type) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? 'Active' : 'Inactive' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="editOrg(record)">
                Edit
              </a-button>
              <a-button type="link" size="small" @click="showAddUserToOrgModal(record)">
                Add User
              </a-button>
              <a-popconfirm
                title="Are you sure you want to delete this organization?"
                @confirm="deleteOrg(record.id)"
              >
                <a-button type="link" size="small" danger>
                  Delete
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>

      <!-- Tree View -->
      <div v-else-if="viewMode === 'tree'" class="tree-layout">
        <a-row :gutter="24">
          <!-- Left Panel: Organization Tree -->
          <a-col :span="12">
            <a-card title="Organization Tree" size="small" :bordered="false">
              <div class="tree-container">
                <div
                  v-for="org in flattenedTree"
                  :key="org.id"
                  class="tree-node"
                  :class="{ 'tree-node-selected': selectedOrgId === org.id }"
                  :style="{ paddingLeft: (org.level || 0) * 20 + 'px' }"
                  @click="selectOrganization(org)"
                >
                  <div class="tree-node-content">
                    <span v-if="org.children && org.children.length > 0" 
                          class="tree-expand-icon" 
                          @click.stop="toggleExpanded(org)">
                      {{ org.expanded ? '▼' : '▶' }}
                    </span>
                    <span v-else class="tree-expand-placeholder"></span>
                    
                    <span class="tree-node-name">{{ org.name }}</span>
                    
                    <a-tag :color="getTypeColor(org.type)" size="small">
                      {{ getTypeName(org.type) }}
                    </a-tag>
                    
                    <div class="tree-node-actions">
                      <a-button type="text" size="small" @click.stop="editOrg(org)">
                        <EditOutlined />
                      </a-button>
                      <a-button type="text" size="small" @click.stop="showAddUserToOrgModal(org)">
                        <UserAddOutlined />
                      </a-button>
                      <a-popconfirm
                        title="Are you sure you want to delete this organization?"
                        @confirm="deleteOrg(org.id)"
                        @click.stop
                      >
                        <a-button type="text" size="small" danger>
                          <DeleteOutlined />
                        </a-button>
                      </a-popconfirm>
                    </div>
                  </div>
                </div>
                
                <a-empty v-if="!loading && flattenedTree.length === 0" description="No organizations found" />
              </div>
            </a-card>
          </a-col>
          
          <!-- Right Panel: Users in Selected Organization -->
          <a-col :span="12">
            <a-card 
              :title="selectedOrgUsers.length > 0 ? `Users in ${selectedOrgName}` : 'Select an organization to view users'" 
              size="small" 
              :bordered="false"
            >
              <template v-if="selectedOrgId">
                <a-table
                  :columns="userColumns"
                  :data-source="selectedOrgUsers"
                  :loading="usersLoading"
                  :pagination="userPagination"
                  @change="handleUserTableChange"
                  row-key="id"
                  size="small"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'avatar'">
                      <UserAvatar :user="record" :size="32" />
                    </template>
                    <template v-else-if="column.key === 'status'">
                      <a-tag :color="record.status === 1 ? 'green' : 'red'">
                        {{ record.status === 1 ? 'Active' : 'Inactive' }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="editUser(record)">
                          Edit
                        </a-button>
                        <a-button type="link" size="small" danger @click="removeUserFromOrg(record.id)">
                          Remove
                        </a-button>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </template>
              <template v-else>
                <a-empty description="Select an organization from the left panel to view its users" />
              </template>
            </a-card>
          </a-col>
        </a-row>
      </div>
    </a-card>

    <!-- Add/Edit Organization Modal -->
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
        <a-form-item label="Code" name="code">
          <a-input v-model:value="formData.code" />
        </a-form-item>
        <a-form-item label="Type" name="type">
          <a-select v-model:value="formData.type">
            <a-select-option :value="1">Headquarters</a-select-option>
            <a-select-option :value="2">Branch</a-select-option>
            <a-select-option :value="3">Department</a-select-option>
            <a-select-option :value="4">Team</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Parent Organization" name="parent_id">
          <a-select v-model:value="formData.parent_id" placeholder="Select parent organization" allow-clear>
            <a-select-option v-for="org in organizations" :key="org.id" :value="org.id">
              {{ org.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Manager" name="manager">
          <a-select v-model:value="formData.manager" placeholder="Select manager" allow-clear>
            <a-select-option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.display_name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="formData.description" />
        </a-form-item>
        <a-form-item label="Location" name="location">
          <a-input v-model:value="formData.location" />
        </a-form-item>
        <a-form-item label="Phone" name="phone">
          <a-input v-model:value="formData.phone" />
        </a-form-item>
        <a-form-item label="Email" name="email">
          <a-input v-model:value="formData.email" />
        </a-form-item>
        <a-form-item label="Status" name="status">
          <a-select v-model:value="formData.status">
            <a-select-option :value="1">Active</a-select-option>
            <a-select-option :value="0">Inactive</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Add User to Organization Modal -->
    <a-modal
      v-model:open="addUserModalVisible"
      :title="`Add User to ${selectedOrg?.name}`"
      width="600px"
      @ok="handleAddUserToOrgOk"
      @cancel="handleAddUserToOrgCancel"
    >
      <a-form layout="vertical">
        <a-form-item label="Select User">
          <a-select
            v-model:value="selectedUserId"
            placeholder="Select a user to add to this organization"
            show-search
            :filter-option="filterUserOption"
            style="width: 100%"
          >
            <a-select-option
              v-for="user in availableUsers"
              :key="user.id"
              :value="user.id"
              :label="user.display_name"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ user.display_name }}</span>
                <span style="color: #999; font-size: 12px;">@{{ user.username }}</span>
              </div>
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item v-if="selectedUserId">
          <a-card size="small" style="background-color: #f9f9f9;">
            <template #title>Selected User</template>
            <div v-if="selectedUser">
              <p><strong>Name:</strong> {{ selectedUser.display_name }}</p>
              <p><strong>Username:</strong> {{ selectedUser.username }}</p>
              <p><strong>Email:</strong> {{ selectedUser.email }}</p>
              <p><strong>Current Organization:</strong> {{ selectedUser.organization?.name || 'None' }}</p>
            </div>
          </a-card>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { message } from 'ant-design-vue'
import { 
  PlusOutlined,
  EditOutlined,
  UserAddOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import type { Organization, User } from '@/types/api'
import { organizationApi, userApi } from '@/api/index'
import type { CreateOrganizationRequest, UpdateOrganizationRequest } from '@/api/organizations'
import UserAvatar from '@/components/UserAvatar.vue'

// Data
const loading = ref(false)
const viewMode = ref<'list' | 'tree'>('list')
const organizations = ref<Organization[]>([])
const organizationTree = ref<any[]>([])
const users = ref<User[]>([])
const modalVisible = ref(false)
const modalTitle = ref('Add Organization')
const formRef = ref()
const editingOrg = ref<Organization | null>(null)

// Add user to organization modal
const addUserModalVisible = ref(false)
const selectedOrg = ref<Organization | null>(null)
const selectedUserId = ref<string | undefined>(undefined)
const availableUsers = ref<User[]>([])
const allUsers = ref<User[]>([])

// Tree view specific data
const selectedOrgId = ref<string>('')
const selectedOrgName = ref<string>('')
const selectedOrgUsers = ref<User[]>([])
const usersLoading = ref(false)

// Computed properties
const selectedUser = computed(() => {
  return allUsers.value.find(user => user.id === selectedUserId.value)
})

const flattenedTree = computed(() => {
  const flattenOrgs = (orgs: any[], level = 0): any[] => {
    let result: any[] = []
    for (const org of orgs) {
      result.push({ ...org, level })
      if (org.expanded && org.children && org.children.length > 0) {
        result.push(...flattenOrgs(org.children, level + 1))
      }
    }
    return result
  }
  
  const sourceData = organizationTree.value.length > 0 ? organizationTree.value : organizations.value
  // 确保每个组织都有 expanded 属性
  const dataWithExpanded = sourceData.map(org => ({
    ...org,
    expanded: org.expanded || false
  }))
  
  return flattenOrgs(dataWithExpanded)
})

const formData = reactive({
  name: '',
  code: '',
  type: 2,
  parent_id: undefined as string | undefined,
  manager: undefined as string | undefined,
  description: '',
  location: '',
  phone: '',
  email: '',
  status: 1
})

const formRules = {
  name: [{ required: true, message: 'Please input organization name!' }],
  code: [{ required: true, message: 'Please input organization code!' }],
  type: [{ required: true, message: 'Please select organization type!' }],
  status: [{ required: true, message: 'Please select status!' }]
}

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  total_pages: 0
})

const userPagination = reactive({
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
    title: 'Code',
    dataIndex: 'code',
    key: 'code'
  },
  {
    title: 'Type',
    dataIndex: 'type',
    key: 'type'
  },
  {
    title: 'Manager',
    dataIndex: 'manager_name',
    key: 'manager_name',
    customRender: ({ record }: { record: Organization }) => {
      return record.manager_name || record.manager || '-'
    }
  },
  {
    title: 'Location',
    dataIndex: 'location',
    key: 'location'
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status'
  },
  {
    title: 'Created At',
    dataIndex: 'created_at',
    key: 'created_at',
    customRender: ({ record }: { record: Organization }) => {
      return new Date(record.created_at).toLocaleDateString()
    }
  },
  {
    title: 'Action',
    key: 'action'
  }
]

const userColumns = [
  {
    title: 'Avatar',
    key: 'avatar',
    width: 60
  },
  {
    title: 'Username',
    dataIndex: 'username',
    key: 'username'
  },
  {
    title: 'Display Name',
    dataIndex: 'display_name',
    key: 'display_name'
  },
  {
    title: 'Email',
    dataIndex: 'email',
    key: 'email'
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    width: 80
  },
  {
    title: 'Action',
    key: 'action',
    width: 120
  }
]

// Methods
const loadOrganizations = async () => {
  loading.value = true
  try {
    if (viewMode.value === 'tree') {
      try {
        const response = await organizationApi.getOrganizationsTree()
        organizationTree.value = convertToTreeData(response)
      } catch (treeError) {
        console.warn('Tree API failed, falling back to list API:', treeError)
        // Fallback to regular list if tree API fails
        const response = await organizationApi.getOrganizations({
          page: pagination.current,
          page_size: pagination.pageSize
        })
        // 为fallback数据也添加展开状态
        const itemsWithExpanded = response.items.map((org: any) => ({
          ...org,
          expanded: org.expanded || false
        }))
        organizations.value = itemsWithExpanded
        organizationTree.value = convertToTreeData(itemsWithExpanded)
        pagination.total = response.total
        pagination.total_pages = response.total_pages
      }
    } else {
      const response = await organizationApi.getOrganizations({
        page: pagination.current,
        page_size: pagination.pageSize
      })
      // 为普通列表数据也添加展开状态
      organizations.value = response.items.map((org: any) => ({
        ...org,
        expanded: org.expanded || false
      }))
      pagination.total = response.total
      pagination.total_pages = response.total_pages
    }
  } catch (error) {
    console.error('Failed to load organizations:', error)
    message.error('Failed to load organizations')
  } finally {
    loading.value = false
  }
}

const convertToTreeData = (organizations: any[]): any[] => {
  if (!organizations || !Array.isArray(organizations)) {
    return []
  }
  return organizations.map(org => ({
    ...org,
    expanded: org.expanded || false, // 确保有展开状态
    children: org.children ? convertToTreeData(org.children) : []
  }))
}



const loadUsers = async () => {
  try {
    const response = await userApi.getUsers({
      page: 1,
      page_size: 100
    })
    users.value = response.items
    allUsers.value = response.items
  } catch (error) {
    message.error('Failed to load users')
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadOrganizations()
}

const showAddOrgModal = () => {
  modalTitle.value = 'Add Organization'
  editingOrg.value = null
  resetForm()
  modalVisible.value = true
}

const editOrg = (org: Organization) => {
  modalTitle.value = 'Edit Organization'
  editingOrg.value = org
  Object.assign(formData, {
    name: org.name,
    code: org.code,
    type: org.type,
    parent_id: org.parent_id,
    manager: org.manager_id || org.manager, // 优先使用manager_id，向后兼容
    description: org.description,
    location: org.location,
    phone: org.phone,
    email: org.email,
    status: org.status
  })
  modalVisible.value = true
}

const resetForm = () => {
  Object.assign(formData, {
    name: '',
    code: '',
    type: 2,
    parent_id: undefined,
    manager: undefined,
    description: '',
    location: '',
    phone: '',
    email: '',
    status: 1
  })
  formRef.value?.resetFields()
}

const handleModalOk = async () => {
  try {
    await formRef.value?.validate()
    
    if (editingOrg.value) {
      // Update organization
      const updateData: UpdateOrganizationRequest = {
        name: formData.name,
        code: formData.code,
        type: formData.type,
        parent_id: formData.parent_id,
        manager: formData.manager,
        description: formData.description,
        location: formData.location,
        phone: formData.phone,
        email: formData.email,
        status: formData.status
      }
      await organizationApi.updateOrganization(editingOrg.value.id, updateData)
      message.success('Organization updated successfully')
    } else {
      // Create organization
      const createData: CreateOrganizationRequest = {
        name: formData.name,
        code: formData.code,
        type: formData.type,
        parent_id: formData.parent_id,
        manager: formData.manager,
        description: formData.description,
        location: formData.location,
        phone: formData.phone,
        email: formData.email,
        status: formData.status
      }
      await organizationApi.createOrganization(createData)
      message.success('Organization created successfully')
    }
    
    modalVisible.value = false
    loadOrganizations()
  } catch (error) {
    message.error('Please check the form')
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
  resetForm()
}

const deleteOrg = async (orgId: string) => {
  try {
    await organizationApi.deleteOrganization(orgId)
    message.success('Organization deleted successfully')
    loadOrganizations()
  } catch (error) {
    message.error('Failed to delete organization')
  }
}

// Add user to organization methods
const showAddUserToOrgModal = (org: Organization) => {
  selectedOrg.value = org
  selectedUserId.value = undefined
  // Filter out users who are already in this organization
  availableUsers.value = allUsers.value.filter(user => 
    user.organization_id !== org.id
  )
  addUserModalVisible.value = true
}

const handleAddUserToOrgOk = async () => {
  if (!selectedUserId.value || !selectedOrg.value) {
    message.error('Please select a user')
    return
  }

  try {
    // Update user's organization
    await userApi.updateUser(selectedUserId.value, {
      organization_id: selectedOrg.value.id
    })
    
    message.success(`User added to ${selectedOrg.value.name} successfully`)
    addUserModalVisible.value = false
    
    // Refresh user list to reflect changes
    await loadUsers()
  } catch (error) {
    message.error('Failed to add user to organization')
  }
}

const handleAddUserToOrgCancel = () => {
  addUserModalVisible.value = false
  selectedOrg.value = null
  selectedUserId.value = undefined
  availableUsers.value = []
}

const filterUserOption = (input: string, option: any) => {
  const user = allUsers.value.find(u => u.id === option.value)
  if (!user) return false
  
  const searchText = input.toLowerCase()
  return (
    user.display_name.toLowerCase().includes(searchText) ||
    user.username.toLowerCase().includes(searchText) ||
    user.email.toLowerCase().includes(searchText)
  )
}

const getTypeColor = (type: number) => {
  switch (type) {
    case 1: return 'blue'
    case 2: return 'green'
    case 3: return 'orange'
    case 4: return 'purple'
    default: return 'default'
  }
}

const getTypeName = (type: number) => {
  switch (type) {
    case 1: return 'Headquarters'
    case 2: return 'Branch'
    case 3: return 'Department'
    case 4: return 'Team'
    default: return 'Unknown'
  }
}



// Tree expand/collapse functionality
const toggleExpanded = (record: any) => {
  if (record.children && record.children.length > 0) {
    // Find and update the item in the source data
    const updateExpandedState = (items: any[], targetId: string, newState: boolean): boolean => {
      for (const item of items) {
        if (item.id === targetId) {
          item.expanded = newState
          return true
        }
        if (item.children && item.children.length > 0) {
          if (updateExpandedState(item.children, targetId, newState)) {
            return true
          }
        }
      }
      return false
    }
    
    const newExpandedState = !record.expanded
    
    // Update in the source data
    if (organizationTree.value.length > 0) {
      updateExpandedState(organizationTree.value, record.id, newExpandedState)
      // Force reactivity update
      organizationTree.value = [...organizationTree.value]
    } else {
      updateExpandedState(organizations.value, record.id, newExpandedState)
      // Force reactivity update
      organizations.value = [...organizations.value]
    }
  }
}

// Tree view specific functions
const selectOrganization = (org: Organization) => {
  selectedOrgId.value = org.id
  selectedOrgName.value = org.name
  loadOrganizationUsers(org.id)
}

const loadOrganizationUsers = async (orgId: string) => {
  usersLoading.value = true
  try {
    const response = await userApi.getUsers({
      organization_id: orgId,
      page: userPagination.current,
      page_size: userPagination.pageSize
    })
    selectedOrgUsers.value = response.items
    userPagination.total = response.total
    userPagination.total_pages = response.total_pages
  } catch (error) {
    console.error('Failed to load organization users:', error)
    message.error('Failed to load organization users')
  } finally {
    usersLoading.value = false
  }
}

const handleUserTableChange = (pag: any) => {
  userPagination.current = pag.current
  userPagination.pageSize = pag.pageSize
  if (selectedOrgId.value) {
    loadOrganizationUsers(selectedOrgId.value)
  }
}

const editUser = (_user: User) => {
  // Navigate to user edit page or open user edit modal
  // This functionality can be implemented later
  message.info('User edit functionality will be implemented')
}

const removeUserFromOrg = async (_userId: string) => {
  try {
    // This API endpoint might need to be implemented in the backend
    // For now, we'll just show a message
    message.info('Remove user from organization functionality will be implemented')
    // After successful removal, reload the user list
    // if (selectedOrgId.value) {
    //   loadOrganizationUsers(selectedOrgId.value)
    // }
  } catch (error) {
    console.error('Failed to remove user from organization:', error)
    message.error('Failed to remove user from organization')
  }
}

// Lifecycle
onMounted(() => {
  loadOrganizations()
  loadUsers()
})

// Define TreeTableRow as a component for use in template

// Watch view mode changes
watch(viewMode, () => {
  loadOrganizations()
})
</script>

<style scoped>
.organizations-page {
  padding: 24px;
}

.tree-layout {
  margin-top: 16px;
}

.tree-container {
  max-height: 600px;
  overflow-y: auto;
}

.tree-node {
  padding: 8px 16px;
  margin: 2px 0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.tree-node:hover {
  background-color: #f5f5f5;
}

.tree-node-selected {
  background-color: #e6f7ff;
  border-color: #1890ff;
}

.tree-node-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tree-expand-icon {
  width: 16px;
  text-align: center;
  cursor: pointer;
  user-select: none;
  color: #666;
  font-size: 12px;
}

.tree-expand-icon:hover {
  color: #1890ff;
}

.tree-expand-placeholder {
  width: 16px;
}

.tree-node-name {
  flex: 1;
  font-weight: 500;
  color: #262626;
}

.tree-node-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.tree-node:hover .tree-node-actions {
  opacity: 1;
}

.tree-node-actions .ant-btn {
  width: 24px;
  height: 24px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Compact style for tree layout */
.tree-layout .ant-card-body {
  padding: 16px;
}

.tree-layout .ant-table-tbody > tr > td {
  padding: 8px 12px;
}

.tree-layout .ant-table-small {
  font-size: 13px;
}

.tree-layout .ant-tag {
  margin: 0;
  font-size: 11px;
  padding: 2px 6px;
  line-height: 16px;
}
</style>
