<template>
  <el-menu
    :default-active="activeItem"
    :collapse="isCollapsed"
    class="sidebar-menu"
    background-color="#1e293b"
    text-color="#e2e8f0"
    active-text-color="#3b82f6"
    @select="handleItemClick"
  >
    <div class="sidebar-brand">
      <i-ep-mostly-cloudy class="brand-logo" />
      <span v-if="!isCollapsed" class="brand-name">HelaList</span>
    </div>

    <el-menu-item index="home">
      <el-icon><i-ep-house /></el-icon>
      <template #title>首页</template>
    </el-menu-item>

    <el-menu-item index="downloads">
      <el-badge :value="driveStore.downloading.length" :hidden="driveStore.downloading.length === 0">
        <el-icon><i-ep-download /></el-icon>
      </el-badge>
      <template #title>下载</template>
    </el-menu-item>

    <el-menu-item index="mounts">
      <el-icon><i-ep-data-line /></el-icon>
      <template #title>挂载</template>
    </el-menu-item>
  </el-menu>
</template>

<script setup lang="ts">
import { useDriveStore } from '@/stores/drive'

const driveStore = useDriveStore()

// Component props
defineProps({
  activeItem: {
    type: String,
    required: true,
    default: 'home',
  },
  isCollapsed: {
    type: Boolean,
    default: false,
  },
})

// Component events
const emit = defineEmits(['item-click'])

// Handle menu item click
const handleItemClick = (itemId: string) => {
  emit('item-click', itemId)
}
</script>

<style scoped>
.sidebar-menu {
  height: 100vh;
  border-right: none; /* Remove the default border */
}

/* Non-collapsed state style */
.sidebar-menu:not(.el-menu--collapse) {
  width: 240px;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 0;
  height: 60px; /* Consistent height */
  box-sizing: border-box;
}

.brand-logo {
  font-size: 28px;
  color: #3b82f6;
}

.brand-name {
  margin-left: 12px;
  font-weight: 600;
  font-size: 20px;
  color: #e2e8f0;
}

.el-menu-item {
  font-weight: 500;
}

.el-menu-item.is-active {
  background-color: rgba(59, 130, 246, 0.1) !important; /* Use !important to override default */
}

.el-menu-item i {
  color: #94a3b8;
}

.el-menu-item.is-active i,
.el-menu-item.is-active .el-tooltip__trigger {
  color: #3b82f6;
}
</style>