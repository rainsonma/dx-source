import { api } from '../../utils/api'
import { gradeLabel } from '../../utils/format'

interface DashboardProfile {
  id: string
  username: string
  nickname: string | null
  grade: string
  level: number
  exp: number
  beans: number
  avatarUrl: string | null
  currentPlayStreak: number
  inviteCode: string
  lastReadNoticeAt: string | null
  vipDueAt: string | null
}

interface Greeting { title: string; subtitle: string }

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

interface DashboardData {
  profile: DashboardProfile
  sessions: SessionProgress[]
  todayAnswers: number
  greeting: Greeting
}

interface RecentSession {
  gameId: string
  gameName: string
  completedLevels: number
}

const app = getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    profile: null as DashboardProfile | null,
    greeting: null as Greeting | null,
    gradeLabelText: '',
    statusBarHeight: 20,
    // marketing sections
    recentSession: null as RecentSession | null,
    vipDueAt: '' as string,
    compactRevealed: false,
    heroBottomPx: 0,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ theme: app.globalData.theme, statusBarHeight })
  },
  onReady() {
    wx.createSelectorQuery()
      .in(this)
      .select('.search-row')
      .boundingClientRect((rect) => {
        if (rect && typeof rect.bottom === 'number') {
          this.setData({ heroBottomPx: rect.bottom })
        }
      })
      .exec()
  },
  onPageScroll(e: WechatMiniprogram.Page.IPageScrollOption) {
    const threshold = this.data.heroBottomPx
    if (threshold <= 0) return
    const shouldReveal = e.scrollTop >= threshold
    if (shouldReveal !== this.data.compactRevealed) {
      this.setData({ compactRevealed: shouldReveal })
    }
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ active: 0, theme: app.globalData.theme }) }
    this.loadData()
  },
  async loadData() {
    this.setData({ loading: true })
    try {
      const dash = await api.get<DashboardData>('/api/hall/dashboard')

      const sessions = dash.sessions || []
      const recentSession: RecentSession | null = sessions.length > 0
        ? {
            gameId: sessions[0].gameId,
            gameName: sessions[0].gameName,
            completedLevels: sessions[0].completedLevels,
          }
        : null

      this.setData({
        loading: false,
        profile: dash.profile,
        greeting: dash.greeting,
        gradeLabelText: gradeLabel(dash.profile.grade),
        recentSession,
        vipDueAt: dash.profile.vipDueAt || '',
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  goSearch() { wx.navigateTo({ url: '/pages/games/search/search' }) },
  goPurchase() { wx.navigateTo({ url: '/pages/me/purchase/purchase' }) },
  goInvite() { wx.navigateTo({ url: '/pages/me/invite/invite' }) },
  goStudy() { wx.navigateTo({ url: '/pages/learn/learn' }) },
  goGroups() { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  goTasks() { wx.navigateTo({ url: '/pages/me/tasks/tasks' }) },
  goNotices() { wx.navigateTo({ url: '/pages/me/notices/notices' }) },
  goFeedback() { wx.navigateTo({ url: '/pages/me/feedback/feedback' }) },
})
