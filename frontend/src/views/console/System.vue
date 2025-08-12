<template>
  <div class="system-page">
    <div class="page-header">
      <h2>System Management</h2>
      <p>Configure system settings, security policies, and administrator access</p>
    </div>

    <!-- System Configuration Tabs -->
    <a-tabs v-model:activeKey="activeTab" type="card">
      <!-- Administrator Management -->
      <a-tab-pane key="admins" tab="Administrators">
        <div class="tab-content">
          <div class="actions-bar">
            <a-button type="primary" @click="showAdminModal">
              <PlusOutlined />
              Add Administrator
            </a-button>
          </div>
          
          <a-table
            :columns="adminColumns"
            :data-source="administrators"
            :loading="loading"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <a-tag :color="record.status === 'active' ? 'green' : 'red'">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'role'">
                <a-tag :color="getRoleColor(record.role)">{{ record.role }}</a-tag>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-button type="link" size="small" @click="editAdmin(record)">
                    Edit
                  </a-button>
                  <a-button type="link" size="small" @click="viewAdminPermissions(record)">
                    Permissions
                  </a-button>
                  <a-popconfirm
                    title="Are you sure you want to remove administrator privileges?"
                    @confirm="removeAdmin(record.id)"
                  >
                    <a-button type="link" size="small" danger>Remove</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </a-tab-pane>

      <!-- Site Configuration -->
      <a-tab-pane key="site" tab="Site Configuration">
        <div class="tab-content">
          <a-form
            ref="siteFormRef"
            :model="siteForm"
            :rules="siteRules"
            layout="vertical"
          >
            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item label="Site Name" name="siteName">
                  <a-input v-model:value="siteForm.siteName" placeholder="Enter site name" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Site URL" name="siteUrl">
                  <a-input v-model:value="siteForm.siteUrl" placeholder="https://example.com" />
                </a-form-item>
              </a-col>
            </a-row>
            
            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item label="Contact Email" name="contactEmail">
                  <a-input v-model:value="siteForm.contactEmail" placeholder="admin@example.com" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Support Email" name="supportEmail">
                  <a-input v-model:value="siteForm.supportEmail" placeholder="support@example.com" />
                </a-form-item>
              </a-col>
            </a-row>

            <a-form-item label="Site Description" name="description">
              <a-textarea v-model:value="siteForm.description" placeholder="Enter site description" />
            </a-form-item>

            <a-form-item label="Logo" name="logo">
              <a-upload
                name="logo"
                list-type="picture-card"
                class="logo-uploader"
                :show-upload-list="false"
                :before-upload="beforeLogoUpload"
                @change="handleLogoChange"
              >
                <div v-if="!siteForm.logo">
                  <PlusOutlined />
                  <div style="margin-top: 8px">Upload Logo</div>
                </div>
                <img v-else :src="siteForm.logo" alt="logo" style="width: 100%" />
              </a-upload>
            </a-form-item>

            <a-form-item>
              <a-button type="primary" @click="saveSiteConfig">
                Save Site Configuration
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </a-tab-pane>

      <!-- Security Configuration -->
      <a-tab-pane key="security" tab="Security Configuration">
        <div class="tab-content">
          <a-form
            ref="securityFormRef"
            :model="securityForm"
            :rules="securityRules"
            layout="vertical"
          >
            <!-- Password Policy -->
            <a-card title="Password Policy" :bordered="false" style="margin-bottom: 24px">
              <a-row :gutter="24">
                <a-col :span="8">
                  <a-form-item label="Minimum Length" name="minPasswordLength">
                    <a-input-number
                      v-model:value="securityForm.minPasswordLength"
                      :min="6"
                      :max="50"
                      style="width: 100%"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="8">
                  <a-form-item label="Maximum Length" name="maxPasswordLength">
                    <a-input-number
                      v-model:value="securityForm.maxPasswordLength"
                      :min="8"
                      :max="128"
                      style="width: 100%"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="8">
                  <a-form-item label="Password Expiry (days)" name="passwordExpiryDays">
                    <a-input-number
                      v-model:value="securityForm.passwordExpiryDays"
                      :min="0"
                      :max="365"
                      style="width: 100%"
                    />
                  </a-form-item>
                </a-col>
              </a-row>

              <a-row :gutter="24">
                <a-col :span="12">
                  <a-form-item label="Password Requirements">
                    <a-checkbox v-model:checked="securityForm.requireUppercase">
                      Require uppercase letters
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.requireLowercase">
                      Require lowercase letters
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.requireNumbers">
                      Require numbers
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.requireSpecialChars">
                      Require special characters
                    </a-checkbox>
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item label="Password History">
                    <a-input-number
                      v-model:value="securityForm.passwordHistoryCount"
                      :min="0"
                      :max="10"
                      style="width: 100%"
                      placeholder="Number of previous passwords to remember"
                    />
                  </a-form-item>
                </a-col>
              </a-row>
            </a-card>

            <!-- Session Management -->
            <a-card title="Session Management" :bordered="false" style="margin-bottom: 24px">
              <a-row :gutter="24">
                <a-col :span="8">
                  <a-form-item label="Session Timeout (minutes)" name="sessionTimeout">
                    <a-input-number
                      v-model:value="securityForm.sessionTimeout"
                      :min="5"
                      :max="1440"
                      style="width: 100%"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="8">
                  <a-form-item label="Max Concurrent Sessions" name="maxConcurrentSessions">
                    <a-input-number
                      v-model:value="securityForm.maxConcurrentSessions"
                      :min="1"
                      :max="10"
                      style="width: 100%"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="8">
                  <a-form-item label="Remember Me Duration (days)" name="rememberMeDays">
                    <a-input-number
                      v-model:value="securityForm.rememberMeDays"
                      :min="1"
                      :max="30"
                      style="width: 100%"
                    />
                  </a-form-item>
                </a-col>
              </a-row>
            </a-card>

            <!-- 2FA Configuration -->
            <a-card title="Two-Factor Authentication" :bordered="false" style="margin-bottom: 24px">
              <a-row :gutter="24">
                <a-col :span="12">
                  <a-form-item label="2FA Settings">
                    <a-checkbox v-model:checked="securityForm.enable2FA">
                      Enable 2FA for all users
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.require2FAForAdmins">
                      Require 2FA for administrators
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.allowBackupCodes">
                      Allow backup codes
                    </a-checkbox>
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item label="2FA Methods">
                    <a-checkbox v-model:checked="securityForm.enableTOTP">
                      TOTP (Time-based One-Time Password)
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.enableSMS">
                      SMS verification
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.enableEmail">
                      Email verification
                    </a-checkbox>
                  </a-form-item>
                </a-col>
              </a-row>
            </a-card>

            <!-- Login Security -->
            <a-card title="Login Security" :bordered="false">
              <a-row :gutter="24">
                <a-col :span="12">
                  <a-form-item label="Login Attempts">
                    <a-input-number
                      v-model:value="securityForm.maxLoginAttempts"
                      :min="3"
                      :max="10"
                      style="width: 100%"
                      placeholder="Maximum failed login attempts"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item label="Lockout Duration (minutes)" name="lockoutDuration">
                    <a-input-number
                      v-model:value="securityForm.lockoutDuration"
                      :min="5"
                      :max="1440"
                      style="width: 100%"
                    />
                  </a-form-item>
                </a-col>
              </a-row>

              <a-row :gutter="24">
                <a-col :span="12">
                  <a-form-item label="Security Features">
                    <a-checkbox v-model:checked="securityForm.enableIPWhitelist">
                      Enable IP whitelist
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.enableGeolocation">
                      Enable geolocation tracking
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.enableDeviceFingerprinting">
                      Enable device fingerprinting
                    </a-checkbox>
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item label="Security Notifications">
                    <a-checkbox v-model:checked="securityForm.notifyFailedLogins">
                      Notify on failed login attempts
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.notifyNewDevices">
                      Notify on new device login
                    </a-checkbox>
                    <br />
                    <a-checkbox v-model:checked="securityForm.notifyPasswordChanges">
                      Notify on password changes
                    </a-checkbox>
                  </a-form-item>
                </a-col>
              </a-row>
            </a-card>

            <a-form-item style="margin-top: 24px">
              <a-button type="primary" @click="saveSecurityConfig">
                Save Security Configuration
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </a-tab-pane>
    </a-tabs>

    <!-- Administrator Modal -->
    <a-modal
      v-model:open="adminModalVisible"
      :title="editingAdmin ? 'Edit Administrator' : 'Add Administrator'"
      @ok="handleAdminSubmit"
      @cancel="handleAdminCancel"
    >
      <a-form
        ref="adminFormRef"
        :model="adminForm"
        :rules="adminRules"
        layout="vertical"
      >
        <a-form-item label="User" name="userId">
          <a-select
            v-model:value="adminForm.userId"
            placeholder="Select user"
            show-search
            :filter-option="filterUserOption"
          >
            <a-select-option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.display_name }} ({{ user.username }})
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Role" name="role">
          <a-select v-model:value="adminForm.role">
            <a-select-option value="super_admin">Super Administrator</a-select-option>
            <a-select-option value="system_admin">System Administrator</a-select-option>
            <a-select-option value="security_admin">Security Administrator</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Status" name="status">
          <a-select v-model:value="adminForm.status">
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
import { systemApi } from '@/api/system'

// Data
const activeTab = ref('admins')
const loading = ref(false)
const adminModalVisible = ref(false)
const editingAdmin = ref(false)

// Mock data
const administrators = ref([
  {
    id: '1',
    username: 'admin',
    displayName: 'Administrator',
    email: 'admin@example.com',
    role: 'super_admin',
    status: 'active',
    lastLogin: '2024-01-01 10:00:00',
    created_at: '2024-01-01'
  },
  {
    id: '2',
    username: 'security_admin',
    displayName: 'Security Admin',
    email: 'security@example.com',
    role: 'security_admin',
    status: 'active',
    lastLogin: '2024-01-01 09:30:00',
    created_at: '2024-01-01'
  }
])

const users = ref([
  { id: '1', username: 'admin', display_name: 'Administrator' },
  { id: '2', username: 'user1', display_name: 'John Doe' },
  { id: '3', username: 'user2', display_name: 'Jane Smith' }
])

// Table columns
const adminColumns = [
  { title: 'Username', dataIndex: 'username', key: 'username' },
  { title: 'Display Name', dataIndex: 'displayName', key: 'displayName' },
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Role', dataIndex: 'role', key: 'role' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Last Login', dataIndex: 'lastLogin', key: 'lastLogin' },
  { title: 'Actions', key: 'actions' }
]

// Forms
const adminForm = reactive({
  userId: '',
  role: 'system_admin',
  status: 'active'
})

const siteForm = reactive({
  siteName: 'EIAM Platform',
  siteUrl: 'https://eiam.example.com',
  contactEmail: 'admin@example.com',
  supportEmail: 'support@example.com',
  description: 'Enterprise Identity and Access Management Platform',
  logo: ''
})

const securityForm = reactive({
  // Password Policy
  minPasswordLength: 8,
  maxPasswordLength: 128,
  passwordExpiryDays: 90,
  requireUppercase: true,
  requireLowercase: true,
  requireNumbers: true,
  requireSpecialChars: true,
  passwordHistoryCount: 5,
  
  // Session Management
  sessionTimeout: 30,
  maxConcurrentSessions: 3,
  rememberMeDays: 7,
  
  // 2FA Configuration
  enable2FA: true,
  require2FAForAdmins: true,
  allowBackupCodes: true,
  enableTOTP: true,
  enableSMS: false,
  enableEmail: true,
  
  // Login Security
  maxLoginAttempts: 5,
  lockoutDuration: 15,
  enableIPWhitelist: false,
  enableGeolocation: true,
  enableDeviceFingerprinting: true,
  notifyFailedLogins: true,
  notifyNewDevices: true,
  notifyPasswordChanges: true
})

// Form rules
const adminRules = {
  userId: [{ required: true, message: 'Please select user' }],
  role: [{ required: true, message: 'Please select role' }],
  status: [{ required: true, message: 'Please select status' }]
}

const siteRules = {
  siteName: [{ required: true, message: 'Please enter site name' }],
  siteUrl: [{ required: true, message: 'Please enter site URL' }],
  contactEmail: [{ required: true, message: 'Please enter contact email' }]
}

const securityRules = {
  minPasswordLength: [{ required: true, message: 'Please set minimum password length' }],
  sessionTimeout: [{ required: true, message: 'Please set session timeout' }]
}

// Methods
const loadData = async () => {
  loading.value = true
  try {
    // Load site settings
    const siteSettings = await systemApi.getSiteSettings()
    Object.assign(siteForm, {
      siteName: siteSettings.site_name,
      siteUrl: siteSettings.site_url,
      contactEmail: siteSettings.contact_email,
      supportEmail: siteSettings.support_email,
      description: siteSettings.description,
      logo: siteSettings.logo
    })

    // Load security settings
    const securitySettings = await systemApi.getSecuritySettings()
    Object.assign(securityForm, {
      minPasswordLength: securitySettings.min_password_length,
      maxPasswordLength: securitySettings.max_password_length,
      passwordExpiryDays: securitySettings.password_expiry_days,
      requireUppercase: securitySettings.require_uppercase,
      requireLowercase: securitySettings.require_lowercase,
      requireNumbers: securitySettings.require_numbers,
      requireSpecialChars: securitySettings.require_special_chars,
      passwordHistoryCount: securitySettings.password_history_count,
      sessionTimeout: securitySettings.session_timeout,
      maxConcurrentSessions: securitySettings.max_concurrent_sessions,
      rememberMeDays: securitySettings.remember_me_days,
      enable2FA: securitySettings.enable_2fa,
      require2FAForAdmins: securitySettings.require_2fa_for_admins,
      allowBackupCodes: securitySettings.allow_backup_codes,
      enableTOTP: securitySettings.enable_totp,
      enableSMS: securitySettings.enable_sms,
      enableEmail: securitySettings.enable_email,
      maxLoginAttempts: securitySettings.max_login_attempts,
      lockoutDuration: securitySettings.lockout_duration,
      enableIPWhitelist: securitySettings.enable_ip_whitelist,
      enableGeolocation: securitySettings.enable_geolocation,
      enableDeviceFingerprinting: securitySettings.enable_device_fingerprinting,
      notifyFailedLogins: securitySettings.notify_failed_logins,
      notifyNewDevices: securitySettings.notify_new_devices,
      notifyPasswordChanges: securitySettings.notify_password_changes
    })
  } catch (error) {
    message.error('Failed to load system settings')
  } finally {
    loading.value = false
  }
}

const getRoleColor = (role: string) => {
  const colors = {
    super_admin: 'red',
    system_admin: 'blue',
    security_admin: 'orange'
  }
  return colors[role as keyof typeof colors] || 'default'
}

// Administrator methods
const showAdminModal = () => {
  editingAdmin.value = false
  Object.assign(adminForm, {
    userId: '',
    role: 'system_admin',
    status: 'active'
  })
  adminModalVisible.value = true
}

const editAdmin = (admin: any) => {
  editingAdmin.value = true
  Object.assign(adminForm, admin)
  adminModalVisible.value = true
}

const handleAdminSubmit = async () => {
  try {
    // TODO: Implement API call
    message.success(editingAdmin.value ? 'Administrator updated successfully' : 'Administrator added successfully')
    adminModalVisible.value = false
    loadData()
  } catch (error) {
    message.error('Failed to save administrator')
  }
}

const handleAdminCancel = () => {
  adminModalVisible.value = false
}

const removeAdmin = async (id: string) => {
  try {
    // TODO: Implement API call
    message.success('Administrator removed successfully')
    loadData()
  } catch (error) {
    message.error('Failed to remove administrator')
  }
}

const viewAdminPermissions = (admin: any) => {
  message.info(`Viewing permissions for administrator: ${admin.displayName}`)
  // TODO: Implement permission view modal
}

// Site configuration methods
const beforeLogoUpload = (file: File) => {
  const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png'
  if (!isJpgOrPng) {
    message.error('You can only upload JPG/PNG files!')
    return false
  }
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isLt2M) {
    message.error('Image must smaller than 2MB!')
    return false
  }
  return true
}

const handleLogoChange = (info: any) => {
  if (info.file.status === 'done') {
    try {
      // TODO: Implement actual file upload to server
      siteForm.logo = URL.createObjectURL(info.file.originFileObj)
      message.success('Logo uploaded successfully')
    } catch (error) {
      message.error('Failed to upload logo')
    }
  }
}

const saveSiteConfig = async () => {
  try {
    const settings = {
      site_name: siteForm.siteName,
      site_url: siteForm.siteUrl,
      contact_email: siteForm.contactEmail,
      support_email: siteForm.supportEmail,
      description: siteForm.description,
      logo: siteForm.logo
    }
    await systemApi.updateSystemSettings(settings)
    message.success('Site configuration saved successfully')
  } catch (error) {
    message.error('Failed to save site configuration')
  }
}

// Security configuration methods
const saveSecurityConfig = async () => {
  try {
    const settings = {
      min_password_length: securityForm.minPasswordLength,
      max_password_length: securityForm.maxPasswordLength,
      password_expiry_days: securityForm.passwordExpiryDays,
      require_uppercase: securityForm.requireUppercase,
      require_lowercase: securityForm.requireLowercase,
      require_numbers: securityForm.requireNumbers,
      require_special_chars: securityForm.requireSpecialChars,
      password_history_count: securityForm.passwordHistoryCount,
      session_timeout: securityForm.sessionTimeout,
      max_concurrent_sessions: securityForm.maxConcurrentSessions,
      remember_me_days: securityForm.rememberMeDays,
      enable_2fa: securityForm.enable2FA,
      require_2fa_for_admins: securityForm.require2FAForAdmins,
      allow_backup_codes: securityForm.allowBackupCodes,
      enable_totp: securityForm.enableTOTP,
      enable_sms: securityForm.enableSMS,
      enable_email: securityForm.enableEmail,
      max_login_attempts: securityForm.maxLoginAttempts,
      lockout_duration: securityForm.lockoutDuration,
      enable_ip_whitelist: securityForm.enableIPWhitelist,
      enable_geolocation: securityForm.enableGeolocation,
      enable_device_fingerprinting: securityForm.enableDeviceFingerprinting,
      notify_failed_logins: securityForm.notifyFailedLogins,
      notify_new_devices: securityForm.notifyNewDevices,
      notify_password_changes: securityForm.notifyPasswordChanges
    }
    await systemApi.updateSystemSettings(settings)
    message.success('Security configuration saved successfully')
  } catch (error) {
    message.error('Failed to save security configuration')
  }
}

// Filter methods
const filterUserOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.system-page {
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

.logo-uploader {
  width: 128px;
  height: 128px;
}

.logo-uploader .ant-upload {
  width: 128px;
  height: 128px;
}

/* Responsive Design */
@media (max-width: 768px) {
  .system-page {
    padding: 16px;
  }
  
  .tab-content {
    padding: 16px;
  }
}
</style>
