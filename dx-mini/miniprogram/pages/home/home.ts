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

interface MasterStats { total: number; thisWeek: number; thisMonth: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }

interface DashboardData {
  profile: DashboardProfile
  masterStats: MasterStats
  reviewStats: ReviewStats
  sessions: SessionProgress[]
  todayAnswers: number
  greeting: Greeting
}

interface UnknownStats { total: number; thisWeek: number; thisMonth: number }

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
    masterTotal: null as number | null,
    reviewPending: null as number | null,
    unknownTotal: null as number | null,
    recentSession: null as RecentSession | null,
    vipDueAt: '' as string,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ theme: app.globalData.theme, statusBarHeight })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ active: 0, theme: app.globalData.theme }) }
    this.loadData()
  },
  async loadData() {
    this.setData({ loading: true })
    const [dashResult, unknownResult] = await Promise.allSettled([
      api.get<DashboardData>('/api/hall/dashboard'),
      api.get<UnknownStats>('/api/tracking/unknown/stats'),
    ])

    if (dashResult.status === 'rejected') {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
      return
    }

    const dash = dashResult.value
    const unknownTotal = unknownResult.status === 'fulfilled'
      ? unknownResult.value.total
      : null

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
      masterTotal: dash.masterStats.total,
      reviewPending: dash.reviewStats.pending,
      unknownTotal,
      recentSession,
      vipDueAt: dash.profile.vipDueAt || '',
    })
  },
  goSearch() { wx.navigateTo({ url: '/pages/games/games' }) },
  goPurchase() { wx.navigateTo({ url: '/pages/me/purchase/purchase' }) },
  goInvite() { wx.navigateTo({ url: '/pages/me/invite/invite' }) },
  goStudy() { wx.navigateTo({ url: '/pages/me/study/study' }) },
  goGroups() { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  goTasks() { wx.navigateTo({ url: '/pages/me/tasks/tasks' }) },
  goCommunity() { wx.navigateTo({ url: '/pages/me/community/community' }) },
  goFeedback() { wx.navigateTo({ url: '/pages/me/feedback/feedback' }) },
})
