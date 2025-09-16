<template>
  <div class="files-view">
    <!-- 面包屑导航 -->
    <el-breadcrumb :separator-icon="ArrowRight">
      <el-breadcrumb-item @click="goRoot">根目录</el-breadcrumb-item>
      <el-breadcrumb-item v-for="(seg, idx) in breadcrumbSegments" :key="idx">{{ seg }}</el-breadcrumb-item>
    </el-breadcrumb>

    <!-- 文件网格 -->
    <div class="toolbar" style="margin: 12px 0; display:flex; gap:8px; align-items:center;">
      <el-button type="info" @click="refresh">刷新</el-button>
      <el-divider direction="vertical" />
      <span style="color: var(--el-text-color-regular); font-size: 14px;">排序方式：</span>
      <el-select v-model="sortBy" placeholder="排序方式" style="width: 120px;" size="small">
        <el-option label="名称" value="name" />
        <el-option label="大小" value="size" />
        <el-option label="修改时间" value="modified" />
      </el-select>
      <el-button-group size="small">
        <el-button 
          :type="sortOrder === 'asc' ? 'primary' : 'default'" 
          @click="sortOrder = 'asc'"
          :icon="SortUp"
        >
          升序
        </el-button>
        <el-button 
          :type="sortOrder === 'desc' ? 'primary' : 'default'" 
          @click="sortOrder = 'desc'"
          :icon="SortDown"
        >
          降序
        </el-button>
      </el-button-group>
    </div>

    <div v-if="sortedFileItems.length > 0" class="file-grid">
      <div
        v-for="(item, index) in sortedFileItems"
        :key="item.id"
        class="file-block"
        @click="openFile(item)"
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

        <!-- 操作下拉菜单 -->
        <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, item, index)" class="file-actions">
          <el-button :icon="MoreFilled" circle link class="more-button" @click.stop />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="preview" :icon="View" v-if="['video', 'image', 'pdf', 'document', 'audio'].includes(item.type)">预览</el-dropdown-item>
              <el-dropdown-item command="download" :icon="Download" :disabled="item.isDir">下载</el-dropdown-item>
              <el-dropdown-item command="rename" :icon="EditPen">重命名</el-dropdown-item>
              <el-dropdown-item command="move" :icon="Rank">移动</el-dropdown-item>
              <el-dropdown-item command="delete" :icon="Delete" divided>删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- 空状态 -->
    <el-empty v-else description="此文件夹为空。" />

    <!-- 文件夹选择器 -->
    <FolderSelector 
      v-model="showFolderSelector" 
      :current-path="currentPath"
      @confirm="handleMoveConfirm"
    />

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useDriveStore } from '@/stores/drive';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  ArrowRight, Download, Delete, Folder, Document, Picture, VideoPlay, Headset, MoreFilled, EditPen, Rank, View, SortUp, SortDown
} from '@element-plus/icons-vue';
import { api } from '@/api';
import { download } from '@/api/utils';
import FolderSelector from '@/components/FolderSelector.vue';

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
const driveStore = useDriveStore();
const currentPath = ref(driveStore.currentPath || ''); // '' 表示 root
const sortBy = ref<'name' | 'size' | 'modified'>('name'); // 默认按名称排序
const sortOrder = ref<'asc' | 'desc'>('asc'); // 默认升序
const showFolderSelector = ref(false); // 文件夹选择器显示状态
const pendingMoveItem = ref<FileItem | null>(null); // 待移动的文件

// computed breadcrumb segments
const breadcrumbSegments = computed(() => {
  if (!currentPath.value) return [];
  return currentPath.value.split('/').filter(Boolean);
});

// computed sorted file items
const sortedFileItems = computed(() => {
  const items = [...fileItems.value];
  
  return items.sort((a, b) => {
    // 首先按文件夹优先排序：文件夹在前，文件在后
    if (a.isDir && !b.isDir) return -1;
    if (!a.isDir && b.isDir) return 1;
    
    // 在同类型（都是文件夹或都是文件）中按选定的条件排序
    let compareValue = 0;
    
    switch (sortBy.value) {
      case 'name':
        compareValue = a.name.localeCompare(b.name, 'zh-CN', { 
          numeric: true, 
          sensitivity: 'base' 
        });
        break;
      case 'size':
        compareValue = a.size - b.size;
        break;
      case 'modified':
        compareValue = a.modified - b.modified;
        break;
      default:
        compareValue = a.name.localeCompare(b.name, 'zh-CN', { 
          numeric: true, 
          sensitivity: 'base' 
        });
    }
    
    return sortOrder.value === 'asc' ? compareValue : -compareValue;
  });
});

// helper: build api url for a given path
function buildListUrl(path: string, refresh = false) {
  // backend route: /api/fs/list/*path  where *path may be omitted or like /a/b
  const q = refresh ? '?refresh=true' : '';
  if (!path || path === '' || path === '/') {
    return `/api/fs/list${q}`;
  }
  const normalized = path.startsWith('/') ? path : `/${path}`;
  return `/api/fs/list${normalized}${q}`;
}

function getFileType(name: string): string {
  const extension = name.split('.').pop()?.toLowerCase();
  if (!extension) return 'file';
  if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'].includes(extension)) {
    return 'image';
  }
  if (['mp4', 'mkv', 'avi', 'mov', 'wmv', 'flv', 'webm'].includes(extension)) {
    return 'video';
  }
  if (['mp3', 'wav', 'flac', 'aac', 'ogg'].includes(extension)) {
    return 'audio';
  }
  if (['pdf'].includes(extension)) {
    return 'pdf';
  }
  if (['txt', 'md', 'json', 'xml', 'html', 'css', 'js'].includes(extension)) {
    return 'document';
  }
  return 'file';
}

async function loadList(path = '', refresh = false) {
  loading.value = true;
  try {
    const url = buildListUrl(path, refresh);
    const data = await api.get<any>(url);
    // data expected to be { content: ObjResp[], total: number, write: boolean }
    const content = data?.content || [];
    fileItems.value = content.map((it: any) => ({
      id: it.id,
      name: it.name,
      type: it.is_dir ? 'folder' : getFileType(it.name),
      modified: it.modified ? new Date(it.modified).getTime() : 0,
      size: it.size || 0,
      path: it.path || '',
      isDir: !!it.is_dir,
    }));
  } catch (err: any) {
    ElMessage.error(err.message || '加载文件失败');
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  // initialize from store
  currentPath.value = driveStore.currentPath || '';
  loadList(currentPath.value, true);
  const handler = () => loadList(currentPath.value, true);
  window.addEventListener('hela-files-updated', handler);
  // cleanup on unmount
  onUnmounted(() => {
    window.removeEventListener('hela-files-updated', handler);
  });
});

// Methods
const getFileIcon = (type: string) => {
  switch (type) {
    case 'folder': return Folder;
    case 'image': return Picture;
    case 'video': return VideoPlay;
    case 'audio': return Headset;
    case 'pdf': return Document; // Or a more specific PDF icon if available
    case 'document': return Document;
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
  if (!item.isDir) return;
  // set current path to item's path and reload
  currentPath.value = item.path || `/${item.name}`;
  driveStore.setPath(currentPath.value);
  loadList(currentPath.value, false);
};

const buildFileUrl = (path: string, download = false) => {
  const url = `/api/fs/download${path}`;
  const params = new URLSearchParams();
  params.append('t', new Date().getTime().toString());
  if (download) {
    params.append('type', 'download');
  }
  return `${url}?${params.toString()}`;
}

const openFile = (item: FileItem) => {
  const previewableTypes = ['video', 'image', 'pdf', 'document', 'audio'];
  if (item.isDir) {
    openFolder(item);
  } else if (previewableTypes.includes(item.type)) {
    window.open(buildFileUrl(item.path, false), '_blank');
  } else {
    toggleSelection(item);
  }
};

const handleCommand = (command: string, item: FileItem, index: number) => {
  // 找到原始数组中的索引
  const originalIndex = fileItems.value.findIndex(f => f.id === item.id);
  
  switch (command) {
    case 'preview':
      openFile(item);
      break;
    case 'download':
      {
        driveStore.addDownloading(item.name)
        download(buildFileUrl(item.path, true), (progress) => {
          driveStore.updateDownloadProgress(item.name, progress)
        }).then(async (response) => {
          const blob = await response.blob()
          const link = document.createElement('a')
          link.href = URL.createObjectURL(blob)
          link.download = item.name
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          driveStore.removeDownloading(item.name)
        }).catch((err) => {
          ElMessage.error(err.message || '下载失败')
          driveStore.removeDownloading(item.name)
        })
      }
      break;
    case 'rename':
      ElMessageBox.prompt('请输入新名称', '重命名', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputValue: item.name,
      }).then(async ({ value }) => {
        try {
          await api.fs.rename(item.path, value);
          if (originalIndex !== -1) {
            fileItems.value[originalIndex].name = value;
          }
          ElMessage.success('重命名成功');
        } catch (err: any) {
          ElMessage.error(err.message || '重命名失败');
        }
      });
      break;
    case 'move':
      pendingMoveItem.value = item;
      showFolderSelector.value = true;
      break;
    case 'delete':
      ElMessageBox.confirm(`确定要删除 "${item.name}" 吗？`, '警告', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(async () => {
        try {
          await api.fs.remove(item.path);
          if (originalIndex !== -1) {
            fileItems.value.splice(originalIndex, 1);
          }
          ElMessage.success('文件已删除');
        } catch (err: any) {
          ElMessage.error(err.message || '删除失败');
        }
      });
      break;
  }
};

const handleMoveConfirm = async (targetPath: string) => {
  if (!pendingMoveItem.value) return;
  
  const item = pendingMoveItem.value;
  try {
    // 目标路径就是目标目录，后端会自动处理文件名
    const dstDirPath = targetPath || '/';
    
    // 检查是否移动到同一位置（比较目录）
    const currentDir = currentPath.value || '/';
    if (currentDir === dstDirPath) {
      ElMessage.warning('文件已在目标位置');
      return;
    }
    
    // 调用移动API，传递源文件路径和目标目录路径
    await api.fs.move(item.path, dstDirPath);
    
    // 从当前列表中移除文件
    const originalIndex = fileItems.value.findIndex(f => f.id === item.id);
    if (originalIndex !== -1) {
      fileItems.value.splice(originalIndex, 1);
    }
    
    ElMessage.success(`"${item.name}" 已移动到 "${targetPath || '根目录'}"`);
  } catch (err: any) {
    ElMessage.error(err.message || '移动失败');
  } finally {
    pendingMoveItem.value = null;
  }
};

const refresh = () => loadList(currentPath.value, true);

// upload handled by DriveMain; FilesView listens for 'hela-files-updated' and refreshes

const goRoot = () => {
  currentPath.value = '';
  driveStore.setPath('');
  loadList('', true);
};
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
