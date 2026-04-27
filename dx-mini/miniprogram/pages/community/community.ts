import { api, PaginatedData } from '../../utils/api'
import { isLoggedIn } from '../../utils/auth'
import { FEED_TABS, type FeedTab } from './types'
import type { Post } from './types'

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    feedTabs: FEED_TABS,
    tab: 'latest' as FeedTab,
    posts: [] as Post[],
    nextCursor: '',
    hasMore: false,
    loading: false,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
    })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
    const tabBar = this.getTabBar() as WechatMiniprogram.Component.TrivialInstance | null
    if (tabBar) tabBar.setData({ active: 3, theme: app.globalData.theme })
    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }
    if (this.data.posts.length === 0 && !this.data.loading) {
      this.loadFeed(true)
    }
  },
  onPullDownRefresh() {
    this.loadFeed(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadFeed(false)
    }
  },
  async loadFeed(reset: boolean) {
    if (this.data.loading) return
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const parts = ['limit=20', `tab=${this.data.tab}`]
    if (cursor) parts.push(`cursor=${encodeURIComponent(cursor)}`)
    try {
      const res = await api.get<PaginatedData<Post>>(`/api/posts?${parts.join('&')}`)
      this.setData({
        posts: reset ? res.items : [...this.data.posts, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        loading: false,
      })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: (err as Error).message || '加载失败', icon: 'none' })
    }
  },
  onTabChange(e: WechatMiniprogram.TouchEvent) {
    const name = (e.detail as { name: string }).name
    this.setData({
      tab: name as FeedTab,
      posts: [],
      nextCursor: '',
      hasMore: false,
    })
    this.loadFeed(true)
  },
  onOpenDetail(e: WechatMiniprogram.CustomEvent) {
    const id = (e.detail as { id: string }).id
    wx.navigateTo({ url: `/pages/community/detail/detail?id=${id}` })
  },
})
