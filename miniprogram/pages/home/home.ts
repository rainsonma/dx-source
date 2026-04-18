import { api } from '../../utils/api'

interface DashboardProfile {
  id: string
  username: string
  nickname: string | null
  grade: string
  exp: number
  beans: number
  avatarUrl: string | null
  currentPlayStreak: number
  inviteCode: string
  lastReadNoticeAt: string | null
}

interface MasterStats { total: number; thisWeek: number; thisMonth: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }
interface Greeting { text: string; emoji: string }

interface DashboardData {
  profile: DashboardProfile
  masterStats: MasterStats
  reviewStats: ReviewStats
  todayAnswers: number
  greeting: Greeting
}

interface HeatmapDay { date: string; count: number }
interface HeatmapData { year: number; days: HeatmapDay[]; accountYear: number }

interface HeatmapCell { date: string; level: number }

const app = getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    profile: null as DashboardProfile | null,
    masterStats: null as MasterStats | null,
    reviewStats: null as ReviewStats | null,
    todayAnswers: 0,
    greeting: null as Greeting | null,
    heatmapCells: [] as HeatmapCell[],
    unreadNotices: false,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    (this.getTabBar() as any)?.setData({ active: 0, theme: app.globalData.theme })
    this.loadData()
  },
  async loadData() {
    this.setData({ loading: true })
    try {
      const [dash, heatmap] = await Promise.all([
        api.get<DashboardData>('/api/hall/dashboard'),
        api.get<HeatmapData>('/api/hall/heatmap'),
      ])
      const cells = this.buildHeatmapCells(heatmap.days)
      this.setData({
        loading: false,
        profile: dash.profile,
        masterStats: dash.masterStats,
        reviewStats: dash.reviewStats,
        todayAnswers: dash.todayAnswers,
        greeting: dash.greeting,
        heatmapCells: cells,
        unreadNotices: false,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  buildHeatmapCells(days: HeatmapDay[]): HeatmapCell[] {
    const map = new Map(days.map(d => [d.date, d.count]))
    const cells: HeatmapCell[] = []
    const today = new Date()
    for (let i = 48; i >= 0; i--) {
      const d = new Date(today)
      d.setDate(d.getDate() - i)
      const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
      const count = map.get(key) ?? 0
      const level = count === 0 ? 0 : count < 3 ? 1 : count < 6 ? 2 : count < 10 ? 3 : 4
      cells.push({ date: key, level })
    }
    return cells
  },
  toggleTheme() {
    const next: 'light' | 'dark' = this.data.theme === 'light' ? 'dark' : 'light'
    wx.setStorageSync('dx_theme', next)
    app.globalData.theme = next
    this.setData({ theme: next });
    (this.getTabBar() as any)?.setData({ theme: next })
  },
  goNotices() {
    wx.navigateTo({ url: '/pages/me/notices/notices' })
  },
  goSearch() {
    wx.navigateTo({ url: '/pages/games/games' })
  },
})
