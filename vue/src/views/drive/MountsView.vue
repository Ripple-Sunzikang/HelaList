<template>
  <div class="mounts-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>存储挂载管理</span>
          <el-button type="primary" :icon="Plus" @click="goToMountPage">
            挂载新存储
          </el-button>
        </div>
      </template>

      <!-- 存储列表 -->
      <div v-if="storages.length > 0" class="storages-grid">
        <div 
          v-for="storage in storages" 
          :key="storage.id"
          class="storage-card"
          :class="{ 'disabled': storage.disabled, 'error': storage.status !== 'work' }"
        >
          <div class="storage-header">
            <div class="storage-info">
              <h3 class="storage-name">{{ storage.remark || storage.mount_path }}</h3>
              <p class="storage-path">{{ storage.mount_path }}</p>
            </div>
            <div class="storage-status">
              <el-tag 
                :type="getStatusTagType(storage.status)" 
                size="small"
              >
                {{ getStatusText(storage.status) }}
              </el-tag>
            </div>
          </div>

          <div class="storage-details">
            <div class="detail-item">
              <span class="label">驱动类型:</span>
              <span class="value">{{ storage.driver.toUpperCase() }}</span>
            </div>
            <div class="detail-item">
              <span class="label">排序优先级:</span>
              <span class="value">{{ storage.order }}</span>
            </div>
            <div class="detail-item">
              <span class="label">缓存过期:</span>
              <span class="value">{{ storage.cache_expiration }}秒</span>
            </div>
            <div class="detail-item">
              <span class="label">最后修改:</span>
              <span class="value">{{ formatDate(storage.modified_time) }}</span>
            </div>
          </div>

          <div class="storage-actions">
            <el-button size="small" type="primary" @click="editStorage(storage)">
              编辑
            </el-button>
            <el-button 
              size="small" 
              :type="storage.disabled ? 'success' : 'warning'"
              @click="toggleStorageStatus(storage)"
            >
              {{ storage.disabled ? '启用' : '禁用' }}
            </el-button>
            <el-button size="small" type="danger" @click="deleteStorage(storage)">
              删除
            </el-button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else class="empty-state">
        <el-empty description="还没有挂载任何存储">
          <el-button type="primary" :icon="Plus" @click="goToMountPage">
            挂载第一个存储
          </el-button>
        </el-empty>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <el-skeleton :rows="3" animated />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { api } from '@/api'

const router = useRouter()

// 数据状态
const storages = ref<any[]>([])
const loading = ref(false)

// 跳转到挂载页面
const goToMountPage = () => {
  router.push('/mount')
}

// 获取存储列表
const fetchStorages = async () => {
  loading.value = true
  try {
    const data = await api.storage.getAll()
    console.log('获取到的存储数据:', data) // 调试日志
    storages.value = Array.isArray(data) ? data : []
  } catch (error: any) {
    console.error('获取存储列表错误:', error) // 调试日志
    ElMessage.error(`获取存储列表失败: ${error.message}`)
  } finally {
    loading.value = false
  }
}

// 获取状态标签类型
const getStatusTagType = (status: string) => {
  switch (status) {
    case 'work':
      return 'success'
    case 'disabled':
      return 'info'
    default:
      return 'danger'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'work':
      return '正常工作'
    case 'disabled':
      return '已禁用'
    default:
      return '错误'
  }
}

// 格式化日期
const formatDate = (dateString: string) => {
  if (!dateString) return '未知'
  try {
    return new Date(dateString).toLocaleString('zh-CN')
  } catch {
    return '格式错误'
  }
}

// 编辑存储
const editStorage = (storage: any) => {
  ElMessage.info('编辑功能暂未实现')
}

// 切换存储状态
const toggleStorageStatus = async (storage: any) => {
  try {
    const action = storage.disabled ? '启用' : '禁用'
    await ElMessageBox.confirm(
      `确定要${action}存储 "${storage.remark || storage.mount_path}" 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    // 这里应该调用更新存储的API
    ElMessage.info('切换状态功能暂未实现')
  } catch {
    // 用户取消操作
  }
}

// 删除存储
const deleteStorage = async (storage: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除存储 "${storage.remark || storage.mount_path}" 吗？\n此操作不可恢复！`,
      '危险操作',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'error',
      }
    )
    
    // 这里应该调用删除存储的API
    ElMessage.info('删除功能暂未实现')
  } catch {
    // 用户取消操作
  }
}

// 组件挂载后获取数据
onMounted(() => {
  fetchStorages()
})
</script>

<style scoped>
.mounts-view {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.storages-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.storage-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
  background: #fff;
  transition: all 0.3s ease;
}

.storage-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.storage-card.disabled {
  opacity: 0.6;
  background: #f5f7fa;
}

.storage-card.error {
  border-color: #f56c6c;
  background: #fef0f0;
}

.storage-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.storage-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.storage-info p {
  margin: 0;
  font-size: 14px;
  color: #909399;
}

.storage-details {
  margin-bottom: 16px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.detail-item .label {
  color: #606266;
  font-weight: 500;
}

.detail-item .value {
  color: #303133;
}

.storage-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
}

.loading-state {
  padding: 20px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .storages-grid {
    grid-template-columns: 1fr;
  }
  
  .storage-actions {
    flex-wrap: wrap;
  }
  
  .storage-actions .el-button {
    flex: 1;
    min-width: 60px;
  }
}
</style>
