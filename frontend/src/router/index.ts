import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { TokenManager } from '@/utils/storage'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/login'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: 'Login', requiresAuth: false }
  },
  {
    path: '/console',
    component: () => import('@/views/Console.vue'),
    meta: { title: 'Console', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Console',
        redirect: '/console/dashboard'
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/console/Dashboard.vue'),
        meta: { title: 'Dashboard', requiresAdmin: true }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/console/Users.vue'),
        meta: { title: 'User Management', requiresAdmin: true }
      },
                  {
              path: 'organizations',
              name: 'Organizations',
              component: () => import('@/views/console/Organizations.vue'),
              meta: { title: 'Organization Management', requiresAdmin: true }
            },
            {
              path: 'permissions',
              name: 'Permissions',
              component: () => import('@/views/console/Permissions.vue'),
              meta: { title: 'Permissions Management', requiresAdmin: true }
            },
            {
              path: 'system',
              name: 'System',
              component: () => import('@/views/console/System.vue'),
              meta: { title: 'System Management', requiresAdmin: true }
            },
            {
              path: 'audit',
              name: 'Audit',
              component: () => import('@/views/console/Audit.vue'),
              meta: { title: 'Audit & Monitoring', requiresAdmin: true }
            },
            {
              path: 'security',
              name: 'Security',
              component: () => import('@/views/console/Security.vue'),
              meta: { title: 'Security Settings', requiresAdmin: true }
            },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('@/views/console/Roles.vue'),
        meta: { title: 'Role Management', requiresAdmin: true }
      },
      {
        path: 'applications',
        name: 'Applications',
        component: () => import('@/views/console/Applications.vue'),
        meta: { title: 'Application Management', requiresAdmin: true }
      },
      {
        path: 'application-groups',
        name: 'ApplicationGroups',
        component: () => import('@/views/console/ApplicationGroups.vue'),
        meta: { title: 'Application Groups', requiresAdmin: true }
      }
    ]
  },
  {
    path: '/portal',
    component: () => import('@/views/Portal.vue'),
    meta: { title: 'Portal', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Portal',
        redirect: '/portal/applications'
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/portal/Profile.vue'),
        meta: { title: 'Profile' }
      },
      {
        path: 'applications',
        name: 'UserApplications',
        component: () => import('@/views/portal/Applications.vue'),
        meta: { title: 'My Applications' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
  // Set page title - 使用动态站点名称
  const siteStore = useSiteStore()
  const siteName = siteStore.siteName || 'EIAM Platform'
  document.title = to.meta.title ? `${to.meta.title} - ${siteName}` : siteName
  
  // Check authentication - 使用新的安全存储
  const token = TokenManager.getAccessToken()
  const userStore = useUserStore()
  
  console.log('路由守卫检查:', {
    to: to.path,
    from: from.path,
    hasToken: !!token,
    requiresAuth: to.meta.requiresAuth,
    requiresAdmin: to.meta.requiresAdmin,
    isAdmin: userStore.isAdmin,
    userRoles: userStore.userRoles,
    user: userStore.user
  })
  
  // 如果有token但没有用户信息，尝试获取用户信息
  if (token && !userStore.user) {
    try {
      console.log('尝试获取用户信息...')
      await userStore.getCurrentUser()
      console.log('用户信息获取成功:', userStore.user)
    } catch (error) {
      console.error('获取用户信息失败:', error)
      // 如果获取用户信息失败，清除token并跳转到登录页
      userStore.clearAuth()
      next('/login')
      return
    }
  }
  
  if (to.meta.requiresAuth && !token) {
    console.log('需要认证但无token，跳转到登录页')
    next('/login')
  } else if (to.meta.requiresAdmin && !userStore.isAdmin) {
    console.log('需要管理员权限但用户不是管理员，跳转到Portal')
    console.log('用户角色:', userStore.userRoles)
    next('/portal')
  } else if (to.path === '/login' && token) {
    console.log('已登录用户访问登录页，跳转到控制台')
    next('/console')
  } else {
    console.log('路由检查通过，继续导航')
    next()
  }
})

export default router
