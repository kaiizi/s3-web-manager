import http from './http'

export interface BucketInfo {
  name: string
  owner: string
  numObjects: number
  sizeBytes: number
}

export function listBuckets() {
  return http.get<BucketInfo[]>('/buckets')
}

export function getBucket(name: string) {
  return http.get<BucketInfo>(`/buckets/${name}`)
}
