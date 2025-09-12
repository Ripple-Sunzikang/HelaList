<template>
  <div class="ai-chat-container">
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
import { ref, nextTick, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

interface ChatMessage {
  type: 'user' | 'ai'
  content: string
  timestamp: Date
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

// 建议列表
const suggestions = [
  '创建一个新文件夹',
  '列出当前目录文件',
  '分析图片内容',
  '帮我整理文件'
]

// 初始化欢迎消息
onMounted(() => {
  addMessage('ai', '你好！我是HelaList AI助手 🤖\n\n我可以帮你：\n• 管理文件和文件夹\n• 分析图片内容\n• 整理存储空间\n\n请告诉我你需要什么帮助！')
})

// 添加消息
const addMessage = (type: 'user' | 'ai', content: string, imagePreview?: any) => {
  messages.value.push({
    type,
    content,
    timestamp: new Date(),
    imagePreview
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
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      },
      body: JSON.stringify({ message: input })
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const result = await response.json()
    
    if (result.code !== 200) {
      throw new Error(result.message || 'API调用失败')
    }

    // 添加AI回复
    addMessage('ai', result.data.reply)

    // 执行操作（如果有）
    if (result.data.actions && result.data.actions.length > 0) {
      for (const action of result.data.actions) {
        await executeAction(action)
      }
    }

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

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const result = await response.json()
    console.log('操作API响应结果:', result)
    
    if (result.code !== 200) {
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
      
    case 'list_files':
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