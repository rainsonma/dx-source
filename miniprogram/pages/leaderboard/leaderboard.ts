import { api } from '../../utils/api'
import { formatNumber } from '../../utils/format'

interface LeaderboardEntry {
  id: string; username: string; nickname: string | null
  avatarUrl: string | null; value: number; rank: number
}
interface LeaderboardResult {
  entries: LeaderboardEntry[]
  myRank: LeaderboardEntry | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: false,
    period: 'month' as 'day' | 'week' | 'month',
    lbType: 'exp' as 'exp' | 'playtime',
    entries: [] as LeaderboardEntry[],
    entries4Plus: [] as LeaderboardEntry[],
    myRank: null as LeaderboardEntry | null,
    formatNumber,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadLeaderboard()
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    (this.getTabBar() as any)?.setData({ active: 2, theme: app.globalData.theme })
  },
  async loadLeaderboard() {
    this.setData({ loading: true })
    try {
      const res = await api.get<LeaderboardResult>(
        `/api/leaderboard?type=${this.data.lbType}&period=${this.data.period}`
      )
      this.setData({ loading: false, entries: res.entries, entries4Plus: res.entries.slice(3), myRank: res.myRank })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  onPeriodChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ period: (e.detail as { name: string }).name as 'day' | 'week' | 'month' })
    this.loadLeaderboard()
  },
  onTypeChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ lbType: (e.detail as { name: string }).name as 'exp' | 'playtime' })
    this.loadLeaderboard()
  },
})
