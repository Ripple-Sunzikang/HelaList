<template>
  <div class="files-view">
    <!-- Breadcrumb -->
    <el-breadcrumb :separator-icon="ArrowRight">
      <el-breadcrumb-item @click="goRoot">Root</el-breadcrumb-item>
      <el-breadcrumb-item v-for="(seg, idx) in breadcrumbSegments" :key="idx">{{ seg }}</el-breadcrumb-item>
    </el-breadcrumb>

    <!-- File Grid -->
    <div class="toolbar" style="margin: 12px 0; display:flex; gap:8px; align-items:center;">
      <el-button type="info" @click="refresh">Refresh</el-button>
    </div>

    <div v-if="fileItems.length > 0" class="file-grid">
      <div
        v-for="(item, index) in fileItems"
        :key="item.id"
        class="file-block"
        @click="item.isDir ? openFolder(item) : toggleSelection(item)"
        :class="{ selected: isSelected(item) }"
      >
        <!-- Selection Checkbox -->
        <el-checkbox :model-value="isSelected(item)" size="large" class="selection-checkbox" />

        <!-- File Icon -->
        <div class="file-icon-container">
          <el-icon :size="64" class="file-icon">
            <component :is="getFileIcon(item.type)" />
          </el-icon>
        </div>

        <!-- File Name -->
        <div class="file-name-container">
          <span class="file-name">{{ item.name }}</span>
        </div>

        <!-- Actions Dropdown -->
  <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, item, index)" class="file-actions">
            <el-button :icon="MoreFilled" circle link class="more-button" @click.stop />
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="download" :icon="Download" :disabled="item.isDir">Download</el-dropdown-item>
                <el-dropdown-item command="rename" :icon="EditPen">Rename</el-dropdown-item>
                <el-dropdown-item command="move" :icon="Rank">Move</el-dropdown-item>
                <el-dropdown-item command="delete" :icon="Delete" divided>Delete</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
      </div>
    </div>

    <!-- Empty State -->
    <el-empty v-else description="This folder is empty." />

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useDriveStore } from '@/stores/drive'
import { ElMessage, ElMessageBox } from 'element-plus';
import { 
  ArrowRight, Download, Delete, Folder, Document, Picture, VideoPlay, Headset, MoreFilled, EditPen, Rank 
} from '@element-plus/icons-vue';
import { api } from '@/api'

// Types
interface FileItem {
  id: string;
  name: string;
  type: string;
  modified: number;
  size: number;
  path: string;
  isDir: boolean;
}

// Refs
const loading = ref(false);
const fileItems = ref<FileItem[]>([]);
const selectedItems = ref<FileItem[]>([]);
const driveStore = useDriveStore()
const currentPath = ref(driveStore.currentPath || '') // '' 表示 root

// computed breadcrumb segments
const breadcrumbSegments = computed(() => {
  if (!currentPath.value) return []
  return currentPath.value.split('/').filter(Boolean)
})

// helper: build api url for a given path
function buildListUrl(path: string, refresh = false) {
  // backend route: /api/fs/list/*path  where *path may be omitted or like /a/b
  const q = refresh ? '?refresh=true' : ''
  if (!path || path === '' || path === '/') {
    return `/api/fs/list${q}`
  }
  const normalized = path.startsWith('/') ? path : `/${path}`
  return `/api/fs/list${normalized}${q}`
}

async function loadList(path = '', refresh = false) {
  loading.value = true
  try {
    const url = buildListUrl(path, refresh)
    const data = await api.get<any>(url)
    // data expected to be { content: ObjResp[], total: number, write: boolean }
    const content = data?.content || []
    fileItems.value = content.map((it: any) => ({
      id: it.id,
      name: it.name,
      type: it.is_dir ? 'folder' : 'file',
      modified: it.modified ? new Date(it.modified).getTime() : 0,
      size: it.size || 0,
      path: it.path || '',
      isDir: !!it.is_dir,
    }))
  } catch (err: any) {
    ElMessage.error(err.message || 'Failed to load files')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  // initialize from store
  currentPath.value = driveStore.currentPath || ''
  loadList(currentPath.value, true)
  const handler = () => loadList(currentPath.value, true)
  window.addEventListener('hela-files-updated', handler)
  // cleanup on unmount
  onUnmounted(() => {
    window.removeEventListener('hela-files-updated', handler)
  })
})

// Methods
const getFileIcon = (type: string) => {
  switch (type) {
    case 'folder': return Folder;
    case 'image': return Picture;
    case 'video': return VideoPlay;
    case 'audio': return Headset;
    default: return Document;
  }
};

const isSelected = (item: FileItem) => {
  return selectedItems.value.some(selected => selected.name === item.name);
};

const toggleSelection = (item: FileItem) => {
  if (isSelected(item)) {
    selectedItems.value = selectedItems.value.filter(selected => selected.name !== item.name);
  } else {
    selectedItems.value.push(item);
  }
};

const openFolder = (item: FileItem) => {
  if (!item.isDir) return
  // set current path to item's path and reload
  currentPath.value = item.path || `/${item.name}`
  driveStore.setPath(currentPath.value)
  loadList(currentPath.value, false)
}

const handleCommand = (command: string, item: FileItem, index: number) => {
  switch (command) {
    case 'download':
      ElMessage.success(`Downloading: ${item.name}`);
      break;
    case 'rename':
      ElMessageBox.prompt('Please enter a new name', 'Rename', {
        confirmButtonText: 'OK',
        cancelButtonText: 'Cancel',
        inputValue: item.name,
      }).then(async ({ value }) => {
        try {
          await api.fs.rename(item.path, value)
          fileItems.value[index].name = value;
          ElMessage.success('Renamed successfully');
        } catch (err: any) {
          ElMessage.error(err.message || 'Failed to rename')
        }
      });
      break;
    case 'move':
      ElMessage.info(`Move: ${item.name}`);
      break;
    case 'delete':
      ElMessageBox.confirm(`Are you sure you want to delete "${item.name}"?`, 'Warning', {
        confirmButtonText: 'Delete',
        cancelButtonText: 'Cancel',
        type: 'warning',
      }).then(async () => {
        try {
          await api.fs.remove(item.path)
          fileItems.value.splice(index, 1);
          ElMessage.success('File deleted');
        } catch (err: any) {
          ElMessage.error(err.message || 'Failed to delete')
        }
      });
      break;
  }
};

const refresh = () => loadList(currentPath.value, true)

// upload handled by DriveMain; FilesView listens for 'hela-files-updated' and refreshes

const goRoot = () => {
  currentPath.value = ''
  driveStore.setPath('')
  loadList('', true)
}

</script>

<style scoped>
.files-view {
  padding: 10px;
}

.file-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 20px;
}

.file-block {
  position: relative;
  width: 160px;
  height: 140px;
  border-radius: 8px;
  border: 1px solid var(--el-border-color-lighter);
  background-color: var(--el-bg-color);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px;
  transition: border-color 0.2s, box-shadow 0.2s;
  overflow: hidden;
}

.file-block:hover {
  border-color: var(--el-color-primary);
  box-shadow: var(--el-box-shadow-lighter);
}

.file-block.selected {
  border-color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

.selection-checkbox {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 2;
  opacity: 0;
  transition: opacity 0.2s;
}

.file-block:hover .selection-checkbox,
.file-block.selected .selection-checkbox {
  opacity: 1;
}

.file-icon-container {
  flex-grow: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-icon {
  color: #606266;
}

.file-block .file-icon {
  color: #5686f5;
}

.file-name-container {
  width: 100%;
  text-align: center;
  margin-top: 8px;
}

.file-name {
  font-size: 14px;
  color: var(--el-text-color-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}

.file-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 2;
  opacity: 0;
  transition: opacity 0.2s;
}

.file-block:hover .file-actions,
.file-block.selected .file-actions {
  opacity: 1;
}

.more-button {
  color: var(--el-text-color-secondary);
}
.more-button:hover {
  color: var(--el-color-primary);
}

</style>