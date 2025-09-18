<template>
  <el-container class="drive-container">
    <!-- Sidebar -->
    <el-aside :width="sidebarCollapsed ? '64px' : '240px'" class="sidebar-container">
      <Sidebar
        :active-item="activeView"
        :is-collapsed="sidebarCollapsed"
        @item-click="changeView"
      />
    </el-aside>

    <!-- Main Content -->
    <el-container>
      <!-- Header -->
      <el-header class="main-header">
        <div class="header-left">
          <el-button :icon="sidebarCollapsed ? Expand : Fold" link @click="toggleSidebar" class="collapse-btn" />
          <h1 class="page-title">{{ currentPageTitle }}</h1>
        </div>
        <div class="header-right">
          <div class="action-buttons" v-if="activeView === 'home'">
            <el-button type="primary" :icon="UploadFilled" @click="showUploadModal = true">上传</el-button>
            <el-button type="success" :icon="FolderAdd" @click="createNewFolder">新建文件夹</el-button>
            <el-button type="info" :icon="ChatDotRound" @click="toggleAIChat" plain>
              AI 助手
            </el-button>
          </div>
        </div>
      </el-header>

      <!-- Main content area -->
      <el-main class="content-area">
        <div class="main-content-wrapper">
          <!-- 文件管理主界面 -->
          <div class="file-management-area" :class="{ 'with-ai-panel': showAIPanel }">
            <component :is="currentViewComponent" />
          </div>
          
          <!-- AI 聊天面板 -->
          <div v-if="showAIPanel" class="ai-chat-panel">
            <div class="ai-panel-header">
              <h3>AI 助手</h3>
              <el-button size="small" text @click="toggleAIChat">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
            <div class="ai-panel-content">
              <AIChat />
            </div>
          </div>
        </div>
      </el-main>
    </el-container>

    <!-- 上传对话框 (使用后端 /api/fs/put) -->
    <el-dialog v-model="showUploadModal" title="上传文件" width="500px">
      <el-upload
        class="upload-dragger"
        drag
        multiple
        :http-request="httpUpload"
        :on-success="handleUploadSuccess"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处或 <em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            拖拽文件到此处上传到当前目录
          </div>
        </template>
      </el-upload>
    </el-dialog>

  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, shallowRef, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useDriveStore } from '@/stores/drive'
import Sidebar from '../components/Sidebar.vue';
import FilesView from './drive/FilesView.vue';
import DownloadsView from './drive/DownloadsView.vue';
import MountsView from './drive/MountsView.vue';
import SettingsView from './drive/SettingsView.vue';
import AIChat from '@/components/AIChat.vue';
import { Expand, Fold, UploadFilled, FolderAdd, ChatDotRound, Close } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

// State
const route = useRoute();
const activeView = ref('home');
const sidebarCollapsed = ref(false);
const showUploadModal = ref(false);
const showAIPanel = ref(false);

// 组件挂载时检查URL参数
onMounted(() => {
  const view = route.query.view as string;
  if (view && ['home', 'downloads', 'mounts', 'settings'].includes(view)) {
    activeView.value = view;
  }
});

const viewComponents = {
  home: FilesView,
  downloads: DownloadsView,
  mounts: MountsView,
  settings: SettingsView,
};

// Computed
const currentViewComponent = computed(() => {
  // @ts-ignore
  return viewComponents[activeView.value] || FilesView; // Default to FilesView
});

const currentPageTitle = computed(() => {
  const titles: Record<string, string> = {
    'home': '文件管理',
    'downloads': '下载管理',
    'mounts': '磁盘挂载',
    'settings': '系统设置',
  };
  return titles[activeView.value] || '文件管理';
});

// Methods
const changeView = (viewId: string) => {
  activeView.value = viewId;
};

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value;
};

const toggleAIChat = () => {
  showAIPanel.value = !showAIPanel.value;
};

import { ElMessageBox } from 'element-plus'
import { api } from '@/api'

const createNewFolder = async () => {
  try {
    const result = await ElMessageBox.prompt('文件夹名称', '创建新文件夹', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputPattern: /\S+/, // non-empty
      inputErrorMessage: '请输入文件夹名称',
    })
    const folderName = (result as any).value
    if (!folderName) return
    const driveStore = useDriveStore()
    const base = driveStore.currentPath || '/'
    // build path: if base is root '/', join as '/name', else '/base/name'
    const target = base === '/' || base === '' ? `/${folderName}` : `${base.replace(/\/$/, '')}/${folderName}`
    await api.post('/api/fs/mkdir', { path: target })
    ElMessage.success('文件夹创建成功')
    // notify FilesView to refresh
    try { window.dispatchEvent(new CustomEvent('hela-files-updated')) } catch (e) {}
  } catch (err: any) {
    if (err === 'cancel' || err === undefined) return
    ElMessage.error(err.message || '创建文件夹失败')
  }
}

const handleUploadSuccess = (response: any, file: any, fileList: any) => {
  ElMessage.success(`${file.name} 上传成功。`);
  // We might need to emit an event to FilesView to refresh the list
  if (fileList.every((f: any) => f.status === 'success')) {
    setTimeout(() => {
      showUploadModal.value = false;
      // notify FilesView to reload
      try { window.dispatchEvent(new CustomEvent('hela-files-updated')) } catch(e) {}
    }, 1000);
  }
};

// custom http upload handler for el-upload
const httpUpload = async (options: any) => {
  // options: { file, onProgress, onSuccess, onError }
  const { file, onProgress, onSuccess, onError } = options
  try {
    const form = new FormData()
  const driveStore = useDriveStore()
  const target = driveStore.currentPath || '/'
  form.append('path', target)
    form.append('file', file)

    const token = localStorage.getItem('token') || ''
    const resp = await fetch('/api/fs/put', {
      method: 'POST',
      body: form,
      headers: token ? { 'Authorization': `Bearer ${token}` } : undefined,
    })

    if (!resp.ok) {
      const text = await resp.text()
      onError(new Error(`Upload failed: ${resp.status} ${text}`))
      return
    }

    const json = await resp.json()
    // backend wraps response as { code, message, data }
    if (json.code && json.code !== 200) {
      onError(new Error(json.message || 'upload error'))
      return
    }
    onSuccess(json.data)
  } catch (err: any) {
    onError(err)
  }
}

</script>

<style scoped>
.drive-container {
  height: 100vh;
  overflow: hidden;
  background-color: #f0f2f5;
}

.sidebar-container {
  background-color: #1e293b;
  transition: width 0.3s ease;
}

.main-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 20px;
  background-color: #ffffff;
  border-bottom: 1px solid #e4e7ed;
}

.header-left {
  display: flex;
  align-items: center;
}

.collapse-btn {
  font-size: 22px;
  margin-right: 15px;
  color: #303133;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
}

.content-area {
  padding: 20px;
  overflow-y: auto;
}

.main-content-wrapper {
  display: flex;
  height: 100%;
  gap: 20px;
}

.file-management-area {
  flex: 1;
  min-width: 0;
  transition: all 0.3s ease;
}

.file-management-area.with-ai-panel {
  max-width: calc(100% - 420px);
}

.ai-chat-panel {
  width: 400px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.ai-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafbfc;
  border-radius: 8px 8px 0 0;
}

.ai-panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.ai-panel-content {
  flex: 1;
  overflow: hidden;
}

.upload-dragger {
  padding: 20px;
}
</style>
