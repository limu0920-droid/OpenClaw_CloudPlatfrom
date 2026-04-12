import { computed, ref, shallowRef } from 'vue'

import { api } from './api'
import type { OEMConfig } from './types'

const defaultOEMConfig: OEMConfig = {
  brand: {
    id: '0',
    code: 'openclaw',
    name: 'OpenClaw',
    status: 'active',
    logoUrl: '',
    faviconUrl: '',
    supportEmail: 'support@openclaw.local',
    supportUrl: '',
    domains: [],
  },
  theme: {
    brandId: '0',
    primaryColor: '#1d6bff',
    secondaryColor: '#5c7bff',
    accentColor: '#22c55e',
    surfaceMode: 'hybrid',
    fontFamily: 'MiSans VF',
    radius: '20px',
  },
  features: {
    brandId: '0',
    portalEnabled: true,
    adminEnabled: true,
    channelsEnabled: true,
    ticketsEnabled: true,
    purchaseEnabled: true,
    runtimeControlEnabled: true,
    auditEnabled: true,
    ssoEnabled: true,
  },
  binding: null,
}

const oemConfig = shallowRef<OEMConfig>(defaultOEMConfig)
const loaded = ref(false)
const loading = ref(false)
const error = ref('')
let loadPromise: Promise<OEMConfig> | null = null

function routeTitle(path: string) {
  if (path.startsWith('/admin')) return 'Admin'
  if (path.startsWith('/portal/artifacts')) return 'Artifacts'
  if (path.startsWith('/portal/instances')) return 'Instances'
  if (path.startsWith('/portal')) return 'Portal'
  if (path.startsWith('/login')) return 'Login'
  return 'Platform'
}

function updateFavicon(href?: string) {
  if (typeof document === 'undefined' || !href) return

  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  link.href = href
}

function applyBrandTheme(themeName: string) {
  if (typeof document === 'undefined') return

  const root = document.documentElement
  const theme = oemConfig.value.theme ?? defaultOEMConfig.theme!
  const brand = oemConfig.value.brand ?? defaultOEMConfig.brand!
  root.style.setProperty('--brand', theme.primaryColor || defaultOEMConfig.theme!.primaryColor!)
  root.style.setProperty('--brand-strong', theme.secondaryColor || defaultOEMConfig.theme!.secondaryColor!)
  root.style.setProperty('--accent', theme.accentColor || defaultOEMConfig.theme!.accentColor!)
  root.style.setProperty('--font-sans', `"${theme.fontFamily || defaultOEMConfig.theme!.fontFamily}" , "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif`)
  if (theme.radius) {
    root.style.setProperty('--radius-lg', theme.radius)
    root.style.setProperty('--radius-md', theme.radius)
  }

  const surfaceMode = theme.surfaceMode || 'hybrid'
  const shouldUseDark = themeName === 'admin'
    ? surfaceMode !== 'light'
    : surfaceMode === 'dark'
  root.classList.toggle('dark', shouldUseDark)

  if (brand.faviconUrl) {
    updateFavicon(brand.faviconUrl)
  }
}

export async function ensureBrandingLoaded(force = false) {
  if (loaded.value && !force) {
    return oemConfig.value
  }
  if (loadPromise && !force) {
    return loadPromise
  }

  loading.value = true
  error.value = ''
  loadPromise = api.getOEMConfig(typeof window === 'undefined' ? undefined : window.location.host)
    .then((response) => {
      oemConfig.value = {
        brand: response.brand ?? defaultOEMConfig.brand,
        theme: response.theme ?? defaultOEMConfig.theme,
        features: response.features ?? defaultOEMConfig.features,
        binding: response.binding,
      }
      loaded.value = true
      return oemConfig.value
    })
    .catch((err) => {
      error.value = err instanceof Error ? err.message : '加载品牌配置失败'
      oemConfig.value = defaultOEMConfig
      loaded.value = true
      return oemConfig.value
    })
    .finally(() => {
      loading.value = false
      loadPromise = null
    })

  return loadPromise
}

export function useBranding() {
  return {
    oemConfig,
    brand: computed(() => oemConfig.value.brand ?? defaultOEMConfig.brand!),
    theme: computed(() => oemConfig.value.theme ?? defaultOEMConfig.theme!),
    features: computed(() => oemConfig.value.features ?? defaultOEMConfig.features!),
    loaded,
    loading,
    error,
  }
}

export function applyBrandingForRoute(path: string, themeName: string) {
  applyBrandTheme(themeName)
  if (typeof document === 'undefined') return
  const brandName = oemConfig.value.brand?.name || defaultOEMConfig.brand!.name
  document.title = `${routeTitle(path)} · ${brandName}`
}

export function canAccessBrandRoute(path: string) {
  const features = oemConfig.value.features ?? defaultOEMConfig.features!
  if (path.startsWith('/admin') && !features.adminEnabled) {
    return false
  }
  if (path.startsWith('/portal') && !features.portalEnabled) {
    return false
  }
  if (path.includes('/channels') && !features.channelsEnabled) {
    return false
  }
  if (path.includes('/tickets') && !features.ticketsEnabled) {
    return false
  }
  if (path.includes('/audit') && !features.auditEnabled) {
    return false
  }
  return true
}
