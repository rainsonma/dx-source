import { api, PaginatedData } from '../../../utils/api'
import { formatRelativeDate } from '../../../utils/format'

interface TrackingContentData { content: string; translation: string | null; contentType: string }
interface MasterItemData {
  id: string
  contentItem: TrackingContentData | null
  gameName: string | null
  masteredAt: string | null
  createdAt: string
}
interface MasterItemView extends MasterItemData {
  timeText: string
}
interface MasterStats { total: number; thisWeek: number; thisMonth: number }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    statusBarHeight: 20,
    loading: true,
    stats: { total: 0, thisWeek: 0, thisMonth: 0 } as MasterStats,
    items: [] as MasterItemView[],
    nextCursor: '',
    hasMore: false,
    selectedIds: [] as string[],
    selectMode: false,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const theme = app.globalData.theme
    this.setData({
      statusBarHeight: sys.statusBarHeight || 20,
      theme,
      primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488',
    })
    this.loadAll(true)
  },
  onPullDownRefresh() {
    this.loadAll(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadAll(false)
  },
  async loadAll(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? '?cursor=' + cursor + '&limit=20' : '?limit=20'
    const results = await Promise.allSettled([
      api.get<PaginatedData<MasterItemData>>('/api/tracking/master' + qs),
      reset ? api.get<MasterStats>('/api/tracking/master/stats') : Promise.resolve(this.data.stats),
    ])

    const list = results[0].status === 'fulfilled'
      ? results[0].value
      : { items: [] as MasterItemData[], nextCursor: '', hasMore: false }
    const newViews: MasterItemView[] = list.items.map((it: MasterItemData) => ({
      ...it,
      timeText: formatRelativeDate(it.masteredAt || it.createdAt),
    }))

    const stats = results[1].status === 'fulfilled' ? results[1].value : this.data.stats

    this.setData({
      loading: false,
      items: reset ? newViews : [...this.data.items, ...newViews],
      nextCursor: list.nextCursor,
      hasMore: list.hasMore,
      stats,
    })

    if (results.some((r) => r.status === 'rejected')) {
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  goBack() { wx.navigateBack() },
  toggleSelectMode() {
    this.setData({ selectMode: !this.data.selectMode, selectedIds: [] })
  },
  onSelectChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ selectedIds: e.detail as string[] })
  },
  onDeleteOne(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string | undefined
    if (!id) return
    wx.showModal({
      title: '确认删除',
      content: '删除该词？',
      success: async (res) => {
        if (!res.confirm) return
        try {
          await api.delete('/api/tracking/master', { ids: [id] })
          this.setData({
            items: this.data.items.filter((i) => i.id !== id),
            stats: { ...this.data.stats, total: Math.max(0, this.data.stats.total - 1) },
          })
          wx.showToast({ title: '已删除', icon: 'none' })
        } catch {
          wx.showToast({ title: '删除失败', icon: 'none' })
        }
      },
    })
  },
  async bulkDelete() {
    const ids = this.data.selectedIds
    if (ids.length === 0) return
    wx.showModal({
      title: '确认删除',
      content: '删除 ' + ids.length + ' 个词？',
      success: async (res) => {
        if (!res.confirm) return
        try {
          await api.delete('/api/tracking/master', { ids })
          this.setData({
            items: this.data.items.filter((i) => !ids.includes(i.id)),
            selectedIds: [],
            selectMode: false,
            stats: { ...this.data.stats, total: Math.max(0, this.data.stats.total - ids.length) },
          })
          wx.showToast({ title: '已删除', icon: 'none' })
        } catch {
          wx.showToast({ title: '删除失败', icon: 'none' })
        }
      },
    })
  },
})
