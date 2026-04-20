import { api } from '../../utils/api'
import { formatDate, gradeLabel } from '../../utils/format'
import { clearToken } from '../../utils/auth'
import { ws } from '../../utils/ws'

interface ProfileData {
  id: string; grade: string; username: string; nickname: string | null
  avatarUrl: string | null; city: string | null; beans: number
  exp: number; level: number; inviteCode: string; currentPlayStreak: number
  vipDueAt: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    arrowColor: '#9ca3af',
    cellIconColor: '#6b7280',
    loading: true,
    profile: null as ProfileData | null,
    avatarChar: '',
    statusBarHeight: 20,
    formatDate,
    gradeLabel,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ statusBarHeight })
  },
  onShow() {
    const theme = app.globalData.theme
    this.setData({
      theme,
      primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488',
      arrowColor: theme === 'dark' ? '#6b7280' : '#9ca3af',
      cellIconColor: theme === 'dark' ? '#9ca3af' : '#6b7280',
    });
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ active: 4, theme }) }
    this.loadProfile()
  },
  async loadProfile() {
    try {
      const profile = await api.get<ProfileData>('/api/user/profile')
      const avatarChar = (profile.nickname || profile.username).charAt(0)
      this.setData({ loading: false, profile, avatarChar })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  toggleTheme() {
    const next: 'light' | 'dark' = this.data.theme === 'light' ? 'dark' : 'light'
    wx.setStorageSync('dx_theme', next)
    app.globalData.theme = next
    this.setData({
      theme: next,
      primaryColor: next === 'dark' ? '#14b8a6' : '#0d9488',
      arrowColor: next === 'dark' ? '#6b7280' : '#9ca3af',
      cellIconColor: next === 'dark' ? '#9ca3af' : '#6b7280',
    })
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ theme: next }) }
  },
  goProfileEdit() { wx.navigateTo({ url: '/pages/me/profile-edit/profile-edit' }) },
  goNotices() { wx.navigateTo({ url: '/pages/me/notices/notices' }) },
  goGroups() { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  goInvite() { wx.navigateTo({ url: '/pages/me/invite/invite' }) },
  goRedeem() { wx.navigateTo({ url: '/pages/me/redeem/redeem' }) },
  goPurchase() { wx.navigateTo({ url: '/pages/me/purchase/purchase' }) },
  logout() {
    wx.showModal({
      title: '退出登录',
      content: '确定退出？',
      success: (res) => {
        if (!res.confirm) return
        clearToken()
        ws.disconnect()
        wx.reLaunch({ url: '/pages/login/login' })
      },
    })
  },
})
