import { api, PaginatedData } from '../../utils/api'

interface Category { id: string; name: string }
interface GameCardData {
  id: string; name: string; description: string | null; mode: string
  coverUrl: string | null; categoryName: string | null; levelCount: number; author: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: false,
    categories: [{ id: '', name: '全部' }] as Category[],
    activeCategoryId: '',
    games: [] as GameCardData[],
    nextCursor: '',
    hasMore: false,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadCategories()
    this.loadGames(true)
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    (this.getTabBar() as any)?.setData({ active: 1, theme: app.globalData.theme })
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
    const cats = await api.get<Category[]>('/api/game-categories').catch(() => [] as Category[])
    this.setData({ categories: [{ id: '', name: '全部' }, ...cats] })
  },
  async loadGames(reset: boolean) {
    if (this.data.loading) return
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const catId = this.data.activeCategoryId
    const qs = new URLSearchParams({ limit: '20' })
    if (cursor) qs.set('cursor', cursor)
    if (catId) qs.set('categoryIds', catId)
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
  onCategoryChange(e: WechatMiniprogram.TouchEvent) {
    const id = (e.detail as { name: string }).name
    this.setData({ activeCategoryId: id })
    this.loadGames(true)
  },
  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    wx.navigateTo({ url: `/pages/games/detail/detail?id=${id}` })
  },
  goFavorites() {
    wx.navigateTo({ url: '/pages/games/favorites/favorites' })
  },
})
