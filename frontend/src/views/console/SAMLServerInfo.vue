<template>
  <div class="saml-server-info">
    <a-card title="SAML Server Information" class="mb-6">
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
          <!-- Basic Information -->
          <a-descriptions title="Basic Information" :column="2" bordered>
            <a-descriptions-item label="Server URL">
              <a-tag color="blue">{{ serverInfo.server_url }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Protocol Version">
              <a-tag color="green">{{ serverInfo.protocol_version }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Entity ID">
              <a-tag color="purple">{{ serverInfo.entity_id }}</a-tag>
            </a-descriptions-item>
          </a-descriptions>

          <!-- SAML Endpoints -->
          <a-divider>SAML Endpoints</a-divider>
          <a-descriptions :column="1" bordered>
            <a-descriptions-item label="Metadata URL">
              <a-input 
                :value="serverInfo.metadata_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.metadata_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Single Sign-On URL (SSO)">
              <a-input 
                :value="serverInfo.sso_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.sso_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Single Logout URL (SLS)">
              <a-input 
                :value="serverInfo.sls_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.sls_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Artifact Resolution URL">
              <a-input 
                :value="serverInfo.artifact_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.artifact_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
          </a-descriptions>

          <!-- Supported Bindings -->
          <a-divider>Supported Bindings</a-divider>
          <div class="bindings-list">
            <a-tag 
              v-for="binding in serverInfo.supported_bindings" 
              :key="binding"
              color="geekblue"
              class="binding-tag"
            >
              {{ binding }}
            </a-tag>
          </div>

          <!-- Supported Features -->
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

          <!-- Certificate Information -->
          <a-divider>Certificate Information</a-divider>
          <a-descriptions :column="2" bordered>
            <a-descriptions-item label="Signing Certificate">
              <a-tag color="orange">{{ serverInfo.certificate_info.signing_certificate }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Encryption Certificate">
              <a-tag color="orange">{{ serverInfo.certificate_info.encryption_certificate }}</a-tag>
            </a-descriptions-item>
          </a-descriptions>

          <!-- Usage Instructions -->
          <a-divider>Usage Instructions</a-divider>
          <a-alert
            message="How to configure your application as a SAML Service Provider"
            type="info"
            show-icon
            class="mb-4"
          >
            <template #description>
              <div class="usage-instructions">
                <p><strong>1. Register your application:</strong></p>
                <ul>
                  <li>Go to <strong>Applications</strong> → <strong>Add Application</strong></li>
                  <li>Select <strong>SAML</strong> as the protocol type</li>
                  <li>Configure <strong>Entity ID</strong> (your application's unique identifier)</li>
                  <li>Configure <strong>ACS URL</strong> (Assertion Consumer Service URL)</li>
                  <li>Configure <strong>SLO URL</strong> (Single Logout URL, optional)</li>
                  <li>Upload or paste your application's <strong>certificate</strong> (for encrypting assertions)</li>
                </ul>
                
                <p><strong>2. Configure your application:</strong></p>
                <ul>
                  <li>Download SAML metadata: <code>{{ serverInfo.metadata_url }}</code></li>
                  <li>Or manually configure the following endpoints:</li>
                  <li>Set IdP Entity ID to: <code>{{ serverInfo.entity_id }}</code></li>
                  <li>Set SSO URL to: <code>{{ serverInfo.sso_url }}</code></li>
                  <li>Set SLO URL to: <code>{{ serverInfo.sls_url }}</code></li>
                  <li>Get signing and encryption certificates from metadata</li>
                </ul>
                
                <p><strong>3. Test integration:</strong></p>
                <ul>
                  <li>Initiate SAML authentication request from your application</li>
                  <li>Users will be redirected to the SAML login page</li>
                  <li>After successful authentication, users will be redirected back to your application with SAML assertion</li>
                  <li>Validate the SAML assertion signature and content</li>
                </ul>
                
                <p><strong>4. Attribute mapping:</strong></p>
                <ul>
                  <li>Configure user attribute mapping to get required user information</li>
                  <li>Common attributes include: username, email, name, department, etc.</li>
                  <li>Support for custom attribute mapping rules</li>
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

interface SAMLServerInfo {
  server_url: string
  entity_id: string
  metadata_url: string
  sso_url: string
  sls_url: string
  artifact_url: string
  protocol_version: string
  supported_bindings: string[]
  supported_features: string[]
  certificate_info: {
    signing_certificate: string
    encryption_certificate: string
  }
}

const loading = ref(false)
const serverInfo = ref<SAMLServerInfo | null>(null)

const fetchServerInfo = async () => {
  loading.value = true
  try {
    const response = await fetch('/public/saml-server-info')
    const data = await response.json()
    if (data.code === 200) {
      serverInfo.value = data.data
    } else {
      message.error('Failed to fetch SAML server information')
    }
  } catch (error) {
    console.error('Error fetching SAML server info:', error)
    message.error('Failed to fetch SAML server information')
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
.saml-server-info {
  padding: 24px;
}

.server-info-content {
  max-width: 100%;
}

.copy-input {
  font-family: 'Courier New', monospace;
}

.bindings-list,
.features-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.binding-tag,
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
