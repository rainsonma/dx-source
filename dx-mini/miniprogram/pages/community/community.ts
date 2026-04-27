const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
    })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
    const tabBar = this.getTabBar() as WechatMiniprogram.Component.TrivialInstance | null
    if (tabBar) tabBar.setData({ active: 3, theme: app.globalData.theme })
  },
})
