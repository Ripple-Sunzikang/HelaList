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
            <el-button type="primary" :icon="UploadFilled" @click="showUploadModal = true">Upload</el-button>
            <el-button type="success" :icon="FolderAdd" @click="createNewFolder">New Folder</el-button>
          </div>
        </div>
      </el-header>

      <!-- Main content area -->
      <el-main class="content-area">
        <component :is="currentViewComponent" />
      </el-main>
    </el-container>

    <!-- Upload Dialog (uses backend /api/fs/put) -->
    <el-dialog v-model="showUploadModal" title="Upload Files" width="500px">
      <el-upload
        class="upload-dragger"
        drag
        multiple
        :http-request="httpUpload"
        :on-success="handleUploadSuccess"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          Drop file here or <em>click to upload</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            Drop files to upload to current directory
          </div>
        </template>
      </el-upload>
    </el-dialog>

  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, shallowRef } from 'vue';
import { useDriveStore } from '@/stores/drive'
import Sidebar from '../components/Sidebar.vue';
import FilesView from './drive/FilesView.vue';
import DownloadsView from './drive/DownloadsView.vue';
import MountsView from './drive/MountsView.vue';
import SettingsView from './drive/SettingsView.vue';
import { Expand, Fold, UploadFilled, FolderAdd } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

// State
const activeView = ref('home');
const sidebarCollapsed = ref(false);
const showUploadModal = ref(false);

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
    'home': 'File Management',
    'downloads': 'Download Management',
    'mounts': 'Disk Mounts',
    'settings': 'System Settings',
  };
  return titles[activeView.value] || 'File Management';
});

// Methods
const changeView = (viewId: string) => {
  activeView.value = viewId;
};

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value;
};

import { ElMessageBox } from 'element-plus'
import { api } from '@/api'

const createNewFolder = async () => {
  try {
    const result = await ElMessageBox.prompt('Folder name', 'Create New Folder', {
      confirmButtonText: 'Create',
      cancelButtonText: 'Cancel',
      inputPattern: /\S+/, // non-empty
      inputErrorMessage: 'Folder name is required',
    })
    const folderName = (result as any).value
    if (!folderName) return
    const driveStore = useDriveStore()
    const base = driveStore.currentPath || '/'
    // build path: if base is root '/', join as '/name', else '/base/name'
    const target = base === '/' || base === '' ? `/${folderName}` : `${base.replace(/\/$/, '')}/${folderName}`
    await api.post('/api/fs/mkdir', { path: target })
    ElMessage.success('Folder created')
    // notify FilesView to refresh
    try { window.dispatchEvent(new CustomEvent('hela-files-updated')) } catch (e) {}
  } catch (err: any) {
    if (err === 'cancel' || err === undefined) return
    ElMessage.error(err.message || 'Create folder failed')
  }
}

const handleUploadSuccess = (response: any, file: any, fileList: any) => {
  ElMessage.success(`${file.name} uploaded successfully.`);
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

.upload-dragger {
  padding: 20px;
}
</style>
