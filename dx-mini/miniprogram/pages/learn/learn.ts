import { api } from '../../utils/api'

interface Stats { total: number; thisWeek: number; thisMonth: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    arrowColor: '#9ca3af',
    accentColors: { teal: '#10b981', amber: '#f59e0b', purple: '#6366f1' } as { teal: string; amber: string; purple: string },
    loading: true,
    masterStats: null as Stats | null,
    unknownStats: null as Stats | null,
    reviewStats: null as ReviewStats | null,
  },
  onShow() {
    this.setData({
      theme: app.globalData.theme,
      arrowColor: app.globalData.theme === 'dark' ? '#6b7280' : '#9ca3af',
    });
    this.loadStats()
  },
  async loadStats() {
    this.setData({ loading: true })
    try {
      const [masterStats, unknownStats, reviewStats] = await Promise.all([
        api.get<Stats>('/api/tracking/master/stats'),
        api.get<Stats>('/api/tracking/unknown/stats'),
        api.get<ReviewStats>('/api/tracking/review/stats'),
      ])
      this.setData({ loading: false, masterStats, unknownStats, reviewStats })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  goMastered() { wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
  goReview() { wx.navigateTo({ url: '/pages/learn/review/review' }) },
})
