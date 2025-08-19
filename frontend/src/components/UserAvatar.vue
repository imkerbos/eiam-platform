<template>
  <div class="user-avatar-container">
    <a-avatar 
      :size="size" 
      :src="showImage ? avatarSrc : undefined"
      :style="avatarStyle"
      :class="avatarClass"
      @error="handleImageError"
    >
      {{ avatarText }}
    </a-avatar>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

interface Props {
  user?: any // 允许任何用户对象类型
  username?: string
  displayName?: string
  avatar?: string
  size?: number | string
  showBorder?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 40,
  showBorder: false
})

// 控制是否显示图片
const showImage = ref(true)

// 处理图片加载错误
const handleImageError = () => {
  console.log('头像图片加载失败，切换到字母头像')
  showImage.value = false
}

// 当头像路径改变时，重置显示状态
watch(() => props.avatar || props.user?.avatar, () => {
  showImage.value = true
})

// 计算头像源地址
const avatarSrc = computed(() => {
  const avatar = props.avatar || props.user?.avatar
  if (!avatar || avatar === '') return undefined
  
  // 如果是完整URL，直接返回
  if (avatar.startsWith('http://') || avatar.startsWith('https://')) {
    return avatar
  }
  
  // 如果是相对路径，直接返回
  if (avatar.startsWith('/')) {
    return avatar
  }
  
  // 如果是文件路径，添加根路径前缀
  return `/${avatar}`
})

// 计算头像文字（显示姓名首字母）
const avatarText = computed(() => {
  if (avatarSrc.value && showImage.value) return '' // 有头像图片且能正常显示时不显示文字
  
  const displayName = props.displayName || props.user?.display_name
  const username = props.username || props.user?.username
  
  const name = displayName || username || 'U'
  
  // 调试信息
  if (import.meta.env.DEV) {
    console.log('UserAvatar debug:', {
      avatarSrc: avatarSrc.value,
      displayName,
      username,
      finalName: name,
      user: props.user
    })
  }
  
  // 如果是中文名，取最后一个字符
  if (/[\u4e00-\u9fa5]/.test(name)) {
    return name.charAt(name.length - 1).toUpperCase()
  }
  
  // 如果是英文名，取第一个字符
  return name.charAt(0).toUpperCase()
})

// 计算头像样式
const avatarStyle = computed(() => {
  const style: Record<string, string> = {}
  
  // 如果没有头像图片或图片加载失败，设置背景色
  if (!avatarSrc.value || !showImage.value) {
    const colors = [
      '#1890ff', '#52c41a', '#faad14', '#f5222d', 
      '#722ed1', '#13c2c2', '#eb2f96', '#fa541c'
    ]
    
    const name = props.displayName || props.user?.display_name || props.username || props.user?.username || 'User'
    const colorIndex = name.charCodeAt(0) % colors.length
    style.backgroundColor = colors[colorIndex]
    style.color = '#fff'
  }
  
  if (props.showBorder) {
    style.border = '2px solid #f0f0f0'
  }
  
  return style
})

// 计算头像CSS类
const avatarClass = computed(() => {
  return {
    'user-avatar': true,
    'user-avatar--with-border': props.showBorder
  }
})
</script>

<style scoped>
.user-avatar-container {
  display: inline-block;
}

.user-avatar {
  transition: all 0.3s ease;
  cursor: default;
}

.user-avatar--with-border {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.user-avatar:hover {
  transform: scale(1.05);
}
</style>
