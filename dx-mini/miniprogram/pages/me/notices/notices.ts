import { api, PaginatedData } from '../../../utils/api'
import { formatRelativeDate } from '../../../utils/format'

interface NoticeItem { id: string; title: string; content: string | null; icon: string | null; createdAt: string }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    loading: true,
    notices: [] as NoticeItem[],
    nextCursor: '',
    hasMore: false,
    formatRelativeDate,
  },
  onLoad() {
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488' })
    this.loadNotices(true)
  },
  onPullDownRefresh() {
    this.loadNotices(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadNotices(false)
  },
  async loadNotices(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${cursor}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<NoticeItem>>(`/api/notices${qs}`)
      this.setData({
        loading: false,
        notices: reset ? res.items : [...this.data.notices, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
      api.post('/api/notices/mark-read', {}).catch(() => {})
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
})
