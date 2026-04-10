<template>
  <n-modal v-model:show="show" title="上传文件" preset="card" style="width: 440px">
    <n-upload
      :custom-request="handleUpload"
      :show-file-list="true"
      multiple
    >
      <n-upload-dragger>
        <div style="padding: 20px 0">
          <n-icon size="48" :depth="3"><CloudUploadOutline /></n-icon>
          <p>点击或拖拽文件到此区域上传</p>
        </div>
      </n-upload-dragger>
    </n-upload>

    <n-progress
      v-if="progress > 0 && progress < 100"
      type="line"
      :percentage="progress"
      style="margin-top: 12px"
    />
  </n-modal>
</template>

<script setup lang="ts">
import { CloudUploadOutline } from '@vicons/ionicons5'
import {
  NIcon,
  NModal,
  NProgress,
  NUpload,
  NUploadDragger,
  useMessage,
} from 'naive-ui'
import type { UploadCustomRequestOptions } from 'naive-ui'
import { ref } from 'vue'
import { getUploadUrl } from '../api/objects'

const props = defineProps<{ bucket: string; prefix: string }>()
const emit = defineEmits<{ (e: 'uploaded'): void }>()

const show = defineModel<boolean>('show', { default: false })
const message = useMessage()
const progress = ref(0)

async function handleUpload({ file, onFinish, onError, onProgress }: UploadCustomRequestOptions) {
  try {
    const key = props.prefix + file.name
    const contentType = file.type || 'application/octet-stream'
    const { data } = await getUploadUrl(props.bucket, key, contentType)

    await new Promise<void>((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
          const pct = Math.round((e.loaded / e.total) * 100)
          progress.value = pct
          onProgress({ percent: pct })
        }
      }
      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve()
        } else {
          reject(new Error(`Upload failed: ${xhr.status}`))
        }
      }
      xhr.onerror = () => reject(new Error('Network error'))
      xhr.open('PUT', data.url)
      xhr.setRequestHeader('Content-Type', contentType)
      xhr.send(file.file)
    })

    onFinish()
    message.success(`${file.name} 上传成功`)
    progress.value = 0
    emit('uploaded')
  } catch (err) {
    onError()
    message.error(`上传失败: ${(err as Error).message}`)
    progress.value = 0
  }
}
</script>
