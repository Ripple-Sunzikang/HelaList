<template>
  <div class="ai-chat-container">
    <!-- 会话管理区域 -->
    <div class="session-header">
      <el-row :gutter="10" align="middle">
        <el-col :span="16">
          <el-select
            v-model="currentSessionId"
            placeholder="选择会话"
            @change="switchSession"
            class="session-select"
            filterable
            clearable
          >
            <el-option
              v-for="session in sessions"
              :key="session.session_id"
              :label="session.title"
              :value="session.session_id"
            >
              <div class="session-option">
                <span>{{ session.title }}</span>
                <span class="session-time">{{ formatSessionTime(session.updated_at) }}</span>
              </div>
            </el-option>
          </el-select>
        </el-col>
        <el-col :span="8">
          <el-button @click="createNewSession" type="primary" size="small">新建对话</el-button>
          <el-button 
            v-if="currentSessionId" 
            @click="deleteCurrentSession" 
            type="danger" 
            size="small"
          >
            删除
          </el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 聊天消息区域 -->
    <div class="chat-messages" ref="chatWindow">
      <div
        v-for="(message, index) in messages"
        :key="`msg-${index}-${message.timestamp.getTime()}`"
        :class="['message', message.type]"
      >
        <div class="message-bubble">
          <div class="message-content" v-html="formatMessage(message.content)"></div>
          
          <!-- 图片预览区域 -->
          <div v-if="message.imagePreview" class="image-preview">
            <img 
              :src="message.imagePreview.url" 
              :alt="message.imagePreview.filename"
              class="preview-image"
              @click="openImageFullscreen(message.imagePreview.url)"
            />
            <p class="image-caption">{{ message.imagePreview.filename }}</p>
          </div>

          <!-- RAG上下文信息 -->
          <div v-if="message.context && message.context.length > 0" class="rag-context">
            <div class="context-header">📚 参考文档：</div>
            <div v-for="(ctx, idx) in message.context" :key="idx" class="context-item">
              <div class="context-source">{{ ctx.source }}</div>
              <div class="context-preview">{{ ctx.content.substring(0, 100) }}...</div>
            </div>
          </div>
          
          <div class="message-time">{{ formatTime(message.timestamp) }}</div>
        </div>
      </div>
      
      <!-- 正在输入状态 -->
      <div v-if="isTyping" class="message ai typing-message">
        <div class="message-bubble">
          <div class="typing-indicator">
            <div class="typing-dot"></div>
            <div class="typing-dot"></div>
            <div class="typing-dot"></div>
          </div>
          <div class="message-time">正在思考...</div>
        </div>
      </div>
    </div>

    <!-- 输入区域 -->
    <div class="chat-input">
      <div class="input-options">
        <el-checkbox v-model="useRAG" size="small">使用文档增强</el-checkbox>
      </div>
      
      <div class="input-container">
        <el-input
          v-model="currentInput"
          placeholder="告诉我你想做什么..."
          @keydown.enter.exact.prevent="sendMessage"
          :disabled="isTyping"
          class="message-input"
          size="large"
        >
          <template #append>
            <el-button 
              type="primary" 
              @click="sendMessage"
              :loading="isTyping"
              :disabled="!currentInput.trim()"
            >
              发送
            </el-button>
          </template>
        </el-input>
      </div>
      
      <!-- 快捷建议 -->
      <div v-if="!messages.length" class="suggestions">
        <el-tag
          v-for="suggestion in suggestions"
          :key="suggestion"
          @click="useSuggestion(suggestion)"
          class="suggestion-item"
          effect="plain"
          type="info"
        >
          {{ suggestion }}
        </el-tag>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  createChatSession, 
  getUserSessions, 
  getChatHistory, 
  chatWithContext, 
  deleteChatSession,
  type ChatSession,
  type ChatMessage as APIChatMessage,
  type ChatResponse
} from '../api/ai'

interface ChatMessage {
  type: 'user' | 'ai'
  content: string
  timestamp: Date
  context?: Array<{
    source: string
    content: string
    similarity: number
  }>
  imagePreview?: {
    url: string
    filename: string
  }
}

// 响应式数据
const messages = ref<ChatMessage[]>([])
const currentInput = ref('')
const isTyping = ref(false)
const chatWindow = ref<HTMLElement>()
const useRAG = ref(false)

// 会话管理
const sessions = ref<ChatSession[]>([])
const currentSessionId = ref<string>('')
const currentUserId = 'default_user' // 实际应用中应该从用户认证获取

// 建议列表
const suggestions = [
  '创建一个新文件夹',
  '列出当前目录文件',
  '分析图片内容',
  '帮我整理文件'
]

// 初始化
onMounted(async () => {
  await loadUserSessions()
  addWelcomeMessage()
})

// 会话管理功能
const loadUserSessions = async () => {
  try {
    const result = await getUserSessions(currentUserId)
    if (result.code === 200) {
      sessions.value = result.data
    }
  } catch (error) {
    console.error('加载会话列表失败:', error)
  }
}

const createNewSession = async () => {
  try {
    const result = await createChatSession({
      user_id: currentUserId,
      title: '新对话'
    })
    
    if (result.code === 200) {
      sessions.value.unshift(result.data)
      currentSessionId.value = result.data.session_id
      messages.value = []
      addWelcomeMessage()
    }
  } catch (error) {
    console.error('创建会话失败:', error)
    ElMessage.error('创建会话失败')
  }
}

const switchSession = async (sessionId: string) => {
  if (!sessionId) {
    messages.value = []
    addWelcomeMessage()
    return
  }
  
  try {
    const result = await getChatHistory(sessionId)
    if (result.code === 200) {
      messages.value = result.data.map(msg => ({
        type: msg.role === 'user' ? 'user' : 'ai',
        content: msg.content,
        timestamp: new Date(msg.created_at),
        context: msg.metadata ? JSON.parse(msg.metadata).rag_contexts : undefined
      }))
      scrollToBottom()
    }
  } catch (error) {
    console.error('加载会话历史失败:', error)
    ElMessage.error('加载会话历史失败')
  }
}

const deleteCurrentSession = async () => {
  if (!currentSessionId.value) return
  
  try {
    await ElMessageBox.confirm('确定要删除这个会话吗？此操作不可恢复。', '确认删除', {
      type: 'warning'
    })
    
    await deleteChatSession(currentSessionId.value)
    sessions.value = sessions.value.filter(s => s.session_id !== currentSessionId.value)
    currentSessionId.value = ''
    messages.value = []
    addWelcomeMessage()
    ElMessage.success('会话已删除')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除会话失败:', error)
      ElMessage.error('删除会话失败')
    }
  }
}

// 添加欢迎消息
const addWelcomeMessage = () => {
  addMessage('ai', '你好！我是HelaList AI助手 🤖\n\n我可以帮你：\n• 管理文件和文件夹\n• 分析图片内容\n• 整理存储空间\n• 记住我们的对话历史\n\n请告诉我你需要什么帮助！')
}

// 添加消息
const addMessage = (type: 'user' | 'ai', content: string, imagePreview?: any, context?: any[]) => {
  messages.value.push({
    type,
    content,
    timestamp: new Date(),
    imagePreview,
    context
  })
  scrollToBottom()
}

// 发送消息
const sendMessage = async () => {
  const input = currentInput.value.trim()
  if (!input || isTyping.value) return

  addMessage('user', input)
  currentInput.value = ''
  isTyping.value = true

  try {
    // 使用新的带上下文的聊天API
    const response = await chatWithContext({
      session_id: currentSessionId.value,
      user_id: currentUserId,
      message: input,
      use_rag: useRAG.value
    })
    
    if (response.code !== 200) {
      throw new Error('API调用失败')
    }

    // 更新当前会话ID（如果是新创建的会话）
    if (!currentSessionId.value && response.data.session_id) {
      currentSessionId.value = response.data.session_id
      await loadUserSessions() // 重新加载会话列表
    }

    // 添加AI回复
    addMessage('ai', response.data.message, undefined, response.data.context)

  } catch (error) {
    console.error('聊天错误:', error)
    addMessage('ai', '抱歉，我遇到了一些问题，请稍后再试 😔')
  } finally {
    isTyping.value = false
  }
}

// 执行AI操作
const executeAction = async (action: any) => {
  try {
    console.log('开始执行操作:', action)
    
    const response = await fetch('/api/ai/execute', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      },
      body: JSON.stringify({
        operation: action.type,
        params: action.params
      })
    })

    console.log('操作API响应状态:', response.status)
    console.log('操作API响应头:', response.headers)

    if (!response.ok) {
      const errorText = await response.text()
      console.error('API响应错误内容:', errorText)
      throw new Error(`HTTP ${response.status}: ${errorText}`)
    }

    const result = await response.json()
    console.log('操作API响应结果:', result)
    console.log('操作API响应结果类型:', typeof result)
    console.log('result.code:', result.code)
    console.log('result.message:', result.message)
    console.log('result.data:', result.data)
    
    if (result.code !== 200) {
      console.error('操作失败，错误信息:', result.message)
      throw new Error(result.message || '操作失败')
    }

    // 处理不同操作的结果
    handleActionResult(action.type, result.data)
    
  } catch (error) {
    console.error('操作执行失败:', error)
    ElMessage.error(`操作失败: ${error instanceof Error ? error.message : '未知错误'}`)
    addMessage('ai', `❌ 操作执行失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

// 处理操作结果
const handleActionResult = (actionType: string, result: any) => {
  console.log('处理操作结果:', actionType, result)
  
  switch (actionType) {
    case 'preview_image':
      // result.result 包含实际的操作结果
      if (result && result.result && result.result.preview_url) {
        addMessage('ai', '📷 图片预览', {
          url: result.result.preview_url,
          filename: result.result.file_path
        })
      }
      break
      
    case 'analyze_image':
      // result.result 包含实际的操作结果
      if (result && result.result && result.result.analysis) {
        // 显示详细的图片分析结果
        addMessage('ai', `🔍 **图片分析结果：**\n\n${result.result.analysis}`)
        console.log('添加图片分析结果消息:', result.result.analysis)
      } else {
        console.error('图片分析结果为空:', result)
        addMessage('ai', '图片分析完成，但没有收到分析结果 😕')
      }
      break
      
    case 'preview_document':
      // 显示文档内容
      console.log('处理preview_document结果:', result)
      console.log('result.result:', result.result)
      if (result && result.result) {
        console.log('result.result.content存在:', !!result.result.content)
        console.log('result.result.content长度:', result.result.content ? result.result.content.length : 'N/A')
        console.log('result.result.file_path:', result.result.file_path)
        console.log('result.result.type:', result.result.type)
      }
      
      if (result && result.result && result.result.content) {
        const maxLength = 2000 // 限制显示长度，避免太长
        let content = result.result.content
        let truncated = false
        
        if (content.length > maxLength) {
          content = content.substring(0, maxLength)
          truncated = true
        }
        
        // 使用代码块格式显示文档内容
        let message = `📄 **文档预览：${result.result.file_path}**\n\n\`\`\`\n${content}\n\`\`\``
        
        if (truncated) {
          message += '\n\n*（内容已截断，显示前2000个字符）*'
        }
        
        addMessage('ai', message)
        console.log('成功添加文档预览消息')
      } else {
        console.error('文档预览结果为空或结构不正确:', result)
        addMessage('ai', '❌ 无法预览文档内容')
      }
      break
      
    case 'list_files':
      // 显示文件列表内容
      console.log('处理list_files结果:', result)
      if (result && result.result) {
        console.log('result.result:', result.result)
        const files = result.result
        if (Array.isArray(files) && files.length > 0) {
          let fileListText = '📁 **目录内容：**\n\n'
          files.forEach((file: any, index: number) => {
            console.log(`文件 ${index}:`, file)
            const icon = file.is_dir ? '📁' : '📄'
            const size = file.is_dir ? '' : ` (${formatFileSize(file.size || 0)})`
            fileListText += `${icon} ${file.name}${size}\n`
          })
          addMessage('ai', fileListText)
          console.log('添加文件列表消息:', fileListText)
        } else {
          console.log('文件列表为空或不是数组:', files)
          addMessage('ai', '📁 目录为空，没有找到任何文件或文件夹。')
        }
      } else {
        console.log('result 或 result.result 为空:', result)
        addMessage('ai', '❌ 无法获取目录内容。')
      }
      // 同时刷新文件列表界面
      setTimeout(() => {
        window.dispatchEvent(new CustomEvent('hela-files-updated'))
      }, 500)
      break
      
    case 'create_folder':
    case 'delete_item':
    case 'rename_item':
    case 'copy_item':
    case 'move_item':
      // 刷新文件列表
      ElMessage.success('操作完成！')
      setTimeout(() => {
        window.dispatchEvent(new CustomEvent('hela-files-updated'))
      }, 500)
      break
      
    default:
      console.log('未处理的操作类型:', actionType, result)
  }
}

// 使用建议
const useSuggestion = (suggestion: string) => {
  currentInput.value = suggestion
  sendMessage()
}

// 格式化文件大小
const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化消息
const formatMessage = (content: string) => {
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}

// 格式化时间
const formatTime = (timestamp: Date) => {
  return timestamp.toLocaleTimeString('zh-CN', { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

// 格式化会话时间
const formatSessionTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diffTime = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
  
  if (diffDays === 0) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  } else if (diffDays === 1) {
    return '昨天'
  } else if (diffDays < 7) {
    return `${diffDays}天前`
  } else {
    return date.toLocaleDateString('zh-CN')
  }
}

// 滚动到底部
const scrollToBottom = async () => {
  await nextTick()
  if (chatWindow.value) {
    chatWindow.value.scrollTop = chatWindow.value.scrollHeight
  }
}

// 全屏查看图片
const openImageFullscreen = (imageUrl: string) => {
  window.open(imageUrl, '_blank')
}
</script>

<style scoped>
.ai-chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #ffffff;
  border-radius: 8px;
  overflow: hidden;
}

/* 会话管理区域 */
.session-header {
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
}

.session-select {
  width: 100%;
}

.session-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.session-time {
  font-size: 12px;
  color: #999;
}

.input-options {
  padding: 8px 16px;
  background: #f8f9fa;
  border-bottom: 1px solid #e4e7ed;
}

/* RAG上下文显示 */
.rag-context {
  margin-top: 8px;
  padding: 8px;
  background: #f0f8ff;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.context-header {
  font-size: 12px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 4px;
}

.context-item {
  margin-bottom: 6px;
  padding: 4px;
  background: white;
  border-radius: 4px;
}

.context-source {
  font-size: 11px;
  color: #666;
  font-weight: bold;
}

.context-preview {
  font-size: 12px;
  color: #333;
  margin-top: 2px;
}

.chat-messages {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  background: linear-gradient(to bottom, #f8f9fa, #ffffff);
}

.message {
  margin-bottom: 16px;
  display: flex;
}

.message.user {
  justify-content: flex-end;
}

.message.ai {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 75%;
  min-width: 120px;
}

.message.user .message-bubble {
  background: #409eff;
  color: white;
  border-radius: 18px 18px 4px 18px;
  padding: 12px 16px;
}

.message.ai .message-bubble {
  background: #f0f0f0;
  color: #333;
  border-radius: 18px 18px 18px 4px;
  padding: 12px 16px;
}

.message-content {
  line-height: 1.5;
  word-break: break-word;
}

.message-time {
  font-size: 11px;
  opacity: 0.7;
  margin-top: 4px;
  text-align: right;
}

.message.ai .message-time {
  text-align: left;
}

.typing-message .message-bubble {
  background: #f0f0f0;
  border-radius: 18px 18px 18px 4px;
  padding: 12px 16px;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  margin-bottom: 4px;
}

.typing-dot {
  width: 8px;
  height: 8px;
  background: #999;
  border-radius: 50%;
  animation: typing-bounce 1.4s infinite ease-in-out;
}

.typing-dot:nth-child(1) { animation-delay: -0.32s; }
.typing-dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes typing-bounce {
  0%, 80%, 100% { 
    transform: scale(0);
  } 
  40% { 
    transform: scale(1);
  }
}

.image-preview {
  margin-top: 8px;
  border-radius: 8px;
  overflow: hidden;
  max-width: 280px;
}

.preview-image {
  width: 100%;
  height: auto;
  max-height: 200px;
  object-fit: cover;
  cursor: pointer;
  transition: transform 0.2s;
}

.preview-image:hover {
  transform: scale(1.02);
}

.image-caption {
  font-size: 12px;
  opacity: 0.8;
  margin: 4px 0 0 0;
  text-align: center;
}

.chat-input {
  border-top: 1px solid #e4e7ed;
  background: white;
  padding: 16px;
}

.input-container {
  margin-bottom: 12px;
}

.message-input {
  --el-input-focus-border-color: #409eff;
}

.suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.suggestion-item {
  cursor: pointer;
  transition: all 0.2s;
  font-size: 12px;
}

.suggestion-item:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.3);
}
</style>