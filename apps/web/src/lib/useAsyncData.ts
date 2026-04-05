import { onMounted, ref } from 'vue'

export function useAsyncData<T>(loader: () => Promise<T>) {
  const data = ref<T | null>(null)
  const error = ref('')
  const loading = ref(true)

  const load = async () => {
    loading.value = true
    error.value = ''

    try {
      data.value = await loader()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载失败'
    } finally {
      loading.value = false
    }
  }

  onMounted(load)

  return {
    data,
    error,
    loading,
    reload: load,
  }
}
