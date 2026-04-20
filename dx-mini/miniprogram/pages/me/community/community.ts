Page({
  data: {
    statusBarHeight: 20,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ statusBarHeight })
  },
  goBack() {
    wx.navigateBack()
  },
})
