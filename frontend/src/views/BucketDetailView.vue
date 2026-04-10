<template>
  <div class="page">
    <n-layout>
      <n-layout-header class="header">
        <span>
          <router-link to="/buckets" style="color:#fff;text-decoration:none">S3 Web Manager</router-link>
          &nbsp;/&nbsp;{{ bucketName }}
        </span>
        <n-button size="small" @click="handleLogout">退出登录</n-button>
      </n-layout-header>
      <n-layout-content class="content">
        <n-card>
          <template #header>
            <!-- Breadcrumb -->
            <n-breadcrumb>
              <n-breadcrumb-item @click="navigateTo('')" style="cursor:pointer">根目录</n-breadcrumb-item>
              <n-breadcrumb-item
                v-for="(part, idx) in breadcrumbParts"
                :key="idx"
                style="cursor:pointer"
                @click="navigateTo(breadcrumbPrefixAt(idx))"
              >
                {{ part }}
              </n-breadcrumb-item>
            </n-breadcrumb>
          </template>
          <template #header-extra>
            <n-button type="primary" @click="showUpload = true">上传文件</n-button>
          </template>

          <ObjectTable
            :bucket="bucketName"
            :prefixes="result.prefixes"
            :objects="result.objects"
            :loading="loading"
            @enter-prefix="enterPrefix"
            @refresh="fetchObjects"
          />
        </n-card>
      </n-layout-content>
    </n-layout>

    <UploadModal
      v-model:show="showUpload"
      :bucket="bucketName"
      :prefix="currentPrefix"
      @uploaded="fetchObjects"
    />
  </div>
</template>

<script setup lang="ts">
import {
  NBreadcrumb,
  NBreadcrumbItem,
  NButton,
  NCard,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  useMessage,
} from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { ListResult } from '../api/objects'
import { listObjects } from '../api/objects'
import { useAuthStore } from '../stores/auth'
import ObjectTable from '../components/ObjectTable.vue'
import UploadModal from '../components/UploadModal.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const message = useMessage()

const bucketName = computed(() => route.params.name as string)
const currentPrefix = ref('')
const loading = ref(false)
const showUpload = ref(false)
const result = ref<ListResult>({ prefixes: [], objects: [] })

const breadcrumbParts = computed(() =>
  currentPrefix.value ? currentPrefix.value.split('/').filter(Boolean) : []
)

function breadcrumbPrefixAt(idx: number): string {
  return breadcrumbParts.value.slice(0, idx + 1).join('/') + '/'
}

function enterPrefix(prefix: string) {
  currentPrefix.value = prefix
  fetchObjects()
}

function navigateTo(prefix: string) {
  currentPrefix.value = prefix
  fetchObjects()
}

async function fetchObjects() {
  loading.value = true
  try {
    const res = await listObjects(bucketName.value, currentPrefix.value)
    result.value = res.data
  } catch {
    message.error('获取对象列表失败')
  } finally {
    loading.value = false
  }
}

function handleLogout() {
  auth.clearToken()
  router.push({ name: 'Login' })
}

onMounted(fetchObjects)
</script>

<style scoped>
.page { min-height: 100vh; }
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 56px;
  background: #18a058;
  color: #fff;
}
.content { padding: 24px; }
</style>
