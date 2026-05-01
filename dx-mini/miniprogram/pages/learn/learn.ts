import { api } from '../../utils/api'

interface Stats { total: number; thisWeek: number; thisMonth: number }
interface UnknownStats { total: number; today: number; lastThreeDays: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }

interface SessionProgress {
  gameId: string
  gameName: string
  gameMode: string
  completedLevels: number
  totalLevels: number
  score: number
  exp: number
  lastPlayedAt: string
}

interface ProgressItem {
  gameId: string
  title: string
  progressPct: number
  barColor: string
  barWidth: string
}

const PAGE_SIZE = 5

const PROGRESS_COLORS = [
  '#14b8a6',
  '#3b82f6',
  '#f59e0b',
  '#ec4899',
  '#8b5cf6',
  '#06b6d4',
]

const GAME_MODE_LABELS: Record<string, string> = {
  'word-sentence': '连词成句',
  'vocab-battle': '词汇对轰',
  'vocab-match': '词汇配对',
  'vocab-elimination': '词汇消消乐',
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    arrowColor: '#9ca3af',
    accentColors: { teal: '#10b981', amber: '#f59e0b', purple: '#6366f1' } as { teal: string; amber: string; purple: string },
    statusBarHeight: 20,
    loading: true,
    masterStats: null as Stats | null,
    unknownStats: null as UnknownStats | null,
    reviewStats: null as ReviewStats | null,
    sessions: [] as SessionProgress[],
    progressItems: [] as ProgressItem[],
    pageItems: [] as ProgressItem[],
    progressPage: 1,
    progressTotalPages: 1,
    hasPagination: false,
    prevDisabled: true,
    nextDisabled: true,
    pageLabel: '第 1 / 1 页',
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({ statusBarHeight: sys.statusBarHeight || 20 })
  },
  onShow() {
    this.setData({
      theme: app.globalData.theme,
      arrowColor: app.globalData.theme === 'dark' ? '#6b7280' : '#9ca3af',
    })
    this.loadAll()
  },
  async loadAll() {
    this.setData({ loading: true })
    const results = await Promise.allSettled([
      api.get<Stats>('/api/tracking/master/stats'),
      api.get<UnknownStats>('/api/tracking/unknown/stats'),
      api.get<ReviewStats>('/api/tracking/review/stats'),
      api.get<SessionProgress[]>('/api/tracking/sessions'),
    ])
    const masterStats = results[0].status === 'fulfilled' ? results[0].value : this.data.masterStats
    const unknownStats = results[1].status === 'fulfilled' ? results[1].value : this.data.unknownStats
    const reviewStats = results[2].status === 'fulfilled' ? results[2].value : this.data.reviewStats
    const sessions = results[3].status === 'fulfilled' ? results[3].value : []

    this.setData({ loading: false, masterStats, unknownStats, reviewStats, sessions })
    this.rebuildProgress(this.data.progressPage)

    if (results.some((r) => r.status === 'rejected')) {
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  rebuildProgress(targetPage: number) {
    const items: ProgressItem[] = this.data.sessions.map((s, i) => {
      const modeText = GAME_MODE_LABELS[s.gameMode] || s.gameMode
      const pct = s.totalLevels === 0 ? 0 : Math.round((s.completedLevels / s.totalLevels) * 100)
      return {
        gameId: s.gameId,
        title: s.gameName + ' · ' + modeText,
        progressPct: pct,
        barColor: PROGRESS_COLORS[i % PROGRESS_COLORS.length],
        barWidth: pct + '%',
      }
    })

    const totalPages = Math.max(1, Math.ceil(items.length / PAGE_SIZE))
    const page = Math.min(Math.max(1, targetPage), totalPages)
    const start = (page - 1) * PAGE_SIZE
    const pageItems = items.slice(start, start + PAGE_SIZE)
    const hasPagination = items.length > PAGE_SIZE

    this.setData({
      progressItems: items,
      pageItems,
      progressPage: page,
      progressTotalPages: totalPages,
      hasPagination,
      prevDisabled: page === 1,
      nextDisabled: page === totalPages,
      pageLabel: '第 ' + page + ' / ' + totalPages + ' 页',
    })
  },
  prevProgressPage() {
    if (this.data.prevDisabled) return
    this.rebuildProgress(this.data.progressPage - 1)
  },
  nextProgressPage() {
    if (this.data.nextDisabled) return
    this.rebuildProgress(this.data.progressPage + 1)
  },
  goGame(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string | undefined
    if (id) wx.navigateTo({ url: '/pages/games/detail/detail?id=' + id })
  },
  goMastered() { wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
  goReview() { wx.navigateTo({ url: '/pages/learn/review/review' }) },
})
