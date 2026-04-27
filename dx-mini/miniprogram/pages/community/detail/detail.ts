const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    postId: '',
  },
  onLoad(query: Record<string, string>) {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
      postId: query.id || '',
    })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
  },
  goBack() {
    wx.navigateBack()
  },
})
