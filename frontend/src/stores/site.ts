import { defineStore } from 'pinia'
import { ref } from 'vue'
import { systemApi } from '@/api/system'

export const useSiteStore = defineStore('site', () => {
  // State
  const siteName = ref('EIAM')
  const siteUrl = ref('')
  const logoUrl = ref('/logo.svg')
  const contactEmail = ref('')
  const supportEmail = ref('')
  const description = ref('')

  // Actions
  const loadPublicSiteInfo = async () => {
    try {
      const info = await systemApi.getPublicSiteInfo()
      siteName.value = info.site_name || 'EIAM'
      logoUrl.value = info.logo_url || '/logo.svg'
    } catch (error) {
      console.error('Failed to load public site info:', error)
      // 保持默认值
    }
  }

  const loadSiteSettings = async () => {
    try {
      const settings = await systemApi.getSiteSettings()
      siteName.value = settings.site_name || 'EIAM'
      siteUrl.value = settings.site_url || ''
      logoUrl.value = settings.logo || '/logo.svg'
      contactEmail.value = settings.contact_email || ''
      supportEmail.value = settings.support_email || ''
      description.value = settings.description || ''
    } catch (error) {
      console.error('Failed to load site settings:', error)
      // 保持默认值
    }
  }

  const updateSiteSettings = (settings: any) => {
    if (settings.site_name) siteName.value = settings.site_name
    if (settings.site_url) siteUrl.value = settings.site_url
    if (settings.logo_url) logoUrl.value = settings.logo_url
    if (settings.logo) logoUrl.value = settings.logo
    if (settings.contact_email) contactEmail.value = settings.contact_email
    if (settings.support_email) supportEmail.value = settings.support_email
    if (settings.description) description.value = settings.description
  }

  return {
    // State
    siteName,
    siteUrl,
    logoUrl,
    contactEmail,
    supportEmail,
    description,

    // Actions
    loadPublicSiteInfo,
    loadSiteSettings,
    updateSiteSettings
  }
})
