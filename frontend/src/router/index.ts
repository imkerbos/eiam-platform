import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { TokenManager } from '@/utils/storage'

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
        meta: { title: 'Dashboard' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/console/Users.vue'),
        meta: { title: 'User Management' }
      },
                  {
              path: 'organizations',
              name: 'Organizations',
              component: () => import('@/views/console/Organizations.vue'),
              meta: { title: 'Organization Management' }
            },
            {
              path: 'permissions',
              name: 'Permissions',
              component: () => import('@/views/console/Permissions.vue'),
              meta: { title: 'Permissions Management' }
            },
            {
              path: 'system',
              name: 'System',
              component: () => import('@/views/console/System.vue'),
              meta: { title: 'System Management' }
            },
            {
              path: 'audit',
              name: 'Audit',
              component: () => import('@/views/console/Audit.vue'),
              meta: { title: 'Audit & Monitoring' }
            },
            {
              path: 'security',
              name: 'Security',
              component: () => import('@/views/console/Security.vue'),
              meta: { title: 'Security Settings' }
            },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('@/views/console/Roles.vue'),
        meta: { title: 'Role Management' }
      },
      {
        path: 'applications',
        name: 'Applications',
        component: () => import('@/views/console/Applications.vue'),
        meta: { title: 'Application Management' }
      },
      {
        path: 'application-groups',
        name: 'ApplicationGroups',
        component: () => import('@/views/console/ApplicationGroups.vue'),
        meta: { title: 'Application Groups' }
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
router.beforeEach((to, from, next) => {
  // Set page title
  document.title = to.meta.title ? `${to.meta.title} - EIAM Platform` : 'EIAM Platform'
  
  // Check authentication - 使用新的安全存储
  const token = TokenManager.getAccessToken()
  console.log('路由守卫检查:', {
    to: to.path,
    from: from.path,
    hasToken: !!token,
    requiresAuth: to.meta.requiresAuth
  })
  
  if (to.meta.requiresAuth && !token) {
    console.log('需要认证但无token，跳转到登录页')
    next('/login')
  } else if (to.path === '/login' && token) {
    console.log('已登录用户访问登录页，跳转到控制台')
    next('/console')
  } else {
    console.log('路由检查通过，继续导航')
    next()
  }
})

export default router
