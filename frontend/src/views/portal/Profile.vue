<template>
  <div class="profile-page">
    <div class="profile-header">
      <a-button @click="goBack" icon="arrow-left">
        Back
      </a-button>
      <h2>Profile</h2>
    </div>
    <a-row :gutter="24">
      <a-col :span="8">
        <a-card title="Profile Information" :bordered="false">
          <div class="profile-avatar">
            <UserAvatar 
              :size="120" 
              :user="profileData"
              :show-border="true"
            />
            <a-upload
              v-model:file-list="fileList"
              name="avatar"
              list-type="picture-card"
              class="avatar-uploader"
              :show-upload-list="false"
              :before-upload="beforeUpload"
              :custom-request="handleAvatarChange"
              accept="image/*"
            >
              <div v-if="!profileData.avatar">
                <PlusOutlined />
                <div style="margin-top: 8px">Upload</div>
              </div>
            </a-upload>
          </div>
          
          <a-descriptions :column="1" bordered>
            <a-descriptions-item label="Username">
              {{ profileData.username }}
            </a-descriptions-item>
            <a-descriptions-item label="Email">
              {{ profileData.email }}
              <a-tag v-if="profileData.emailVerified" color="green">Verified</a-tag>
              <a-tag v-else color="red">Unverified</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Display Name">
              {{ profileData.display_name }}
            </a-descriptions-item>
            <a-descriptions-item label="Phone">
              {{ profileData.phone || 'Not set' }}
              <a-tag v-if="profileData.phoneVerified" color="green">Verified</a-tag>
              <a-tag v-else-if="profileData.phone" color="red">Unverified</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Organization">
              {{ profileData.organizationName }}
            </a-descriptions-item>
            <a-descriptions-item label="Status">
              <a-tag :color="profileData.status === 'active' ? 'green' : 'red'">
                {{ profileData.status }}
              </a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="2FA Enabled">
              <a-tag :color="profileData.enableOTP ? 'green' : 'red'">
                {{ profileData.enableOTP ? 'Enabled' : 'Disabled' }}
              </a-tag>
            </a-descriptions-item>
          </a-descriptions>
        </a-card>
      </a-col>

      <a-col :span="16">
        <a-card title="Edit Profile" :bordered="false">
          <a-form
            ref="formRef"
            :model="formData"
            :rules="formRules"
            layout="vertical"
          >
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item label="Display Name" name="displayName">
                  <a-input v-model:value="formData.displayName" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Phone" name="phone">
                  <a-input v-model:value="formData.phone" />
                </a-form-item>
              </a-col>
            </a-row>
            <a-form-item label="Email" name="email">
              <a-input v-model:value="formData.email" disabled />
              <a-button v-if="!profileData.emailVerified" type="link" @click="verifyEmail">
                Verify Email
              </a-button>
            </a-form-item>
            <a-form-item>
              <a-button type="primary" @click="updateProfile">
                Update Profile
              </a-button>
            </a-form-item>
          </a-form>
        </a-card>

        <a-card title="Security Settings" :bordered="false" style="margin-top: 24px">
          <a-space direction="vertical" style="width: 100%">
            <a-button @click="showChangePasswordModal">
              Change Password
            </a-button>
            <a-button @click="showOTPSetupModal">
              {{ profileData.enableOTP ? 'Manage 2FA' : 'Setup 2FA' }}
            </a-button>
            <a-button @click="showBackupCodesModal">
              Backup Codes
            </a-button>
          </a-space>
        </a-card>

        <a-card title="Recent Activities" :bordered="false" style="margin-top: 24px">
          <a-timeline>
            <a-timeline-item v-for="activity in recentActivities" :key="activity.id">
              <template #dot>
                <component :is="activity.icon" />
              </template>
              <div class="activity-content">
                <div class="activity-title">{{ activity.title }}</div>
                <div class="activity-time">{{ activity.time }}</div>
                <div class="activity-description">{{ activity.description }}</div>
              </div>
            </a-timeline-item>
          </a-timeline>
        </a-card>
      </a-col>
    </a-row>

    <!-- Change Password Modal -->
    <a-modal
      v-model:open="passwordModalVisible"
      title="Change Password"
      @ok="handlePasswordChange"
      @cancel="handlePasswordCancel"
    >
      <a-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        layout="vertical"
      >
        <a-form-item label="Current Password" name="currentPassword">
          <a-input-password v-model:value="passwordForm.currentPassword" />
        </a-form-item>
        <a-form-item label="New Password" name="newPassword">
          <a-input-password v-model:value="passwordForm.newPassword" />
        </a-form-item>
        <a-form-item label="Confirm New Password" name="confirmPassword">
          <a-input-password v-model:value="passwordForm.confirmPassword" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- OTP Setup Modal -->
    <a-modal
      v-model:open="otpModalVisible"
      title="Two-Factor Authentication"
      width="600px"
      @ok="handleOTPSetup"
      @cancel="handleOTPCancel"
    >
      <div v-if="!profileData.enableOTP">
        <p>Scan the QR code with your authenticator app:</p>
        <div class="qr-code-container">
          <img :src="otpData.qrCode" alt="QR Code" />
        </div>
        <a-form layout="vertical">
          <a-form-item label="Verification Code">
            <a-input v-model:value="otpData.verificationCode" placeholder="Enter 6-digit code" />
          </a-form-item>
        </a-form>
      </div>
      <div v-else>
        <p>Two-factor authentication is currently enabled.</p>
        <a-button danger @click="disableOTP">
          Disable 2FA
        </a-button>
      </div>
    </a-modal>

    <!-- Backup Codes Modal -->
    <a-modal
      v-model:open="backupCodesModalVisible"
      title="Backup Codes"
      @ok="handleBackupCodesOk"
      @cancel="handleBackupCodesCancel"
    >
      <div v-if="backupCodes.length > 0">
        <p>Save these backup codes in a secure location. You can use them to access your account if you lose your 2FA device.</p>
        <a-alert
          message="Warning"
          description="Each code can only be used once. Generate new codes if you run out."
          type="warning"
          show-icon
          style="margin-bottom: 16px"
        />
        <div class="backup-codes">
          <div v-for="code in backupCodes" :key="code" class="backup-code">
            {{ code }}
          </div>
        </div>
      </div>
      <div v-else>
        <p>No backup codes available.</p>
        <a-button type="primary" @click="generateBackupCodes">
          Generate New Codes
        </a-button>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, LoginOutlined, KeyOutlined, SafetyOutlined, ArrowLeftOutlined } from '@ant-design/icons-vue'
import { http } from '@/api/request'
import { useUserStore } from '@/stores/user'
import UserAvatar from '@/components/UserAvatar.vue'

const userStore = useUserStore()
const router = useRouter()

// Data
const fileList = ref([])
const passwordModalVisible = ref(false)
const otpModalVisible = ref(false)
const backupCodesModalVisible = ref(false)
const formRef = ref()
const passwordFormRef = ref()

const profileData = reactive({
  username: userStore.currentUser?.username || 'admin',
  email: userStore.currentUser?.email || 'admin@example.com',
  emailVerified: userStore.currentUser?.email_verified || true,
  display_name: userStore.currentUser?.display_name || 'Administrator', // 改为下划线格式，匹配API和UserAvatar组件
  phone: userStore.currentUser?.phone || '+1234567890',
  phoneVerified: userStore.currentUser?.phone_verified || false,
  organizationName: 'Headquarters',
  status: userStore.currentUser?.status || 'active',
  enableOTP: userStore.currentUser?.enable_otp || false,
  avatar: userStore.currentUser?.avatar || ''
})

const formData = reactive({
  displayName: userStore.currentUser?.display_name || '',
  phone: userStore.currentUser?.phone || '',
  email: userStore.currentUser?.email || ''
})

const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const otpData = reactive({
  qrCode: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==',
  verificationCode: ''
})

const backupCodes = ref(['12345678', '87654321', '11111111', '22222222', '33333333'])

const recentActivities = ref([
  {
    id: '1',
    title: 'Login',
    time: '2024-01-01 10:00:00',
    description: 'Logged in from 192.168.1.100',
    icon: LoginOutlined
  },
  {
    id: '2',
    title: 'Password Changed',
    time: '2024-01-01 09:30:00',
    description: 'Password was changed successfully',
    icon: KeyOutlined
  },
  {
    id: '3',
    title: '2FA Setup',
    time: '2024-01-01 09:00:00',
    description: 'Two-factor authentication was enabled',
    icon: SafetyOutlined
  }
])

const formRules = {
  displayName: [{ required: true, message: 'Please input display name!' }],
  phone: [{ pattern: /^\+?[\d\s\-\(\)]+$/, message: 'Please input valid phone number!' }]
}

const passwordRules = {
  currentPassword: [{ required: true, message: 'Please input current password!' }],
  newPassword: [
    { required: true, message: 'Please input new password!' },
    { min: 8, message: 'Password must be at least 8 characters!' }
  ],
  confirmPassword: [
    { required: true, message: 'Please confirm new password!' },
    {
      validator: (rule: any, value: string) => {
        if (value !== passwordForm.newPassword) {
          return Promise.reject('Passwords do not match!')
        }
        return Promise.resolve()
      }
    }
  ]
}

// Methods
const loadProfile = async () => {
  try {
    const response = await http.get('/portal/profile')
    const data = response
    
    // Update profile data
    Object.assign(profileData, {
      username: data.username || '',
      email: data.email || '',
      display_name: data.display_name || '',
      phone: data.phone || '',
      avatar: data.avatar || '',
      status: data.status || '',
      emailVerified: data.email_verified || false,
      phoneVerified: data.phone_verified || false,
      enableOTP: data.enable_otp || false,
      organizationName: data.organization_name || ''
    })
    
    // Update form data
    Object.assign(formData, {
      displayName: data.display_name || '',
      phone: data.phone || '',
      email: data.email || ''
    })
  } catch (error) {
    message.error('Failed to load profile')
  }
}

const beforeUpload = (file: File) => {
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



const handleAvatarChange = async (options: any) => {
  const { file, onSuccess, onError } = options
  
  try {
    console.log('开始上传头像...', file)
    
    const formData = new FormData()
    formData.append('avatar', file)
    
    const response = await http.post('/portal/profile/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    console.log('头像上传响应:', response)
    
    // 更新头像URL - 检查多个可能的字段
    const avatarUrl = response.avatar || response.data?.avatar || response.avatar_url || response.data?.avatar_url
    
    if (avatarUrl) {
      profileData.avatar = avatarUrl
      
      // 同步更新用户store中的头像
      if (userStore.currentUser) {
        userStore.setUser({
          ...userStore.currentUser,
          avatar: avatarUrl
        })
      }
      
      message.success('Avatar uploaded successfully')
      
      // 通知上传成功
      onSuccess && onSuccess(response)
      
      // 重新加载profile以确保数据一致性
      await loadProfile()
    } else {
      console.error('头像上传响应数据:', response)
      const errorMsg = 'Upload successful but no avatar URL returned'
      message.error(errorMsg)
      onError && onError(new Error(errorMsg))
    }
  } catch (error: any) {
    console.error('头像上传失败:', error)
    const errorMsg = error.message || 'Failed to upload avatar'
    message.error(errorMsg)
    onError && onError(error)
  }
}

const updateProfile = async () => {
  try {
    await formRef.value?.validate()
    
    // 调用API更新用户资料
    const response = await http.put('/portal/profile', {
      display_name: formData.displayName,
      phone: formData.phone
    })
    
    if (response.code === 200) {
      // 更新本地数据
      Object.assign(profileData, {
        display_name: formData.displayName,
        phone: formData.phone
      })
      
      // 关键修复：同步更新用户store中的用户信息
      if (userStore.currentUser) {
        userStore.setUser({
          ...userStore.currentUser,
          display_name: formData.displayName,
          phone: formData.phone
        })
      }
      
      message.success('Profile updated successfully')
    } else {
      message.error(response.message || 'Failed to update profile')
    }
  } catch (error) {
    console.error('Update profile error:', error)
    message.error('Failed to update profile')
  }
}

const verifyEmail = async () => {
  try {
    // TODO: Implement API call
    message.success('Verification email sent')
  } catch (error) {
    message.error('Failed to send verification email')
  }
}

const showChangePasswordModal = () => {
  passwordModalVisible.value = true
}

const handlePasswordChange = async () => {
  try {
    await passwordFormRef.value?.validate()
    // TODO: Implement API call
    message.success('Password changed successfully')
    passwordModalVisible.value = false
    Object.assign(passwordForm, {
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    })
  } catch (error) {
    message.error('Please check the form')
  }
}

const handlePasswordCancel = () => {
  passwordModalVisible.value = false
  Object.assign(passwordForm, {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
}

const showOTPSetupModal = () => {
  otpModalVisible.value = true
}

const handleOTPSetup = async () => {
  try {
    // TODO: Implement API call
    profileData.enableOTP = true
    message.success('2FA enabled successfully')
    otpModalVisible.value = false
  } catch (error) {
    message.error('Failed to enable 2FA')
  }
}

const handleOTPCancel = () => {
  otpModalVisible.value = false
  otpData.verificationCode = ''
}

const disableOTP = async () => {
  try {
    // TODO: Implement API call
    profileData.enableOTP = false
    message.success('2FA disabled successfully')
    otpModalVisible.value = false
  } catch (error) {
    message.error('Failed to disable 2FA')
  }
}

const showBackupCodesModal = () => {
  backupCodesModalVisible.value = true
}

const handleBackupCodesOk = () => {
  backupCodesModalVisible.value = false
}

const handleBackupCodesCancel = () => {
  backupCodesModalVisible.value = false
}

const generateBackupCodes = async () => {
  try {
    // TODO: Implement API call
    backupCodes.value = ['11111111', '22222222', '33333333', '44444444', '55555555']
    message.success('New backup codes generated')
  } catch (error) {
    message.error('Failed to generate backup codes')
  }
}

// Navigation
const goBack = () => {
  router.back()
}

// Lifecycle
onMounted(() => {
  loadProfile()
})
</script>

<style scoped>
.profile-page {
  padding: 24px;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.profile-header h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  color: #1890ff;
}

.profile-header .ant-btn {
  border-radius: 6px;
  height: 36px;
  padding: 0 16px;
  font-weight: 500;
}

.profile-avatar {
  text-align: center;
  margin-bottom: 24px;
}

.avatar-uploader {
  margin-top: 16px;
}

.activity-content {
  margin-left: 16px;
}

.activity-title {
  font-weight: 500;
  color: #333;
}

.activity-time {
  font-size: 12px;
  color: #999;
  margin: 4px 0;
}

.activity-description {
  color: #666;
}

.qr-code-container {
  text-align: center;
  margin: 16px 0;
}

.qr-code-container img {
  max-width: 200px;
  border: 1px solid #d9d9d9;
}

.backup-codes {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-top: 16px;
}

.backup-code {
  background: #f5f5f5;
  padding: 8px 12px;
  text-align: center;
  font-family: monospace;
  font-weight: bold;
  border-radius: 4px;
}
</style>
