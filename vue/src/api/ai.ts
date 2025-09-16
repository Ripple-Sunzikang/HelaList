// AI API 接口
export interface AIResponse {
  code: number
  message: string
  data: {
    reply: string
    actions?: Array<{
      type: string
      params: any
    }>
  }
}

export interface AIExecuteResponse {
  code: number
  message: string
  data: any
}

// 新的聊天接口类型定义
export interface ChatSession {
  id: number
  session_id: string
  user_id: string
  title: string
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  id: number
  session_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  metadata?: string
  created_at: string
}

export interface ChatRequest {
  session_id?: string
  user_id?: string
  message: string
  use_rag?: boolean
}

export interface ChatResponse {
  session_id: string
  message: string
  context?: Array<{
    source: string
    content: string
    similarity: number
  }>
}

export interface CreateSessionRequest {
  user_id: string
  title?: string
}

// 创建新会话
export const createChatSession = async (request: CreateSessionRequest): Promise<{ code: number; data: ChatSession }> => {
  const response = await fetch('/api/chat/sessions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
    },
    body: JSON.stringify(request)
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return await response.json()
}

// 获取用户会话列表
export const getUserSessions = async (userId: string): Promise<{ code: number; data: ChatSession[] }> => {
  const response = await fetch(`/api/chat/sessions?user_id=${encodeURIComponent(userId)}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
    }
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return await response.json()
}

// 获取会话历史
export const getChatHistory = async (sessionId: string): Promise<{ code: number; data: ChatMessage[] }> => {
  const response = await fetch(`/api/chat/sessions/${sessionId}/history`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
    }
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return await response.json()
}

// 发送带上下文的聊天消息
export const chatWithContext = async (request: ChatRequest): Promise<{ code: number; data: ChatResponse }> => {
  const response = await fetch('/api/chat/message', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
    },
    body: JSON.stringify(request)
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return await response.json()
}

// 删除会话
export const deleteChatSession = async (sessionId: string): Promise<{ code: number }> => {
  const response = await fetch(`/api/chat/sessions/${sessionId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
    }
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return await response.json()
}

// 更新会话标题
export const updateSessionTitle = async (sessionId: string, title: string): Promise<{ code: number }> => {
  const response = await fetch(`/api/chat/sessions/${sessionId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
    },
    body: JSON.stringify({ title })
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return await response.json()
}

// AI 聊天API
export const chatWithAI = async (message: string): Promise<AIResponse> => {
  const response = await fetch('/api/ai/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
    },
    body: JSON.stringify({ message })
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return await response.json()
}

// 执行AI操作
export const executeAIAction = async (operation: string, params: any): Promise<AIExecuteResponse> => {
  const response = await fetch('/api/ai/execute', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
    },
    body: JSON.stringify({
      operation,
      params
    })
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return await response.json()
}