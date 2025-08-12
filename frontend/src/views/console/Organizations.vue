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
      <div v-else-if="viewMode === 'tree'" class="tree-view">
        <a-tree
          :tree-data="organizationTree"
          :loading="loading"
          :default-expand-all="true"
          @select="handleTreeSelect"
        >
          <template #title="{ title, record }">
            <span>{{ title }}</span>
            <a-space style="margin-left: 8px">
              <a-tag :color="getTypeColor(record.type)" size="small">
                {{ getTypeName(record.type) }}
              </a-tag>
              <a-tag :color="record.status === 1 ? 'green' : 'red'" size="small">
                {{ record.status === 1 ? 'Active' : 'Inactive' }}
              </a-tag>
              <a-button type="link" size="small" @click.stop="editOrg(record)">
                Edit
              </a-button>
              <a-popconfirm
                title="Are you sure you want to delete this organization?"
                @confirm="deleteOrg(record.id)"
              >
                <a-button type="link" size="small" danger @click.stop>
                  Delete
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </a-tree>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import type { Organization, User } from '@/types/api'
import { organizationApi, userApi } from '@/api/index'
import type { CreateOrganizationRequest, UpdateOrganizationRequest } from '@/api/organizations'

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
    dataIndex: 'manager',
    key: 'manager'
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

// Methods
const loadOrganizations = async () => {
  loading.value = true
  try {
    if (viewMode.value === 'tree') {
      const response = await organizationApi.getOrganizationsTree()
      organizationTree.value = convertToTreeData(response)
    } else {
      const response = await organizationApi.getOrganizations({
        page: pagination.current,
        page_size: pagination.pageSize
      })
      organizations.value = response.items
      pagination.total = response.total
      pagination.total_pages = response.total_pages
    }
  } catch (error) {
    message.error('Failed to load organizations')
  } finally {
    loading.value = false
  }
}

const convertToTreeData = (organizations: any[]): any[] => {
  return organizations.map(org => ({
    key: org.id,
    title: org.name,
    record: org,
    children: org.children ? convertToTreeData(org.children) : []
  }))
}

const handleTreeSelect = (selectedKeys: string[], info: any) => {
  // Handle tree selection if needed
  console.log('Selected:', selectedKeys, info)
}

const loadUsers = async () => {
  try {
    const response = await userApi.getUsers({
      page: 1,
      page_size: 100
    })
    users.value = response.items
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
    manager: org.manager,
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

// Lifecycle
onMounted(() => {
  loadOrganizations()
  loadUsers()
})

// Watch view mode changes
watch(viewMode, () => {
  loadOrganizations()
})
</script>

<style scoped>
.organizations-page {
  padding: 24px;
}
</style>
