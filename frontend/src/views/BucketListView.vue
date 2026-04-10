<template>
  <div class="page">
    <n-layout>
      <n-layout-header class="header">
        <span class="title">S3 Web Manager</span>
        <n-button size="small" @click="handleLogout">退出登录</n-button>
      </n-layout-header>
      <n-layout-content class="content">
        <n-card title="Bucket 列表">
          <n-data-table
            :columns="columns"
            :data="buckets"
            :loading="loading"
            :row-props="rowProps"
            striped
          />
        </n-card>
      </n-layout-content>
    </n-layout>
  </div>
</template>

<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { h, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { type BucketInfo, listBuckets } from '../api/buckets'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const message = useMessage()
const loading = ref(false)
const buckets = ref<BucketInfo[]>([])

const columns: DataTableColumns<BucketInfo> = [
  { title: 'Bucket 名称', key: 'name' },
  { title: '所有者', key: 'owner' },
  {
    title: '对象数量',
    key: 'numObjects',
    render: (row) => h('span', row.numObjects.toLocaleString()),
  },
  {
    title: '占用空间',
    key: 'sizeBytes',
    render: (row) => h('span', formatBytes(row.sizeBytes)),
  },
]

function rowProps(row: BucketInfo) {
  return {
    style: 'cursor: pointer',
    onClick: () => router.push({ name: 'BucketDetail', params: { name: row.name } }),
  }
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`
}

async function fetchBuckets() {
  loading.value = true
  try {
    const res = await listBuckets()
    buckets.value = res.data
  } catch {
    message.error('获取 bucket 列表失败')
  } finally {
    loading.value = false
  }
}

function handleLogout() {
  auth.clearToken()
  router.push({ name: 'Login' })
}

onMounted(fetchBuckets)
</script>

<style scoped>
.page {
  min-height: 100vh;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 56px;
  background: #18a058;
  color: #fff;
}
.title {
  font-size: 18px;
  font-weight: 600;
}
.content {
  padding: 24px;
}
</style>
