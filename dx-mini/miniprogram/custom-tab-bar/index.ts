import { api } from '../utils/api'
import { isLoggedIn } from '../utils/auth'

interface TabItem {
  icon: string
  text: string
  path: string
}

Component({
  data: {
    active: 0,
    theme: 'light' as 'light' | 'dark',
    unread: false,
    tabs: [
      { icon: 'home',          text: '首页',   path: '/pages/home/home' },
      { icon: 'book-text',     text: '课程',   path: '/pages/games/games' },
      { icon: 'trophy',        text: '排行榜', path: '/pages/leaderboard/leaderboard' },
      { icon: 'bell',          text: '消息',   path: '/pages/notices/notices' },
      { icon: 'user',          text: '我的',   path: '/pages/me/me' },
    ] as TabItem[],
  },
  lifetimes: {
    attached() {
      const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()
      this.setData({ theme: app.globalData.theme })
      this.refreshUnread()
    },
  },
  methods: {
    switchTab(e: WechatMiniprogram.TouchEvent) {
      const path = e.currentTarget.dataset['path'] as string
      wx.switchTab({ url: path })
    },
    async refreshUnread() {
      if (!isLoggedIn()) return
      try {
        const [profile, list] = await Promise.all([
          api.get<{ last_read_notice_at: string | null }>('/api/user/profile'),
          api.get<{
            items: { id: string; createdAt: string }[]
            nextCursor: string
            hasMore: boolean
          }>('/api/notices?limit=1'),
        ])
        const latest = list.items[0]
        if (!latest) return
        const lastRead = profile.last_read_notice_at
        const unread = !lastRead || new Date(lastRead).getTime() < new Date(latest.createdAt).getTime()
        this.setData({ unread })
      } catch {
        // ignore — leave dot in its previous state
      }
    },
    clearUnread() {
      this.setData({ unread: false })
    },
  },
})
