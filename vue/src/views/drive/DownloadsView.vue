<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>下载管理</span>
      </div>
    </template>
    <div v-if="driveStore.downloading.length > 0">
      <div v-for="download in driveStore.downloading" :key="download.name" class="download-item">
        <span class="download-name">{{ download.name }}</span>
        <el-progress :percentage="download.progress" class="download-progress" />
      </div>
    </div>
    <el-empty v-else description="暂无活动下载。" />
  </el-card>
</template>

<script setup lang="ts">
import { useDriveStore } from '@/stores/drive';

const driveStore = useDriveStore();
</script>

<style scoped>
.download-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.download-name {
  width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.download-progress {
  flex-grow: 1;
  margin-left: 20px;
}
</style>
