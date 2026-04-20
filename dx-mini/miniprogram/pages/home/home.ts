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
}

interface Greeting { title: string; subtitle: string }

interface DashboardData {
  profile: DashboardProfile
  greeting: Greeting
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
    try {
      const dash = await api.get<DashboardData>('/api/hall/dashboard')
      this.setData({
        loading: false,
        profile: dash.profile,
        greeting: dash.greeting,
        gradeLabelText: gradeLabel(dash.profile.grade),
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
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
