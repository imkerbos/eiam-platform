<template>
  <div class="applications-page">
    <a-tabs v-model:activeKey="activeTab" type="card">
      <a-tab-pane key="applications" tab="应用程序管理">
        <ApplicationsManagement />
      </a-tab-pane>
      <a-tab-pane key="cas-info" tab="CAS 服务器信息">
        <CASServerInfo />
      </a-tab-pane>
      <a-tab-pane key="saml-info" tab="SAML 服务器信息">
        <SAMLServerInfo />
      </a-tab-pane>
      <a-tab-pane key="oidc-info" tab="OIDC 服务器信息">
        <OIDCServerInfo />
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ApplicationsManagement from './ApplicationsManagement.vue'
import CASServerInfo from './CASServerInfo.vue'
import SAMLServerInfo from './SAMLServerInfo.vue'
import OIDCServerInfo from './OIDCServerInfo.vue'

const activeTab = ref('applications')
</script>

<style scoped>
.applications-page {
  padding: 24px;
}
</style>