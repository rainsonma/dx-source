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
    loading: true,
    profile: null as ProfileData | null,
    formatDate,
    gradeLabel,
  },
  onShow() {
    const theme = app.globalData.theme
    this.setData({
      theme,
      primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488',
      arrowColor: theme === 'dark' ? '#6b7280' : '#9ca3af',
    });
    (this.getTabBar() as any)?.setData({ active: 4, theme })
    this.loadProfile()
  },
  async loadProfile() {
    try {
      const profile = await api.get<ProfileData>('/api/user/profile')
      this.setData({ loading: false, profile })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
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
