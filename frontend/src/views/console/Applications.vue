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
          <template v-if="column.key === 'protocol'">
            <a-tag :color="getTypeColor(record.protocol)">
              {{ record.protocol }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? 'Active' : 'Inactive' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'group'">
            {{ record.group?.name || '-' }}
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ new Date(record.created_at).toLocaleDateString() }}
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
              <a-select v-model:value="formData.type" @change="onTypeChange">
                <a-select-option value="oauth2">OAuth2</a-select-option>
                <a-select-option value="saml">SAML</a-select-option>
                <a-select-option value="cas">CAS</a-select-option>
                <a-select-option value="oidc">OpenID Connect</a-select-option>
                <a-select-option value="ldap">LDAP</a-select-option>
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
        
        <!-- Dynamic Configuration Fields based on Type -->
        <a-divider>Configuration</a-divider>
        
        <!-- OAuth2 Configuration -->
        <div v-if="formData.type === 'oauth2'">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Client ID" name="clientId">
                <a-input v-model:value="formData.config.clientId" placeholder="Enter client ID" />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Client Secret" name="clientSecret">
                <a-input-password v-model:value="formData.config.clientSecret" placeholder="Enter client secret" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="Redirect URIs" name="redirectUris">
            <a-textarea v-model:value="formData.config.redirectUris" :rows="3" placeholder="Enter redirect URIs (one per line)" />
          </a-form-item>
          <a-form-item label="Scopes" name="scopes">
            <a-select v-model:value="formData.config.scopes" mode="multiple" placeholder="Select scopes">
              <a-select-option value="openid">openid</a-select-option>
              <a-select-option value="profile">profile</a-select-option>
              <a-select-option value="email">email</a-select-option>
              <a-select-option value="read">read</a-select-option>
              <a-select-option value="write">write</a-select-option>
            </a-select>
          </a-form-item>
        </div>
        
        <!-- SAML Configuration -->
        <div v-if="formData.type === 'saml'">
          <a-form-item label="Entity ID" name="entityId">
            <a-input v-model:value="formData.config.entityId" placeholder="Enter entity ID" />
          </a-form-item>
          <a-form-item label="ACS URL" name="acsUrl">
            <a-input v-model:value="formData.config.acsUrl" placeholder="Enter ACS URL" />
          </a-form-item>
          <a-form-item label="SLO URL" name="sloUrl">
            <a-input v-model:value="formData.config.sloUrl" placeholder="Enter SLO URL" />
          </a-form-item>
          <a-form-item label="Certificate" name="certificate">
            <a-textarea v-model:value="formData.config.certificate" :rows="5" placeholder="Enter certificate" />
          </a-form-item>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Signature Algorithm" name="signatureAlgorithm">
                <a-select v-model:value="formData.config.signatureAlgorithm">
                  <a-select-option value="sha256">SHA-256</a-select-option>
                  <a-select-option value="sha512">SHA-512</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Digest Algorithm" name="digestAlgorithm">
                <a-select v-model:value="formData.config.digestAlgorithm">
                  <a-select-option value="sha256">SHA-256</a-select-option>
                  <a-select-option value="sha512">SHA-512</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
          </a-row>
        </div>
        
        <!-- CAS Configuration -->
        <div v-if="formData.type === 'cas'">
          <a-form-item label="Service URL" name="serviceUrl">
            <a-input v-model:value="formData.config.serviceUrl" placeholder="Enter service URL" />
          </a-form-item>
          <a-form-item label="Gateway" name="gateway">
            <a-switch v-model:checked="formData.config.gateway" />
          </a-form-item>
          <a-form-item label="Renew" name="renew">
            <a-switch v-model:checked="formData.config.renew" />
          </a-form-item>
        </div>
        
        <!-- OpenID Connect Configuration -->
        <div v-if="formData.type === 'oidc'">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Client ID" name="clientId">
                <a-input v-model:value="formData.config.clientId" placeholder="Enter client ID" />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Client Secret" name="clientSecret">
                <a-input-password v-model:value="formData.config.clientSecret" placeholder="Enter client secret" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="Redirect URIs" name="redirectUris">
            <a-textarea v-model:value="formData.config.redirectUris" :rows="3" placeholder="Enter redirect URIs (one per line)" />
          </a-form-item>
          <a-form-item label="Scopes" name="scopes">
            <a-select v-model:value="formData.config.scopes" mode="multiple" placeholder="Select scopes">
              <a-select-option value="openid">openid</a-select-option>
              <a-select-option value="profile">profile</a-select-option>
              <a-select-option value="email">email</a-select-option>
              <a-select-option value="address">address</a-select-option>
              <a-select-option value="phone">phone</a-select-option>
            </a-select>
          </a-form-item>
        </div>
        
        <!-- LDAP Configuration -->
        <div v-if="formData.type === 'ldap'">
          <a-form-item label="LDAP URL" name="ldapUrl">
            <a-input v-model:value="formData.config.ldapUrl" placeholder="Enter LDAP URL" />
          </a-form-item>
          <a-form-item label="Base DN" name="baseDn">
            <a-input v-model:value="formData.config.baseDn" placeholder="Enter base DN" />
          </a-form-item>
          <a-form-item label="Bind DN" name="bindDn">
            <a-input v-model:value="formData.config.bindDn" placeholder="Enter bind DN" />
          </a-form-item>
          <a-form-item label="Bind Password" name="bindPassword">
            <a-input-password v-model:value="formData.config.bindPassword" placeholder="Enter bind password" />
          </a-form-item>
        </div>
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
import { ref, reactive, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import type { Application, Pagination } from '@/types/api'
import { applicationApi } from '@/api/applications'

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
  type: 'oauth2',
  description: '',
  groupId: undefined,
  status: 'active',
  homepageUrl: '',
  logoUrl: '',
  config: {
    // OAuth2/OIDC fields
    clientId: '',
    clientSecret: '',
    redirectUris: '',
    scopes: [],
    // SAML fields
    entityId: '',
    acsUrl: '',
    sloUrl: '',
    certificate: '',
    signatureAlgorithm: 'sha256',
    digestAlgorithm: 'sha256',
    // CAS fields
    serviceUrl: '',
    gateway: false,
    renew: false,
    // LDAP fields
    ldapUrl: '',
    baseDn: '',
    bindDn: '',
    bindPassword: ''
  }
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
  status: [{ required: true, message: 'Please select status!' }],
  // 配置字段验证规则
  'config.clientId': [{ required: false }],
  'config.clientSecret': [{ required: false }],
  'config.redirectUris': [{ required: false }],
  'config.scopes': [{ required: false }],
  'config.entityId': [{ required: false }],
  'config.acsUrl': [{ required: false }],
  'config.sloUrl': [{ required: false }],
  'config.certificate': [{ required: false }],
  'config.signatureAlgorithm': [{ required: false }],
  'config.digestAlgorithm': [{ required: false }],
  'config.serviceUrl': [{ required: false }],
  'config.gateway': [{ required: false }],
  'config.renew': [{ required: false }],
  'config.ldapUrl': [{ required: false }],
  'config.baseDn': [{ required: false }],
  'config.bindDn': [{ required: false }],
  'config.bindPassword': [{ required: false }]
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
    dataIndex: 'protocol',
    key: 'protocol'
  },
  {
    title: 'Group',
    dataIndex: 'group',
    key: 'group'
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status'
  },
  {
    title: 'Homepage URL',
    dataIndex: 'home_page_url',
    key: 'home_page_url'
  },
  {
    title: 'Created At',
    dataIndex: 'created_at',
    key: 'created_at'
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
    const response = await applicationApi.getApplications({
      page: pagination.current,
      page_size: pagination.pageSize
    })
    applications.value = response.items || []
    pagination.total = response.total
    pagination.total_pages = response.total_pages
  } catch (error: any) {
    console.error('Failed to load applications:', error)
    message.error(error.response?.data?.message || 'Failed to load applications')
  } finally {
    loading.value = false
  }
}

const loadAppGroups = async () => {
  try {
    const response = await applicationApi.getApplicationGroups({
      page: 1,
      page_size: 100
    })
    appGroups.value = response.items || []
  } catch (error: any) {
    console.error('Failed to load application groups:', error)
    message.error(error.response?.data?.message || 'Failed to load application groups')
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadApplications()
}

const getTypeColor = (type: string) => {
  const colors = {
    oauth2: 'blue',
    saml: 'green',
    cas: 'orange',
    oidc: 'purple',
    ldap: 'cyan'
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
    type: app.protocol,
    description: app.description,
    groupId: app.group_id,
    status: app.status === 1 ? 'active' : 'inactive',
    homepageUrl: app.home_page_url,
    logoUrl: app.logo
  })
  modalVisible.value = true
}

const onTypeChange = (type: string) => {
  // Reset config when type changes
  formData.config = {
    clientId: '',
    clientSecret: '',
    redirectUris: '',
    scopes: [],
    entityId: '',
    acsUrl: '',
    sloUrl: '',
    certificate: '',
    signatureAlgorithm: 'sha256',
    digestAlgorithm: 'sha256',
    serviceUrl: '',
    gateway: false,
    renew: false,
    ldapUrl: '',
    baseDn: '',
    bindDn: '',
    bindPassword: ''
  }
}

const resetForm = () => {
  Object.assign(formData, {
    name: '',
    type: 'oauth2',
    description: '',
    groupId: undefined,
    status: 'active',
    homepageUrl: '',
    logoUrl: '',
    config: {
      clientId: '',
      clientSecret: '',
      redirectUris: '',
      scopes: [],
      entityId: '',
      acsUrl: '',
      sloUrl: '',
      certificate: '',
      signatureAlgorithm: 'sha256',
      digestAlgorithm: 'sha256',
      serviceUrl: '',
      gateway: false,
      renew: false,
      ldapUrl: '',
      baseDn: '',
      bindDn: '',
      bindPassword: ''
    }
  })
  formRef.value?.resetFields()
}

const handleModalOk = async () => {
  try {
    await formRef.value?.validate()
    
    const requestData = {
      name: formData.name,
      type: formData.type,
      description: formData.description,
      groupId: formData.groupId,
      status: formData.status === 'active' ? 1 : 0,
      homepageUrl: formData.homepageUrl,
      logoUrl: formData.logoUrl,
      config: formData.config
    }
    
    if (editingApp.value) {
      await applicationApi.updateApplication(editingApp.value.id, requestData)
      message.success('Application updated successfully')
    } else {
      await applicationApi.createApplication(requestData)
      message.success('Application created successfully')
    }
    
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
    await applicationApi.deleteApplication(appId)
    message.success('Application deleted successfully')
    loadApplications()
  } catch (error: any) {
    console.error('Failed to delete application:', error)
    message.error(error.response?.data?.message || 'Failed to delete application')
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
