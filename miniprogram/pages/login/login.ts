import { api } from '../../utils/api'
import { setToken, setUserId } from '../../utils/auth'
import { ws } from '../../utils/ws'

interface AuthResponse {
  token: string
  user: { id: string }
}

Page({
  data: {
    loading: false,
    theme: 'light' as 'light' | 'dark',
  },
  onLoad() {
    const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()
    this.setData({ theme: app.globalData.theme })
  },
  handleLogin() {
    if (this.data.loading) return
    this.setData({ loading: true })
    wx.login({
      success: (res) => {
        api.post<AuthResponse>('/api/auth/wechat-mini', { code: res.code })
          .then((data) => {
            setToken(data.token)
            setUserId(data.user.id)
            const app = getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()
            app.globalData.userId = data.user.id
            ws.connect(data.token)
            ws.subscribe(`user::${data.user.id}`)
            wx.reLaunch({ url: '/pages/home/home' })
          })
          .catch((err: Error) => {
            wx.showToast({ title: err.message || '登录失败', icon: 'none' })
            this.setData({ loading: false })
          })
      },
      fail: () => {
        wx.showToast({ title: '获取登录凭证失败', icon: 'none' })
        this.setData({ loading: false })
      },
    })
  },
})
