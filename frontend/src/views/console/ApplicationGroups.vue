<template>
  <div class="application-groups-page">
    <a-card title="Application Groups" :bordered="false">
      <template #extra>
        <a-button type="primary" @click="showAddGroupModal">
          <PlusOutlined />
          Add Group
        </a-button>
      </template>

      <a-table
        :columns="columns"
        :data-source="groups"
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
          <template v-else-if="column.key === 'color'">
            <div class="color-preview" :style="{ backgroundColor: record.color }"></div>
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ new Date(record.created_at).toLocaleDateString() }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="editGroup(record)">
                Edit
              </a-button>
              <a-button type="link" size="small" danger @click="deleteGroup(record.id)">
                Delete
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- Add/Edit Group Modal -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
      width="600px"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        layout="vertical"
      >
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Name" name="name">
              <a-input v-model:value="formData.name" placeholder="Enter group name" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Code" name="code">
              <a-input v-model:value="formData.code" placeholder="Enter group code" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="formData.description" placeholder="Enter description" :rows="3" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Icon" name="icon">
              <a-input v-model:value="formData.icon" placeholder="Enter icon class" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Color" name="color">
              <a-input v-model:value="formData.color" placeholder="#1890ff" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Sort" name="sort">
              <a-input-number v-model:value="formData.sort" :min="0" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Status" name="status">
              <a-select v-model:value="formData.status">
                <a-select-option :value="1">Active</a-select-option>
                <a-select-option :value="0">Inactive</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'

// Data
const loading = ref(false)
const modalVisible = ref(false)
const modalTitle = ref('Add Application Group')
const editingGroup = ref<any>(null)
const formRef = ref()

const groups = ref<any[]>([])
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  total_pages: 0
})

const formData = reactive({
  name: '',
  code: '',
  description: '',
  icon: '',
  color: '#1890ff',
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: 'Please enter group name' }],
  code: [{ required: true, message: 'Please enter group code' }],
  color: [{ required: true, message: 'Please enter color' }]
}

// Table columns
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
    title: 'Description',
    dataIndex: 'description',
    key: 'description',
    ellipsis: true
  },
  {
    title: 'Color',
    dataIndex: 'color',
    key: 'color',
    width: 80
  },
  {
    title: 'Sort',
    dataIndex: 'sort',
    key: 'sort',
    width: 80
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    width: 100
  },
  {
    title: 'Created At',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 120
  },
  {
    title: 'Action',
    key: 'action',
    width: 150
  }
]

// Methods
const loadGroups = async () => {
  try {
    loading.value = true
    // TODO: Implement API call
    // const response = await applicationGroupApi.getGroups({
    //   page: pagination.current,
    //   page_size: pagination.pageSize
    // })
    // groups.value = response.items
    // pagination.total = response.total
    // pagination.total_pages = response.total_pages
    
    // Mock data
    groups.value = [
      {
        id: '1',
        name: 'Internal Tools',
        code: 'internal',
        description: 'Internal company tools and applications',
        icon: 'tool',
        color: '#1890ff',
        sort: 1,
        status: 1,
        created_at: '2024-01-01T00:00:00Z'
      },
      {
        id: '2',
        name: 'External Services',
        code: 'external',
        description: 'External third-party services',
        icon: 'cloud',
        color: '#52c41a',
        sort: 2,
        status: 1,
        created_at: '2024-01-01T00:00:00Z'
      }
    ]
    pagination.total = 2
  } catch (error) {
    message.error('Failed to load application groups')
  } finally {
    loading.value = false
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadGroups()
}

const showAddGroupModal = () => {
  modalTitle.value = 'Add Application Group'
  editingGroup.value = null
  resetForm()
  modalVisible.value = true
}

const editGroup = (group: any) => {
  modalTitle.value = 'Edit Application Group'
  editingGroup.value = group
  Object.assign(formData, {
    name: group.name,
    code: group.code,
    description: group.description,
    icon: group.icon,
    color: group.color,
    sort: group.sort,
    status: group.status
  })
  modalVisible.value = true
}

const resetForm = () => {
  Object.assign(formData, {
    name: '',
    code: '',
    description: '',
    icon: '',
    color: '#1890ff',
    sort: 0,
    status: 1
  })
  formRef.value?.resetFields()
}

const handleModalOk = async () => {
  try {
    await formRef.value?.validate()
    
    if (editingGroup.value) {
      // Update group
      // TODO: Implement API call
      // await applicationGroupApi.updateGroup(editingGroup.value.id, formData)
      message.success('Application group updated successfully')
    } else {
      // Create group
      // TODO: Implement API call
      // await applicationGroupApi.createGroup(formData)
      message.success('Application group created successfully')
    }
    
    modalVisible.value = false
    loadGroups()
  } catch (error) {
    message.error('Please check the form')
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
  resetForm()
}

const deleteGroup = async (groupId: string) => {
  try {
    // TODO: Implement API call
    // await applicationGroupApi.deleteGroup(groupId)
    message.success('Application group deleted successfully')
    loadGroups()
  } catch (error) {
    message.error('Failed to delete application group')
  }
}

// Lifecycle
onMounted(() => {
  loadGroups()
})
</script>

<style scoped>
.application-groups-page {
  padding: 24px;
}

.color-preview {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid #d9d9d9;
}
</style>
