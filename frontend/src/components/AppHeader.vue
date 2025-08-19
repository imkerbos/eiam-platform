<template>
  <div class="app-header">
    <div class="header-container">
      <!-- Logo and Title -->
      <div class="header-left">
        <div class="logo" @click="goToHome" style="cursor: pointer;">
          <div class="logo-img">
            <img :src="siteStore.logoUrl" :alt="siteStore.siteName" class="logo-svg" />
          </div>
          <h1 class="platform-title">{{ siteStore.siteName }}</h1>
        </div>
      </div>

      <!-- Navigation Tabs -->
      <div class="header-center">
        <div class="nav-tabs">
          <div 
            class="nav-tab" 
            :class="{ active: currentTab === 'portal' }"
            @click="switchTab('portal')"
          >
            Portal Center
          </div>
          <div 
            class="nav-tab" 
            :class="{ active: currentTab === 'console' }"
            @click="switchTab('console')"
          >
            Backend Management
          </div>
        </div>
      </div>

      <!-- Right Side -->
      <div class="header-right">
        <!-- Utility Icons -->
        <div class="utility-icons">
          <a-button type="text" class="icon-btn">
            <InfoCircleOutlined />
          </a-button>
          <a-button type="text" class="icon-btn">
            <QuestionCircleOutlined />
          </a-button>
          <a-button type="text" class="icon-btn" @click="refresh">
            <ReloadOutlined />
          </a-button>
        </div>

        <!-- User Avatar Dropdown -->
        <div class="user-section">
          <a-dropdown :trigger="['hover', 'click']" placement="bottomRight">
            <div class="user-avatar-wrapper">
              <UserAvatar 
                :user="userStore.user"
                :size="32"
                class="user-avatar"
              />
              <span class="username">{{ userStore.user?.display_name || 'admin' }}</span>
              <DownOutlined class="dropdown-arrow" />
            </div>
            <template #overlay>
              <a-menu class="user-dropdown-menu">
                <a-menu-item key="profile" @click="goToProfile">
                  <UserOutlined />
                  <span>Profile</span>
                </a-menu-item>
                <a-menu-item key="applications" @click="goToApplications">
                  <AppstoreOutlined />
                  <span>My Applications</span>
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout" @click="handleLogout">
                  <LogoutOutlined />
                  <span>Logout</span>
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  UserOutlined,
  AppstoreOutlined,
  LogoutOutlined,
  DownOutlined,
  InfoCircleOutlined,
  QuestionCircleOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import UserAvatar from '@/components/UserAvatar.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const siteStore = useSiteStore()

// Current tab based on route
const currentTab = computed(() => {
  if (route.path.startsWith('/console')) {
    return 'console'
  }
  return 'portal'
})

// Methods
const goToHome = () => {
  // 根据当前标签跳转到对应的首页
  if (currentTab.value === 'console') {
    router.push('/console')
  } else {
    router.push('/portal/applications')
  }
}

const switchTab = (tab: string) => {
  if (tab === 'portal') {
    router.push('/portal/applications')
  } else if (tab === 'console') {
    router.push('/console')
  }
}

const goToProfile = () => {
  router.push('/portal/profile')
}

const goToApplications = () => {
  router.push('/portal/applications')
}

const handleLogout = () => {
  Modal.confirm({
    title: 'Confirm Logout',
    content: 'Are you sure you want to logout?',
    onOk: async () => {
      try {
        await userStore.logout()
        message.success('Logged out successfully')
        router.push('/login')
      } catch (error) {
        message.error('Logout failed')
      }
    }
  })
}

const refresh = () => {
  window.location.reload()
}
</script>

<style scoped>
.app-header {
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  height: 64px;
}

.header-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 24px;
  max-width: 1400px;
  margin: 0 auto;
}

/* Left Section - Logo and Title */
.header-left {
  flex: 0 0 auto;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-img {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-svg {
  width: 100%;
  height: 100%;
}

.platform-title {
  font-size: 18px;
  font-weight: 600;
  color: #1890ff;
  margin: 0;
  white-space: nowrap;
}

/* Center Section - Navigation Tabs */
.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.nav-tabs {
  display: flex;
  background: #f5f5f5;
  border-radius: 6px;
  padding: 4px;
  gap: 2px;
}

.nav-tab {
  padding: 8px 24px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #666;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.nav-tab:hover {
  color: #1890ff;
  background: rgba(24, 144, 255, 0.1);
}

.nav-tab.active {
  background: #1890ff;
  color: #fff;
}

/* Right Section - Utility Icons and User */
.header-right {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 16px;
}

.utility-icons {
  display: flex;
  gap: 8px;
}

.icon-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  transition: all 0.3s ease;
}

.icon-btn:hover {
  background: #f5f5f5;
  color: #1890ff;
}

.user-section {
  display: flex;
  align-items: center;
}

.user-avatar-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.user-avatar-wrapper:hover {
  background: #f5f5f5;
}

.user-avatar {
  /* UserAvatar组件自己处理大小，这里不需要重复设置 */
}

.username {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dropdown-arrow {
  font-size: 12px;
  color: #999;
  transition: transform 0.3s ease;
}

.user-avatar-wrapper:hover .dropdown-arrow {
  transform: rotate(180deg);
}

.user-dropdown-menu {
  min-width: 160px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.user-dropdown-menu .ant-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 14px;
}

.user-dropdown-menu .ant-menu-item:hover {
  background: #f5f5f5;
}

/* Responsive Design */
@media (max-width: 1200px) {
  .platform-title {
    font-size: 16px;
  }
  
  .nav-tab {
    padding: 6px 16px;
    font-size: 13px;
  }
}

@media (max-width: 768px) {
  .header-container {
    padding: 0 16px;
  }
  
  .platform-title {
    display: none;
  }
  
  .utility-icons {
    display: none;
  }
  
  .username {
    display: none;
  }
}
</style>
