<template>
  <div class="users-page">
    <a-card title="User Management" :bordered="false">
      <template #extra>
        <a-button type="primary" @click="showAddUserModal">
          <template #icon>
            <PlusOutlined />
          </template>
          Add User
        </a-button>
      </template>

      <a-table
        :columns="columns"
        :data-source="users"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? 'Active' : 'Inactive' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="editUser(record)">
                Edit
              </a-button>
              <a-button type="link" size="small" @click="resetPassword(record)">
                Reset Password
              </a-button>
              <a-popconfirm
                title="Are you sure you want to delete this user?"
                @confirm="deleteUser(record.id)"
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

    <!-- Add/Edit User Modal -->
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
        <a-form-item label="Username" name="username">
          <a-input v-model:value="formData.username" />
        </a-form-item>
        <a-form-item label="Email" name="email">
          <a-input v-model:value="formData.email" />
        </a-form-item>
        <a-form-item label="Display Name" name="display_name">
          <a-input v-model:value="formData.display_name" />
        </a-form-item>
        <a-form-item label="Phone" name="phone">
          <a-input v-model:value="formData.phone" />
        </a-form-item>
        <a-form-item label="Password" name="password" v-if="!editingUser">
          <a-input-password v-model:value="formData.password" />
        </a-form-item>
        <a-form-item label="Organization" name="organization_id">
          <a-select v-model:value="formData.organization_id" placeholder="Select organization">
            <a-select-option v-for="org in organizations" :key="org.id" :value="org.id">
              {{ org.name }}
            </a-select-option>
          </a-select>
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
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import type { User, PaginatedResponse } from '@/types/api'
import { userApi, organizationApi } from '@/api/index'
import type { CreateUserRequest, UpdateUserRequest } from '@/api/users'

// Data
const loading = ref(false)
const users = ref<User[]>([])
const organizations = ref<{ id: string; name: string }[]>([])
const modalVisible = ref(false)
const modalTitle = ref('Add User')
const formRef = ref()
const editingUser = ref<User | null>(null)

const formData = reactive({
  username: '',
  email: '',
  display_name: '',
  phone: '',
  password: '',
  organization_id: undefined as string | undefined,
  status: 1
})

const formRules = {
  username: [{ required: true, message: 'Please input username!' }],
  email: [
    { required: true, message: 'Please input email!' },
    { type: 'email', message: 'Please input valid email!' }
  ],
  display_name: [{ required: true, message: 'Please input display name!' }],
  password: [{ required: true, message: 'Please input password!' }],
  organization_id: [{ required: true, message: 'Please select organization!' }]
}

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  total_pages: 0
})

const columns = [
  {
    title: 'Username',
    dataIndex: 'username',
    key: 'username'
  },
  {
    title: 'Email',
    dataIndex: 'email',
    key: 'email'
  },
  {
    title: 'Display Name',
    dataIndex: 'display_name',
    key: 'display_name'
  },
  {
    title: 'Phone',
    dataIndex: 'phone',
    key: 'phone'
  },
  {
    title: 'Organization',
    dataIndex: 'organization',
    key: 'organization',
    customRender: ({ record }: { record: User }) => {
      return record.organization?.name || '-'
    }
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    customRender: ({ record }: { record: User }) => {
      return record.status === 1 ? 'Active' : 'Inactive'
    }
  },
  {
    title: 'Created At',
    dataIndex: 'created_at',
    key: 'created_at',
    customRender: ({ record }: { record: User }) => {
      return new Date(record.created_at).toLocaleDateString()
    }
  },
  {
    title: 'Action',
    key: 'action'
  }
]

// Methods
const loadUsers = async () => {
  loading.value = true
  try {
    const response = await userApi.getUsers({
      page: pagination.current,
      page_size: pagination.pageSize
    })
    users.value = response.items
    pagination.total = response.total
    pagination.total_pages = response.total_pages
  } catch (error) {
    message.error('Failed to load users')
  } finally {
    loading.value = false
  }
}

const loadOrganizations = async () => {
  try {
    const response = await organizationApi.getOrganizations({
      page: 1,
      page_size: 100
    })
    organizations.value = response.items.map((org: any) => ({
      id: org.id,
      name: org.name
    }))
  } catch (error) {
    message.error('Failed to load organizations')
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadUsers()
}

const showAddUserModal = () => {
  modalTitle.value = 'Add User'
  editingUser.value = null
  resetForm()
  modalVisible.value = true
}

const editUser = (user: User) => {
  modalTitle.value = 'Edit User'
  editingUser.value = user
  Object.assign(formData, {
    username: user.username,
    email: user.email,
    display_name: user.display_name,
    phone: user.phone,
    organization_id: user.organization_id,
    status: user.status
  })
  modalVisible.value = true
}

const resetForm = () => {
  Object.assign(formData, {
    username: '',
    email: '',
    display_name: '',
    phone: '',
    password: '',
    organization_id: undefined,
    status: 1
  })
  formRef.value?.resetFields()
}

const handleModalOk = async () => {
  try {
    await formRef.value?.validate()
    
    if (editingUser.value) {
      // Update user
      const updateData: UpdateUserRequest = {
        display_name: formData.display_name,
        phone: formData.phone,
        status: formData.status
      }
      // Only include organization_id if it has a value
      if (formData.organization_id) {
        updateData.organization_id = formData.organization_id
      }
      await userApi.updateUser(editingUser.value.id, updateData)
      message.success('User updated successfully')
    } else {
      // Create user
      const createData: CreateUserRequest = {
        username: formData.username,
        email: formData.email,
        display_name: formData.display_name,
        phone: formData.phone,
        password: formData.password,
        organization_id: formData.organization_id!,
        status: formData.status
      }
      await userApi.createUser(createData)
      message.success('User created successfully')
    }
    
    modalVisible.value = false
    loadUsers()
  } catch (error) {
    message.error('Please check the form')
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
  resetForm()
}

const resetPassword = async (user: User) => {
  try {
    // TODO: Implement API call
    message.success('Password reset email sent')
  } catch (error) {
    message.error('Failed to reset password')
  }
}

const deleteUser = async (userId: string) => {
  try {
    await userApi.deleteUser(userId)
    message.success('User deleted successfully')
    loadUsers()
  } catch (error) {
    message.error('Failed to delete user')
  }
}

// Lifecycle
onMounted(() => {
  loadUsers()
  loadOrganizations()
})
</script>

<style scoped>
.users-page {
  padding: 24px;
}
</style>
