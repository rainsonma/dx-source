const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

const TIERS = [
  { id: 'monthly', name: '月度会员', price: '¥19', desc: '30天无限访问' },
  { id: 'quarterly', name: '季度会员', price: '¥49', desc: '90天无限访问' },
  { id: 'yearly', name: '年度会员', price: '¥149', desc: '365天无限访问' },
]

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    tiers: TIERS,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
  },
  onBuy() {
    wx.showToast({ title: '即将开放', icon: 'none' })
  },
})
