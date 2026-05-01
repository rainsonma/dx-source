import { api, PaginatedData } from '../../utils/api'
import { isLoggedIn } from '../../utils/auth'
import { formatRelativeDate } from '../../utils/format'
import { resolveNoticeIcon } from './icons'

interface NoticeRaw {
  id: string
  title: string
  content: string | null
  icon: string | null
  createdAt: string
}

interface NoticeView extends NoticeRaw {
  iconName: string
  timeText: string
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    statusBarHeight: 20,
    firstLoading: true,
    loadingMore: false,
    notices: [] as NoticeView[],
    nextCursor: '',
    hasMore: false,
    markedRead: false,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({ statusBarHeight: sys.statusBarHeight || 20 })
  },
  onShow() {
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488' })
    const tabBar = this.getTabBar() as WechatMiniprogram.Component.TrivialInstance | null
    if (tabBar) tabBar.setData({ active: 3, theme })
    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }
    if (this.data.notices.length === 0 && this.data.firstLoading) {
      this.loadNotices(true)
    }
  },
  onPullDownRefresh() {
    this.loadNotices(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loadingMore) {
      this.loadNotices(false)
    }
  },
  async loadNotices(reset: boolean) {
    if (reset) this.setData({ firstLoading: this.data.notices.length === 0 })
    else this.setData({ loadingMore: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${encodeURIComponent(cursor)}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<NoticeRaw>>(`/api/notices${qs}`)
      const view: NoticeView[] = res.items.map((n) => ({
        id: n.id,
        title: n.title,
        content: n.content,
        icon: n.icon,
        createdAt: n.createdAt,
        iconName: resolveNoticeIcon(n.icon),
        timeText: formatRelativeDate(n.createdAt),
      }))
      this.setData({
        firstLoading: false,
        loadingMore: false,
        notices: reset ? view : [...this.data.notices, ...view],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
      if (reset && !this.data.markedRead) this.markRead()
    } catch {
      this.setData({ firstLoading: false, loadingMore: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  markRead() {
    api.post('/api/notices/mark-read', {}).then(() => {
      this.setData({ markedRead: true })
      const tabBar = this.getTabBar() as
        | (WechatMiniprogram.Component.TrivialInstance & { clearUnread?: () => void })
        | null
      if (tabBar && tabBar.clearUnread) tabBar.clearUnread()
    }).catch(() => {})
  },
})
