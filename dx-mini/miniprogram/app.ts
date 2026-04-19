import { isLoggedIn, getToken, getUserId, clearToken } from './utils/auth'
import { ws } from './utils/ws'

interface GlobalData {
  theme: 'light' | 'dark'
  userId: string
}

App<{ globalData: GlobalData }>({
  globalData: {
    theme: 'light',
    userId: '',
  },
  onLaunch() {
    const stored = wx.getStorageSync('dx_theme') as 'light' | 'dark' | ''
    const sys = wx.getSystemSetting()
    this.globalData.theme = stored || ((sys.theme as 'light' | 'dark') || 'light')

    wx.onThemeChange(({ theme }) => {
      if (!wx.getStorageSync('dx_theme')) {
        this.globalData.theme = theme as 'light' | 'dark'
      }
    })

    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }

    this.globalData.userId = getUserId() || ''
    const token = getToken()!
    ws.connect(token)
    ws.subscribe(`user::${this.globalData.userId}`)
    ws.on('session_replaced', () => {
      clearToken()
      wx.reLaunch({ url: '/pages/login/login' })
    })
  },
})
