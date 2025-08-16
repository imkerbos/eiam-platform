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
    
    console.log('开始登录...')
    const response = await userStore.login(
      formData.username,
      formData.password,
      showOtp.value ? formData.otp_code : undefined
    )
    
    console.log('登录成功，响应数据:', response)
    console.log('用户存储状态:', {
      isLoggedIn: userStore.isLoggedIn,
      user: userStore.currentUser,
      token: userStore.token
    })
    
    message.success('Login successful')
    
    // Redirect to console
    console.log('准备跳转到控制台...')
    await router.push('/console')
    console.log('跳转完成')
  } catch (error: any) {
    console.error('登录失败:', error)
    
    // Handle different error types based on error code and message
    if (error.message?.includes('OTP') || error.message?.includes('require_otp')) {
      showOtp.value = true
      message.info('Please enter your OTP code')
    } else if (error.message?.includes('Account is locked') || error.message?.includes('账户已被锁定')) {
      message.error('Account is locked due to multiple failed login attempts. Please contact administrator or try again later.')
    } else if (error.message?.includes('Invalid credentials') || error.message?.includes('用户名或密码错误')) {
      message.error('Invalid username or password. Please check your credentials.')
    } else if (error.message?.includes('User inactive') || error.message?.includes('用户已停用')) {
      message.error('Your account has been deactivated. Please contact administrator.')
    } else if (error.message?.includes('Network Error') || error.message?.includes('timeout')) {
      message.error('Network connection failed. Please check your connection and try again.')
    } else {
      // Generic error message
      message.error(error.message || 'Login failed. Please try again.')
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
