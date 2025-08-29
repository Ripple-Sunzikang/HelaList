<template>
  <div class="drive-container">
    <!-- 侧边栏组件 -->
    <Sidebar
      :active-item="activeView"
      :is-collapsed="sidebarCollapsed"
      @item-click="changeView"
      @toggle-collapse="toggleSidebar"
    />

    <!-- 主内容区 -->
    <main class="main-content" :style="{ marginLeft: sidebarCollapsed ? '60px' : '10px' }">
      <!-- 顶部工具栏 -->
      <header class="main-header">
        <h1 class="page-title">{{ currentPageTitle }}</h1>

        <div class="action-buttons" v-if="activeView === 'home'">
          <button class="btn upload-btn" @click="showUploadModal = true">
            <i class="fas fa-upload"></i>
            <span>上传文件</span>
          </button>
          <button class="btn new-folder-btn" @click="createNewFolder">
            <i class="fas fa-folder-plus"></i>
            <span>新建文件夹</span>
          </button>
        </div>
      </header>

      <!-- 内容区域 -->
      <div class="content-area">
        <!-- 主页/文件列表视图 -->
        <div v-if="activeView === 'home'" class="file-manager">
          <!-- 路径导航 -->
          <div class="breadcrumb">
            <a href="#" class="breadcrumb-item">我的文件</a>
            <i class="fas fa-chevron-right separator"></i>
            <span class="current-path">全部文件</span>
          </div>

          <!-- 文件列表 -->
          <div class="file-list">
<!--            <div class="file-list-header">-->
<!--              <div class="header-column name-column">文件名</div>-->
<!--              <div class="header-column type-column">类型</div>-->
<!--              <div class="header-column date-column">修改日期</div>-->
<!--              <div class="header-column size-column">大小</div>-->
<!--              <div class="header-column actions-column">操作</div>-->
<!--            </div>-->

            <div class="file-grid">
              <div
                v-for="(item, index) in fileItems"
                :key="index"
                class="file-card"
                :class="{ 'is-locked': item.isLocked }"
              >
                <i :class="getItemIcon(item.type)" class="file-icon"></i>
                <span class="file-name">{{ item.name }}</span>
                <i v-if="item.isLocked" class="fas fa-lock lock-icon" title="已加密"></i>
                <div class="file-actions">
                  <button @click.stop="toggleLock(item)" :title="item.isLocked ? '解锁文件' : '锁定文件'">
                    <i class="fas" :class="item.isLocked ? 'fa-unlock' : 'fa-lock'"></i>
                  </button>
                  <button @click.stop="downloadFile(item)" :disabled="item.type === 'folder'" title="下载">
                    <i class="fas fa-download"></i>
                  </button>
                  <button @click.stop="deleteFile(index)" title="删除">
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
              </div>

              <!-- 空状态 -->
              <div v-if="fileItems.length === 0" class="empty-state">
                <i class="fas fa-folder-open"></i>
                <p>没有文件或文件夹</p>
                <button class="btn primary" @click="showUploadModal = true">上传文件</button>
              </div>
            </div>
          </div>
        </div>

        <!-- 下载管理视图 -->
        <div v-else-if="activeView === 'downloads'" class="downloads-view">
          <h2 class="section-title">下载管理</h2>
          <div class="downloads-container">
            <div class="download-item header">
              <div class="download-name">文件名</div>
              <div class="download-progress">进度</div>
              <div class="download-speed">速度</div>
              <div class="download-status">状态</div>
              <div class="download-actions">操作</div>
            </div>

            <div class="download-item" v-for="(dl, index) in downloads" :key="index">
              <div class="download-name">{{ dl.name }}</div>
              <div class="download-progress">
                <div class="progress-bar">
                  <div class="progress-fill" :style="{ width: dl.progress + '%' }"></div>
                </div>
                <span class="progress-text">{{ dl.progress }}%</span>
              </div>
              <div class="download-speed">{{ dl.speed }}</div>
              <div class="download-status">{{ dl.status }}</div>
              <div class="download-actions">
                <button class="action-btn" @click="cancelDownload(index)">
                  <i class="fas fa-times"></i>
                </button>
              </div>
            </div>

            <div v-if="downloads.length === 0" class="empty-state">
              <i class="fas fa-download"></i>
              <p>没有正在进行的下载任务</p>
            </div>
          </div>
        </div>

        <!-- 磁盘挂载视图 -->
        <div v-else-if="activeView === 'mounts'" class="mounts-view">
          <h2 class="section-title">磁盘挂载</h2>
          <div class="mounts-container">
            <div class="mount-item" v-for="(mount, index) in mountedDisks" :key="index">
              <div class="mount-info">
                <i class="fas fa-hdd"></i>
                <div class="mount-details">
                  <h3 class="mount-name">{{ mount.name }}</h3>
                  <p class="mount-path">{{ mount.path }}</p>
                </div>
              </div>
              <div class="mount-stats">
                <div class="mount-usage">
                  <span>{{ formatFileSize(mount.used) }} / {{ formatFileSize(mount.total) }}</span>
                  <div class="usage-bar">
                    <div class="usage-fill" :style="{ width: mount.usagePercent + '%' }"></div>
                  </div>
                </div>
              </div>
              <button
                class="mount-action-btn"
                @click="toggleMount(index)"
              >
                {{ mount.mounted ? '卸载' : '挂载' }}
              </button>
            </div>

            <button class="btn add-mount-btn" @click="addNewMount">
              <i class="fas fa-plus"></i> 添加磁盘
            </button>
          </div>
        </div>

        <!-- 设置视图 -->
        <div v-else-if="activeView === 'settings'" class="settings-view">
          <h2 class="section-title">设置</h2>
          <div class="settings-container">
            <div class="settings-section">
              <h3 class="section-heading">账户设置</h3>
              <div class="setting-item">
                <label class="setting-label">用户名</label>
                <input type="text" class="setting-input" v-model="username" />
              </div>
              <div class="setting-item">
                <label class="setting-label">电子邮件</label>
                <input type="email" class="setting-input" v-model="email" />
              </div>
            </div>

            <div class="settings-section">
              <h3 class="section-heading">存储空间</h3>
              <div class="storage-info">
                <div class="storage-summary">
                  <span>{{ formatFileSize(usedStorage) }} / {{ formatFileSize(totalStorage) }}</span>
                  <span class="storage-percent">{{ storagePercent }}%</span>
                </div>
                <div class="storage-bar">
                  <div class="storage-fill" :style="{ width: storagePercent + '%' }"></div>
                </div>
              </div>
              <button class="btn upgrade-storage-btn">升级存储空间</button>
            </div>

            <div class="settings-section">
              <h3 class="section-heading">外观设置</h3>
              <div class="setting-item">
                <label class="setting-label">主题模式</label>
                <select class="setting-select" v-model="themeMode">
                  <option value="light">浅色模式</option>
                  <option value="dark">深色模式</option>
                  <option value="system">跟随系统</option>
                </select>
              </div>
              <div class="setting-item">
                <label class="setting-label">显示文件扩展名</label>
                <label class="switch">
                  <input type="checkbox" v-model="showFileExtensions" />
                  <span class="slider round"></span>
                </label>
              </div>
            </div>

            <button class="btn save-settings-btn" @click="saveSettings">保存设置</button>
          </div>
        </div>
      </div>
    </main>

    <!-- 上传文件模态框 -->
    <div class="modal-overlay" v-if="showUploadModal" @click="showUploadModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>上传文件</h3>
          <button class="modal-close" @click="showUploadModal = false">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="upload-area" @click="$refs.fileInput.click()">
            <i class="fas fa-cloud-upload-alt"></i>
            <p>点击或拖放文件到此处上传</p>
            <input
              type="file"
              ref="fileInput"
              class="file-input"
              multiple
              @change="handleFileUpload"
              hidden
            />
          </div>

          <!-- 上传进度 -->
          <div class="upload-progress" v-if="uploading">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: uploadProgress + '%' }"></div>
            </div>
            <div class="progress-info">
              <span class="progress-filename">{{ uploadingFileName }}</span>
              <span class="progress-percent">{{ uploadProgress }}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue';
import Sidebar from '../components/Sidebar.vue';

// 状态管理
const activeView = ref('home'); // 当前激活的视图
const sidebarCollapsed = ref(false); // 侧边栏是否折叠
const showUploadModal = ref(false); // 上传模态框显示状态

// 文件列表数据
const fileItems = ref([
  {
    name: '工作文档',
    type: 'folder',
    modified: new Date('2023-06-15').getTime(),
    size: 0,
    isLocked: false
  },
  {
    name: '项目计划书.docx',
    type: 'doc',
    modified: new Date('2023-06-20').getTime(),
    size: 2097152, // 2MB
    isLocked: true
  },
  {
    name: '财务报表.xlsx',
    type: 'xls',
    modified: new Date('2023-06-22').getTime(),
    size: 3145728, // 3MB
    isLocked: false
  },
  {
    name: '产品演示.pptx',
    type: 'ppt',
    modified: new Date('2023-06-25').getTime(),
    size: 10485760, // 10MB
    isLocked: false
  },
  {
    name: '会议记录.pdf',
    type: 'pdf',
    modified: new Date('2023-06-28').getTime(),
    size: 1572864, // 1.5MB
    isLocked: true
  },
  {
    name: '设计图.png',
    type: 'image',
    modified: new Date('2023-07-01').getTime(),
    size: 4194304, // 4MB
    isLocked: false
  },
  {
    name: '宣传视频.mp4',
    type: 'video',
    modified: new Date('2023-07-05').getTime(),
    size: 52428800, // 50MB
    isLocked: false
  }
]);

// 下载任务数据
const downloads = ref([
  {
    name: '大型数据集.zip',
    progress: 65,
    speed: '2.4 MB/s',
    status: '下载中'
  },
  {
    name: '培训视频.mp4',
    progress: 100,
    speed: '0 B/s',
    status: '已完成'
  }
]);

// 挂载磁盘数据
const mountedDisks = ref([
  {
    name: '本地磁盘 (C:)',
    path: '/mnt/c',
    used: 42949672960, // 40GB
    total: 107374182400, // 100GB
    usagePercent: 40,
    mounted: true
  },
  {
    name: '移动硬盘 (D:)',
    path: '/mnt/d',
    used: 107374182400, // 100GB
    total: 536870912000, // 500GB
    usagePercent: 20,
    mounted: false
  }
]);

// 设置页面数据
const username = ref('cloud_user');
const email = ref('user@clouddrive.com');
const usedStorage = ref(64424509440); // 60GB
const totalStorage = ref(107374182400); // 100GB
const themeMode = ref('dark');
const showFileExtensions = ref(true);

// 上传相关状态
const uploading = ref(false);
const uploadProgress = ref(0);
const uploadingFileName = ref('');

// 计算属性
const currentPageTitle = computed(() => {
  const titles: Record<string, string> = {
    'home': '文件管理',
    'downloads': '下载管理',
    'mounts': '磁盘挂载',
    'settings': '系统设置'
  };
  return titles[activeView.value] || '文件管理';
});

const storagePercent = computed(() => {
  return Math.round((usedStorage.value / totalStorage.value) * 100);
});

// 方法：切换视图
const changeView = (viewId: string) => {
  activeView.value = viewId;
};

// 方法：切换侧边栏折叠状态
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value;
};

// 方法：获取文件图标
const getItemIcon = (type: string): string => {
  const icons: Record<string, string> = {
    'folder': 'fas fa-folder text-yellow-500',
    'doc': 'fas fa-file-word text-blue-500',
    'xls': 'fas fa-file-excel text-green-500',
    'ppt': 'fas fa-file-powerpoint text-orange-500',
    'pdf': 'fas fa-file-pdf text-red-500',
    'image': 'fas fa-file-image text-purple-500',
    'video': 'fas fa-file-video text-indigo-500',
    'audio': 'fas fa-file-audio text-pink-500',
    'zip': 'fas fa-file-archive text-teal-500'
  };
  return icons[type] || 'fas fa-file text-gray-500';
};

// 方法：获取文件类型文本
const getItemTypeText = (type: string): string => {
  const types: Record<string, string> = {
    'folder': '文件夹',
    'doc': 'Word 文档',
    'xls': 'Excel 表格',
    'ppt': 'PowerPoint 演示',
    'pdf': 'PDF 文档',
    'image': '图片文件',
    'video': '视频文件',
    'audio': '音频文件',
    'zip': '压缩文件'
  };
  return types[type] || '未知文件';
};

// 方法：格式化日期
const formatDate = (timestamp: number): string => {
  const date = new Date(timestamp);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
};

// 方法：格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// 方法：切换文件锁定状态
const toggleLock = (item: any) => {
  item.isLocked = !item.isLocked;
  // 这里可以添加实际的加密/解密逻辑
};

// 方法：下载文件
const downloadFile = (item: any) => {
  // 模拟添加到下载列表
  downloads.value.unshift({
    name: item.name,
    progress: 0,
    speed: '0 B/s',
    status: '准备中'
  });

  // 切换到下载视图
  changeView('downloads');

  // 模拟下载进度
  simulateDownloadProgress(0);
};

// 方法：删除文件
const deleteFile = (index: number) => {
  if (confirm('确定要删除这个文件吗？')) {
    fileItems.value.splice(index, 1);
  }
};

// 方法：创建新文件夹
const createNewFolder = () => {
  const folderName = prompt('请输入文件夹名称：', '新建文件夹');
  if (folderName && folderName.trim() !== '') {
    fileItems.value.unshift({
      name: folderName,
      type: 'folder',
      modified: Date.now(),
      size: 0,
      isLocked: false
    });
  }
};

// 方法：处理文件上传
const handleFileUpload = (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (input.files && input.files.length > 0) {
    const file = input.files[0];
    uploadingFileName.value = file.name;
    uploading.value = true;
    uploadProgress.value = 0;

    // 清空输入值，允许重复上传同一文件
    input.value = '';

    // 模拟上传进度
    simulateUploadProgress();
  }
};

// 方法：模拟上传进度
const simulateUploadProgress = () => {
  const interval = setInterval(() => {
    uploadProgress.value += 5;

    if (uploadProgress.value >= 100) {
      clearInterval(interval);

      // 上传完成后添加文件到列表
      setTimeout(() => {
        uploading.value = false;
        showUploadModal.value = false;

        // 确定文件类型
        let fileType = 'unknown';
        const fileName = uploadingFileName.value;

        if (fileName.endsWith('.docx') || fileName.endsWith('.doc')) fileType = 'doc';
        else if (fileName.endsWith('.xlsx') || fileName.endsWith('.xls')) fileType = 'xls';
        else if (fileName.endsWith('.pptx') || fileName.endsWith('.ppt')) fileType = 'ppt';
        else if (fileName.endsWith('.pdf')) fileType = 'pdf';
        else if (fileName.endsWith('.png') || fileName.endsWith('.jpg') || fileName.endsWith('.jpeg')) fileType = 'image';
        else if (fileName.endsWith('.mp4') || fileName.endsWith('.mov') || fileName.endsWith('.avi')) fileType = 'video';
        else if (fileName.endsWith('.zip') || fileName.endsWith('.rar') || fileName.endsWith('.7z')) fileType = 'zip';

        // 添加到文件列表
        fileItems.value.unshift({
          name: fileName,
          type: fileType,
          modified: Date.now(),
          size: 1024 * 1024 * (Math.floor(Math.random() * 10) + 1), // 1-10MB
          isLocked: false
        });
      }, 500);
    }
  }, 300);
};

// 方法：模拟下载进度
const simulateDownloadProgress = (index: number) => {
  const interval = setInterval(() => {
    if (index >= downloads.value.length) {
      clearInterval(interval);
      return;
    }

    downloads.value[index].progress += 1;
    downloads.value[index].speed = `${(Math.random() * 3 + 1).toFixed(1)} MB/s`;
    downloads.value[index].status = '下载中';

    if (downloads.value[index].progress >= 100) {
      clearInterval(interval);
      downloads.value[index].progress = 100;
      downloads.value[index].speed = '0 B/s';
      downloads.value[index].status = '已完成';
    }
  }, 200);
};

// 方法：取消下载
const cancelDownload = (index: number) => {
  downloads.value.splice(index, 1);
};

// 方法：切换磁盘挂载状态
const toggleMount = (index: number) => {
  mountedDisks.value[index].mounted = !mountedDisks.value[index].mounted;
};

// 方法：添加新磁盘
const addNewMount = () => {
  const diskName = prompt('请输入磁盘名称：', '新磁盘');
  if (diskName && diskName.trim() !== '') {
    mountedDisks.value.push({
      name: diskName,
      path: `/mnt/${diskName.toLowerCase()}`,
      used: 0,
      total: 214748364800, // 200GB
      usagePercent: 0,
      mounted: true
    });
  }
};

// 方法：保存设置
const saveSettings = () => {
  // 这里可以添加保存设置的逻辑
  alert('设置已保存！');
};

// 页面加载时启动已有的下载任务模拟
onMounted(() => {
  downloads.value.forEach((_, index) => {
    if (downloads.value[index].status === '下载中') {
      simulateDownloadProgress(index);
    }
  });
});
</script>

<style scoped>
.drive-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background-color: #f8fafc;
  color: #1e293b;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100vh;
  transition: margin-left 0.3s ease;
  overflow: hidden;
}

.main-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background-color: #ffffff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  z-index: 5;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1e293b;
}

.action-buttons {
  display: flex;
  gap: 0.75rem;
}

.btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn i {
  font-size: 1rem;
}

.upload-btn {
  background-color: #3b82f6;
  color: #ffffff;
}

.upload-btn:hover {
  background-color: #2563eb;
}

.new-folder-btn {
  background-color: #10b981;
  color: #ffffff;
}

.new-folder-btn:hover {
  background-color: #059669;
}

.content-area {
  flex: 1;
  padding: 1.5rem 2rem;
  overflow-y: auto;
}

/* 文件管理器样式 */
.file-manager {
  background-color: #ffffff;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.breadcrumb {
  display: flex;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
  font-size: 0.9rem;
}

.breadcrumb-item {
  color: #3b82f6;
  text-decoration: none;
  transition: color 0.2s ease;
}

.breadcrumb-item:hover {
  color: #2563eb;
  text-decoration: underline;
}

.separator {
  margin: 0 0.5rem;
  color: #94a3b8;
  font-size: 0.75rem;
}

.current-path {
  color: #64748b;
  font-weight: 500;
}

.file-list {
  width: 100%;
}

.file-list-header {
  display: flex;
  background-color: #f1f5f9;
  border-bottom: 1px solid #e2e8f0;
  font-weight: 600;
  font-size: 0.9rem;
}

.header-column {
  padding: 0.75rem 1.5rem;
  color: #64748b;
}

.name-column {
  flex: 3;
}

.type-column, .date-column {
  flex: 2;
}

.size-column {
  flex: 1;
}

.actions-column {
  flex: 1;
  text-align: right;
}

.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 0.3fr));
  gap: 1.5rem;
  padding: 2rem;
}

.file-card {
  position: relative;
  background: #f8fafc;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  padding: 1.2rem 1rem 2.5rem 1rem;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  min-height: 140px;
  transition: box-shadow 0.2s;
}

.file-card:hover {
  box-shadow: 0 4px 16px rgba(59,130,246,0.12);
}

.file-card.is-locked {
  background: #f0f9ff;
}

.file-icon {
  font-size: 3.5rem;
  margin-bottom: 0.7rem;

}

.file-name {
  font-weight: 500;
  font-size: 1rem;
  margin-bottom: 0.5rem;
  word-break: break-all;
}

.lock-icon {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  color: #3b82f6;
  font-size: 1rem;
}

.file-actions {
  position: absolute;
  right: 0.7rem;
  bottom: 0.7rem;
  display: flex;
  gap: 0.5rem;
}

.file-actions button {
  background: #e2e8f0;
  border: none;
  border-radius: 50%;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
}

.file-actions button:hover {
  background: #3b82f6;
  color: #fff;
}


.file-icon {
  margin-right: 0.75rem;
  font-size: 1.25rem;
}

.file-name {
  flex: 1;
}

.lock-icon {
  margin-left: 0.5rem;
  color: #3b82f6;
  font-size: 0.85rem;
}

.file-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 50%;
  background: none;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  color: #64748b;
}

.action-btn:hover {
  background-color: #f1f5f9;
  color: #1e293b;
}



.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  color: #64748b;
}

.empty-state i {
  font-size: 3rem;
  margin-bottom: 1rem;
  color: #94a3b8;
}

.empty-state p {
  margin-bottom: 1.5rem;
  font-size: 1.1rem;
}

/* 下载管理视图样式 */
.downloads-view {
  background-color: #ffffff;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
}

.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
  color: #1e293b;
}

.downloads-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.download-item {
  display: flex;
  align-items: center;
  padding: 1rem;
  border-radius: 0.5rem;
  background-color: #f8fafc;
}

.download-item.header {
  font-weight: 600;
  color: #64748b;
  background-color: #f1f5f9;
}

.download-name {
  flex: 3;
}

.download-progress {
  flex: 2;
  padding: 0 1rem;
}

.progress-bar {
  height: 0.5rem;
  background-color: #e2e8f0;
  border-radius: 0.25rem;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: #3b82f6;
  border-radius: 0.25rem;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 0.85rem;
  margin-top: 0.25rem;
  color: #64748b;
}

.download-speed, .download-status {
  flex: 1;
}

.download-actions {
  flex: 0.5;
  text-align: right;
}

/* 磁盘挂载视图样式 */
.mounts-view {
  background-color: #ffffff;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
}

.mounts-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.mount-item {
  display: flex;
  align-items: center;
  padding: 1rem;
  border-radius: 0.5rem;
  background-color: #f8fafc;
  gap: 1rem;
}

.mount-info {
  display: flex;
  align-items: center;
  flex: 2;
}

.mount-info i {
  font-size: 1.5rem;
  color: #3b82f6;
  margin-right: 1rem;
}

.mount-details {
  flex: 1;
}

.mount-name {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.mount-path {
  font-size: 0.85rem;
  color: #64748b;
}

.mount-stats {
  flex: 3;
}

.mount-usage {
  width: 100%;
}

.mount-usage span {
  display: block;
  font-size: 0.85rem;
  margin-bottom: 0.25rem;
  color: #64748b;
}

.usage-bar {
  height: 0.5rem;
  background-color: #e2e8f0;
  border-radius: 0.25rem;
  overflow: hidden;
}

.usage-fill {
  height: 100%;
  background-color: #10b981;
  border-radius: 0.25rem;
}

.mount-action-btn {
  padding: 0.5rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  background-color: #ffffff;
  cursor: pointer;
  transition: all 0.2s ease;
}

.mount-action-btn:hover {
  background-color: #f1f5f9;
}

.add-mount-btn {
  margin-top: 1rem;
  background-color: #f1f5f9;
  color: #3b82f6;
  border: 1px dashed #94a3b8;
}

.add-mount-btn:hover {
  background-color: #e2e8f0;
}

/* 设置视图样式 */
.settings-view {
  background-color: #ffffff;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
}

.settings-container {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.settings-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-heading {
  font-size: 1.1rem;
  font-weight: 600;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e2e8f0;
  color: #1e293b;
}

.setting-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.5rem 0;
}

.setting-label {
  width: 150px;
  font-weight: 500;
  color: #475569;
}

.setting-input, .setting-select {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  font-size: 0.95rem;
}

.setting-input:focus, .setting-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.storage-info {
  flex: 1;
}

.storage-summary {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}

.storage-percent {
  font-weight: 600;
}

.storage-bar {
  height: 0.75rem;
  background-color: #e2e8f0;
  border-radius: 0.375rem;
  overflow: hidden;
}

.storage-fill {
  height: 100%;
  background-color: #3b82f6;
  border-radius: 0.375rem;
}

.upgrade-storage-btn {
  margin-top: 1rem;
  background-color: #f1f5f9;
  color: #3b82f6;
  width: fit-content;
}

.upgrade-storage-btn:hover {
  background-color: #e2e8f0;
}

/* 开关样式 */
.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #e2e8f0;
  transition: .4s;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .4s;
}

input:checked + .slider {
  background-color: #3b82f6;
}

input:checked + .slider:before {
  transform: translateX(26px);
}

.slider.round {
  border-radius: 24px;
}

.slider.round:before {
  border-radius: 50%;
}

.save-settings-btn {
  background-color: #3b82f6;
  color: #ffffff;
  align-self: flex-start;
}

.save-settings-btn:hover {
  background-color: #2563eb;
}

/* 模态框样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  animation: fadeIn 0.3s ease;
}

.modal-content {
  background-color: #ffffff;
  border-radius: 0.5rem;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  animation: slideIn 0.3s ease;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1e293b;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #64748b;
  transition: color 0.2s ease;
}

.modal-close:hover {
  color: #ef4444;
}

.modal-body {
  padding: 1.5rem;
}

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 2rem;
  border: 2px dashed #94a3b8;
  border-radius: 0.5rem;
  color: #64748b;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.upload-area:hover {
  border-color: #3b82f6;
  background-color: #f0f9ff;
}

.upload-area i {
  font-size: 3rem;
  margin-bottom: 1rem;
  color: #94a3b8;
  transition: color 0.2s ease;
}

.upload-area:hover i {
  color: #3b82f6;
}

.upload-progress {
  margin-top: 1.5rem;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  margin-top: 0.5rem;
  font-size: 0.9rem;
}

.progress-filename {
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.progress-percent {
  font-weight: 600;
  margin-left: 1rem;
}

/* 动画效果 */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .date-column {
    display: none;
  }

  .mount-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .mount-info, .mount-stats {
    width: 100%;
    margin-bottom: 1rem;
  }
}

@media (max-width: 768px) {
  .main-header {
    padding: 1rem;
  }

  .page-title {
    font-size: 1.25rem;
  }

  .btn span {
    display: none;
  }

  .btn {
    padding: 0.5rem;
  }

  .content-area {
    padding: 1rem;
  }

  .type-column {
    display: none;
  }

  .download-speed {
    display: none;
  }

  .setting-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .setting-label {
    width: 100%;
    margin-bottom: 0.5rem;
  }
}

@media (max-width: 480px) {
  .size-column {
    display: none;
  }

  .download-status {
    display: none;
  }

  .header-column, .file-item {
    padding: 0.75rem;
  }




}
</style>
