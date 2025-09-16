<template>
  <div class="cas-server-info">
    <a-card title="CAS Server Information" class="mb-6">
      <template #extra>
        <a-button type="primary" @click="refreshInfo">
          <template #icon>
            <ReloadOutlined />
          </template>
          Refresh
        </a-button>
      </template>
      
      <a-spin :spinning="loading">
        <div v-if="serverInfo" class="server-info-content">
          <!-- 基本信息 -->
          <a-descriptions title="Basic Information" :column="2" bordered>
            <a-descriptions-item label="Server URL">
              <a-tag color="blue">{{ serverInfo.server_url }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Protocol Version">
              <a-tag color="green">{{ serverInfo.protocol_version }}</a-tag>
            </a-descriptions-item>
          </a-descriptions>

          <!-- CAS端点 -->
          <a-divider>CAS Endpoints</a-divider>
          <a-descriptions :column="1" bordered>
            <a-descriptions-item label="Login URL">
              <a-input 
                :value="serverInfo.login_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.login_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Validate URL (CAS 1.0)">
              <a-input 
                :value="serverInfo.validate_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.validate_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Service Validate URL (CAS 2.0)">
              <a-input 
                :value="serverInfo.service_validate_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.service_validate_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Proxy Validate URL (CAS 2.0)">
              <a-input 
                :value="serverInfo.proxy_validate_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.proxy_validate_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Proxy URL (CAS 2.0)">
              <a-input 
                :value="serverInfo.proxy_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.proxy_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Logout URL">
              <a-input 
                :value="serverInfo.logout_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.logout_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
          </a-descriptions>

          <!-- 支持的功能 -->
          <a-divider>Supported Features</a-divider>
          <div class="features-list">
            <a-tag 
              v-for="feature in serverInfo.supported_features" 
              :key="feature"
              color="cyan"
              class="feature-tag"
            >
              {{ feature }}
            </a-tag>
          </div>

          <!-- 使用说明 -->
          <a-divider>Usage Instructions</a-divider>
          <a-alert
            message="How to configure your application as a CAS Service Provider"
            type="info"
            show-icon
            class="mb-4"
          >
            <template #description>
              <div class="usage-instructions">
                <p><strong>1. Register your application:</strong></p>
                <ul>
                  <li>Go to <strong>Applications</strong> → <strong>Add Application</strong></li>
                  <li>Select <strong>CAS</strong> as the protocol type</li>
                  <li>Configure the <strong>Service URL</strong> (your application's callback URL)</li>
                  <li>Enable <strong>Gateway</strong> mode if you want automatic authentication</li>
                  <li>Enable <strong>Renew</strong> mode if you want to force re-authentication</li>
                </ul>
                
                <p><strong>2. Configure your application:</strong></p>
                <ul>
                  <li>Set the CAS server URL to: <code>{{ serverInfo.server_url }}</code></li>
                  <li>Set the login URL to: <code>{{ serverInfo.login_url }}</code></li>
                  <li>Set the validate URL to: <code>{{ serverInfo.validate_url }}</code> (CAS 1.0) or <code>{{ serverInfo.service_validate_url }}</code> (CAS 2.0)</li>
                  <li>Set the logout URL to: <code>{{ serverInfo.logout_url }}</code></li>
                </ul>
                
                <p><strong>3. Test the integration:</strong></p>
                <ul>
                  <li>Access your application with the service parameter: <code>?service=YOUR_SERVICE_URL</code></li>
                  <li>You will be redirected to the CAS login page</li>
                  <li>After successful authentication, you'll be redirected back with a ticket</li>
                  <li>Validate the ticket using the validate endpoint</li>
                </ul>
              </div>
            </template>
          </a-alert>
        </div>
      </a-spin>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { ReloadOutlined, CopyOutlined } from '@ant-design/icons-vue'
import { systemApi } from '@/api/system'

interface CASServerInfo {
  server_url: string
  login_url: string
  validate_url: string
  service_validate_url: string
  proxy_validate_url: string
  proxy_url: string
  logout_url: string
  protocol_version: string
  supported_features: string[]
}

const loading = ref(false)
const serverInfo = ref<CASServerInfo | null>(null)

const fetchServerInfo = async () => {
  loading.value = true
  try {
    const response = await fetch('/public/cas-server-info')
    const data = await response.json()
    if (data.code === 200) {
      serverInfo.value = data.data
    } else {
      message.error('Failed to fetch CAS server information')
    }
  } catch (error) {
    console.error('Error fetching CAS server info:', error)
    message.error('Failed to fetch CAS server information')
  } finally {
    loading.value = false
  }
}

const refreshInfo = () => {
  fetchServerInfo()
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    message.success('Copied to clipboard!')
  } catch (error) {
    console.error('Failed to copy to clipboard:', error)
    message.error('Failed to copy to clipboard')
  }
}

onMounted(() => {
  fetchServerInfo()
})
</script>

<style scoped>
.cas-server-info {
  padding: 24px;
}

.server-info-content {
  max-width: 100%;
}

.copy-input {
  font-family: 'Courier New', monospace;
}

.features-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.feature-tag {
  margin: 2px;
}

.usage-instructions {
  text-align: left;
}

.usage-instructions ul {
  margin: 8px 0;
  padding-left: 20px;
}

.usage-instructions li {
  margin: 4px 0;
}

.usage-instructions code {
  background-color: #f5f5f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}
</style>
