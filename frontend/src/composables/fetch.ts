import { ref } from 'vue'

export const useFetch = (url: RequestInfo, config = {}) => {
    const data = ref<any>(null)
    const response = ref<Response | null>(null)
    const error = ref<unknown | null>(null)
    const loading = ref(false)

    const fetchWrapper = async () => {
        loading.value = true
        try {
            const result = await fetch(url, config)
            response.value = result
            data.value = result
        } catch (err: unknown) {
            error.value = err
        } finally {
            loading.value = false
        }
    }
    fetchWrapper()

    return { data, response, error, loading, fetcher: fetchWrapper }
}
