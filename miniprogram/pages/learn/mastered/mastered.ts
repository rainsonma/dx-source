import { api, PaginatedData } from '../../../utils/api'

interface TrackingContentData { content: string; translation: string | null; contentType: string }
interface TrackingItemData {
  id: string; contentItem: TrackingContentData | null
  gameName: string | null; masteredAt: string | null; createdAt: string
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    loading: true,
    items: [] as TrackingItemData[],
    nextCursor: '',
    hasMore: false,
    selectedIds: [] as string[],
    selectMode: false,
  },
  onLoad() {
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488' })
    this.loadItems(true)
  },
  onPullDownRefresh() {
    this.loadItems(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadItems(false)
  },
  async loadItems(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${cursor}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<TrackingItemData>>(`/api/tracking/master${qs}`)
      this.setData({
        loading: false,
        items: reset ? res.items : [...this.data.items, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  toggleSelectMode() {
    this.setData({ selectMode: !this.data.selectMode, selectedIds: [] })
  },
  onSelectChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ selectedIds: e.detail as string[] })
  },
  async bulkDelete() {
    if (this.data.selectedIds.length === 0) return
    wx.showModal({
      title: '确认删除',
      content: `删除 ${this.data.selectedIds.length} 个词？`,
      success: async (res) => {
        if (!res.confirm) return
        try {
          await api.delete('/api/tracking/master', { ids: this.data.selectedIds })
          this.setData({
            items: this.data.items.filter(i => !this.data.selectedIds.includes(i.id)),
            selectedIds: [],
            selectMode: false,
          })
          wx.showToast({ title: '已删除', icon: 'none' })
        } catch {
          wx.showToast({ title: '删除失败', icon: 'none' })
        }
      },
    })
  },
})
