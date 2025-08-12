<template>
  <div class="applications-page">
    <a-card title="Application Management" :bordered="false">
      <template #extra>
        <a-button type="primary" @click="showAddAppModal">
          <template #icon>
            <PlusOutlined />
          </template>
          Add Application
        </a-button>
      </template>

      <a-table
        :columns="columns"
        :data-source="applications"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            <a-tag :color="getTypeColor(record.type)">
              {{ record.type }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 'active' ? 'green' : 'red'">
              {{ record.status }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="editApp(record)">
                Edit
              </a-button>
              <a-button type="link" size="small" @click="viewConfig(record)">
                Config
              </a-button>
              <a-popconfirm
                title="Are you sure you want to delete this application?"
                @confirm="deleteApp(record.id)"
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

    <!-- Add/Edit Application Modal -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      width="800px"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
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
              <a-input v-model:value="formData.name" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Type" name="type">
              <a-select v-model:value="formData.type">
                <a-select-option value="web">Web Application</a-select-option>
                <a-select-option value="mobile">Mobile Application</a-select-option>
                <a-select-option value="api">API Service</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="formData.description" :rows="3" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Group" name="groupId">
              <a-select v-model:value="formData.groupId" placeholder="Select group">
                <a-select-option v-for="group in appGroups" :key="group.id" :value="group.id">
                  {{ group.name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Status" name="status">
              <a-select v-model:value="formData.status">
                <a-select-option value="active">Active</a-select-option>
                <a-select-option value="inactive">Inactive</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Homepage URL" name="homepageUrl">
          <a-input v-model:value="formData.homepageUrl" />
        </a-form-item>
        <a-form-item label="Logo URL" name="logoUrl">
          <a-input v-model:value="formData.logoUrl" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Config Modal -->
    <a-modal
      v-model:open="configModalVisible"
      title="Application Configuration"
      width="1000px"
      @ok="handleConfigOk"
      @cancel="handleConfigCancel"
    >
      <a-tabs v-model:activeKey="activeConfigTab">
        <a-tab-pane key="oauth2" tab="OAuth2 Configuration">
          <a-form layout="vertical">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item label="Client ID">
                  <a-input v-model:value="configData.oauth2.clientId" readonly />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Client Secret">
                  <a-input-password v-model:value="configData.oauth2.clientSecret" readonly />
                </a-form-item>
              </a-col>
            </a-row>
            <a-form-item label="Redirect URIs">
              <a-textarea v-model:value="configData.oauth2.redirectUris" :rows="3" />
            </a-form-item>
            <a-form-item label="Grant Types">
              <a-checkbox-group v-model:value="configData.oauth2.grantTypes">
                <a-checkbox value="authorization_code">Authorization Code</a-checkbox>
                <a-checkbox value="client_credentials">Client Credentials</a-checkbox>
                <a-checkbox value="password">Password</a-checkbox>
              </a-checkbox-group>
            </a-form-item>
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item label="Access Token TTL (seconds)">
                  <a-input-number v-model:value="configData.oauth2.accessTokenTTL" :min="1" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Refresh Token TTL (seconds)">
                  <a-input-number v-model:value="configData.oauth2.refreshTokenTTL" :min="1" />
                </a-form-item>
              </a-col>
            </a-row>
          </a-form>
        </a-tab-pane>
        <a-tab-pane key="saml" tab="SAML Configuration">
          <a-form layout="vertical">
            <a-form-item label="Entity ID">
              <a-input v-model:value="configData.saml.entityId" />
            </a-form-item>
            <a-form-item label="ACS URL">
              <a-input v-model:value="configData.saml.acsUrl" />
            </a-form-item>
            <a-form-item label="SLO URL">
              <a-input v-model:value="configData.saml.sloUrl" />
            </a-form-item>
            <a-form-item label="Certificate">
              <a-textarea v-model:value="configData.saml.certificate" :rows="5" />
            </a-form-item>
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item label="Signature Algorithm">
                  <a-select v-model:value="configData.saml.signatureAlgorithm">
                    <a-select-option value="sha256">SHA-256</a-select-option>
                    <a-select-option value="sha512">SHA-512</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Digest Algorithm">
                  <a-select v-model:value="configData.saml.digestAlgorithm">
                    <a-select-option value="sha256">SHA-256</a-select-option>
                    <a-select-option value="sha512">SHA-512</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
            </a-row>
          </a-form>
        </a-tab-pane>
      </a-tabs>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import type { Application, Pagination } from '@/types/api'

// Data
const loading = ref(false)
const applications = ref<Application[]>([])
const appGroups = ref([])
const modalVisible = ref(false)
const configModalVisible = ref(false)
const modalTitle = ref('Add Application')
const activeConfigTab = ref('oauth2')
const formRef = ref()
const editingApp = ref<Application | null>(null)
const currentApp = ref<Application | null>(null)

const formData = reactive({
  name: '',
  type: 'web',
  description: '',
  groupId: undefined,
  status: 'active',
  homepageUrl: '',
  logoUrl: ''
})

const configData = reactive({
  oauth2: {
    clientId: '',
    clientSecret: '',
    redirectUris: '',
    grantTypes: ['authorization_code'],
    accessTokenTTL: 3600,
    refreshTokenTTL: 86400
  },
  saml: {
    entityId: '',
    acsUrl: '',
    sloUrl: '',
    certificate: '',
    signatureAlgorithm: 'sha256',
    digestAlgorithm: 'sha256'
  }
})

const formRules = {
  name: [{ required: true, message: 'Please input application name!' }],
  type: [{ required: true, message: 'Please select application type!' }],
  status: [{ required: true, message: 'Please select status!' }]
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
    title: 'Type',
    dataIndex: 'type',
    key: 'type'
  },
  {
    title: 'Group',
    dataIndex: 'groupName',
    key: 'groupName'
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status'
  },
  {
    title: 'Homepage URL',
    dataIndex: 'homepageUrl',
    key: 'homepageUrl'
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
const loadApplications = async () => {
  loading.value = true
  try {
    // Mock data for now
    applications.value = [
      {
        id: '1',
        name: 'HR System',
        type: 'web',
        groupName: 'Internal Apps',
        status: 'active',
        homepageUrl: 'https://hr.example.com',
        createdAt: '2024-01-01T00:00:00Z'
      },
      {
        id: '2',
        name: 'CRM System',
        type: 'web',
        groupName: 'Business Apps',
        status: 'active',
        homepageUrl: 'https://crm.example.com',
        createdAt: '2024-01-01T00:00:00Z'
      }
    ]
    pagination.total = 2
    pagination.total_pages = 1
  } catch (error) {
    message.error('Failed to load applications')
  } finally {
    loading.value = false
  }
}

const loadAppGroups = async () => {
  try {
    // Mock data for now
    appGroups.value = [
      { id: '1', name: 'Internal Apps' },
      { id: '2', name: 'Business Apps' },
      { id: '3', name: 'External Apps' }
    ]
  } catch (error) {
    message.error('Failed to load application groups')
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadApplications()
}

const getTypeColor = (type: string) => {
  const colors = {
    web: 'blue',
    mobile: 'green',
    api: 'orange'
  }
  return colors[type as keyof typeof colors] || 'default'
}

const showAddAppModal = () => {
  modalTitle.value = 'Add Application'
  editingApp.value = null
  resetForm()
  modalVisible.value = true
}

const editApp = (app: Application) => {
  modalTitle.value = 'Edit Application'
  editingApp.value = app
  Object.assign(formData, {
    name: app.name,
    type: app.type,
    description: app.description,
    groupId: app.groupId,
    status: app.status,
    homepageUrl: app.homepageUrl,
    logoUrl: app.logoUrl
  })
  modalVisible.value = true
}

const resetForm = () => {
  Object.assign(formData, {
    name: '',
    type: 'web',
    description: '',
    groupId: undefined,
    status: 'active',
    homepageUrl: '',
    logoUrl: ''
  })
  formRef.value?.resetFields()
}

const handleModalOk = async () => {
  try {
    await formRef.value?.validate()
    // TODO: Implement API call
    message.success(editingApp.value ? 'Application updated successfully' : 'Application created successfully')
    modalVisible.value = false
    loadApplications()
  } catch (error) {
    message.error('Please check the form')
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
  resetForm()
}

const viewConfig = (app: Application) => {
  currentApp.value = app
  // TODO: Load current app configuration
  configData.oauth2.clientId = 'client_' + app.id
  configData.oauth2.clientSecret = 'secret_' + app.id
  configModalVisible.value = true
}

const handleConfigOk = async () => {
  try {
    // TODO: Implement API call to update app configuration
    message.success('Configuration updated successfully')
    configModalVisible.value = false
  } catch (error) {
    message.error('Failed to update configuration')
  }
}

const handleConfigCancel = () => {
  configModalVisible.value = false
  currentApp.value = null
}

const deleteApp = async (appId: string) => {
  try {
    // TODO: Implement API call
    message.success('Application deleted successfully')
    loadApplications()
  } catch (error) {
    message.error('Failed to delete application')
  }
}

// Lifecycle
onMounted(() => {
  loadApplications()
  loadAppGroups()
})
</script>

<style scoped>
.applications-page {
  padding: 24px;
}
</style>
