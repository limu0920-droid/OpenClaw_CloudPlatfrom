import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { api } from '../lib/api'
import type { AuthSession } from '../lib/types'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'marketing-home',
    component: () => import('../features/marketing/HomePage.vue'),
    meta: { theme: 'marketing', requiredRole: 'guest' },
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../features/auth/LoginPage.vue'),
    meta: { theme: 'portal', requiredRole: 'guest' },
  },
  {
    path: '/portal',
    component: () => import('../layouts/PortalLayout.vue'),
    meta: { theme: 'portal', requiredRole: 'user' },
    children: [
      { path: '', name: 'portal-overview', component: () => import('../features/portal/OverviewPage.vue') },
      { path: 'instances', name: 'portal-instances', component: () => import('../features/portal/InstancesPage.vue') },
      {
        path: 'instances/:id',
        name: 'portal-instance-detail',
        component: () => import('../features/portal/InstanceDetailPage.vue'),
      },
      {
        path: 'instances/:id/workspace',
        name: 'portal-instance-workspace',
        component: () => import('../features/workspace/InstanceWorkspacePage.vue'),
      },
      { path: 'operations', name: 'portal-operations', component: () => import('../features/portal/OperationsPage.vue') },
      { path: 'artifacts', name: 'portal-artifacts', component: () => import('../features/artifacts/ArtifactsPage.vue') },
      {
        path: 'artifacts/:id',
        name: 'portal-artifact-detail',
        component: () => import('../features/artifacts/ArtifactDetailPage.vue'),
      },
      { path: 'channels', name: 'portal-channels', component: () => import('../features/portal/ChannelsPage.vue') },
      {
        path: 'channels/:id',
        name: 'portal-channel-detail',
        component: () => import('../features/portal/ChannelDetailPage.vue'),
      },
      { path: 'jobs', name: 'portal-jobs', component: () => import('../features/portal/JobsPage.vue') },
      { path: 'tickets', name: 'portal-tickets', component: () => import('../features/portal/TicketsPage.vue') },
      { path: 'logs', name: 'portal-logs', component: () => import('../features/portal/LogsPage.vue') },
    ],
  },
  {
    path: '/admin',
    component: () => import('../layouts/AdminLayout.vue'),
    meta: { theme: 'admin', requiredRole: 'admin' },
    children: [
      { path: '', name: 'admin-overview', component: () => import('../features/admin/OverviewPage.vue') },
      { path: 'tenants', name: 'admin-tenants', component: () => import('../features/admin/TenantsPage.vue') },
      { path: 'instances', name: 'admin-instances', component: () => import('../features/admin/InstancesPage.vue') },
      {
        path: 'instances/:id',
        name: 'admin-instance-detail',
        component: () => import('../features/admin/InstanceDetailPage.vue'),
      },
      {
        path: 'instances/:id/workspace',
        name: 'admin-instance-workspace',
        component: () => import('../features/workspace/InstanceWorkspacePage.vue'),
      },
      { path: 'artifacts', name: 'admin-artifacts', component: () => import('../features/artifacts/ArtifactsPage.vue') },
      {
        path: 'artifacts/:id',
        name: 'admin-artifact-detail',
        component: () => import('../features/artifacts/ArtifactDetailPage.vue'),
      },
      { path: 'channels', name: 'admin-channels', component: () => import('../features/admin/ChannelsPage.vue') },
      {
        path: 'channels/:id',
        name: 'admin-channel-detail',
        component: () => import('../features/admin/ChannelDetailPage.vue'),
      },
      { path: 'jobs', name: 'admin-jobs', component: () => import('../features/admin/JobsPage.vue') },
      { path: 'tickets', name: 'admin-tickets', component: () => import('../features/admin/TicketsPage.vue') },
      { path: 'oem/brands', name: 'admin-oem-brands', component: () => import('../features/admin/OEMBrandsPage.vue') },
      { path: 'approvals', name: 'admin-approvals', component: () => import('../features/admin/ApprovalsPage.vue') },
      { path: 'diagnostics', name: 'admin-diagnostics', component: () => import('../features/admin/DiagnosticsPage.vue') },
      { path: 'alerts', name: 'admin-alerts', component: () => import('../features/admin/AlertsPage.vue') },
      { path: 'audit', name: 'admin-audit', component: () => import('../features/admin/AuditPage.vue') },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  const requiredRole = to.meta.requiredRole as string || 'user'
  
  // 公开页面不需要认证
  if (requiredRole === 'guest') {
    next()
    return
  }
  
  try {
    const session: AuthSession = await api.getAuthSession()
    
    if (!session.authenticated) {
      next('/login')
      return
    }
    
    const userRole = session.user?.role || ''
    
    // 权限检查
    if (requiredRole === 'admin') {
      if (userRole !== 'admin') {
        next('/portal')
        return
      }
    } else if (requiredRole === 'user') {
      if (!['admin', 'portal', 'user'].includes(userRole)) {
        next('/login')
        return
      }
    }
    
    next()
  } catch (error) {
    next('/login')
  }
})

export default router
