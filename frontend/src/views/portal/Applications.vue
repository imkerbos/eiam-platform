<template>
  <div class="applications-page">
    <div class="page-header">
      <div class="header-content">
        <div class="header-text">
          <h2>My Applications</h2>
          <p>Access your authorized applications and services</p>
        </div>
        <div class="search-section">
          <a-input-search
            v-model:value="searchText"
            placeholder="Search applications..."
            style="width: 300px"
            @search="handleSearch"
            @change="handleSearchChange"
          />
        </div>
      </div>
    </div>

    <!-- Application Groups -->
    <div class="application-groups">
      <div 
        v-for="group in applicationGroups" 
        :key="group.id" 
        class="application-group"
      >
        <div class="group-header">
          <div class="group-info">
            <div class="group-icon" :style="{ backgroundColor: group.color }">
              <component :is="group.icon" />
            </div>
            <div class="group-details">
              <h3 class="group-name">{{ group.name }}</h3>
              <p class="group-description">{{ group.description }}</p>
            </div>
          </div>
          <div class="group-count">
            {{ group.applications.length }} apps
          </div>
        </div>

        <div class="applications-grid">
          <div 
            v-for="app in group.applications" 
            :key="app.id" 
            class="application-card"
            @click="openApplication(app)"
          >
            <div class="app-logo">
              <img v-if="app.logo" :src="app.logo" :alt="app.name" />
              <div v-else class="app-logo-placeholder">
                {{ app.name.charAt(0).toUpperCase() }}
              </div>
            </div>
            <div class="app-info">
              <h4 class="app-name">{{ app.name }}</h4>
              <p class="app-description">{{ app.description }}</p>
              <div class="app-meta">
                <span class="app-type">{{ app.type }}</span>
                <span class="app-status" :class="app.status">
                  {{ app.status }}
                </span>
              </div>
            </div>
            <div class="app-actions">
              <a-button type="primary" size="small">
                Launch
              </a-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="applicationGroups.length === 0" class="empty-state">
      <div class="empty-icon">
        <AppstoreOutlined />
      </div>
      <h3>No Applications Available</h3>
      <p>You don't have access to any applications yet. Contact your administrator to get access.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  AppstoreOutlined,
  ToolOutlined,
  CloudOutlined,
  DesktopOutlined,
  MobileOutlined,
  ApiOutlined
} from '@ant-design/icons-vue'

// Search functionality
const searchText = ref('')

// Application groups with mock data
const originalApplicationGroups = ref([
  {
    id: '1',
    name: 'DevOps Tools',
    description: 'Development and operations tools',
    color: '#1890ff',
    icon: ToolOutlined,
    applications: [
      {
        id: '1',
        name: 'Jenkins',
        description: 'Continuous integration and delivery',
        logo: '',
        type: 'Web',
        status: 'active',
        url: 'https://jenkins.example.com'
      },
      {
        id: '2',
        name: 'GitLab',
        description: 'Git repository management',
        logo: '',
        type: 'Web',
        status: 'active',
        url: 'https://gitlab.example.com'
      },
      {
        id: '3',
        name: 'Docker Registry',
        description: 'Container image registry',
        logo: '',
        type: 'API',
        status: 'active',
        url: 'https://registry.example.com'
      }
    ]
  },
  {
    id: '2',
    name: 'Office Applications',
    description: 'Office productivity tools',
    color: '#52c41a',
    icon: DesktopOutlined,
    applications: [
      {
        id: '4',
        name: 'Office 365',
        description: 'Microsoft Office suite',
        logo: '',
        type: 'Web',
        status: 'active',
        url: 'https://office365.example.com'
      },
      {
        id: '5',
        name: 'Google Workspace',
        description: 'Google productivity tools',
        logo: '',
        type: 'Web',
        status: 'active',
        url: 'https://workspace.google.com'
      }
    ]
  },
  {
    id: '3',
    name: 'Cloud Services',
    description: 'Cloud infrastructure and services',
    color: '#722ed1',
    icon: CloudOutlined,
    applications: [
      {
        id: '6',
        name: 'AWS Console',
        description: 'Amazon Web Services management',
        logo: '',
        type: 'Web',
        status: 'active',
        url: 'https://console.aws.amazon.com'
      },
      {
        id: '7',
        name: 'Azure Portal',
        description: 'Microsoft Azure management',
        logo: '',
        type: 'Web',
        status: 'active',
        url: 'https://portal.azure.com'
      }
    ]
  },
  {
    id: '4',
    name: 'Mobile Apps',
    description: 'Mobile applications',
    color: '#fa8c16',
    icon: MobileOutlined,
    applications: [
      {
        id: '8',
        name: 'Mobile VPN',
        description: 'Secure mobile access',
        logo: '',
        type: 'Mobile',
        status: 'active',
        url: 'mobile://vpn'
      }
    ]
  },
  {
    id: '5',
    name: 'API Services',
    description: 'Internal API services',
    color: '#eb2f96',
    icon: ApiOutlined,
    applications: [
      {
        id: '9',
        name: 'User API',
        description: 'User management API',
        logo: '',
        type: 'API',
        status: 'active',
        url: 'https://api.example.com/users'
      },
      {
        id: '10',
        name: 'Auth API',
        description: 'Authentication API',
        logo: '',
        type: 'API',
        status: 'active',
        url: 'https://api.example.com/auth'
      }
    ]
  }
])

// Computed property for filtered application groups
const applicationGroups = computed(() => {
  if (!searchText.value) {
    return originalApplicationGroups.value
  }
  
  const searchLower = searchText.value.toLowerCase()
  return originalApplicationGroups.value.map(group => {
    const filteredApps = group.applications.filter(app =>
      app.name.toLowerCase().includes(searchLower) ||
      app.description.toLowerCase().includes(searchLower) ||
      app.type.toLowerCase().includes(searchLower)
    )
    
    if (filteredApps.length === 0) {
      return null
    }
    
    return {
      ...group,
      applications: filteredApps
    }
  }).filter((group): group is NonNullable<typeof group> => group !== null)
})

// Search methods
const handleSearch = (value: string) => {
  searchText.value = value
}

const handleSearchChange = (e: any) => {
  searchText.value = e.target.value
}

// Methods
const openApplication = (app: any) => {
  try {
    if (app.type === 'Mobile') {
      // Handle mobile app launch
      message.info(`Launching ${app.name} mobile app...`)
    } else if (app.type === 'API') {
      // Handle API service
      message.info(`Accessing ${app.name} API...`)
    } else {
      // Handle web application
      window.open(app.url, '_blank')
    }
  } catch (error) {
    message.error(`Failed to launch ${app.name}`)
  }
}

// Load applications
const loadApplications = async () => {
  try {
    // TODO: Implement API call to load user's applications
    // const response = await applicationApi.getUserApplications()
    // applicationGroups.value = response.data
  } catch (error) {
    message.error('Failed to load applications')
  }
}

onMounted(() => {
  loadApplications()
})
</script>

<style scoped>
.applications-page {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 32px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.header-text {
  flex: 1;
}

.header-text h2 {
  font-size: 28px;
  font-weight: 600;
  color: #1890ff;
  margin: 0 0 8px 0;
}

.header-text p {
  color: #666;
  margin: 0;
  font-size: 16px;
}

.search-section {
  flex-shrink: 0;
}

.application-groups {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.application-group {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border-bottom: 1px solid #f0f0f0;
}

.group-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.group-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
}

.group-details {
  flex: 1;
}

.group-name {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0 0 4px 0;
}

.group-description {
  color: #666;
  margin: 0;
  font-size: 14px;
}

.group-count {
  background: #1890ff;
  color: #fff;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 12px;
  font-weight: 500;
}

.applications-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
  padding: 24px;
}

.application-card {
  background: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 12px;
}

.application-card:hover {
  border-color: #1890ff;
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.15);
  transform: translateY(-2px);
}

.app-logo {
  flex-shrink: 0;
}

.app-logo img {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: cover;
}

.app-logo-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #1890ff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
}

.app-info {
  flex: 1;
  min-width: 0;
}

.app-name {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin: 0 0 4px 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.app-description {
  font-size: 12px;
  color: #666;
  margin: 0 0 8px 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.app-meta {
  display: flex;
  gap: 8px;
  align-items: center;
}

.app-type {
  background: #f5f5f5;
  color: #666;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  text-transform: uppercase;
}

.app-status {
  font-size: 10px;
  font-weight: 500;
  padding: 2px 6px;
  border-radius: 4px;
}

.app-status.active {
  background: #f6ffed;
  color: #52c41a;
}

.app-status.inactive {
  background: #fff2f0;
  color: #ff4d4f;
}

.app-actions {
  flex-shrink: 0;
}

.empty-state {
  text-align: center;
  padding: 80px 24px;
  color: #666;
}

.empty-icon {
  font-size: 64px;
  color: #d9d9d9;
  margin-bottom: 16px;
}

.empty-state h3 {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin: 0 0 8px 0;
}

.empty-state p {
  font-size: 14px;
  margin: 0;
}

/* Responsive Design */
@media (max-width: 768px) {
  .applications-page {
    padding: 16px;
  }
  
  .applications-grid {
    grid-template-columns: 1fr;
    padding: 16px;
  }
  
  .group-header {
    padding: 16px;
  }
  
  .group-info {
    gap: 12px;
  }
  
  .group-icon {
    width: 40px;
    height: 40px;
    font-size: 16px;
  }
  
  .group-name {
    font-size: 16px;
  }
}
</style>
