import { api, PaginatedData } from '../../../utils/api'
import { loadHistory, pushHistory, clearHistory } from './history'

type Mode = 'idle' | 'loading' | 'results' | 'empty'

interface GameCardData {
  id: string
  name: string
  description: string | null
  mode: string
  coverUrl: string | null
  categoryName: string | null
  levelCount: number
  author: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    autoFocus: true,
    query: '',
    mode: 'idle' as Mode,
    recents: [] as string[],
    suggestions: [] as string[],
    suggestionsLoading: false,
    games: [] as GameCardData[],
    nextCursor: '',
    hasMore: false,
    loadingMore: false,
  },

  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight,
      recents: loadHistory(),
    })
    this.loadSuggestions()
  },

  onShow() {
    this.setData({ theme: app.globalData.theme })
  },

  onReachBottom() {
    if (this.data.mode !== 'results') return
    if (!this.data.hasMore || this.data.loadingMore) return
    this.loadMore()
  },

  async loadSuggestions() {
    this.setData({ suggestionsLoading: true })
    try {
      const items = await api.get<string[]>('/api/games/search-suggestions')
      this.setData({
        suggestions: Array.isArray(items) ? items : [],
        suggestionsLoading: false,
      })
    } catch {
      // Suggestions are non-critical; fall back to empty list silently.
      this.setData({ suggestions: [], suggestionsLoading: false })
    }
  },

  onInput(e: WechatMiniprogram.Input) {
    const value = (e.detail as { value: string }).value
    const next: { query: string; mode?: Mode } = { query: value }
    if (value.trim() === '' && (this.data.mode === 'results' || this.data.mode === 'empty')) {
      next.mode = 'idle'
    }
    this.setData(next)
  },

  onClear() {
    this.setData({ query: '', mode: 'idle', autoFocus: false })
    // Re-trigger focus on next tick.
    setTimeout(() => {
      this.setData({ autoFocus: true })
    }, 0)
  },

  onSubmit(e: WechatMiniprogram.Input) {
    const raw = (e.detail as { value: string }).value
    const term = (raw || this.data.query || '').trim()
    if (!term) return
    this.runSearch(term)
  },

  onSearchTap() {
    const term = (this.data.query || '').trim()
    if (!term) return
    this.runSearch(term)
  },

  onChipTap(e: WechatMiniprogram.TouchEvent) {
    const term = String(e.currentTarget.dataset['term'] || '').trim()
    if (!term) return
    this.runSearch(term)
  },

  async runSearch(term: string) {
    const recents = pushHistory(term)
    this.setData({
      query: term,
      mode: 'loading',
      recents,
      games: [],
      nextCursor: '',
      hasMore: false,
    })
    try {
      const qs = `q=${encodeURIComponent(term)}&limit=20`
      const res = await api.get<PaginatedData<GameCardData>>(`/api/games?${qs}`)
      this.setData({
        games: res.items,
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        mode: res.items.length > 0 ? 'results' : 'empty',
      })
    } catch {
      this.setData({ mode: 'idle' })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  async loadMore() {
    if (this.data.loadingMore) return
    this.setData({ loadingMore: true })
    try {
      const qs = `q=${encodeURIComponent(this.data.query)}&limit=20&cursor=${encodeURIComponent(this.data.nextCursor)}`
      const res = await api.get<PaginatedData<GameCardData>>(`/api/games?${qs}`)
      this.setData({
        games: [...this.data.games, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        loadingMore: false,
      })
    } catch {
      this.setData({ loadingMore: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  onClearHistoryTap() {
    wx.showModal({
      title: '清除最近搜索?',
      confirmText: '清除',
      cancelText: '取消',
      success: (res) => {
        if (res.confirm) {
          clearHistory()
          this.setData({ recents: [] })
        }
      },
    })
  },

  onCancel() {
    const pages = getCurrentPages()
    if (pages.length > 1) {
      wx.navigateBack({ delta: 1 })
    } else {
      wx.switchTab({ url: '/pages/games/games' })
    }
  },

  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = String(e.currentTarget.dataset['id'] || '')
    if (!id) return
    wx.navigateTo({ url: `/pages/games/detail/detail?id=${id}` })
  },
})
