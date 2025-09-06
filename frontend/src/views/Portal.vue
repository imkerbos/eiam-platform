<template>
  <a-layout class="portal-layout">
    <!-- New Header -->
    <AppHeader />

    <a-layout-content class="portal-content">
      <router-view />
    </a-layout-content>

    <a-layout-footer class="portal-footer">
      <div class="footer-content">
        <p>&copy; 2024 EIAM Platform. All rights reserved.</p>
      </div>
    </a-layout-footer>
  </a-layout>
</template>

<script setup lang="ts">
// import { useRouter } from 'vue-router'
// import { message } from 'ant-design-vue'
// import { useUserStore } from '@/stores/user'
import AppHeader from '@/components/AppHeader.vue'

// const router = useRouter()
// const userStore = useUserStore()
</script>

<style scoped>
.portal-layout {
  min-height: 100vh;
  padding-top: 64px; /* Account for fixed header */
}

.portal-content {
  min-height: calc(100vh - 64px - 70px);
  background: #f0f2f5;
}

.portal-footer {
  background: #fff;
  text-align: center;
  padding: 16px 24px;
}

.footer-content p {
  margin: 0;
  color: #666;
}
</style>
