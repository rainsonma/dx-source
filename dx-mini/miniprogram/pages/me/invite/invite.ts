import { api } from '../../../utils/api'
import { formatRelativeDate } from '../../../utils/format'

interface InviteStats { total: number; pending: number; paid: number; rewarded: number }
interface ReferralInvitee { id: string; username: string; nickname: string | null; grade: string }
interface ReferralItem {
  id: string; status: string; rewardAmount: number; rewardedAt: string | null; createdAt: string
  invitee: ReferralInvitee | null
}
interface InviteData {
  inviteCode: string; stats: InviteStats; referrals: ReferralItem[]; totalReferrals: number
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    loading: true,
    inviteCode: '',
    stats: null as InviteStats | null,
    referrals: [] as ReferralItem[],
    totalReferrals: 0,
    formatRelativeDate,
    statusBarHeight: 20,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488', statusBarHeight })
    this.loadData()
  },
  goBack() { wx.navigateBack() },
  async loadData() {
    try {
      const data = await api.get<InviteData>('/api/invite')
      this.setData({
        loading: false,
        inviteCode: data.inviteCode,
        stats: data.stats,
        referrals: data.referrals,
        totalReferrals: data.totalReferrals,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  copyCode() {
    wx.setClipboardData({ data: this.data.inviteCode })
    wx.showToast({ title: '邀请码已复制', icon: 'none' })
  },
})
