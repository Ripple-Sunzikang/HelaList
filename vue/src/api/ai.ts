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