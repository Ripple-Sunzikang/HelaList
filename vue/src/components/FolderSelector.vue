<template>
  <el-dialog v-model="visible" title="选择目标文件夹" width="500px" @close="handleClose">
    <div class="folder-selector">
      <!-- 当前路径显示 -->
      <div class="current-path">
        <el-breadcrumb :separator-icon="ArrowRight" style="margin-bottom: 16px;">
          <el-breadcrumb-item @click="navigateTo('')">根目录</el-breadcrumb-item>
          <el-breadcrumb-item 
            v-for="(seg, idx) in pathSegments" 
            :key="idx"
            @click="navigateTo(pathSegments.slice(0, idx + 1).join('/'))"
          >
            {{ seg }}
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>

      <!-- 文件夹列表 -->
      <div class="folder-list" v-loading="loading">
        <div 
          v-for="folder in folders" 
          :key="folder.name"
          class="folder-item"
          @click="navigateTo(folder.path)"
        >
          <el-icon class="folder-icon"><Folder /></el-icon>
          <span class="folder-name">{{ folder.name }}</span>
        </div>
        <el-empty v-if="!loading && folders.length === 0" description="此文件夹中没有子文件夹" />
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleConfirm">
          移动到此处
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowRight, Folder } from '@element-plus/icons-vue'
import { api } from '@/api'

interface FolderItem {
  name: string
  path: string
}

// Props
interface Props {
  modelValue: boolean
  currentPath?: string
}

const props = withDefaults(defineProps<Props>(), {
  currentPath: ''
})

// Emits
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'confirm': [targetPath: string]
}>()

// Data
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const loading = ref(false)
const folders = ref<FolderItem[]>([])
const selectedPath = ref('')

// Computed
const pathSegments = computed(() => {
  if (!selectedPath.value) return []
  return selectedPath.value.split('/').filter(Boolean)
})

// Methods
const loadFolders = async (path: string) => {
  loading.value = true
  try {
    const url = path ? `/api/fs/list/${path}` : '/api/fs/list'
    const data = await api.get<any>(url)
    const content = data?.content || []
    
    // 只获取文件夹
    folders.value = content
      .filter((item: any) => item.is_dir)
      .map((item: any) => ({
        name: item.name,
        path: item.path || (path ? `${path}/${item.name}` : `/${item.name}`)
      }))
  } catch (err: any) {
    ElMessage.error(err.message || '加载文件夹失败')
    folders.value = []
  } finally {
    loading.value = false
  }
}

const navigateTo = (path: string) => {
  selectedPath.value = path
  loadFolders(path)
}

const handleClose = () => {
  visible.value = false
}

const handleConfirm = () => {
  const targetPath = selectedPath.value || '/'
  emit('confirm', targetPath)
  visible.value = false
}

// Watch
watch(visible, (newVal) => {
  if (newVal) {
    selectedPath.value = props.currentPath || ''
    loadFolders(selectedPath.value)
  }
})
</script>

<style scoped>
.folder-selector {
  min-height: 300px;
  max-height: 400px;
}

.current-path {
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.folder-list {
  margin-top: 16px;
  max-height: 300px;
  overflow-y: auto;
}

.folder-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.folder-item:hover {
  background-color: var(--el-color-primary-light-9);
}

.folder-icon {
  color: var(--el-color-primary);
  margin-right: 8px;
  font-size: 16px;
}

.folder-name {
  color: var(--el-text-color-primary);
  font-size: 14px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>