<template>
  <div class="files-view">
    <!-- Breadcrumb -->
    <el-breadcrumb :separator-icon="ArrowRight">
      <el-breadcrumb-item :to="{ path: '/' }">Root</el-breadcrumb-item>
      <!-- Dynamic breadcrumbs here -->
    </el-breadcrumb>

    <!-- File Grid -->
    <div v-if="fileItems.length > 0" class="file-grid">
      <div
        v-for="(item, index) in fileItems"
        :key="item.name"
        class="file-block"
        @click="toggleSelection(item)"
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
              <el-dropdown-item command="download" :icon="Download" :disabled="item.type === 'folder'">Download</el-dropdown-item>
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
import { ref, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { 
  ArrowRight, Download, Delete, Folder, Document, Picture, VideoPlay, Headset, MoreFilled, EditPen, Rank 
} from '@element-plus/icons-vue';

// Types
interface FileItem {
  name: string;
  type: string;
  modified: number;
  size: number;
}

// Refs
const loading = ref(false);
const fileItems = ref<FileItem[]>([
  { name: 'Work Documents', type: 'folder', modified: new Date('2023-06-15').getTime(), size: 0 },
  { name: 'project-plan.docx', type: 'doc', modified: new Date('2023-06-20').getTime(), size: 2097152 },
  { name: 'financials.xlsx', type: 'xls', modified: new Date('2023-06-22').getTime(), size: 3145728 },
  { name: 'product-demo.pptx', type: 'ppt', modified: new Date('2023-06-25').getTime(), size: 10485760 },
  { name: 'meeting-notes.pdf', type: 'pdf', modified: new Date('2023-06-28').getTime(), size: 1572864 },
  { name: 'design.png', type: 'image', modified: new Date('2023-07-01').getTime(), size: 4194304 },
  { name: 'promo.mp4', type: 'video', modified: new Date('2023-07-05').getTime(), size: 52428800 },
]);
const selectedItems = ref<FileItem[]>([]);

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
      }).then(({ value }) => {
        fileItems.value[index].name = value;
        ElMessage.success('Renamed successfully');
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
      }).then(() => {
        fileItems.value.splice(index, 1);
        ElMessage.success('File deleted');
      });
      break;
  }
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