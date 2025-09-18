<template>
  <div class="mount-storage-page">
    <div class="container">
      <div class="header">
        <h1 class="title">挂载WebDAV存储</h1>
        <p class="description">连接您的WebDAV网盘到HelaList，统一管理云端文件</p>
      </div>
      
      <form @submit.prevent="submitForm" class="mount-form">
        <!-- 基本信息 -->
        <div class="form-section">
          <h2 class="section-title">基本信息</h2>
          <div class="form-group">
            <label for="mountPath">挂载路径 *</label>
            <input 
              type="text" 
              id="mountPath"
              v-model="formData.mount_path"
              placeholder="例如: /my-webdav"
              required
              class="form-input"
            />
            <small class="help-text">在HelaList中显示的虚拟路径，以/开头</small>
          </div>
          
          <div class="form-group">
            <label for="remark">存储名称</label>
            <input 
              type="text" 
              id="remark"
              v-model="formData.remark"
              placeholder="例如: 我的WebDAV网盘"
              class="form-input"
            />
            <small class="help-text">可选，用于描述这个存储</small>
          </div>
        </div>

        <!-- WebDAV连接信息 -->
        <div class="form-section">
          <h2 class="section-title">WebDAV连接信息</h2>
          <div class="form-group">
            <label for="address">服务器地址 *</label>
            <input 
              type="url" 
              id="address"
              v-model="formData.addition.address"
              placeholder="https://example.com/webdav"
              required
              class="form-input"
            />
            <small class="help-text">WebDAV服务器的完整URL地址</small>
          </div>
          
          <div class="form-group">
            <label for="username">用户名 *</label>
            <input 
              type="text" 
              id="username"
              v-model="formData.addition.username"
              placeholder="输入您的用户名"
              required
              class="form-input"
            />
          </div>
          
          <div class="form-group">
            <label for="password">密码 *</label>
            <div class="password-wrapper">
              <input 
                :type="showPassword ? 'text' : 'password'"
                id="password"
                v-model="formData.addition.password"
                placeholder="输入您的密码"
                required
                class="form-input"
              />
              <button type="button" @click="showPassword = !showPassword" class="password-toggle">
                {{ showPassword ? '👁️' : '👁️‍🗨️' }}
              </button>
            </div>
          </div>
          
          <div class="form-group">
            <label for="rootPath">根目录路径</label>
            <input 
              type="text" 
              id="rootPath"
              v-model="formData.addition.root_folder_path"
              placeholder="可选，默认为根目录"
              class="form-input"
            />
            <small class="help-text">WebDAV服务器上的起始目录，可以为空</small>
          </div>
        </div>

        <!-- 高级设置 -->
        <div class="form-section">
          <h2 class="section-title">高级设置</h2>
          <div class="checkbox-group">
            <label class="checkbox-label">
              <input 
                type="checkbox" 
                v-model="formData.addition.tls_insecure_skip_verify"
                class="checkbox"
              />
              <span class="checkmark"></span>
              跳过TLS证书验证
            </label>
            <small class="help-text">仅在使用自签名证书或测试环境时启用</small>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="form-actions">
          <button type="submit" :disabled="loading" class="btn btn-primary">
            {{ loading ? '挂载中...' : '确认挂载' }}
          </button>
          <button type="button" @click="goBack" class="btn btn-secondary">
            取消返回
          </button>
        </div>
      </form>
      
      <!-- 状态消息 -->
      <div v-if="message" class="message" :class="messageType">
        {{ message }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api'

const router = useRouter()

const formData = reactive({
  mount_path: '',
  driver: 'webdav',
  order: 1,
  cache_expiration: 3600,
  remark: '',
  disabled: false,
  web_proxy: true,
  webdav_policy: '',
  proxy_range: false,
  down_proxy_url: '',
  disable_proxy_sign: false,
  addition: {
    vendor: '',
    address: '',
    username: '',
    password: '',
    root_folder_path: '',
    tls_insecure_skip_verify: false
  }
})

const loading = ref(false)
const message = ref('')
const messageType = ref('info')
const showPassword = ref(false)

const submitForm = async () => {
  if (!formData.mount_path || !formData.addition.address || !formData.addition.username || !formData.addition.password) {
    message.value = '请填写必填字段'
    return
  }

  loading.value = true
  try {
    // 创建要发送的数据，将addition对象转换为JSON字符串
    const submitData = {
      ...formData,
      addition: JSON.stringify(formData.addition)
    }
    
    await api.storage.create(submitData)
    message.value = '挂载成功！正在跳转...'
    setTimeout(() => {
      router.push('/home?view=mounts')
    }, 2000)
  } catch (error: any) {
    message.value = `挂载失败: ${error.message}`
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/home?view=mounts')
}
</script>

<style scoped>
.mount-storage-page {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.container {
  max-width: 700px;
  margin: 0 auto;
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 40px 30px;
  text-align: center;
}

.title {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0 0 10px 0;
}

.description {
  font-size: 1.1rem;
  opacity: 0.9;
  margin: 0;
}

.mount-form {
  padding: 30px;
}

.form-section {
  margin-bottom: 30px;
  padding: 20px;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
}

.section-title {
  font-size: 1.3rem;
  font-weight: 600;
  color: #2d3748;
  margin: 0 0 20px 0;
  padding-bottom: 10px;
  border-bottom: 2px solid #667eea;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
  font-size: 0.95rem;
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 16px;
  transition: all 0.3s ease;
  box-sizing: border-box;
  background: white;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  transform: translateY(-1px);
}

.password-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-toggle {
  position: absolute;
  right: 12px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  padding: 4px;
  color: #6b7280;
  transition: color 0.2s;
}

.password-toggle:hover {
  color: #667eea;
}

.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  cursor: pointer;
  font-weight: 500;
  color: #374151;
}

.checkbox {
  width: 18px;
  height: 18px;
  margin-right: 10px;
  accent-color: #667eea;
}

.help-text {
  display: block;
  font-size: 0.85rem;
  color: #6b7280;
  margin-top: 6px;
  line-height: 1.4;
}

.form-actions {
  display: flex;
  gap: 15px;
  justify-content: center;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 14px 28px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 140px;
  position: relative;
  overflow: hidden;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
}

.btn-secondary {
  background: #f8fafc;
  color: #4a5568;
  border: 2px solid #e2e8f0;
}

.btn-secondary:hover:not(:disabled) {
  background: #edf2f7;
  border-color: #cbd5e0;
  transform: translateY(-1px);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: none !important;
}

.message {
  margin: 20px 30px;
  padding: 15px 20px;
  border-radius: 8px;
  font-weight: 500;
  text-align: center;
  animation: slideIn 0.3s ease;
}

.message.info {
  background: #dbeafe;
  color: #1e40af;
  border: 1px solid #93c5fd;
}

.message.success {
  background: #d1fae5;
  color: #065f46;
  border: 1px solid #a7f3d0;
}

.message.error {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fca5a5;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .mount-storage-page {
    padding: 10px;
    align-items: flex-start;
    padding-top: 20px;
  }
  
  .container {
    max-width: 100%;
    border-radius: 12px;
  }
  
  .header {
    padding: 30px 20px;
  }
  
  .title {
    font-size: 2rem;
  }
  
  .mount-form {
    padding: 20px;
  }
  
  .form-section {
    padding: 15px;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
  }
  
  .message {
    margin: 15px 20px;
  }
}
</style>