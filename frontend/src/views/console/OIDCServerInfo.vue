<template>
  <div class="oidc-server-info">
    <a-card title="OIDC Server Information" class="mb-6">
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
            <a-descriptions-item label="Issuer">
              <a-tag color="purple">{{ serverInfo.issuer }}</a-tag>
            </a-descriptions-item>
          </a-descriptions>

          <!-- OIDC Endpoints -->
          <a-divider>OIDC Endpoints</a-divider>
          <a-descriptions :column="1" bordered>
            <a-descriptions-item label="Discovery Endpoint">
              <a-input 
                :value="serverInfo.discovery_url" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.discovery_url)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Authorization Endpoint">
              <a-input 
                :value="serverInfo.authorization_endpoint" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.authorization_endpoint)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Token Endpoint">
              <a-input 
                :value="serverInfo.token_endpoint" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.token_endpoint)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="UserInfo Endpoint">
              <a-input 
                :value="serverInfo.userinfo_endpoint" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.userinfo_endpoint)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="JWKS Endpoint">
              <a-input 
                :value="serverInfo.jwks_uri" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.jwks_uri)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Logout Endpoint">
              <a-input 
                :value="serverInfo.end_session_endpoint" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.end_session_endpoint)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Token Introspection Endpoint">
              <a-input 
                :value="serverInfo.introspection_endpoint" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.introspection_endpoint)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
            
            <a-descriptions-item label="Token Revocation Endpoint">
              <a-input 
                :value="serverInfo.revocation_endpoint" 
                readonly 
                class="copy-input"
              >
                <template #addonAfter>
                  <a-button 
                    type="text" 
                    @click="copyToClipboard(serverInfo.revocation_endpoint)"
                    title="Copy to clipboard"
                  >
                    <CopyOutlined />
                  </a-button>
                </template>
              </a-input>
            </a-descriptions-item>
          </a-descriptions>

          <!-- Supported Response Types -->
          <a-divider>Supported Response Types</a-divider>
          <div class="response-types-list">
            <a-tag 
              v-for="type in serverInfo.supported_response_types" 
              :key="type"
              color="geekblue"
              class="type-tag"
            >
              {{ type }}
            </a-tag>
          </div>

          <!-- Supported Grant Types -->
          <a-divider>Supported Grant Types</a-divider>
          <div class="grant-types-list">
            <a-tag 
              v-for="type in serverInfo.supported_grant_types" 
              :key="type"
              color="orange"
              class="type-tag"
            >
              {{ type }}
            </a-tag>
          </div>

          <!-- Supported Scopes -->
          <a-divider>Supported Scopes</a-divider>
          <div class="scopes-list">
            <a-tag 
              v-for="scope in serverInfo.supported_scopes" 
              :key="scope"
              color="purple"
              class="scope-tag"
            >
              {{ scope }}
            </a-tag>
          </div>

          <!-- Supported Claims -->
          <a-divider>Supported Claims</a-divider>
          <div class="claims-list">
            <a-tag 
              v-for="claim in serverInfo.supported_claims" 
              :key="claim"
              color="cyan"
              class="claim-tag"
            >
              {{ claim }}
            </a-tag>
          </div>

          <!-- Supported Features -->
          <a-divider>Supported Features</a-divider>
          <div class="features-list">
            <a-tag 
              v-for="feature in serverInfo.supported_features" 
              :key="feature"
              color="green"
              class="feature-tag"
            >
              {{ feature }}
            </a-tag>
          </div>

          <!-- Technical Specifications -->
          <a-divider>Technical Specifications</a-divider>
          <a-descriptions :column="2" bordered>
            <a-descriptions-item label="Subject Types">
              <a-tag color="blue">{{ serverInfo.supported_subject_types.join(', ') }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="ID Token Signing Algorithms">
              <a-tag color="blue">{{ serverInfo.supported_id_token_signing_alg.join(', ') }}</a-tag>
            </a-descriptions-item>
          </a-descriptions>

          <!-- Usage Instructions -->
          <a-divider>Usage Instructions</a-divider>
          <a-alert
            message="How to configure your application as an OIDC client"
            type="info"
            show-icon
            class="mb-4"
          >
            <template #description>
              <div class="usage-instructions">
                <p><strong>1. Register your application:</strong></p>
                <ul>
                  <li>Go to <strong>Applications</strong> → <strong>Add Application</strong></li>
                  <li>Select <strong>OIDC</strong> as the protocol type</li>
                  <li>Configure <strong>Redirect URIs</strong> (your application's callback URLs)</li>
                  <li>Select appropriate <strong>Grant Types</strong> (Authorization Code, Implicit, etc.)</li>
                  <li>Select required <strong>Scopes</strong> (openid, profile, email, etc.)</li>
                  <li>Set access token and refresh token <strong>TTL</strong></li>
                </ul>
                
                <p><strong>2. Configure your application:</strong></p>
                <ul>
                  <li>Use discovery endpoint for auto-configuration: <code>{{ serverInfo.discovery_url }}</code></li>
                  <li>Or manually configure the following endpoints:</li>
                  <li>Set authorization endpoint to: <code>{{ serverInfo.authorization_endpoint }}</code></li>
                  <li>Set token endpoint to: <code>{{ serverInfo.token_endpoint }}</code></li>
                  <li>Set userinfo endpoint to: <code>{{ serverInfo.userinfo_endpoint }}</code></li>
                  <li>Set JWKS endpoint to: <code>{{ serverInfo.jwks_uri }}</code></li>
                  <li>Set logout endpoint to: <code>{{ serverInfo.end_session_endpoint }}</code></li>
                </ul>
                
                <p><strong>3. Implement authentication flow:</strong></p>
                <ul>
                  <li><strong>授权码流程（推荐）：</strong>重定向用户到授权端点 → 获取授权码 → 交换访问令牌</li>
                  <li><strong>隐式流程：</strong>直接在授权端点获取访问令牌（适用于SPA）</li>
                  <li><strong>混合流程：</strong>结合授权码和隐式流程的特点</li>
                  <li>使用PKCE增强安全性（推荐用于公共客户端）</li>
                </ul>
                
                <p><strong>4. 验证和使用令牌：</strong></p>
                <ul>
                  <li>验证ID Token的签名和声明</li>
                  <li>使用访问令牌调用用户信息端点获取用户详情</li>
                  <li>使用刷新令牌获取新的访问令牌</li>
                  <li>实现适当的令牌存储和管理策略</li>
                </ul>
                
                <p><strong>5. 注销处理：</strong></p>
                <ul>
                  <li>实现本地注销（清除本地会话和令牌）</li>
                  <li>实现远程注销（重定向到注销端点）</li>
                  <li>处理单点注销回调</li>
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

interface OIDCServerInfo {
  server_url: string
  issuer: string
  discovery_url: string
  authorization_endpoint: string
  token_endpoint: string
  userinfo_endpoint: string
  jwks_uri: string
  end_session_endpoint: string
  introspection_endpoint: string
  revocation_endpoint: string
  protocol_version: string
  supported_response_types: string[]
  supported_grant_types: string[]
  supported_scopes: string[]
  supported_subject_types: string[]
  supported_id_token_signing_alg: string[]
  supported_claims: string[]
  supported_features: string[]
}

const loading = ref(false)
const serverInfo = ref<OIDCServerInfo | null>(null)

const fetchServerInfo = async () => {
  loading.value = true
  try {
    const response = await fetch('/public/oidc-server-info')
    const data = await response.json()
    if (data.code === 200) {
      serverInfo.value = data.data
    } else {
      message.error('Failed to fetch OIDC server information')
    }
  } catch (error) {
    console.error('Error fetching OIDC server info:', error)
    message.error('Failed to fetch OIDC server information')
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
.oidc-server-info {
  padding: 24px;
}

.server-info-content {
  max-width: 100%;
}

.copy-input {
  font-family: 'Courier New', monospace;
}

.response-types-list,
.grant-types-list,
.scopes-list,
.claims-list,
.features-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.type-tag,
.scope-tag,
.claim-tag,
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
