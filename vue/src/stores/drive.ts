import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Downloading {
  name: string
  progress: number
}

export const useDriveStore = defineStore('drive', () => {
  // currentPath: '' means root
  const currentPath = ref('')
  const downloading = ref<Downloading[]>([])

  function setPath(p: string) {
    currentPath.value = p || ''
  }

  function addDownloading(name: string) {
    downloading.value.push({ name, progress: 0 })
  }

  function updateDownloadProgress(name: string, progress: number) {
    const download = downloading.value.find((d) => d.name === name)
    if (download) {
      download.progress = progress
    }
  }

  function removeDownloading(name: string) {
    downloading.value = downloading.value.filter((d) => d.name !== name)
  }

  return {
    currentPath,
    setPath,
    downloading,
    addDownloading,
    updateDownloadProgress,
    removeDownloading,
  }
})
