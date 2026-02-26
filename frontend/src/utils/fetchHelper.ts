// API request helper
/**
 * Makes a generic HTTP request and parses the response.
 * @template T - The expected response data type
 * @param input - The URL or RequestInfo object for the request
 * @param init - Optional RequestInit object to configure the request (method, headers, body, etc.)
 * @returns A promise that resolves to the parsed response data of type T
 * @throws Error if the response status is not ok, or if the response body cannot be parsed as JSON
 * @remarks
 * - Handles 204 No Content responses by returning undefined
 * - Attempts to extract error messages from the API response
 * - Returns undefined if the response body is empty but status is ok
 */
export async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const res = await fetch(input, init)

  if (!res.ok) {
    const text = await res.text().catch(() => '')
    try {
      const data = text ? JSON.parse(text) : null
      const apiMsg = data?.error?.message
      if (apiMsg) throw new Error(apiMsg)
    } catch {
      // fall through
    }
    throw new Error(text || `Response status: ${res.status}`)
  }

  // Handle 204 No Content
  if (res.status === 204) {
    return undefined as unknown as T
  }

  // avoids JSON parse crash if server returns no body but not 204
  const text = await res.text()
  if (!text) {
    return undefined as unknown as T
  }

  return JSON.parse(text) as T
}
