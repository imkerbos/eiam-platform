<template>
  <div class="security-page">
    <div class="page-header">
      <h2>Security Settings</h2>
      <p>Configure password policies, authentication settings, and security controls</p>
    </div>

    <a-tabs v-model:activeKey="activeTab" type="card">
      <!-- Password Policy -->
      <a-tab-pane key="password" tab="Password Policy">
        <div class="tab-content">
          <a-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            layout="vertical"
          >
            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item label="Minimum Password Length" name="minLength">
                  <a-input-number v-model:value="passwordForm.minLength" :min="6" :max="32" style="width: 100%" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Maximum Password Length" name="maxLength">
                  <a-input-number v-model:value="passwordForm.maxLength" :min="8" :max="128" style="width: 100%" />
                </a-form-item>
              </a-col>
            </a-row>

            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item label="Password Expiry (Days)" name="expiryDays">
                  <a-input-number v-model:value="passwordForm.expiryDays" :min="0" :max="365" style="width: 100%" />
                  <div class="form-hint">0 = Never expires</div>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Password History Count" name="historyCount">
                  <a-input-number v-model:value="passwordForm.historyCount" :min="0" :max="24" style="width: 100%" />
                  <div class="form-hint">Prevent reusing recent passwords</div>
                </a-form-item>
              </a-col>
            </a-row>

            <a-divider>Password Requirements</a-divider>

            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item>
                  <a-checkbox v-model:checked="passwordForm.requireUppercase">
                    Require uppercase letters (A-Z)
                  </a-checkbox>
                </a-form-item>
                <a-form-item>
                  <a-checkbox v-model:checked="passwordForm.requireLowercase">
                    Require lowercase letters (a-z)
                  </a-checkbox>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item>
                  <a-checkbox v-model:checked="passwordForm.requireNumbers">
                    Require numbers (0-9)
                  </a-checkbox>
                </a-form-item>
                <a-form-item>
                  <a-checkbox v-model:checked="passwordForm.requireSpecialChars">
                    Require special characters (!@#$%^&*)
                  </a-checkbox>
                </a-form-item>
              </a-col>
            </a-row>

            <a-form-item>
              <a-button type="primary" @click="savePasswordPolicy">
                Save Password Policy
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </a-tab-pane>

      <!-- Session Management -->
      <a-tab-pane key="session" tab="Session Management">
        <div class="tab-content">
          <a-form
            ref="sessionFormRef"
            :model="sessionForm"
            layout="vertical"
          >
            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item label="Session Timeout (Minutes)" name="timeout">
                  <a-input-number v-model:value="sessionForm.timeout" :min="5" :max="1440" style="width: 100%" />
                  <div class="form-hint">Auto logout after inactivity</div>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Max Concurrent Sessions" name="maxSessions">
                  <a-input-number v-model:value="sessionForm.maxSessions" :min="1" :max="10" style="width: 100%" />
                  <div class="form-hint">Maximum sessions per user</div>
                </a-form-item>
              </a-col>
            </a-row>

            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item label="Remember Me Duration (Days)" name="rememberMeDays">
                  <a-input-number v-model:value="sessionForm.rememberMeDays" :min="1" :max="90" style="width: 100%" />
                </a-form-item>
              </a-col>
            </a-row>

            <a-form-item>
              <a-button type="primary" @click="saveSessionSettings">
                Save Session Settings
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </a-tab-pane>

      <!-- Two-Factor Authentication -->
      <a-tab-pane key="2fa" tab="Two-Factor Authentication">
        <div class="tab-content">
          <a-form
            ref="twoFAFormRef"
            :model="twoFAForm"
            layout="vertical"
          >
            <a-form-item>
              <a-checkbox v-model:checked="twoFAForm.enable2FA">
                Enable Two-Factor Authentication
              </a-checkbox>
            </a-form-item>

            <a-form-item>
              <a-checkbox v-model:checked="twoFAForm.require2FAForAdmins">
                Require 2FA for Administrators
              </a-checkbox>
            </a-form-item>

            <a-divider>Authentication Methods</a-divider>

            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item>
                  <a-checkbox v-model:checked="twoFAForm.enableTOTP">
                    TOTP (Time-based One-Time Password)
                  </a-checkbox>
                </a-form-item>
                <a-form-item>
                  <a-checkbox v-model:checked="twoFAForm.enableSMS">
                    SMS Authentication
                  </a-checkbox>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item>
                  <a-checkbox v-model:checked="twoFAForm.enableEmail">
                    Email Authentication
                  </a-checkbox>
                </a-form-item>
                <a-form-item>
                  <a-checkbox v-model:checked="twoFAForm.allowBackupCodes">
                    Allow Backup Codes
                  </a-checkbox>
                </a-form-item>
              </a-col>
            </a-row>

            <a-form-item>
              <a-button type="primary" @click="save2FASettings">
                Save 2FA Settings
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </a-tab-pane>

      <!-- Account Security -->
      <a-tab-pane key="account" tab="Account Security">
        <div class="tab-content">
          <a-form
            ref="accountFormRef"
            :model="accountForm"
            layout="vertical"
          >
            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item label="Max Login Attempts" name="maxAttempts">
                  <a-input-number v-model:value="accountForm.maxAttempts" :min="3" :max="10" style="width: 100%" />
                  <div class="form-hint">Account locked after failed attempts</div>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="Lockout Duration (Minutes)" name="lockoutDuration">
                  <a-input-number v-model:value="accountForm.lockoutDuration" :min="5" :max="1440" style="width: 100%" />
                </a-form-item>
              </a-col>
            </a-row>

            <a-divider>Security Features</a-divider>

            <a-form-item>
              <a-checkbox v-model:checked="accountForm.enableIPWhitelist">
                Enable IP Address Whitelist
              </a-checkbox>
            </a-form-item>

            <a-form-item>
              <a-checkbox v-model:checked="accountForm.enableGeolocation">
                Enable Geolocation Verification
              </a-checkbox>
            </a-form-item>

            <a-form-item>
              <a-checkbox v-model:checked="accountForm.enableDeviceFingerprinting">
                Enable Device Fingerprinting
              </a-checkbox>
            </a-form-item>

            <a-divider>Notifications</a-divider>

            <a-form-item>
              <a-checkbox v-model:checked="accountForm.notifyFailedLogins">
                Notify on Failed Login Attempts
              </a-checkbox>
            </a-form-item>

            <a-form-item>
              <a-checkbox v-model:checked="accountForm.notifyNewDevices">
                Notify on New Device Login
              </a-checkbox>
            </a-form-item>

            <a-form-item>
              <a-button type="primary" @click="saveAccountSecurity">
                Save Account Security
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { systemApi } from '@/api/system'

// Data
const activeTab = ref('password')
const loading = ref(false)

// Form refs
const passwordFormRef = ref()
const sessionFormRef = ref()
const twoFAFormRef = ref()
const accountFormRef = ref()

// Forms
const passwordForm = reactive({
  minLength: 8,
  maxLength: 64,
  expiryDays: 90,
  historyCount: 5,
  requireUppercase: true,
  requireLowercase: true,
  requireNumbers: true,
  requireSpecialChars: true
})

const sessionForm = reactive({
  timeout: 30,
  maxSessions: 3,
  rememberMeDays: 30
})

const twoFAForm = reactive({
  enable2FA: false,
  require2FAForAdmins: true,
  enableTOTP: true,
  enableSMS: false,
  enableEmail: true,
  allowBackupCodes: true
})

const accountForm = reactive({
  maxAttempts: 5,
  lockoutDuration: 15,
  enableIPWhitelist: false,
  enableGeolocation: false,
  enableDeviceFingerprinting: false,
  notifyFailedLogins: true,
  notifyNewDevices: true
})

// Form rules
const passwordRules = {
  minLength: [{ required: true, message: 'Please set minimum password length!' }],
  maxLength: [{ required: true, message: 'Please set maximum password length!' }]
}

// Methods
const loadSecuritySettings = async () => {
  loading.value = true
  try {
    const settings = await systemApi.getSecuritySettings()
    
    // Update forms with loaded data
    Object.assign(passwordForm, {
      minLength: settings.min_password_length,
      maxLength: settings.max_password_length,
      expiryDays: settings.password_expiry_days,
      historyCount: settings.password_history_count,
      requireUppercase: settings.require_uppercase,
      requireLowercase: settings.require_lowercase,
      requireNumbers: settings.require_numbers,
      requireSpecialChars: settings.require_special_chars
    })
    
    Object.assign(sessionForm, {
      timeout: settings.session_timeout,
      maxSessions: settings.max_concurrent_sessions,
      rememberMeDays: settings.remember_me_days
    })
    
    Object.assign(twoFAForm, {
      enable2FA: settings.enable_2fa,
      require2FAForAdmins: settings.require_2fa_for_admins,
      enableTOTP: settings.enable_totp,
      enableSMS: settings.enable_sms,
      enableEmail: settings.enable_email,
      allowBackupCodes: settings.allow_backup_codes
    })
    
    Object.assign(accountForm, {
      maxAttempts: settings.max_login_attempts,
      lockoutDuration: settings.lockout_duration
    })
    
  } catch (error) {
    message.error('Failed to load security settings')
  } finally {
    loading.value = false
  }
}

const savePasswordPolicy = async () => {
  try {
    await passwordFormRef.value?.validate()
    
    const settings = {
      min_password_length: passwordForm.minLength,
      max_password_length: passwordForm.maxLength,
      password_expiry_days: passwordForm.expiryDays,
      password_history_count: passwordForm.historyCount,
      require_uppercase: passwordForm.requireUppercase,
      require_lowercase: passwordForm.requireLowercase,
      require_numbers: passwordForm.requireNumbers,
      require_special_chars: passwordForm.requireSpecialChars
    }
    
    await systemApi.updateSecuritySettings(settings)
    message.success('Password policy saved successfully')
  } catch (error) {
    message.error('Failed to save password policy')
  }
}

const saveSessionSettings = async () => {
  try {
    const settings = {
      session_timeout: sessionForm.timeout,
      max_concurrent_sessions: sessionForm.maxSessions,
      remember_me_days: sessionForm.rememberMeDays
    }
    
    await systemApi.updateSecuritySettings(settings)
    message.success('Session settings saved successfully')
  } catch (error) {
    message.error('Failed to save session settings')
  }
}

const save2FASettings = async () => {
  try {
    const settings = {
      enable_2fa: twoFAForm.enable2FA,
      require_2fa_for_admins: twoFAForm.require2FAForAdmins,
      enable_totp: twoFAForm.enableTOTP,
      enable_sms: twoFAForm.enableSMS,
      enable_email: twoFAForm.enableEmail,
      allow_backup_codes: twoFAForm.allowBackupCodes
    }
    
    await systemApi.updateSecuritySettings(settings)
    message.success('2FA settings saved successfully')
  } catch (error) {
    message.error('Failed to save 2FA settings')
  }
}

const saveAccountSecurity = async () => {
  try {
    const settings = {
      max_login_attempts: accountForm.maxAttempts,
      lockout_duration: accountForm.lockoutDuration
    }
    
    await systemApi.updateSecuritySettings(settings)
    message.success('Account security settings saved successfully')
  } catch (error) {
    message.error('Failed to save account security settings')
  }
}

// Lifecycle
onMounted(() => {
  loadSecuritySettings()
})
</script>

<style scoped>
.security-page {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
  color: #262626;
  font-size: 20px;
  font-weight: 500;
}

.page-header p {
  margin: 8px 0 0 0;
  color: #8c8c8c;
  font-size: 14px;
}

.tab-content {
  padding: 24px 0;
}

.form-hint {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 4px;
}
</style>
