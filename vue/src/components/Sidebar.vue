<template>
  <div class="sidebar" :class="{ collapsed: isCollapsed }">
    <!-- 品牌标识 -->
    <div class="sidebar-brand">
      <i class="fas fa-cloud text-primary"></i>
      <span class="brand-name" v-if="!isCollapsed">CloudDrive</span>
    </div>

    <!-- 功能菜单 -->
    <nav class="sidebar-menu">
      <button
        v-for="item in menuItems"
        :key="item.id"
        class="menu-item"
        :class="{ active: activeItem === item.id }"
        @click="handleItemClick(item.id)"
      >
        <i :class="item.icon"></i>
        <span class="menu-text" >{{ item.name }}</span>
      </button>
    </nav>

    <!-- 折叠控制按钮 -->
    <button class="collapse-toggle" @click="toggleCollapse">
      <i class="fas" :class="isCollapsed ? 'fa-angle-right' : 'fa-angle-left'"></i>
    </button>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue';

// 定义组件属性
const props = defineProps({
  activeItem: {
    type: String,
    required: true,
    default: 'home'
  },
  isCollapsed: {
    type: Boolean,
    default: false
  }
});

// 定义组件事件
const emit = defineEmits(['item-click', 'toggle-collapse']);

// 菜单数据
const menuItems = [
  { id: 'home', name: '主页', icon: 'fas fa-home' },
  { id: 'downloads', name: '文件下载', icon: 'fas fa-download' },
  { id: 'mounts', name: '磁盘挂载', icon: 'fas fa-hdd' },
  { id: 'settings', name: '设置', icon: 'fas fa-cog' }
];

// 处理菜单点击
const handleItemClick = (itemId: string) => {
  emit('item-click', itemId);
};

// 处理折叠状态切换
const toggleCollapse = () => {
  emit('toggle-collapse');
};
</script>

<style scoped>
.sidebar {
  width: 240px;
  height: 100vh;
  background-color: #1e293b;
  color: #e2e8f0;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

.sidebar.collapsed {
  width: 60px;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.sidebar-brand i {
  font-size: 1.5rem;
  color: #3b82f6;
}

.brand-name {
  margin-left: 0.75rem;
  font-weight: 600;
  font-size: 1.1rem;
  transition: opacity 0.3s ease;
}

.sidebar-menu {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem 0;
}

.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  width: 100%;
  height: 80px;
  background: none;
  border: none;
  cursor: pointer;
}

.menu-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
  color: #ffffff;
}

.menu-item.active {
  background-color: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  border-left: 3px solid #3b82f6;
}

.menu-item i {
  font-size: 2.8rem;
}

.menu-text {
  position: absolute;
  left: 70px;
  top: 50%;
  transform: translateY(-50%);
  background: #1e293b;
  color: #fff;
  padding: 0.3rem 0.8rem;
  border-radius: 0.375rem;
  white-space: nowrap;
  z-index: 100;
  font-size: 1.1rem;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s;
}
.menu-item:hover .menu-text {
  opacity: 1;
  pointer-events: auto;
}

.collapse-toggle {
  width: 100%;
  padding: 0.75rem;
  background-color: rgba(255, 255, 255, 0.05);
  border: none;
  color: #cbd5e1;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.collapse-toggle:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: #ffffff;
}
</style>
