<template>
  <a-layout class="console-layout">
    <!-- New Header -->
    <AppHeader />

    <a-layout>
      <!-- Sidebar -->
      <a-layout-sider
        v-model:collapsed="collapsed"
        :trigger="null"
        collapsible
        class="sidebar"
      >
        <a-menu
          v-model:selectedKeys="selectedKeys"
          v-model:openKeys="openKeys"
          mode="inline"
          theme="light"
          @click="handleMenuClick"
        >
          <a-menu-item key="dashboard">
            <DashboardOutlined />
            <span>Dashboard</span>
          </a-menu-item>
          
          <a-sub-menu key="user-management">
            <template #title>
              <UserOutlined />
              <span>User Management</span>
            </template>
            <a-menu-item key="users">Users</a-menu-item>
            <a-menu-item key="organizations">Organizations</a-menu-item>
          </a-sub-menu>
          
          <a-sub-menu key="access-control">
            <template #title>
              <SafetyOutlined />
              <span>Access Control</span>
            </template>
            <a-menu-item key="roles">Roles</a-menu-item>
            <a-menu-item key="permissions">Permissions</a-menu-item>
          </a-sub-menu>
          
          <a-sub-menu key="applications">
            <template #title>
              <AppstoreOutlined />
              <span>Applications</span>
            </template>
            <a-menu-item key="applications">Applications</a-menu-item>
            <a-menu-item key="application-groups">Application Groups</a-menu-item>
          </a-sub-menu>
          
          <a-menu-item key="audit">
            <AuditOutlined />
            <span>Audit & Monitoring</span>
          </a-menu-item>

          <a-menu-item key="security">
            <SecurityScanOutlined />
            <span>Security Settings</span>
          </a-menu-item>
          
          <a-menu-item key="system">
            <SettingOutlined />
            <span>System Settings</span>
          </a-menu-item>
        </a-menu>
      </a-layout-sider>

      <!-- Content -->
      <a-layout-content class="content">
        <div class="content-wrapper">
          <router-view />
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  DashboardOutlined,
  UserOutlined,
  SafetyOutlined,
  AppstoreOutlined,
  SettingOutlined,
  AuditOutlined,
  SecurityScanOutlined
} from '@ant-design/icons-vue'
import { useUserStore } from '@/stores/user'
import AppHeader from '@/components/AppHeader.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

// Layout state
const collapsed = ref(false)
const selectedKeys = ref<string[]>([])
const openKeys = ref<string[]>([])

// Update selected keys based on current route
watch(
  () => route.path,
  (path) => {
    const key = path.split('/').pop() || 'dashboard'
    selectedKeys.value = [key]
    
    // Set open keys for sub-menus
    if (path.includes('/users') || path.includes('/organizations')) {
      openKeys.value = ['user-management']
    } else if (path.includes('/permissions') || path.includes('/roles')) {
      openKeys.value = ['access-control']
    } else if (path.includes('/applications')) {
      openKeys.value = ['applications']
    }
    // audit, security, system现在是独立菜单项，不需要openKeys
  },
  { immediate: true }
)

// Handle menu click
const handleMenuClick = ({ key }: { key: string }) => {
  router.push(`/console/${key}`)
}

// Handle logout
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


</script>

<style scoped>
.console-layout {
  min-height: 100vh;
  padding-top: 64px; /* Account for fixed header */
}

.sidebar {
  background: #fff;
  border-right: 1px solid #f0f0f0;
}

.content {
  background: #f0f2f5;
  overflow: auto;
}

.content-wrapper {
  padding: 24px;
  min-height: calc(100vh - 64px);
}
</style>
