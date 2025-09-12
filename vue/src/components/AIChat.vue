<template>
  <div class="ai-chat-container">
    <!-- 聊天窗口 -->
    <div class="chat-window" ref="chatWindow">
      <div
        v-for="(message, index) in messages"
        :key="index"
        :class="['message', message.type]"
      >
        <div class="message-avatar">
          <el-icon v-if="message.type === 'user'">
            <User />
          </el-icon>
          <el-icon v-else>
            <ChatDotRound />
          </el-icon>
        </div>
        <div class="message-content">
          <div class="message-text" v-html="formatMessage(message.content)"></div>
          <div class="message-time">{{ formatTime(message.timestamp) }}</div>
        </div>
      </div>
      
      <!-- 正在输入指示器 -->
      <div v-if="isTyping" class="message ai typing">
        <div class="message-avatar">
          <el-icon><ChatDotRound /></el-icon>
        </div>
        <div class="message-content">
          <div class="typing-indicator">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </div>
      </div>
    </div>

    <!-- 输入区域 -->
    <div class="chat-input-area">
      <div class="input-row">
        <el-input
          v-model="currentInput"
          type="textarea"
          :rows="2"
          placeholder="告诉我你想做什么... 例如：创建一个名为'项目资料'的文件夹"
          @keydown.enter.ctrl="sendMessage"
          @keydown.enter.exact.prevent="sendMessage"
          :disabled="isTyping"
          class="chat-input"
        />
        <el-button
          type="primary"
          :icon="ChatDotRound"
          @click="sendMessage"
          :loading="isTyping"
          class="send-button"
        >
          发送
        </el-button>
      </div>
      
      <!-- 快捷操作建议 -->
      <div class="quick-actions" v-if="!messages.length">
        <el-tag
          v-for="suggestion in quickSuggestions"
          :key="suggestion"
          class="suggestion-tag"
          @click="selectSuggestion(suggestion)"
          type="info"
          effect="plain"
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
import { User, ChatDotRound } from '@element-plus/icons-vue'
import { aiApi } from '@/api/ai'

interface ChatMessage {
  type: 'user' | 'ai'
  content: string
  timestamp: Date
  action?: {
    type: string
    params: any
    result?: any
  }
}

// 响应式数据
const messages = ref<ChatMessage[]>([])
const currentInput = ref('')
const isTyping = ref(false)
const chatWindow = ref<HTMLElement>()

// 快捷建议
const quickSuggestions = [
  '创建一个新文件夹',
  '列出当前目录的所有文件',
  '查找包含"项目"的文件',
  '创建一个名为"资料"的存储',
  '显示所有用户信息'
]

// 方法
const sendMessage = async () => {
  const input = currentInput.value.trim()
  if (!input || isTyping.value) return

  // 添加用户消息
  messages.value.push({
    type: 'user',
    content: input,
    timestamp: new Date()
  })

  currentInput.value = ''
  isTyping.value = true
  
  // 滚动到底部
  await scrollToBottom()

  try {
    // 调用 AI API
    const response = await aiApi.chat(input, messages.value)
    
    // 添加 AI 回复
    messages.value.push({
      type: 'ai',
      content: response.reply,
      timestamp: new Date(),
      action: response.actions && response.actions.length > 0 ? response.actions[0] : undefined
    })

    // 如果有操作结果，执行对应的文件系统操作
    if (response.actions && response.actions.length > 0) {
      for (const action of response.actions) {
        await executeAction(action)
      }
    }
    
  } catch (error) {
    console.error('AI 聊天错误:', error)
    messages.value.push({
      type: 'ai',
      content: '抱歉，我遇到了一些问题。请稍后再试。',
      timestamp: new Date()
    })
  } finally {
    isTyping.value = false
    await scrollToBottom()
  }
}

const executeAction = async (action: any) => {
  try {
    console.log('执行操作:', action)
    
    // 根据操作类型执行相应的API调用
    switch (action.type) {
      case 'list_files':
        // 刷新文件列表
        window.dispatchEvent(new CustomEvent('hela-files-updated'))
        break
        
      case 'create_folder':
      case 'delete_item':
      case 'rename_item':
      case 'copy_item':  
      case 'move_item':
        // 调用文件操作API
        const result = await aiApi.executeFileOperation(action.type, action.params)
        console.log('操作结果:', result)
        
        // 操作完成后刷新文件列表
        setTimeout(() => {
          window.dispatchEvent(new CustomEvent('hela-files-updated'))
        }, 500)
        break
        
      default:
        console.warn('未知操作类型:', action.type)
    }
  } catch (error) {
    console.error('执行操作失败:', error)
    // 显示错误消息
    ElMessage.error(`执行操作失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

const selectSuggestion = (suggestion: string) => {
  currentInput.value = suggestion
  sendMessage()
}

const formatMessage = (content: string) => {
  // 支持简单的 Markdown 格式
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}

const formatTime = (timestamp: Date) => {
  return timestamp.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

const scrollToBottom = async () => {
  await nextTick()
  if (chatWindow.value) {
    chatWindow.value.scrollTop = chatWindow.value.scrollHeight
  }
}

// 初始化欢迎消息
onMounted(() => {
  messages.value.push({
    type: 'ai',
    content: '你好！我是 HelaList AI 助手。我可以帮你管理文件、创建文件夹、查看存储信息等。请告诉我你想做什么！',
    timestamp: new Date()
  })
})
</script>

<style scoped>
.ai-chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f8f9fa;
  border-radius: 8px;
  overflow: hidden;
}

.chat-window {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: white;
}

.message {
  display: flex;
  margin-bottom: 16px;
  align-items: flex-start;
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 8px;
  font-size: 16px;
}

.message.user .message-avatar {
  background: #409eff;
  color: white;
}

.message.ai .message-avatar {
  background: #67c23a;
  color: white;
}

.message-content {
  max-width: 70%;
  min-width: 100px;
}

.message.user .message-content {
  text-align: right;
}

.message-text {
  background: #f0f0f0;
  padding: 12px 16px;
  border-radius: 12px;
  line-height: 1.4;
  word-wrap: break-word;
}

.message.user .message-text {
  background: #409eff;
  color: white;
}

.message.ai .message-text {
  background: #f0f0f0;
  color: #333;
}

.message-time {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.typing-indicator {
  display: flex;
  padding: 12px 16px;
  background: #f0f0f0;
  border-radius: 12px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #999;
  margin: 0 2px;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) { animation-delay: -0.32s; }
.typing-indicator span:nth-child(2) { animation-delay: -0.16s; }

@keyframes typing {
  0%, 80%, 100% { 
    transform: scale(0);
    opacity: 0.5;
  }
  40% { 
    transform: scale(1);
    opacity: 1;
  }
}

.chat-input-area {
  padding: 16px;
  background: white;
  border-top: 1px solid #eee;
}

.input-row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.chat-input {
  flex: 1;
}

.send-button {
  height: 40px;
}

.quick-actions {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.suggestion-tag {
  cursor: pointer;
  transition: all 0.2s;
}

.suggestion-tag:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
</style>