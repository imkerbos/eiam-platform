<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>EIAM Platform</h1>
        <p>Enterprise Identity and Access Management</p>
      </div>
      
      <a-form
        :model="formData"
        :rules="rules"
        @finish="handleLogin"
        layout="vertical"
        class="login-form"
      >
        <a-form-item name="username" label="Username">
          <a-input
            v-model:value="formData.username"
            size="large"
            placeholder="Enter your username"
          >
            <template #prefix>
              <UserOutlined />
            </template>
          </a-input>
        </a-form-item>
        
        <a-form-item name="password" label="Password">
          <a-input-password
            v-model:value="formData.password"
            size="large"
            placeholder="Enter your password"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>
        
        <a-form-item v-if="showOtp" name="otp_code" label="OTP Code">
          <a-input
            v-model:value="formData.otp_code"
            size="large"
            placeholder="Enter OTP code"
            maxlength="6"
          >
            <template #prefix>
              <SafetyOutlined />
            </template>
          </a-input>
        </a-form-item>
        
        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
            block
          >
            {{ showOtp ? 'Verify & Login' : 'Login' }}
          </a-button>
        </a-form-item>
        
        <div class="login-options">
          <a-checkbox v-model:checked="rememberMe">
            Remember me
          </a-checkbox>
          <a @click="forgotPassword">Forgot password?</a>
        </div>
      </a-form>
      
      <div class="login-footer">
        <p>&copy; 2024 EIAM Platform. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined, SafetyOutlined } from '@ant-design/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

// Form data
const formData = reactive({
  username: '',
  password: '',
  otp_code: ''
})

// Form state
const loading = ref(false)
const showOtp = ref(false)
const rememberMe = ref(false)

// Form validation rules
const rules = {
  username: [
    { required: true, message: 'Please enter your username' },
    { min: 3, message: 'Username must be at least 3 characters' }
  ],
  password: [
    { required: true, message: 'Please enter your password' },
    { min: 6, message: 'Password must be at least 6 characters' }
  ],
  otp_code: [
    { required: true, message: 'Please enter OTP code', trigger: 'blur' }
  ]
}

// Handle login
const handleLogin = async () => {
  try {
    loading.value = true
    
    const response = await userStore.login(
      formData.username,
      formData.password,
      showOtp.value ? formData.otp_code : undefined
    )
    
    message.success('Login successful')
    
    // Redirect to console
    router.push('/console')
  } catch (error: any) {
    // Handle OTP requirement
    if (error.message?.includes('OTP')) {
      showOtp.value = true
      message.info('Please enter your OTP code')
    } else {
      message.error(error.message || 'Login failed')
    }
  } finally {
    loading.value = false
  }
}

// Handle forgot password
const forgotPassword = () => {
  message.info('Forgot password feature coming soon')
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  padding: 40px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  color: #1890ff;
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 8px;
}

.login-header p {
  color: #666;
  font-size: 14px;
  margin: 0;
}

.login-form {
  margin-bottom: 24px;
}

.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}

.login-footer {
  text-align: center;
  color: #999;
  font-size: 12px;
}
</style>
