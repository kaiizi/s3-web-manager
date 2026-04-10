<template>
  <n-data-table
    :columns="columns"
    :data="tableData"
    :loading="loading"
    striped
    :row-key="(row: TableRow) => row.key"
  />
</template>

<script setup lang="ts">
import { DownloadOutline, FolderOutline, TrashOutline } from '@vicons/ionicons5'
import { NButton, NDataTable, NIcon, NSpace, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { computed, h } from 'vue'
import type { ObjectInfo } from '../api/objects'
import { deleteObject, getDownloadUrl } from '../api/objects'

interface PrefixRow { type: 'prefix'; key: string; name: string }
interface ObjectRow { type: 'object'; key: string; name: string; size: number; lastModified: string }
type TableRow = PrefixRow | ObjectRow

const props = defineProps<{
  bucket: string
  prefixes: string[]
  objects: ObjectInfo[]
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'enter-prefix', prefix: string): void
  (e: 'refresh'): void
}>()

const message = useMessage()
const dialog = useDialog()

const tableData = computed<TableRow[]>(() => [
  ...props.prefixes.map((p) => ({ type: 'prefix' as const, key: p, name: p })),
  ...props.objects.map((o) => ({
    type: 'object' as const,
    key: o.key,
    name: o.key,
    size: o.size,
    lastModified: o.lastModified,
  })),
])

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`
}

async function handleDownload(row: ObjectRow) {
  try {
    const { data } = await getDownloadUrl(props.bucket, row.key)
    window.location.href = data.url
  } catch {
    message.error('获取下载链接失败')
  }
}

function handleDelete(row: ObjectRow) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除 "${row.key}" 吗？此操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteObject(props.bucket, row.key)
        message.success('删除成功')
        emit('refresh')
      } catch {
        message.error('删除失败')
      }
    },
  })
}

const columns: DataTableColumns<TableRow> = [
  {
    title: '名称',
    key: 'name',
    render: (row) => {
      if (row.type === 'prefix') {
        return h(NSpace, { align: 'center', style: 'cursor:pointer' }, {
          default: () => [
            h(NIcon, null, { default: () => h(FolderOutline) }),
            h('span', {
              onClick: () => emit('enter-prefix', row.name),
              style: 'color: #18a058; cursor: pointer',
            }, row.name),
          ],
        })
      }
      return h('span', row.name)
    },
  },
  {
    title: '大小',
    key: 'size',
    width: 120,
    render: (row) => row.type === 'object' ? formatBytes(row.size) : '—',
  },
  {
    title: '最后修改时间',
    key: 'lastModified',
    width: 200,
    render: (row) =>
      row.type === 'object'
        ? new Date(row.lastModified).toLocaleString('zh-CN')
        : '—',
  },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    render: (row) => {
      if (row.type !== 'object') return null
      return h(NSpace, null, {
        default: () => [
          h(NButton, {
            size: 'small',
            onClick: () => handleDownload(row as ObjectRow),
          }, {
            default: () => h(NIcon, null, { default: () => h(DownloadOutline) }),
          }),
          h(NButton, {
            size: 'small',
            type: 'error',
            onClick: () => handleDelete(row as ObjectRow),
          }, {
            default: () => h(NIcon, null, { default: () => h(TrashOutline) }),
          }),
        ],
      })
    },
  },
]
</script>
