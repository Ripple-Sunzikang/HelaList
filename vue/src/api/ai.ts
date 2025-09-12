import type { ApiResult } from './index'

export interface ChatMessage {
  type: 'user' | 'ai'
  content: string
  timestamp: Date
  action?: {
    type: string
    params: any
    result?: any
  }
}

export interface AIResponse {
  reply: string
  actions?: {
    type: string
    params: any
  }[]
  error?: string
}

async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
  }

  // 合并 headers
  if (init && init.headers) {
    Object.assign(headers, init.headers as Record<string, string>)
  }

  const token = localStorage.getItem('token') || ''
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
    const body = (await resp.json()) as ApiResult<any>
    if (body.code && body.code !== 200) {
      throw new Error(body.message || 'api error')
    }
    return body.data as T
  }

  // 非JSON返回，直接返回文本
  const text = await resp.text()
  return text as unknown as T
}

export const aiApi = {
  // AI 聊天接口
  async chat(message: string, history: ChatMessage[]): Promise<AIResponse> {
    try {
      const response = await request<AIResponse>('/api/ai/chat', {
        method: 'POST',
        body: JSON.stringify({
          message,
          history: history.slice(-10) // 只保留最近10条消息作为上下文
        })
      })
      
      // 检查是否有错误
      if (response.error) {
        throw new Error(response.error)
      }
      
      return response
    } catch (error) {
      console.error('AI 聊天请求失败:', error)
      throw error
    }
  },

  // 执行文件系统操作
  async executeFileOperation(operation: string, params: any): Promise<any> {
    try {
      const response = await request<any>('/api/ai/execute', {
        method: 'POST',
        body: JSON.stringify({
          operation,
          params
        })
      })
      return response
    } catch (error) {
      console.error('执行文件操作失败:', error)
      throw error
    }
  }
}