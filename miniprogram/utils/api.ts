import { config } from './config'
import { getToken, clearToken } from './auth'

export interface PaginatedData<T> {
  items: T[]
  nextCursor: string
  hasMore: boolean
}

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

function request<T>(method: string, path: string, data?: object): Promise<T> {
  return new Promise((resolve, reject) => {
    const token = getToken()
    wx.request({
      url: config.apiBaseUrl + path,
      method: method as 'GET' | 'POST' | 'PUT' | 'DELETE',
      data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success(res) {
        const body = res.data as ApiResponse<T>
        if (res.statusCode === 401) {
          clearToken()
          if (body?.code === 40104) {
            wx.showModal({
              title: '提示',
              content: '账号已在其他设备登录',
              showCancel: false,
              complete() {
                wx.reLaunch({ url: '/pages/login/login' })
              },
            })
          } else {
            wx.reLaunch({ url: '/pages/login/login' })
          }
          return reject(new Error('unauthorized'))
        }
        if (body?.code !== 0) {
          return reject(new Error(body?.message || '请求失败'))
        }
        resolve(body.data)
      },
      fail(err) {
        reject(new Error(err.errMsg || '网络错误'))
      },
    })
  })
}

export const api = {
  get<T>(path: string): Promise<T> {
    return request<T>('GET', path)
  },
  post<T>(path: string, data: object): Promise<T> {
    return request<T>('POST', path, data)
  },
  put<T>(path: string, data: object): Promise<T> {
    return request<T>('PUT', path, data)
  },
  delete<T>(path: string, data?: object): Promise<T> {
    return request<T>('DELETE', path, data)
  },
}
