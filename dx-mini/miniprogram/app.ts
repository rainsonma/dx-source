import { isLoggedIn, getToken, getUserId, clearToken } from './utils/auth'
import { ws } from './utils/ws'

// Suppress a known WeChat DevTools startup artifact (dev only):
//   "Error: SystemError (appServiceSDKScriptError)\ntimeout"
// Emitted from WAServiceMainContext.js during the IDE's internal SDK handshake.
// Not produced by app code; does not appear on real devices. Active only in
// develop envVersion — release / trial / 体验版 are untouched.
{
  const { envVersion } = wx.getAccountInfoSync().miniProgram
  if (envVersion === 'develop') {
    const origConsoleError = console.error.bind(console)
    console.error = (...args: unknown[]) => {
      for (const a of args) {
        const text = a instanceof Error ? a.name + ' ' + a.message : String(a)
        if (text.includes('appServiceSDKScriptError') && text.toLowerCase().includes('timeout')) {
          return
        }
      }
      origConsoleError(...args)
    }
  }
}

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
    this.globalData.theme = stored || ((sys.theme as 'light' | 'dark') ?? 'light')

    wx.onThemeChange(({ theme }) => {
      if (!wx.getStorageSync('dx_theme')) {
        this.globalData.theme = theme as 'light' | 'dark'
      }
    })

    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }

    this.globalData.userId = getUserId() ?? ''
    const token = getToken()!
    ws.connect(token)
    ws.subscribe(`user::${this.globalData.userId}`)
    ws.on('session_replaced', () => {
      clearToken()
      wx.reLaunch({ url: '/pages/login/login' })
    })
  },
})
