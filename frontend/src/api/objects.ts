import http from './http'

export interface ObjectInfo {
  key: string
  size: number
  lastModified: string
}

export interface ListResult {
  prefixes: string[]
  objects: ObjectInfo[]
}

export function listObjects(bucket: string, prefix = '') {
  return http.get<ListResult>(`/buckets/${bucket}/objects`, { params: { prefix } })
}

export function getUploadUrl(bucket: string, key: string, contentType: string) {
  return http.post<{ url: string; method: string }>(`/buckets/${bucket}/objects/upload-url`, {
    key,
    contentType,
  })
}

export function getDownloadUrl(bucket: string, key: string) {
  return http.get<{ url: string }>(`/buckets/${bucket}/objects/download-url`, {
    params: { key },
  })
}

export function deleteObject(bucket: string, key: string) {
  return http.delete(`/buckets/${bucket}/objects`, { params: { key } })
}
