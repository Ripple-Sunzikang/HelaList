import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useDriveStore = defineStore('drive', () => {
  // currentPath: '' means root
  const currentPath = ref('')

  function setPath(p: string) {
    currentPath.value = p || ''
  }

  return { currentPath, setPath }
})
