<template>
  <div class="applications-management">
    <a-card title="应用程序管理" :bordered="false">
      <template #extra>
        <a-button type="primary" @click="showAddAppModal">
          <template #icon>
            <PlusOutlined />
          </template>
          添加应用程序
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
              {{ record.status === 1 ? '活跃' : '非活跃' }}
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
                编辑
              </a-button>
              <a-button type="link" size="small" @click="viewConfig(record)">
                配置
              </a-button>
              <a-popconfirm
                title="确定要删除这个应用程序吗？"
                @confirm="deleteApp(record.id)"
              >
                <a-button type="link" size="small" danger>
                  删除
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
            <a-form-item label="名称" name="name">
              <a-input v-model:value="formData.name" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="类型" name="type">
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
        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="formData.description" :rows="3" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="分组" name="groupId">
              <a-select v-model:value="formData.groupId" placeholder="选择分组">
                <a-select-option v-for="group in appGroups" :key="group.id" :value="group.id">
                  {{ group.name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="状态" name="status">
              <a-select v-model:value="formData.status">
                <a-select-option value="active">活跃</a-select-option>
                <a-select-option value="inactive">非活跃</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="主页URL" name="homepageUrl">
          <a-input v-model:value="formData.homepageUrl" />
        </a-form-item>
        <a-form-item label="Logo URL" name="logoUrl">
          <a-input v-model:value="formData.logoUrl" />
        </a-form-item>
        
        <!-- Dynamic Configuration Fields based on Type -->
        <a-divider>配置</a-divider>
        
        <!-- OAuth2 Configuration -->
        <div v-if="formData.type === 'oauth2'">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="客户端ID" name="clientId">
                <a-input v-model:value="formData.config.clientId" placeholder="输入客户端ID" />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="客户端密钥" name="clientSecret">
                <a-input-password v-model:value="formData.config.clientSecret" placeholder="输入客户端密钥" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="重定向URI" name="redirectUris">
            <a-textarea v-model:value="formData.config.redirectUris" :rows="3" placeholder="输入重定向URI（每行一个）" />
          </a-form-item>
          <a-form-item label="作用域" name="scopes">
            <a-select v-model:value="formData.config.scopes" mode="multiple" placeholder="选择作用域">
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
          <a-form-item label="实体ID" name="entityId">
            <a-input v-model:value="formData.config.entityId" placeholder="输入实体ID" />
          </a-form-item>
          <a-form-item label="ACS URL" name="acsUrl">
            <a-input v-model:value="formData.config.acsUrl" placeholder="输入ACS URL" />
          </a-form-item>
          <a-form-item label="SLO URL" name="sloUrl">
            <a-input v-model:value="formData.config.sloUrl" placeholder="输入SLO URL" />
          </a-form-item>
          <a-form-item label="证书" name="certificate">
            <a-textarea v-model:value="formData.config.certificate" :rows="5" placeholder="输入证书" />
          </a-form-item>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="签名算法" name="signatureAlgorithm">
                <a-select v-model:value="formData.config.signatureAlgorithm">
                  <a-select-option value="sha256">SHA-256</a-select-option>
                  <a-select-option value="sha512">SHA-512</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="摘要算法" name="digestAlgorithm">
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
          <a-form-item label="服务URL" name="serviceUrl">
            <a-input v-model:value="formData.config.serviceUrl" placeholder="输入服务URL" />
          </a-form-item>
          <a-form-item label="网关模式" name="gateway">
            <a-switch v-model:checked="formData.config.gateway" />
          </a-form-item>
          <a-form-item label="重新认证" name="renew">
            <a-switch v-model:checked="formData.config.renew" />
          </a-form-item>
        </div>
        
        <!-- OpenID Connect Configuration -->
        <div v-if="formData.type === 'oidc'">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="客户端ID" name="clientId">
                <a-input v-model:value="formData.config.clientId" placeholder="输入客户端ID" />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="客户端密钥" name="clientSecret">
                <a-input-password v-model:value="formData.config.clientSecret" placeholder="输入客户端密钥" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="重定向URI" name="redirectUris">
            <a-textarea v-model:value="formData.config.redirectUris" :rows="3" placeholder="输入重定向URI（每行一个）" />
          </a-form-item>
          <a-form-item label="作用域" name="scopes">
            <a-select v-model:value="formData.config.scopes" mode="multiple" placeholder="选择作用域">
              <a-select-option value="openid">openid</a-select-option>
              <a-select-option value="profile">profile</a-select-option>
              <a-select-option value="email">email</a-select-option>
              <a-select-option value="address">address</a-select-option>
              <a-select-option value="phone">phone</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="授权类型" name="grantTypes">
            <a-select v-model:value="formData.config.grantTypes" mode="multiple" placeholder="选择授权类型">
              <a-select-option value="authorization_code">授权码</a-select-option>
              <a-select-option value="refresh_token">刷新令牌</a-select-option>
              <a-select-option value="client_credentials">客户端凭证</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="响应类型" name="responseTypes">
            <a-select v-model:value="formData.config.responseTypes" mode="multiple" placeholder="选择响应类型">
              <a-select-option value="code">Code</a-select-option>
              <a-select-option value="token">Token</a-select-option>
            </a-select>
          </a-form-item>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="访问令牌TTL（秒）" name="accessTokenTTL">
                <a-input-number v-model:value="formData.config.accessTokenTTL" :min="1" />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="刷新令牌TTL（秒）" name="refreshTokenTTL">
                <a-input-number v-model:value="formData.config.refreshTokenTTL" :min="1" />
              </a-form-item>
            </a-col>
          </a-row>
        </div>
        
        <!-- LDAP Configuration -->
        <div v-if="formData.type === 'ldap'">
          <a-form-item label="LDAP URL" name="ldapUrl">
            <a-input v-model:value="formData.config.ldapUrl" placeholder="输入LDAP URL" />
          </a-form-item>
          <a-form-item label="基础DN" name="baseDn">
            <a-input v-model:value="formData.config.baseDn" placeholder="输入基础DN" />
          </a-form-item>
          <a-form-item label="绑定DN" name="bindDn">
            <a-input v-model:value="formData.config.bindDn" placeholder="输入绑定DN" />
          </a-form-item>
          <a-form-item label="绑定密码" name="bindPassword">
            <a-input-password v-model:value="formData.config.bindPassword" placeholder="输入绑定密码" />
          </a-form-item>
        </div>
      </a-form>
    </a-modal>

    <!-- Config Modal -->
    <a-modal
      v-model:open="configModalVisible"
      title="应用程序配置"
      width="1000px"
      @ok="handleConfigOk"
      @cancel="handleConfigCancel"
    >
      <a-tabs v-model:activeKey="activeConfigTab">
        <a-tab-pane key="oauth2" tab="OAuth2 配置">
          <a-form layout="vertical">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item label="客户端ID">
                  <a-input v-model:value="configData.oauth2.clientId" readonly />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="客户端密钥">
                  <a-input-password v-model:value="configData.oauth2.clientSecret" readonly />
                </a-form-item>
              </a-col>
            </a-row>
            <a-form-item label="重定向URI">
              <a-textarea v-model:value="configData.oauth2.redirectUris" :rows="3" />
            </a-form-item>
            <a-form-item label="授权类型">
              <a-checkbox-group v-model:value="configData.oauth2.grantTypes">
                <a-checkbox value="authorization_code">授权码</a-checkbox>
                <a-checkbox value="client_credentials">客户端凭证</a-checkbox>
                <a-checkbox value="password">密码</a-checkbox>
              </a-checkbox-group>
            </a-form-item>
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item label="访问令牌TTL（秒）">
                  <a-input-number v-model:value="configData.oauth2.accessTokenTTL" :min="1" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="刷新令牌TTL（秒）">
                  <a-input-number v-model:value="configData.oauth2.refreshTokenTTL" :min="1" />
                </a-form-item>
              </a-col>
            </a-row>
          </a-form>
        </a-tab-pane>
        <a-tab-pane key="cas" tab="CAS 配置">
          <a-form layout="vertical">
            <a-form-item label="服务URL">
              <a-input v-model:value="configData.cas.serviceUrl" placeholder="输入服务URL" />
            </a-form-item>
            <a-form-item label="网关模式">
              <a-switch v-model:checked="configData.cas.gateway" />
              <div class="text-gray-500 text-sm mt-1">启用网关模式进行自动身份验证</div>
            </a-form-item>
            <a-form-item label="重新认证">
              <a-switch v-model:checked="configData.cas.renew" />
              <div class="text-gray-500 text-sm mt-1">即使用户已经登录也强制重新认证</div>
            </a-form-item>
          </a-form>
        </a-tab-pane>
        <a-tab-pane key="saml" tab="SAML 配置">
          <a-form layout="vertical">
            <a-form-item label="实体ID">
              <a-input v-model:value="configData.saml.entityId" />
            </a-form-item>
            <a-form-item label="ACS URL">
              <a-input v-model:value="configData.saml.acsUrl" />
            </a-form-item>
            <a-form-item label="SLO URL">
              <a-input v-model:value="configData.saml.sloUrl" />
            </a-form-item>
            <a-form-item label="证书">
              <a-textarea v-model:value="configData.saml.certificate" :rows="5" />
            </a-form-item>
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item label="签名算法">
                  <a-select v-model:value="configData.saml.signatureAlgorithm">
                    <a-select-option value="sha256">SHA-256</a-select-option>
                    <a-select-option value="sha512">SHA-512</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="摘要算法">
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
import type { Application } from '@/api/applications'
import { applicationApi } from '@/api/applications'

// Data
const loading = ref(false)
const applications = ref<any[]>([])
const appGroups = ref<any[]>([])
const modalVisible = ref(false)
const configModalVisible = ref(false)
const modalTitle = ref('添加应用程序')
const activeConfigTab = ref('oauth2')
const formRef = ref()
const editingApp = ref<Application | null>(null)
const currentApp = ref<Application | null>(null)

const formData = reactive({
  name: '',
  type: 'oauth2',
  description: '',
  groupId: '' as string | undefined,
  status: 'active',
  homepageUrl: '',
  logoUrl: '',
  config: {
    // OAuth2/OIDC fields
    clientId: '',
    clientSecret: '',
    redirectUris: '',
    scopes: [],
    grantTypes: '',
    responseTypes: '',
    accessTokenTTL: 3600,
    refreshTokenTTL: 604800,
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
  cas: {
    serviceUrl: '',
    gateway: false,
    renew: false
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
  name: [{ required: true, message: '请输入应用程序名称！' }],
  type: [{ required: true, message: '请选择应用程序类型！' }],
  status: [{ required: true, message: '请选择状态！' }],
  groupId: [{ required: false }], // 应用组是可选的
  // 配置字段验证规则
  'config.clientId': [{ required: false }],
  'config.clientSecret': [{ required: false }],
  'config.redirectUris': [{ required: false }],
  'config.scopes': [{ required: false }],
  'config.grantTypes': [{ required: false }],
  'config.responseTypes': [{ required: false }],
  'config.accessTokenTTL': [{ required: false }],
  'config.refreshTokenTTL': [{ required: false }],
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

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  total_pages: 0
})

const columns = [
  {
    title: '名称',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: '类型',
    dataIndex: 'protocol',
    key: 'protocol'
  },
  {
    title: '分组',
    dataIndex: 'group',
    key: 'group'
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status'
  },
  {
    title: '主页URL',
    dataIndex: 'home_page_url',
    key: 'home_page_url'
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at'
  },
  {
    title: '操作',
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
    message.error(error.response?.data?.message || '加载应用程序失败')
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
    message.error(error.response?.data?.message || '加载应用程序分组失败')
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
  modalTitle.value = '添加应用程序'
  editingApp.value = null
  resetForm()
  modalVisible.value = true
}

const editApp = (app: any) => {
  modalTitle.value = '编辑应用程序'
  editingApp.value = app
  
  // 填充基本数据
  Object.assign(formData, {
    name: app.name,
    type: app.protocol || app.type,
    description: app.description,
    groupId: app.group_id,
    status: app.status === 1 ? 'active' : 'inactive',
    homepageUrl: app.home_page_url || app.homepage_url,
    logoUrl: app.logo
  })
  
  // 填充配置数据 - 从应用对象中直接获取配置字段
  const config = {
    clientId: app.client_id || '',
    clientSecret: app.client_secret || '',
    redirectUris: app.redirect_uris || '',
    scopes: app.scopes ? (Array.isArray(app.scopes) ? app.scopes : app.scopes.split(',')) : [],
    grantTypes: app.grant_types || '',
    responseTypes: app.response_types || '',
    accessTokenTTL: app.access_token_ttl || 3600,
    refreshTokenTTL: app.refresh_token_ttl || 604800,
    entityId: app.entity_id || '',
    acsUrl: app.acs_url || '',
    sloUrl: app.slo_url || '',
    certificate: app.certificate || '',
    signatureAlgorithm: app.signature_algorithm || 'sha256',
    digestAlgorithm: app.digest_algorithm || 'sha256',
    serviceUrl: app.service_url || '',
    gateway: app.gateway || false,
    renew: app.renew || false,
    ldapUrl: app.ldap_url || '',
    baseDn: app.base_dn || '',
    bindDn: app.bind_dn || '',
    bindPassword: app.bind_password || ''
  }
  Object.assign(formData.config, config)
  
  modalVisible.value = true
}

const onTypeChange = () => {
  // Reset config when type changes
  formData.config = {
    clientId: '',
    clientSecret: '',
    redirectUris: '',
    scopes: [],
    grantTypes: '',
    responseTypes: '',
    accessTokenTTL: 3600,
    refreshTokenTTL: 604800,
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
      grantTypes: '',
      responseTypes: '',
      accessTokenTTL: 3600,
      refreshTokenTTL: 604800,
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
      protocol: formData.type, // 添加协议字段
      description: formData.description,
      status: formData.status === 'active' ? 1 : 0,
      homepageUrl: formData.homepageUrl,
      logoUrl: formData.logoUrl,
      // 根据应用类型添加相应的配置字段
      ...(formData.type === 'oauth2' && {
        clientId: formData.config.clientId,
        clientSecret: formData.config.clientSecret,
        redirectUris: formData.config.redirectUris,
        scopes: Array.isArray(formData.config.scopes) ? formData.config.scopes.join(' ') : formData.config.scopes
      }),
      ...(formData.type === 'saml' && {
        entity_id: formData.config.entityId,
        acs_url: formData.config.acsUrl,
        slo_url: formData.config.sloUrl,
        certificate: formData.config.certificate,
        signature_algorithm: formData.config.signatureAlgorithm,
        digest_algorithm: formData.config.digestAlgorithm
      }),
      ...(formData.type === 'cas' && {
        service_url: formData.config.serviceUrl,
        gateway: formData.config.gateway,
        renew: formData.config.renew
      }),
      ...(formData.type === 'oidc' && {
        clientId: formData.config.clientId,
        clientSecret: formData.config.clientSecret,
        redirectUris: formData.config.redirectUris,
        scopes: Array.isArray(formData.config.scopes) ? formData.config.scopes.join(' ') : formData.config.scopes,
        grantTypes: Array.isArray(formData.config.grantTypes) ? formData.config.grantTypes.join(' ') : formData.config.grantTypes,
        responseTypes: Array.isArray(formData.config.responseTypes) ? formData.config.responseTypes.join(' ') : formData.config.responseTypes,
        accessTokenTTL: formData.config.accessTokenTTL,
        refreshTokenTTL: formData.config.refreshTokenTTL
      }),
      ...(formData.type === 'ldap' && {
        ldapUrl: formData.config.ldapUrl,
        baseDn: formData.config.baseDn,
        bindDn: formData.config.bindDn,
        bindPassword: formData.config.bindPassword
      })
    }
    
    // 只有当groupId不为空时才添加到请求中
    if (formData.groupId && formData.groupId.trim() !== '') {
      (requestData as any).groupId = formData.groupId
    }
    
    // 调试：打印发送的数据
    console.log('Sending request data:', JSON.stringify(requestData, null, 2))
    
    if (editingApp.value) {
      await applicationApi.updateApplication(editingApp.value.id, requestData)
      message.success('应用程序更新成功')
    } else {
      await applicationApi.createApplication(requestData)
      message.success('应用程序创建成功')
    }
    
    modalVisible.value = false
    loadApplications()
  } catch (error: any) {
    console.error('Error details:', error)
    console.error('Error response:', error.response?.data)
    message.error(error.response?.data?.message || '请检查表单')
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
    message.success('配置更新成功')
    configModalVisible.value = false
  } catch (error) {
    message.error('配置更新失败')
  }
}

const handleConfigCancel = () => {
  configModalVisible.value = false
  currentApp.value = null
}

const deleteApp = async (appId: string) => {
  try {
    await applicationApi.deleteApplication(appId)
    message.success('应用程序删除成功')
    loadApplications()
  } catch (error: any) {
    console.error('Failed to delete application:', error)
    message.error(error.response?.data?.message || '删除应用程序失败')
  }
}

// Lifecycle
onMounted(() => {
  loadApplications()
  loadAppGroups()
})
</script>

<style scoped>
.applications-management {
  /* No padding here since it's handled by the parent */
}
</style>
