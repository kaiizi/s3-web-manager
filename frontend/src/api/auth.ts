import http from './http'

export function login(username: string, password: string) {
  return http.post<{ token: string }>('/auth/login', { username, password })
}
