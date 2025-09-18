// 简单的 fetch 封装，适配后端 {code,message,data} 响应格式
export type ApiResult<T> = {
  code: number
  message: string
  data: T
}

function getToken() {
  return localStorage.getItem('token') || ''
}

async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Accept': 'application/json',
  }

  // 合并 headers
  if (init && init.headers) {
    Object.assign(headers, init.headers as Record<string, string>)
  }

  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const resp = await fetch(input, {
    ...init,
    headers,
  })

  const contentType = resp.headers.get('content-type') || ''
  if (!resp.ok) {
    const text = await resp.text()
    throw new Error(`HTTP ${resp.status}: ${text}`)
  }

  if (contentType.includes('application/json')) {
    const body = await resp.json()
    
    // 如果响应直接是数组或其他数据，直接返回
    if (Array.isArray(body) || typeof body !== 'object' || !body.hasOwnProperty('code')) {
      return body as T
    }
    
    // 处理标准的{code, message, data}格式
    if (body.code && body.code !== 200) {
      throw new Error(body.message || 'api error')
    }
    return body.data as T
  }

  // 非JSON返回，直接返回文本
  const text = await resp.text()
  return text as unknown as T
}

export const api = {
  get: <T = any>(url: string) => request<T>(url, { method: 'GET' }),
  post: <T = any>(url: string, data?: any) => {
    const isForm = data instanceof FormData
    const headersRecord: Record<string, string> = {}
    if (!isForm) headersRecord['Content-Type'] = 'application/json'
    const body = isForm ? data : JSON.stringify(data)
    return request<T>(url, { method: 'POST', body, headers: headersRecord })
  },
  delete: <T = any>(url: string) => request<T>(url, { method: 'DELETE' }),
  fs: {
    rename: (path: string, name: string) => {
      return api.post('/api/fs/rename', { path, name })
    },
    remove: (path: string) => {
      return api.post('/api/fs/remove', { path })
    },
    move: (srcPath: string, dstPath: string) => {
      return api.post('/api/fs/move', { src_path: srcPath, dst_path: dstPath })
    },
  },
  storage: {
    create: (storage: any) => {
      return api.post('/api/storage/create', storage)
    },
    getAll: () => {
      return api.get('/api/storage/all')
    },
    load: (storage: any) => {
      return api.post('/api/storage/load', storage)
    },
    delete: (id: string) => {
      return api.delete(`/api/storage/${id}`)
    },
  },
}
