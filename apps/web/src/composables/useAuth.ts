import { ref, computed, type Ref, type ComputedRef } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../lib/api'
import type { AuthSession } from '../lib/types'

interface UseAuthReturn {
  session: Ref<AuthSession | null>
  loading: Ref<boolean>
  error: Ref<string | null>
  isAuthenticated: ComputedRef<boolean>
  isAdmin: ComputedRef<boolean>
  isPortalUser: ComputedRef<boolean>
  checkPermission: (requiredRole: string) => boolean
  refreshSession: () => Promise<void>
  logout: () => Promise<void>
}

export function useAuth(): UseAuthReturn {
  const router = useRouter()
  const session = ref<AuthSession | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => session.value?.authenticated ?? false)
  const isAdmin = computed(() => session.value?.user?.role === 'admin')
  const isPortalUser = computed(() => session.value?.user?.role === 'portal' || session.value?.user?.role === 'user')

  async function fetchSession() {
    loading.value = true
    error.value = null
    try {
      session.value = await api.getAuthSession()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取会话信息失败'
      session.value = null
    } finally {
      loading.value = false
    }
  }

  function checkPermission(requiredRole: string): boolean {
    if (!session.value?.authenticated) {
      return false
    }
    
    const userRole = session.value.user?.role
    
    // 角色权限层级：admin > portal > user
    if (requiredRole === 'admin') {
      return userRole === 'admin'
    }
    if (requiredRole === 'portal') {
      return userRole === 'admin' || userRole === 'portal'
    }
    if (requiredRole === 'user') {
      return userRole === 'admin' || userRole === 'portal' || userRole === 'user'
    }
    
    return false
  }

  async function refreshSession() {
    await fetchSession()
  }

  async function logout() {
    loading.value = true
    error.value = null
    try {
      await api.logout()
      session.value = null
      await router.push('/login')
    } catch (err) {
      error.value = err instanceof Error ? err.message : '登出失败'
    } finally {
      loading.value = false
    }
  }

  // 初始化时获取会话信息
  fetchSession()

  return {
    session,
    loading,
    error,
    isAuthenticated,
    isAdmin,
    isPortalUser,
    checkPermission,
    refreshSession,
    logout
  }
}
