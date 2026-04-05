import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import PortalLayout from '../layouts/PortalLayout.vue'
import AdminLayout from '../layouts/AdminLayout.vue'
import LoginPage from '../features/auth/LoginPage.vue'
import PortalOverview from '../features/portal/OverviewPage.vue'
import PortalInstances from '../features/portal/InstancesPage.vue'
import PortalInstanceDetail from '../features/portal/InstanceDetailPage.vue'
import PortalJobs from '../features/portal/JobsPage.vue'
import PortalLogs from '../features/portal/LogsPage.vue'
import PortalTickets from '../features/portal/TicketsPage.vue'
import AdminOverview from '../features/admin/OverviewPage.vue'
import AdminTenants from '../features/admin/TenantsPage.vue'
import AdminInstances from '../features/admin/InstancesPage.vue'
import AdminInstanceDetail from '../features/admin/InstanceDetailPage.vue'
import AdminJobs from '../features/admin/JobsPage.vue'
import AdminAlerts from '../features/admin/AlertsPage.vue'
import AdminAudit from '../features/admin/AuditPage.vue'
import AdminTickets from '../features/admin/TicketsPage.vue'
import PortalChannels from '../features/portal/ChannelsPage.vue'
import PortalChannelDetail from '../features/portal/ChannelDetailPage.vue'
import AdminChannels from '../features/admin/ChannelsPage.vue'
import AdminChannelDetail from '../features/admin/ChannelDetailPage.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/login' },
  { path: '/login', name: 'login', component: LoginPage, meta: { theme: 'portal' } },
  {
    path: '/portal',
    component: PortalLayout,
    meta: { theme: 'portal' },
    children: [
      { path: '', name: 'portal-overview', component: PortalOverview },
      { path: 'instances', name: 'portal-instances', component: PortalInstances },
      { path: 'instances/:id', name: 'portal-instance-detail', component: PortalInstanceDetail },
      { path: 'channels', name: 'portal-channels', component: PortalChannels },
      { path: 'channels/:id', name: 'portal-channel-detail', component: PortalChannelDetail },
      { path: 'jobs', name: 'portal-jobs', component: PortalJobs },
      { path: 'tickets', name: 'portal-tickets', component: PortalTickets },
      { path: 'logs', name: 'portal-logs', component: PortalLogs },
    ],
  },
  {
    path: '/admin',
    component: AdminLayout,
    meta: { theme: 'admin' },
    children: [
      { path: '', name: 'admin-overview', component: AdminOverview },
      { path: 'tenants', name: 'admin-tenants', component: AdminTenants },
      { path: 'instances', name: 'admin-instances', component: AdminInstances },
      { path: 'instances/:id', name: 'admin-instance-detail', component: AdminInstanceDetail },
      { path: 'channels', name: 'admin-channels', component: AdminChannels },
      { path: 'channels/:id', name: 'admin-channel-detail', component: AdminChannelDetail },
      { path: 'jobs', name: 'admin-jobs', component: AdminJobs },
      { path: 'tickets', name: 'admin-tickets', component: AdminTickets },
      { path: 'alerts', name: 'admin-alerts', component: AdminAlerts },
      { path: 'audit', name: 'admin-audit', component: AdminAudit },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
