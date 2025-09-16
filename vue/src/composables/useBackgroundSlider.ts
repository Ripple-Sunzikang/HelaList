import { ref, onMounted, onUnmounted } from 'vue'

export const useBackgroundSlider = (interval: number) => {
  const backgroundImages = ref([
    // 确保路径正确，建议使用 import 方式加载图片
    new URL('../picture/1.png', import.meta.url).href,
    new URL('../picture/2.png', import.meta.url).href,
    new URL('../picture/3.png', import.meta.url).href,
    new URL('../picture/4.png', import.meta.url).href
  ])
  const currentImageIndex = ref(0)
  let intervalId: number | null = null

  const nextImage = () => {
    currentImageIndex.value = (currentImageIndex.value + 1) % backgroundImages.value.length
  }

  onMounted(() => {
    intervalId = window.setInterval(nextImage, interval)
  })

  onUnmounted(() => {
    if (intervalId) {
      window.clearInterval(intervalId)
    }
  })

  return {
    currentImageIndex,
    backgroundImages
  }
}
