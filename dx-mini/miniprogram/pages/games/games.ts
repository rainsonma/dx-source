import { api, PaginatedData } from '../../utils/api'

interface Category {
  id: string
  name: string
  depth: number
  isLeaf: boolean
}

interface GameCardData {
  id: string; name: string; description: string | null; mode: string
  coverUrl: string | null; categoryName: string | null; levelCount: number; author: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

const ALL_TAB: Category = { id: '', name: '全部', depth: 0, isLeaf: true }

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: false,
    categoriesAll: [] as Category[],
    topTabs: [ALL_TAB] as Category[],
    subTabs: [] as Category[],
    activeTopId: '',
    activeSubId: '',
    showSubTabs: false,
    games: [] as GameCardData[],
    nextCursor: '',
    hasMore: false,
    statusBarHeight: 20,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ theme: app.globalData.theme, statusBarHeight })
    this.loadCategories()
    this.loadGames(true)
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ active: 1, theme: app.globalData.theme }) }
  },
  onPullDownRefresh() {
    this.loadGames(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadGames(false)
    }
  },
  async loadCategories() {
    const raw = await api.get<Category[]>('/api/game-categories').catch(() => [] as Category[])
    const safe: Category[] = raw.map((c) => ({
      id: c.id,
      name: c.name,
      depth: typeof c.depth === 'number' ? c.depth : 0,
      isLeaf: typeof c.isLeaf === 'boolean' ? c.isLeaf : true,
    }))
    const topLevel = safe.filter((c) => c.depth === 0)
    this.setData({
      categoriesAll: safe,
      topTabs: [ALL_TAB, ...topLevel],
    })
  },
  computeCategoryIds(): string {
    const list = this.data.categoriesAll
    const subId = this.data.activeSubId
    const topId = this.data.activeTopId
    if (subId) return subId
    if (!topId) return ''
    const idx = list.findIndex((c) => c.id === topId)
    if (idx < 0) return topId
    const parentDepth = list[idx].depth
    const ids = [topId]
    for (let i = idx + 1; i < list.length; i++) {
      const c = list[i]
      if (c.depth <= parentDepth) break
      ids.push(c.id)
    }
    return ids.join(',')
  },
  async loadGames(reset: boolean) {
    if (this.data.loading) return
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const categoryIds = this.computeCategoryIds()
    const parts: string[] = ['limit=20']
    if (cursor) parts.push(`cursor=${encodeURIComponent(cursor)}`)
    if (categoryIds) parts.push(`categoryIds=${encodeURIComponent(categoryIds)}`)
    const qs = parts.join('&')
    try {
      const res = await api.get<PaginatedData<GameCardData>>(`/api/games?${qs}`)
      this.setData({
        games: reset ? res.items : [...this.data.games, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        loading: false,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  onTopChange(e: WechatMiniprogram.TouchEvent) {
    const id = (e.detail as { name: string }).name
    const list = this.data.categoriesAll
    let subTabs: Category[] = []
    if (id) {
      const idx = list.findIndex((c) => c.id === id)
      if (idx >= 0 && !list[idx].isLeaf) {
        const parentDepth = list[idx].depth
        const children: Category[] = []
        for (let i = idx + 1; i < list.length; i++) {
          const c = list[i]
          if (c.depth <= parentDepth) break
          if (c.depth === parentDepth + 1) children.push(c)
        }
        if (children.length > 0) {
          const parentAll: Category = { id: '', name: '全部', depth: 0, isLeaf: true }
          subTabs = [parentAll, ...children]
        }
      }
    }
    this.setData({
      activeTopId: id,
      activeSubId: '',
      subTabs,
      showSubTabs: subTabs.length > 0,
    })
    this.loadGames(true)
  },
  onSubChange(e: WechatMiniprogram.TouchEvent) {
    const id = (e.detail as { name: string }).name
    this.setData({ activeSubId: id })
    this.loadGames(true)
  },
  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    wx.navigateTo({ url: `/pages/games/detail/detail?id=${id}` })
  },
  goSearch() {
    wx.navigateTo({ url: '/pages/games/search/search' })
  },
})
