import { api } from '../../../utils/api'
import { formatDate } from '../../../utils/format'

interface RedeemItem { id: string; code: string; grade: string; redeemedAt: string }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    loading: true,
    code: '',
    redeeming: false,
    history: [] as RedeemItem[],
    formatDate,
  },
  onLoad() {
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488' })
    this.loadHistory()
  },
  async loadHistory() {
    try {
      const res = await api.get<{ items: RedeemItem[] }>('/api/redeems')
      const items = Array.isArray(res) ? (res as unknown as RedeemItem[]) : res.items ?? []
      this.setData({ loading: false, history: items })
    } catch {
      this.setData({ loading: false })
    }
  },
  onCodeChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ code: (e.detail as { value: string }).value })
  },
  async redeem() {
    if (!this.data.code.trim() || this.data.redeeming) return
    this.setData({ redeeming: true })
    try {
      const res = await api.post<{ grade: string }>('/api/redeems', { code: this.data.code.trim() })
      this.setData({ code: '', redeeming: false })
      wx.showModal({ title: '兑换成功', content: `已升级为 ${res.grade}`, showCancel: false })
      this.loadHistory()
    } catch (err) {
      this.setData({ redeeming: false })
      wx.showToast({ title: (err as Error).message || '兑换失败', icon: 'none' })
    }
  },
})
